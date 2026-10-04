package passphrase

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/mkorun/lottabits/internal/draws"
	"github.com/mkorun/lottabits/vectors"
	"github.com/mkorun/lottabits/wordlists"
)

// SPEC.md 6.1: the 7744 coordinates map one-to-one onto the indexes 0..7743.
func TestIndexIsABijection(t *testing.T) {
	var seen [Entries]bool
	for a := 1; a <= Chips; a++ {
		for b := 1; b <= Chips; b++ {
			i := Index(a, b)
			if i < 0 || i >= Entries || seen[i] {
				t.Fatalf("Index(%d, %d) = %d is out of range or repeated", a, b, i)
			}
			seen[i] = true
		}
	}
}

func list(t *testing.T, name string) wordlists.List {
	t.Helper()
	l, err := wordlists.Passphrase(name)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestEveryCoordinateSelectsItsWord(t *testing.T) {
	for _, name := range wordlists.PassphraseNames() {
		l := list(t, name)
		for a := 1; a <= Chips; a++ {
			for b := 1; b <= Chips; b++ {
				words, err := FromDraws([]int{a, b}, l)
				if err != nil || words[0].Text != l.Word(Index(a, b)) {
					t.Fatalf("%s %02d-%02d: %v, %v", name, a, b, words, err)
				}
			}
		}
	}
}

func TestVectors(t *testing.T) {
	set, err := vectors.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range set.Passphrase {
		words, err := FromDraws(v.Draws, list(t, v.Wordlist))
		if err != nil {
			t.Fatalf("%s: %v", v.Name, err)
		}
		var texts, coordinates []string
		for _, w := range words {
			texts = append(texts, w.Text)
			coordinates = append(coordinates, w.Coordinate())
		}
		if strings.Join(texts, " ") != v.Passphrase || !slices.Equal(coordinates, v.Coordinates) {
			t.Errorf("%s: got %q %v", v.Name, texts, coordinates)
		}
	}
}

// The same draws give different words in different lists; the list, not the user-interface language, decides.
func TestWordListsAreIndependentData(t *testing.T) {
	en, _ := FromDraws([]int{1, 1}, list(t, "en"))
	de, _ := FromDraws([]int{1, 1}, list(t, "de"))
	if en[0].Text != "abacus" || de[0].Text != "aalen" {
		t.Errorf("01-01: en %q de %q", en[0].Text, de[0].Text)
	}
}

func TestRejectsInvalidInput(t *testing.T) {
	cases := map[string]struct {
		draws []int
		kind  draws.Kind
	}{
		"no draws": {nil, draws.EvenCount},
		"one draw": {[]int{1}, draws.EvenCount},
		"odd":      {[]int{1, 2, 3}, draws.EvenCount},
		"draw 0":   {[]int{1, 0}, draws.OutOfRange},
		"draw 89":  {[]int{89, 1}, draws.OutOfRange},
	}
	for name, c := range cases {
		_, err := FromDraws(c.draws, list(t, "en"))
		var e *draws.Error
		if !errors.As(err, &e) || e.Kind != c.kind {
			t.Errorf("%s: got %v, want kind %d", name, err, c.kind)
		}
	}
	if _, err := FromDraws([]int{1, 1}, wordlists.BIP39English()); err == nil {
		t.Error("a 2048-entry list must be rejected")
	}
}

func TestBits(t *testing.T) {
	if got := Bits(RecommendedMinimum); got < 77.5 || got > 77.6 {
		t.Errorf("Bits(6) = %f, want about 77.5", got)
	}
}
