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
	files, err := Render(lang, Options{Version: "v-test", TokenMM: DefaultTokenMM})
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
	ppEntryRE   = regexp.MustCompile(`<td class="b"><u>(\d\d)</u>-<u>(\d\d)</u></td><td class="w">([a-z-]+)</td>`)
	cellRE      = regexp.MustCompile(`<span class="n"><u>(\d\d)</u></span><span class="k">([DULS])</span><span class="glyph">(.*?)</span>`)
	tokenRE     = regexp.MustCompile(`<div class="token( high)?"[^>]*><u>(\d\d)</u></div>`)
	hashRE      = regexp.MustCompile(`<span class="hash">([0-9a-f ]+)</span>`)
	footRE      = regexp.MustCompile(`<footer class="foot"><span>LottaBits v-test`)
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
			if got := len(footRE.FindAllString(text, -1)); got != len(f.Pages) || got == 0 {
				t.Errorf("%s/%s: version in %d footers, %d pages", lang, name, got, len(f.Pages))
			}
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
		if p.Hash != "" {
			out = append(out, p.Hash)
		}
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
			if seen[number] || list.Word(number-1) != m[3] || (number-1)%bip39.Tokens != second-1 {
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
			matches := ppEntryRE.FindAllStringSubmatch(string(files["passphrase-booklet-"+name+".html"].HTML), -1)
			if len(matches) != 7744 {
				t.Fatalf("%s/%s: %d entries, want 7744", lang, name, len(matches))
			}
			for _, m := range matches {
				first, _ := strconv.Atoi(m[1])
				second, _ := strconv.Atoi(m[2])
				if list.Word((first-1)*88+second-1) != m[3] {
					t.Fatalf("%s/%s: %s-%s %q is wrong", lang, name, m[1], m[2], m[3])
				}
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

func TestTokenSheets(t *testing.T) {
	for _, name := range []string{"token-inventory.html", "token-cutout.html"} {
		matches := tokenRE.FindAllStringSubmatch(string(render(t, "en")[name].HTML), -1)
		if len(matches) != 88 {
			t.Fatalf("%s: %d tokens, want 88", name, len(matches))
		}
		for i, m := range matches {
			if m[2] != fmt.Sprintf("%02d", i+1) || (m[1] != "") != (i+1 > 64) {
				t.Errorf("%s: token %d rendered as %q high=%q", name, i+1, m[2], m[1])
			}
		}
	}
}

func TestTokenDiameterOption(t *testing.T) {
	for _, mm := range []float64{20, 30, 40} {
		pages, _, err := tokenPages(mm)
		if err != nil {
			t.Fatalf("%v mm: %v", mm, err)
		}
		n := 0
		for _, p := range pages {
			tokens := p.Data.([]token)
			if tokens[0].N != 1 && tokens[0].N != 65 && tokens[0].High != tokens[len(tokens)-1].High {
				t.Errorf("%v mm: page mixes token sets", mm)
			}
			n += len(tokens)
		}
		if n != 88 {
			t.Errorf("%v mm: %d tokens", mm, n)
		}
	}
	if _, err := Render("en", Options{TokenMM: 250}); err == nil {
		t.Error("a 250 mm token must be rejected")
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
	if _, err := Render("fr", Options{TokenMM: DefaultTokenMM}); err == nil {
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
