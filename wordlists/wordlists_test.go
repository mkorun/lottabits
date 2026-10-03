package wordlists

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// Upstream files as pinned in SPEC.md section 7.2.
const (
	effUpstreamFile     = "upstream/eff_large_wordlist.txt"
	effUpstreamSHA256   = "addd35536511597a02fa0a9ff1e5284677b8883b83e986e43f15a3db996b903e"
	dys2pUpstreamFile   = "upstream/dys2p-de-7776-v1.txt"
	dys2pUpstreamSHA256 = "440fa02c65591328d6351435d3824c27b483a049f4eca0b13456d8c5090442e7"
	derivedEntries      = 7744
	upstreamEntries     = 7776
)

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestEmbeddedFilesMatchSpecifiedHashes(t *testing.T) {
	for _, l := range All() {
		if got := l.Sum(); got != l.Expected {
			t.Errorf("%s: SHA-256 %s, specified %s", l.ID, got, l.Expected)
		}
	}
}

func TestFileFormat(t *testing.T) {
	for _, l := range All() {
		if strings.HasPrefix(l.file, "\xef\xbb\xbf") || strings.Contains(l.file, "\r") || !strings.HasSuffix(l.file, "\n") {
			t.Errorf("%s: not UTF-8 without BOM, LF line endings and a final LF", l.ID)
		}
		if slices.Contains(l.words, "") {
			t.Errorf("%s: contains an empty line", l.ID)
		}
	}
}

func TestSizesAndUniqueness(t *testing.T) {
	want := map[string]int{BIP39EnglishID: 2048, English7744ID: derivedEntries, German7744ID: derivedEntries}
	for _, l := range All() {
		if l.Len() != want[l.ID] {
			t.Errorf("%s: %d entries, want %d", l.ID, l.Len(), want[l.ID])
		}
		sorted := slices.Clone(l.words)
		slices.Sort(sorted)
		if len(slices.Compact(sorted)) != l.Len() {
			t.Errorf("%s: contains duplicates", l.ID)
		}
	}
}

// SPEC.md 4.5 relies on this (informatively): candidates in alphabetical order are in block order.
func TestBIP39EnglishIsSorted(t *testing.T) {
	if !slices.IsSorted(bip39English.words) {
		t.Error("bip39-english is not sorted")
	}
}

func readUpstream(t *testing.T, path, wantSHA256 string) []string {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is one of the pinned upstream constants above
	if err != nil {
		t.Fatal(err)
	}
	if got := sha256Hex(data); got != wantSHA256 {
		t.Fatalf("%s: SHA-256 %s, pinned %s", path, got, wantSHA256)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != upstreamEntries {
		t.Fatalf("%s: %d lines, want %d", path, len(lines), upstreamEntries)
	}
	return lines
}

func joinList(words []string) string { return strings.Join(words, "\n") + "\n" }

func TestEnglishIsDerivedFromPinnedEFFList(t *testing.T) {
	lines := readUpstream(t, effUpstreamFile, effUpstreamSHA256)
	words := make([]string, 0, derivedEntries)
	for _, line := range lines[:derivedEntries] {
		_, word, ok := strings.Cut(line, "\t")
		if !ok {
			t.Fatalf("unexpected EFF line %q", line)
		}
		words = append(words, word)
	}
	if joinList(words) != english7744.file {
		t.Error("tokenware-en-7744-v1 differs from the first 7744 EFF words")
	}
}

func TestGermanIsDerivedFromPinnedDys2pList(t *testing.T) {
	lines := readUpstream(t, dys2pUpstreamFile, dys2pUpstreamSHA256)
	if joinList(lines[:derivedEntries]) != german7744.file {
		t.Error("tokenware-de-7744-v1 differs from the first 7744 dys2p lines")
	}
}

func TestCharacterSets(t *testing.T) {
	lower := regexp.MustCompile(`^[a-z]+$`)
	for _, w := range german7744.words {
		if !lower.MatchString(w) {
			t.Errorf("tokenware-de-7744-v1: %q is not a-z only", w)
		}
	}
	var hyphenated []string
	for _, w := range english7744.words {
		if !lower.MatchString(w) {
			hyphenated = append(hyphenated, w)
		}
	}
	if want := []string{"drop-down", "felt-tip", "t-shirt"}; !slices.Equal(hyphenated, want) {
		t.Errorf("tokenware-en-7744-v1: entries outside a-z %q, want %q", hyphenated, want)
	}
}

// In sorted order a word that is a prefix of another is directly followed by a word with that prefix.
func TestPassphraseListsArePrefixFree(t *testing.T) {
	for _, l := range []List{english7744, german7744} {
		sorted := slices.Clone(l.words)
		slices.Sort(sorted)
		for i := 1; i < len(sorted); i++ {
			if strings.HasPrefix(sorted[i], sorted[i-1]) {
				t.Errorf("%s: %q is a prefix of %q", l.ID, sorted[i-1], sorted[i])
			}
		}
	}
}

func TestPassphraseLookup(t *testing.T) {
	for _, name := range PassphraseNames() {
		if _, err := Passphrase(name); err != nil {
			t.Errorf("Passphrase(%q): %v", name, err)
		}
	}
	if _, err := Passphrase("fr"); err == nil {
		t.Error("Passphrase(\"fr\") must fail")
	}
}
