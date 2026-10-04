// Package password implements lottabits-password-88-v1 (SPEC.md section 5): one draw selects one character.
package password

import (
	"math"

	"github.com/mkorun/lottabits/internal/draws"
)

// Parameters of lottabits-password-88-v1.
const (
	ID    = "lottabits-password-88-v1"
	Chips = 88
	// Alphabet in canonical order: digits, upper case, lower case, symbols (SPEC.md 5.1).
	Alphabet       = "0123456789" + "ABCDEFGHJKLMNOPQRSTUVWXYZ" + "abcdefghijkmnopqrstuvwxyz" + "!\"#$%&()*+,-./:;<=>?@[]^_{}~"
	AlphabetSHA256 = "a608b1e22ae80afcdbb5989a971da973d1632e38fca4ddcfbe57e919dab9badf"
	// RecommendedMinimum is the recommended minimum number of characters (SPEC.md 5.4).
	RecommendedMinimum = 12
)

// Character classes shown under a password (SPEC.md 5.3).
const (
	Digit  = 'D'
	Upper  = 'U'
	Lower  = 'L'
	Symbol = 'S'
)

// Password is the result of one or more draws.
type Password struct {
	Text    string // the password
	Classes string // one class letter per character
	Draws   []int
}

// ClassOf returns the class letter of an alphabet character.
func ClassOf(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return Digit
	case c >= 'A' && c <= 'Z':
		return Upper
	case c >= 'a' && c <= 'z':
		return Lower
	}
	return Symbol
}

// FromDraws maps each draw in 1..88 to one character.
func FromDraws(d []int) (Password, error) {
	if len(d) == 0 {
		return Password{}, &draws.Error{Kind: draws.MinimumCount, Want: 1, Got: 0}
	}
	if err := draws.Validate(d, Chips); err != nil {
		return Password{}, err
	}
	text := make([]byte, len(d))
	classes := make([]byte, len(d))
	for i, n := range d {
		text[i] = Alphabet[n-1]
		classes[i] = ClassOf(text[i])
	}
	return Password{Text: string(text), Classes: string(classes), Draws: append([]int(nil), d...)}, nil
}

// Bits returns the entropy of a password with the given number of draws.
func Bits(n int) float64 { return float64(n) * math.Log2(Chips) }
