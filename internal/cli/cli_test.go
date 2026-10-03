package cli

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/mkorun/lottabits/vectors"
)

type result struct {
	stdout, stderr string
	code           int
}

func run(args []string, stdin string, interactive bool) result {
	var stdout, stderr bytes.Buffer
	code := Run(args, Env{
		Stdin: strings.NewReader(stdin), Stdout: &stdout, Stderr: &stderr,
		Interactive: interactive, Version: "test",
	})
	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func joinDraws(d []int) string {
	parts := make([]string, len(d))
	for i, v := range d {
		parts[i] = fmt.Sprintf("%02d", v)
	}
	return strings.Join(parts, " ")
}

func loadVectors(t *testing.T) vectors.Set {
	t.Helper()
	set, err := vectors.Load()
	if err != nil {
		t.Fatal(err)
	}
	return set
}

func expect(t *testing.T, name string, r result, code int, stdout, stderr []string) {
	t.Helper()
	if r.code != code {
		t.Errorf("%s: exit %d, want %d\nstdout: %s\nstderr: %s", name, r.code, code, r.stdout, r.stderr)
	}
	for _, s := range stdout {
		if !strings.Contains(r.stdout, s) {
			t.Errorf("%s: stdout lacks %q:\n%s", name, s, r.stdout)
		}
	}
	for _, s := range stderr {
		if !strings.Contains(r.stderr, s) {
			t.Errorf("%s: stderr lacks %q:\n%s", name, s, r.stderr)
		}
	}
}

func TestUsageAndErrors(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		code   int
		stdout []string
		stderr []string
	}{
		{"no command", nil, exitUsage, nil, []string{"Usage:"}},
		{"help", []string{"--help"}, exitOK, []string{"Usage:"}, nil},
		{"help de", []string{"help", "--lang", "de"}, exitOK, []string{"Aufruf:"}, nil},
		{"command help", []string{"seed", "-h"}, exitOK, []string{"Usage:"}, nil},
		{"unknown command", []string{"dice"}, exitUsage, nil, []string{`unknown command "dice"`}},
		{"unknown option", []string{"seed", "--fast"}, exitUsage, nil, []string{"invalid option"}},
		{"draws as arguments", []string{"seed", "53", "08"}, exitUsage, nil, []string{"not from command-line arguments"}},
		{"unknown language", []string{"seed", "--lang", "fr"}, exitUsage, nil, []string{`unknown language "fr"`}},
		{"language without value", []string{"seed", "--lang"}, exitUsage, nil, []string{"--lang needs a value"}},
		{"missing word list", []string{"passphrase"}, exitUsage, nil, []string{"--wordlist is required"}},
		{"unknown word list", []string{"passphrase", "--wordlist", "fr"}, exitUsage, nil, []string{`unknown word list "fr"`}},
	}
	for _, c := range cases {
		expect(t, c.name, run(c.args, "", false), c.code, c.stdout, c.stderr)
	}
}

func TestSeedVectors(t *testing.T) {
	for _, v := range loadVectors(t).Seed {
		r := run([]string{"seed", "--details"}, joinDraws(v.Draws), false)
		want := []string{v.Entropy, fmt.Sprintf("Checksum: 0x%02x", v.Checksum), "Extra bits: " + v.ExtraBits,
			"Software-generated randomness: none"}
		for i, word := range strings.Fields(v.Mnemonic) {
			want = append(want, fmt.Sprintf("%02d", i+1), fmt.Sprintf("%04d    %s\n", v.Numbers[i], word))
		}
		expect(t, "seed "+v.Name, r, exitOK, want, nil)
	}
}

func TestSeedRejectsInvalidBatchInput(t *testing.T) {
	expect(t, "45 draws", run([]string{"seed"}, strings.Repeat("01 ", 45), false), exitFailure, nil,
		[]string{"46 draws needed, got 45"})
	expect(t, "draw 65", run([]string{"seed"}, "01 65", false), exitFailure, nil,
		[]string{`draw 2: "65" is not a number from 01 to 64`})
	expect(t, "German", run([]string{"seed", "--lang", "de"}, "01 x", false), exitFailure, nil,
		[]string{`Ziehung 2: "x" ist keine Zahl von 01 bis 64`})
	expect(t, "too large", run([]string{"seed"}, strings.Repeat(" ", maxInput+1), false), exitFailure, nil,
		[]string{"input is larger than"})
}

func TestSeedInteractiveAsksAgainAfterInvalidEntries(t *testing.T) {
	v := loadVectors(t).Seed[2] // spec-mixed
	var lines []string
	for k := 0; k < len(v.Draws); k += 2 {
		if k == 0 {
			lines = append(lines, "53 65", "53", "") // out of range, one draw only, empty line
		}
		lines = append(lines, fmt.Sprintf("%d,%d", v.Draws[k], v.Draws[k+1]))
	}
	r := run([]string{"seed"}, strings.Join(lines, "\n")+"\n", true)
	expect(t, "interactive seed", r, exitOK, []string{"resemble"},
		[]string{"Word 01 – two draws (01–64): ", `"65" is not a number from 01 to 64`, "Please enter exactly 2 draws",
			"Word 23 – two draws", "clear the terminal scroll-back"})
	if n := strings.Count(r.stderr, "Word 01 –"); n != 4 {
		t.Errorf("word 01 asked %d times, want 4", n)
	}
}

