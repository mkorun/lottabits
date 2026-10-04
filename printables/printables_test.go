package printables

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/mkorun/lottabits/internal/bip39"
	"github.com/mkorun/lottabits/internal/password"
	"github.com/mkorun/lottabits/wordlists"
)

func render(t *testing.T, lang string) map[string]File {
	t.Helper()
	files, err := Render(lang, Options{Version: "v-test", ChipMM: DefaultChipMM})
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]File{}
	for _, f := range files {
		out[f.Name] = f
	}
	return out
}

var (
	seedEntryRE = regexp.MustCompile(`<td class="b"><u>(\d\d)</u></td><td class="n">(\d{4})</td><td class="w">([a-z]+)</td>`)
	ppEntryRE   = regexp.MustCompile(`<td class="b"><u>(\d\d)</u></td><td class="w">([a-z-]+)</td>`)
	ppFirstRE   = regexp.MustCompile(`<u class="first">(\d\d)</u>`)
	cellRE      = regexp.MustCompile(`<span class="n"><u>(\d\d)</u></span><span class="k">([DULS])</span><span class="glyph">(.*?)</span>`)
	chipRE      = regexp.MustCompile(`<div class="chip( high)?"[^>]*><u>(\d\d)</u></div>`)
	hashRE      = regexp.MustCompile(`<span class="hash">([0-9a-f ]+)</span>`)
	externalRE  = regexp.MustCompile(`(?i)(src|href)\s*=\s*"(https?:|//)|url\(\s*["']?(https?:|//)|<script|<link`)
)

func TestBothLanguagesHaveTheSameFiles(t *testing.T) {
	en, de := render(t, "en"), render(t, "de")
	if names := slices.Sorted(maps.Keys(en)); !slices.Equal(names, slices.Sorted(maps.Keys(de))) || len(names) != 12 {
		t.Fatalf("files differ or wrong count: %v", names)
	}
}

func TestLocalesHaveTheSameKeys(t *testing.T) {
	en, _ := loadLocale("en")
	de, _ := loadLocale("de")
	if !slices.Equal(slices.Sorted(maps.Keys(en)), slices.Sorted(maps.Keys(de))) {
		t.Error("locales/en.json and locales/de.json have different keys")
	}
}

func TestNoExternalResourcesAndVersionOnEveryPage(t *testing.T) {
	for _, lang := range Languages() {
		for name, f := range render(t, lang) {
			text := string(f.HTML)
			if externalRE.MatchString(text) {
				t.Errorf("%s/%s references an external resource or script", lang, name)
			}
			checkPageVersions(t, lang+"/"+name, text, len(f.Pages))
			if got, want := printedHashes(text), pageHashes(f); !slices.Equal(got, want) {
				t.Errorf("%s/%s: printed hashes %v, page data %v", lang, name, got, want)
			}
		}
	}
}

func printedHashes(text string) []string {
	var out []string
	for _, h := range hashRE.FindAllStringSubmatch(text, -1) {
		out = append(out, h[1])
	}
	return out
}

func pageHashes(f File) []string {
	var out []string
	for _, p := range f.Pages {
		out = append(out, p.Hashes...)
	}
	return out
}

// Every BIP39 word number appears exactly once in the rendered seed booklet, with the right word and second draw,
// and both languages print identical entries.
func TestSeedBookletEntries(t *testing.T) {
	list := wordlists.BIP39English()
	var previous [][]string
	for _, lang := range Languages() {
		matches := seedEntryRE.FindAllStringSubmatch(string(render(t, lang)["seed-booklet.html"].HTML), -1)
		if len(matches) != 2048 {
			t.Fatalf("%s: %d entries, want 2048", lang, len(matches))
		}
		seen := map[int]bool{}
		for _, m := range matches {
			second, _ := strconv.Atoi(m[1])
			number, _ := strconv.Atoi(m[2])
			if seen[number] || list.Word(number-1) != m[3] || (number-1)%bip39.Chips != second-1 {
				t.Fatalf("%s: entry %v is wrong or repeated", lang, m[1:])
			}
			seen[number] = true
		}
		if previous != nil && !slices.EqualFunc(previous, matches, slices.Equal) {
			t.Error("seed booklet entries differ between languages")
		}
		previous = matches
	}
}

