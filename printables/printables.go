// Package printables renders the LottaBits printables (SPEC.md section 9) as self-contained HTML in English and German
// from the same data the CLI uses. Translations change text only; every number and word comes from the core packages.
package printables

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/mkorun/lottabits/internal/bip39"
	"github.com/mkorun/lottabits/internal/passphrase"
	"github.com/mkorun/lottabits/internal/password"
	"github.com/mkorun/lottabits/wordlists"
)

var (
	//go:embed templates/*.tmpl
	templateFiles embed.FS
	//go:embed locales/*.json
	localeFiles embed.FS
	//go:embed style.css
	styleSheet string
	//go:embed fonts/AtkinsonHyperlegibleMono-wght.ttf
	fontMono []byte
	//go:embed fonts/AtkinsonHyperlegibleNext-wght.ttf
	fontText []byte
)

// DefaultChipMM is the default chip diameter of the inventory and cut-out sheets.
const DefaultChipMM = 25.0

// chipFontRatio is the height of a chip number relative to the chip diameter.
const chipFontRatio = 0.36

// split divides xs into n consecutive parts of nearly equal length.
func split(xs []int, n int) [][]int {
	parts := make([][]int, 0, n)
	size := (len(xs) + n - 1) / n
	for start := 0; start < len(xs); start += size {
		parts = append(parts, xs[start:min(start+size, len(xs))])
	}
	return parts
}

// Options configure a rendering.
type Options struct {
	Version     string  // release version shown on every page
	ChipMM      float64 // chip diameter in millimetres
	LayoutCheck bool    // add a script that reports overflowing boxes (for tools/check_layout.py only, never shipped)
}

// File is one rendered printable.
type File struct {
	Lang  string
	Name  string // file name, for example "seed-booklet.html"
	HTML  []byte
	Pages []PageInfo
}

// PageInfo lists the data hashes printed on one A4 page, in document order.
type PageInfo struct {
	Number int      `json:"number"`
	Hashes []string `json:"data_hashes,omitempty"`
}

// Languages lists the printable languages.
func Languages() []string { return []string{"en", "de"} }

// document is the data of one printable passed to its template.
type document struct {
	Name      string
	Template  string
	Lang      string
	Title     string // locale key
	Sub       string // locale key of the page subtitle
	Landscape bool
	Calibrate bool
	Booklet   string // "seed" or "passphrase" for booklets, empty otherwise
	Check     bool   // include the layout-check script
	Version   string
	IDs       []string
	Style     template.CSS
	Pages     []page
	Extra     any
}

// halfContext gives a booklet page access to its document.
type halfContext struct {
	Doc  *document
	Half a5
}

// pageContext gives the shared page header and footer access to the document and the page.
type pageContext struct {
	Doc  *document
	Page page
}

func loadLocale(lang string) (map[string]string, error) {
	data, err := localeFiles.ReadFile("locales/" + lang + ".json")
	if err != nil {
		return nil, fmt.Errorf("unknown printable language %q", lang)
	}
	var texts map[string]string
	if err := json.Unmarshal(data, &texts); err != nil {
		return nil, fmt.Errorf("locales/%s.json: %w", lang, err)
	}
	return texts, nil
}

func style() template.CSS {
	face := func(name string, font []byte) string {
		return fmt.Sprintf("@font-face { font-family: %q; font-weight: 100 900; src: url(data:font/ttf;base64,%s) format(\"truetype\"); }\n",
			name, base64.StdEncoding.EncodeToString(font))
	}
	return template.CSS(face("LB Text", fontText) + face("LB Mono", fontMono) + styleSheet) //nolint:gosec // G203: embedded, trusted style sheet and fonts
}

func templates(lang string, texts map[string]string) (*template.Template, error) {
	funcs := template.FuncMap{
		"t": func(key string) (string, error) {
			text, ok := texts[key]
			if !ok {
				return "", fmt.Errorf("locale %s lacks %q", lang, key)
			}
			return text, nil
		},
		"two":  func(n int) string { return fmt.Sprintf("%02d", n) },
		"four": func(n int) string { return fmt.Sprintf("%04d", n) },
		"odd":  func(i int) bool { return i%2 == 1 },
		"dec": func(x float64) string {
			s := fmt.Sprintf("%.1f", x)
			if lang == "de" {
				s = strings.Replace(s, ".", ",", 1)
			}
			return s
		},
		"mm":       func(x float64) template.CSS { return template.CSS(fmt.Sprintf("%.2fmm", x)) }, //nolint:gosec // G203: number only
		"ctx":      func(d *document, p page) pageContext { return pageContext{Doc: d, Page: p} },
		"half":     func(d *document, h a5) halfContext { return halfContext{Doc: d, Half: h} },
		"fontSize": func(diameter float64) float64 { return diameter * chipFontRatio },
		"split":    split,
		"desc":     func(title string) string { return strings.Replace(title, "title_", "desc_", 1) },
	}
	return template.New(lang).Option("missingkey=error").Funcs(funcs).ParseFS(templateFiles, "templates/*.tmpl")
}

