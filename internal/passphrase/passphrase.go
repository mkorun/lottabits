// Package passphrase implements tokenware-passphrase-7744-v1 (SPEC.md section 6): two draws select one word.
package passphrase

import (
	"fmt"
	"math"
	"slices"

	"github.com/mkorun/tokenware/internal/draws"
	"github.com/mkorun/tokenware/wordlists"
)

// Parameters of tokenware-passphrase-7744-v1.
const (
	ID      = "tokenware-passphrase-7744-v1"
	Tokens  = 88
	Entries = Tokens * Tokens
	// RecommendedMinimum is the recommended minimum number of words (SPEC.md 6.2).
	RecommendedMinimum = 6
)

// Word is one passphrase word with its two draws.
type Word struct {
	First, Second int
	Text          string
}

// Coordinate returns the printed coordinate "AA-BB".
func (w Word) Coordinate() string { return fmt.Sprintf("%02d-%02d", w.First, w.Second) }

// Index returns the 0-based list index of a valid pair (SPEC.md 6.1).
func Index(first, second int) int { return (first-1)*Tokens + (second - 1) }

// FromDraws maps each pair of draws in 1..88 to a word of the given 7744-entry list.
func FromDraws(d []int, list wordlists.List) ([]Word, error) {
	if list.Len() != Entries {
		return nil, fmt.Errorf("word list %s has %d entries, want %d", list.ID, list.Len(), Entries)
	}
	if len(d) < 2 || len(d)%2 != 0 {
		return nil, &draws.Error{Kind: draws.EvenCount, Want: 2, Got: len(d)}
	}
	if err := draws.Validate(d, Tokens); err != nil {
		return nil, err
	}
	words := make([]Word, 0, len(d)/2)
	for pair := range slices.Chunk(d, 2) {
		first, second := pair[0], pair[1]
		words = append(words, Word{First: first, Second: second, Text: list.Word(Index(first, second))})
	}
	return words, nil
}

// Bits returns the entropy of a passphrase with the given number of words.
func Bits(words int) float64 { return float64(words) * math.Log2(Entries) }