func TestPassphraseBookletEntries(t *testing.T) {
	for _, lang := range Languages() {
		files := render(t, lang)
		for _, name := range wordlists.PassphraseNames() {
			list, _ := wordlists.Passphrase(name)
			pages := strings.Split(string(files["passphrase-booklet-"+name+".html"].HTML), `<div class="a5 content">`)[1:]
			if len(pages) != 88 {
				t.Fatalf("%s/%s: %d content pages, want 88", lang, name, len(pages))
			}
			for _, p := range pages {
				checkPassphrasePage(t, lang+"/"+name, p, list)
			}
		}
	}
}

func TestPasswordMap(t *testing.T) {
	for _, lang := range Languages() {
		cells := cellRE.FindAllStringSubmatch(string(render(t, lang)["password-map.html"].HTML), -1)
		var text, classes strings.Builder
		for i, c := range cells {
			if c[1] != fmt.Sprintf("%02d", i+1) {
				t.Fatalf("%s: cell %d has number %s", lang, i+1, c[1])
			}
			text.WriteString(html.UnescapeString(c[3]))
			classes.WriteString(c[2])
		}
		if text.String() != password.Alphabet {
			t.Errorf("%s: map shows %q, want the alphabet", lang, text.String())
		}
		for i := range len(password.Alphabet) {
			if classes.String()[i] != password.ClassOf(password.Alphabet[i]) {
				t.Errorf("%s: wrong class at %d", lang, i+1)
			}
		}
	}
}

func TestChipSheets(t *testing.T) {
	for _, name := range []string{"chip-inventory.html", "chip-cutout.html"} {
		matches := chipRE.FindAllStringSubmatch(string(render(t, "en")[name].HTML), -1)
		if len(matches) != 88 {
			t.Fatalf("%s: %d chips, want 88", name, len(matches))
		}
		for i, m := range matches {
			if m[2] != fmt.Sprintf("%02d", i+1) || (m[1] != "") != (i+1 > 64) {
				t.Errorf("%s: chip %d rendered as %q high=%q", name, i+1, m[2], m[1])
			}
		}
	}
}

func TestChipDiameterOption(t *testing.T) {
	for _, mm := range []float64{20, 30, 40} {
		pages, _, err := chipPages(mm)
		if err != nil {
			t.Fatalf("%v mm: %v", mm, err)
		}
		n := 0
		for _, p := range pages {
			chips := p.Data.([]chip)
			if chips[0].N != 1 && chips[0].N != 65 && chips[0].High != chips[len(chips)-1].High {
				t.Errorf("%v mm: page mixes chip sets", mm)
			}
			n += len(chips)
		}
		if n != 88 {
			t.Errorf("%v mm: %d chips", mm, n)
		}
	}
	if _, err := Render("en", Options{ChipMM: 250}); err == nil {
		t.Error("a 250 mm chip must be rejected")
	}
}

func TestBlockTable(t *testing.T) {
	rows, _ := blocks()
	want := [][2]string{{"abandon", "cable"}, {"cactus", "divide"}, {"divorce", "garment"}, {"gas", "lend"},
		{"length", "paper"}, {"parade", "say"}, {"scale", "that"}, {"theme", "zoo"}}
	for i, r := range rows {
		if [2]string{r.FirstWord, r.LastWord} != want[i] {
			t.Errorf("block %d: %s–%s, want %v", i, r.FirstWord, r.LastWord, want[i])
		}
	}
}

func TestUnknownLanguage(t *testing.T) {
	if _, err := Render("fr", Options{ChipMM: DefaultChipMM}); err == nil {
		t.Error("unknown language must fail")
	}
	data, _ := json.Marshal(PageInfo{Number: 1})
	if string(data) != `{"number":1}` {
		t.Errorf("empty hash must be omitted: %s", data)
	}
}

