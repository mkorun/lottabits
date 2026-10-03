// Package wordlists embeds the word lists defined in SPEC.md section 7.
// It is data for the lottabits command, not a stable API.
package wordlists

import (
	"crypto/sha256"
	_ "embed" // word lists are compiled into the binary
	"encoding/hex"
	"fmt"
	"strings"
)

// Identifiers and SHA-256 of the embedded files (SPEC.md sections 1 and 7.2).
const (
	BIP39EnglishID     = "bip39-english"
	bip39EnglishSHA256 = "2f5eed53a4727b4bf8880d8f3f199efc90e58503646d9ff8eff3a2ed3b24dbda"
	English7744ID      = "lottabits-en-7744-v1"
	english7744SHA256  = "5c4caefc140efbf20d30e481123fb4beadb1324c12c315971f2bfe600ddafcdb"
	German7744ID       = "lottabits-de-7744-v1"
	german7744SHA256   = "8023bf123831341641b1e33a26abdfb4b1739a5e01b3fbad62f35e77704ea156"
)

var (
	//go:embed bip39-english.txt
	bip39EnglishFile string
	//go:embed lottabits-en-7744-v1.txt
	english7744File string
	//go:embed lottabits-de-7744-v1.txt
	german7744File string
)

// List is an immutable word list.
type List struct {
	ID       string // identifier from SPEC.md section 1
	Expected string // SHA-256 of the file as specified (hex)
	file     string
	words    []string
}

func newList(id, expected, file string) List {
	return List{ID: id, Expected: expected, file: file, words: strings.Split(strings.TrimSuffix(file, "\n"), "\n")}
}

var (
	bip39English = newList(BIP39EnglishID, bip39EnglishSHA256, bip39EnglishFile)
	english7744  = newList(English7744ID, english7744SHA256, english7744File)
	german7744   = newList(German7744ID, german7744SHA256, german7744File)
)

// Len returns the number of entries.
func (l List) Len() int { return len(l.words) }

// Word returns the entry with the given 0-based index. The caller validates the index.
func (l List) Word(index int) string { return l.words[index] }

// Sum returns the SHA-256 of the embedded file (hex), for comparison with Expected.
func (l List) Sum() string {
	sum := sha256.Sum256([]byte(l.file))
	return hex.EncodeToString(sum[:])
}

// BIP39English returns the English BIP39 word list.
func BIP39English() List { return bip39English }

// PassphraseNames lists the accepted values of the --wordlist flag.
func PassphraseNames() []string { return []string{"de", "en"} }

// Passphrase returns the passphrase word list for a --wordlist value ("en" or "de").
func Passphrase(name string) (List, error) {
	switch name {
	case "en":
		return english7744, nil
	case "de":
		return german7744, nil
	}
	return List{}, fmt.Errorf("unknown word list %q (expected one of: %s)", name, strings.Join(PassphraseNames(), ", "))
}

// All returns every embedded list, for the self-test and the version output.
func All() []List { return []List{bip39English, english7744, german7744} }
