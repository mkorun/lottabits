// Package selftest recomputes the embedded test vectors and checks every embedded data hash (SPEC.md 8.4).
package selftest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/mkorun/lottabits/internal/bip39"
	"github.com/mkorun/lottabits/internal/passphrase"
	"github.com/mkorun/lottabits/internal/password"
	"github.com/mkorun/lottabits/vectors"
	"github.com/mkorun/lottabits/wordlists"
)

// Result is the outcome of one check. Err is nil when the check passed.
type Result struct {
	Name string
	Err  error
}

// Run performs all checks and returns one result per check.
func Run() []Result {
	results := dataChecks()
	results = append(results, bip39Checks()...)
	set, err := vectors.Load()
	if err != nil {
		return append(results, Result{Name: "vectors.json", Err: err})
	}
	results = append(results, seedChecks(set.Seed)...)
	results = append(results, passwordChecks(set.Password)...)
	return append(results, passphraseChecks(set.Passphrase)...)
}

func mismatch(got, want any) error {
	return fmt.Errorf("got %v, want %v", got, want)
}

func dataChecks() []Result {
	sum := sha256.Sum256([]byte(password.Alphabet))
	var err error
	if got := hex.EncodeToString(sum[:]); got != password.AlphabetSHA256 {
		err = mismatch(got, password.AlphabetSHA256)
	}
	results := []Result{{Name: password.ID + " alphabet SHA-256", Err: err}}
	for _, l := range wordlists.All() {
		var err error
		if got := l.Sum(); got != l.Expected {
			err = mismatch(got, l.Expected)
		}
		results = append(results, Result{Name: l.ID + " SHA-256", Err: err})
	}
	return results
}

func bip39Checks() []Result {
	official, err := vectors.TrezorEnglish256()
	if err != nil {
		return []Result{{Name: "trezor-vectors.json", Err: err}}
	}
	list := wordlists.BIP39English()
	var results []Result
	for _, v := range official {
		var entropy [bip39.EntropyBytes]byte
		_, err := hex.Decode(entropy[:], []byte(v.Entropy))
		if err == nil {
			words := make([]string, 0, bip39.Words)
			for _, index := range bip39.Mnemonic(entropy) {
				words = append(words, list.Word(index))
			}
			if got := strings.Join(words, " "); got != v.Mnemonic {
				err = mismatch(got, v.Mnemonic)
			}
		}
		results = append(results, Result{Name: "BIP39 official vector " + v.Entropy[:16], Err: err})
	}
	return results
}

func seedChecks(set []vectors.Seed) []Result {
	var results []Result
	for _, v := range set {
		seed, err := bip39.FromDraws(v.Draws)
		if err == nil {
			numbers := make([]int, 0, bip39.Words)
			words := make([]string, 0, bip39.Words)
			for _, w := range seed.Words {
				numbers = append(numbers, w.Number)
				words = append(words, w.Text)
			}
			if !slices.Equal(numbers, v.Numbers) || strings.Join(words, " ") != v.Mnemonic {
				err = mismatch(numbers, v.Numbers)
			}
		}
		results = append(results, Result{Name: bip39.ID + " vector " + v.Name, Err: err})
	}
	return results
}

func passwordChecks(set []vectors.Password) []Result {
	var results []Result
	for _, v := range set {
		p, err := password.FromDraws(v.Draws)
		if err == nil && (p.Text != v.Password || p.Classes != v.Classes) {
			err = mismatch(p.Text, v.Password)
		}
		results = append(results, Result{Name: password.ID + " vector " + v.Name, Err: err})
	}
	return results
}

func passphraseChecks(set []vectors.Passphrase) []Result {
	var results []Result
	for _, v := range set {
		list, err := wordlists.Passphrase(v.Wordlist)
		var words []passphrase.Word
		if err == nil {
			words, err = passphrase.FromDraws(v.Draws, list)
		}
		if err == nil {
			texts := make([]string, 0, len(words))
			for _, w := range words {
				texts = append(texts, w.Text)
			}
			if got := strings.Join(texts, " "); got != v.Passphrase {
				err = mismatch(got, v.Passphrase)
			}
		}
		results = append(results, Result{Name: passphrase.ID + " vector " + v.Name, Err: err})
	}
	return results
}