// The embedded fonts are the pinned upstream files (printables/fonts/README.md).
func TestFontsArePinned(t *testing.T) {
	for name, want := range map[string]string{
		"mono": "5ce8b1698d1ded7dff2178c1a3ad159470085a58ea239e8b2cb88f4fb4a6f646",
		"text": "5a455d1cfa099b601ab70751bb9673e8fe1854dc4500c80e1a220d0d75e31745",
	} {
		font := map[string][]byte{"mono": fontMono, "text": fontText}[name]
		if got := fmt.Sprintf("%x", sha256.Sum256(font)); got != want {
			t.Errorf("%s font SHA-256 %s, pinned %s", name, got, want)
		}
	}
}

// Saddle stitching: sheet i carries (n−2i | 2i+1) on the front and (2i+2 | n−2i−1) on the back, as in the prototype.
func TestImposition(t *testing.T) {
	if got := impose(32)[:4]; !slices.Equal(got, [][2]int{{32, 1}, {2, 31}, {30, 3}, {4, 29}}) {
		t.Errorf("first sides of a 32-page booklet: %v", got)
	}
	for _, n := range []int{4, 32, 88} {
		sides := impose(n)
		seen := map[int]bool{}
		for i := 0; i < len(sides); i += 2 {
			sheet := i / 2
			front, back := sides[i], sides[i+1]
			if front != [2]int{n - 2*sheet, 2*sheet + 1} || back != [2]int{2*sheet + 2, n - 2*sheet - 1} {
				t.Fatalf("n=%d sheet %d: front %v back %v", n, sheet, front, back)
			}
			for _, p := range [4]int{front[0], front[1], back[0], back[1]} {
				seen[p] = true
			}
		}
		if len(seen) != n {
			t.Errorf("n=%d: %d distinct pages", n, len(seen))
		}
	}
	if _, err := booklet(make([]a5, 6), a5{}); err == nil {
		t.Error("a page count that is not a multiple of 4 must be rejected")
	}
}

func TestBookletSizes(t *testing.T) {
	files := render(t, "en")
	for name, sides := range map[string]int{"seed-booklet.html": 18, "passphrase-booklet-en.html": 46, "passphrase-booklet-de.html": 46} {
		if got := len(files[name].Pages); got != sides {
			t.Errorf("%s: %d A4 sides, want %d (cover sheet plus content)", name, got, sides)
		}
	}
}

func TestLayoutCheckScriptOnlyOnRequest(t *testing.T) {
	files, err := Render("en", Options{ChipMM: DefaultChipMM, LayoutCheck: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if !strings.Contains(string(f.HTML), `setAttribute("data-layout"`) {
			t.Errorf("%s: layout-check build lacks the script", f.Name)
		}
	}
}

func checkPageVersions(t *testing.T, label, text string, pages int) {
	t.Helper()
	sections := strings.Split(text, `<section class="page`)[1:]
	if len(sections) != pages || pages == 0 {
		t.Errorf("%s: %d pages rendered, %d expected", label, len(sections), pages)
	}
	for i, section := range sections {
		if !strings.Contains(section, "LottaBits v-test") {
			t.Errorf("%s: page %d lacks the version", label, i+1)
		}
	}
}

func checkPassphrasePage(t *testing.T, label, page string, list wordlists.List) {
	t.Helper()
	first, _ := strconv.Atoi(ppFirstRE.FindStringSubmatch(page)[1])
	entries := ppEntryRE.FindAllStringSubmatch(page, -1)
	if len(entries) != 88 {
		t.Fatalf("%s page %d: %d entries", label, first, len(entries))
	}
	for _, m := range entries {
		second, _ := strconv.Atoi(m[1])
		if list.Word((first-1)*88+second-1) != m[2] {
			t.Fatalf("%s: %02d-%s %q is wrong", label, first, m[1], m[2])
		}
	}
}