// Render renders every printable in one language.
func Render(lang string, opts Options) ([]File, error) {
	texts, err := loadLocale(lang)
	if err != nil {
		return nil, err
	}
	tmpl, err := templates(lang, texts)
	if err != nil {
		return nil, err
	}
	docs, err := documents(lang, opts)
	if err != nil {
		return nil, err
	}
	files := make([]File, 0, len(docs))
	for _, d := range docs {
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, d.Template, d); err != nil {
			return nil, fmt.Errorf("%s/%s: %w", lang, d.Name, err)
		}
		f := File{Lang: lang, Name: d.Name + ".html", HTML: buf.Bytes()}
		for _, p := range d.Pages {
			f.Pages = append(f.Pages, PageInfo{Number: p.Number, Hashes: p.Hashes})
		}
		files = append(files, f)
	}
	return files, nil
}

// documents builds the data of every printable (SPEC.md section 9).
func documents(lang string, opts Options) ([]*document, error) {
	chips, grid, err := chipPages(opts.ChipMM)
	if err != nil {
		return nil, err
	}
	en, _ := wordlists.Passphrase("en")
	de, _ := wordlists.Passphrase("de")
	blockRows, blockLines := blocks()
	single := numberPages([]page{{}})
	seedPages, err := seedBooklet()
	if err != nil {
		return nil, err
	}
	ppStrength := strengthRows([]int{4, 5, 6, 7, 8}, passphrase.Bits)
	enPages, err := passphraseBooklet(en, ppStrength)
	if err != nil {
		return nil, err
	}
	dePages, err := passphraseBooklet(de, ppStrength)
	if err != nil {
		return nil, err
	}
	seedIDs := []string{bip39.ID, wordlists.BIP39EnglishID}
	docs := []*document{
		{Name: "quick-reference", Template: "quick", Title: "title_quick", Sub: "print_portrait", Pages: single},
		{Name: "chip-inventory", Template: "chips", Title: "title_inventory", Sub: "inventory_instruction",
			Landscape: true, Calibrate: true, Pages: chips, Extra: map[string]any{"Class": "inventory", "Grid": grid}},
		{Name: "chip-cutout", Template: "chips", Title: "title_cutout", Sub: "cutout_instruction",
			Landscape: true, Calibrate: true, Pages: chips, Extra: map[string]any{"Class": "cutout", "Grid": grid}},
		{Name: "seed-booklet", Template: "booklet", Title: "title_seed_booklet", Sub: "seed_booklet_sub",
			Landscape: true, Booklet: "seed", IDs: seedIDs, Pages: seedPages},
		{Name: "seed-record", Template: "seed-record", Title: "title_seed_record", Sub: "print_portrait",
			IDs: seedIDs, Pages: single, Extra: rows(bip39.Draws / 2)},
		{Name: "seed-reference", Template: "seed-reference", Title: "title_seed_reference", Sub: "print_portrait",
			IDs: seedIDs, Pages: numberPages([]page{{Hashes: []string{dataHash(blockLines)}}}), Extra: blockRows},
		{Name: "password-map", Template: "password-map", Title: "title_password_map", Sub: "map_sub",
			IDs: []string{password.ID}, Pages: passwordMapPage()},
		{Name: "password-record", Template: "password-record", Title: "title_password_record", Sub: "print_portrait",
			IDs: []string{password.ID}, Pages: single,
			Extra: map[string]any{"Rows": rows(32), "Strength": strengthRows([]int{8, 12, 16, 20, 24}, password.Bits)}},
		{Name: "passphrase-booklet-en", Template: "booklet", Title: "title_passphrase_booklet", Sub: "pp_booklet_sub",
			Landscape: true, Booklet: "passphrase", IDs: []string{passphrase.ID, en.ID}, Pages: enPages, Extra: en.ID},
		{Name: "passphrase-booklet-de", Template: "booklet", Title: "title_passphrase_booklet", Sub: "pp_booklet_sub",
			Landscape: true, Booklet: "passphrase", IDs: []string{passphrase.ID, de.ID}, Pages: dePages, Extra: de.ID},
		{Name: "passphrase-record", Template: "passphrase-record", Title: "title_passphrase_record", Sub: "print_portrait",
			IDs: []string{passphrase.ID}, Pages: single, Extra: map[string]any{"Rows": rows(10), "Lists": []string{en.ID, de.ID},
				"Strength": ppStrength}},
	}
	docs = append([]*document{{Name: "index", Template: "index", Title: "title_index", Sub: "print_portrait",
		Pages: single, Extra: docs}}, docs...)
	css := style()
	for _, d := range docs {
		d.Lang, d.Version, d.Style, d.Check = lang, opts.Version, css, opts.LayoutCheck
	}
	return docs, nil
}