func TestSeedInteractiveEndOfInputAborts(t *testing.T) {
	expect(t, "early end", run([]string{"seed"}, "01 01\n02 02\n", true), exitFailure, nil,
		[]string{"input ended before all draws were entered"})
}

func TestPassword(t *testing.T) {
	for _, v := range loadVectors(t).Password {
		r := run([]string{"password"}, joinDraws(v.Draws), false)
		expect(t, "password "+v.Name, r, exitOK, []string{"Password:  " + v.Password + "\n", "Classes:   " + v.Classes}, nil)
	}
	r := run([]string{"password", "--lang", "de"}, "1\n11\n36\n\n", true)
	expect(t, "interactive German", r, exitOK, []string{"Passwort:   0Aa\n", "Klassen:    DUL", "3 Ziehungen ≈ 19,4 Bit"},
		[]string{"Zeichen 04 – Ziehung", "Warnung: weniger als die empfohlenen 12 Zeichen."})
	expect(t, "no draws", run([]string{"password"}, "", false), exitFailure, nil,
		[]string{"at least 1 draw needed, got 0"})
	expect(t, "draw 89", run([]string{"password"}, "89", false), exitFailure, nil,
		[]string{`draw 1: "89" is not a number from 01 to 88`})
}

func TestPassphrase(t *testing.T) {
	for _, v := range loadVectors(t).Passphrase {
		r := run([]string{"passphrase", "--wordlist", v.Wordlist}, joinDraws(v.Draws), false)
		want := []string{"Passphrase:  " + v.Passphrase + "\n", "lottabits-" + v.Wordlist + "-7744-v1"}
		for i, c := range v.Coordinates {
			want = append(want, fmt.Sprintf("%02d  %s  %s\n", i+1, c, strings.Fields(v.Passphrase)[i]))
		}
		expect(t, "passphrase "+v.Name, r, exitOK, want, nil)
	}
	expect(t, "odd", run([]string{"passphrase", "--wordlist", "en"}, "1 2 3", false), exitFailure, nil,
		[]string{"an even number of draws (at least 2) needed, got 3"})
	expect(t, "interactive", run([]string{"passphrase", "--wordlist", "en"}, "1 1\n88\n88 88\n", true), exitOK,
		[]string{"abacus yiddish"}, []string{"Please enter exactly 2 draws", "fewer than the recommended 6 words"})
}

// SPEC.md 6.3 and 8: the user-interface language never selects or changes a word list.
func TestLanguageNeverChangesWordList(t *testing.T) {
	r := run([]string{"passphrase", "--lang", "de", "--wordlist", "en"}, "1 1", false)
	expect(t, "German UI, English list", r, exitOK, []string{"Passphrase:  abacus\n", "Wortliste:", "lottabits-en-7744-v1"}, nil)
	r = run([]string{"--lang", "en", "passphrase", "--wordlist", "de"}, "1 1", false)
	expect(t, "English UI, German list", r, exitOK, []string{"Passphrase:  aalen\n", "Word list:", "lottabits-de-7744-v1"}, nil)
	en := run([]string{"seed"}, joinDraws(loadVectors(t).Seed[2].Draws), false)
	de := run([]string{"seed", "--lang", "de"}, joinDraws(loadVectors(t).Seed[2].Draws), false)
	rows := regexp.MustCompile(`(?m)^\d\d\s.*?(\d{4})\s+([a-z]+)$`)
	pairs := func(out string) []string {
		var p []string
		for _, m := range rows.FindAllStringSubmatch(out, -1) {
			p = append(p, m[1]+" "+m[2])
		}
		return p
	}
	if got := pairs(en.stdout); len(got) != 24 || !slices.Equal(got, pairs(de.stdout)) {
		t.Errorf("seed words differ between languages:\n%v\n%v", got, pairs(de.stdout))
	}
}

func TestSelftestAndVersion(t *testing.T) {
	expect(t, "selftest", run([]string{"selftest"}, "", false), exitOK, []string{"Self-test passed: 36 checks."}, nil)
	expect(t, "version", run([]string{"version"}, "", false), exitOK,
		[]string{"lottabits  test", "lottabits-seed-v1", "lottabits-password-88-v1", "lottabits-passphrase-7744-v1",
			"bip39-english", "lottabits-en-7744-v1", "lottabits-de-7744-v1"}, nil)
}

// Translations change text only: both languages have the same keys and the same format verbs in the same order.
func TestCatalogsMatch(t *testing.T) {
	verbs := regexp.MustCompile(`%[-+# 0-9.]*[a-zA-Z%]`)
	en, de := catalogs["en"], catalogs["de"]
	if len(en) != len(de) {
		t.Errorf("en has %d messages, de has %d", len(en), len(de))
	}
	for key, text := range en {
		other, ok := de[key]
		if !ok {
			t.Errorf("de lacks %q", key)
			continue
		}
		if !slices.Equal(verbs.FindAllString(text, -1), verbs.FindAllString(other, -1)) {
			t.Errorf("%q: format verbs differ between en and de", key)
		}
	}
}
