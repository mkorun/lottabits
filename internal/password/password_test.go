package password

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/mkorun/tokenware/internal/draws"
	"github.com/mkorun/tokenware/vectors"
)

func TestAlphabetIsFrozen(t *testing.T) {
	sum := sha256.Sum256([]byte(Alphabet))
	if got := hex.EncodeToString(sum[:]); got != AlphabetSHA256 {
		t.Fatalf("alphabet SHA-256 %s, specified %s", got, AlphabetSHA256)
	}
	if len(Alphabet) != Tokens {
		t.Fatalf("alphabet has %d characters, want %d", len(Alphabet), Tokens)
	}
}

// SPEC.md 5.1: printable ASCII 0x21..0x7E without six characters, each exactly once.
func TestAlphabetIsPrintableASCIIWithoutExclusions(t *testing.T) {
	const excluded = "'I\\`l|"
	for c := byte(0x21); c <= 0x7e; c++ {
		count := strings.Count(Alphabet, string(c))
		want := 1
		if strings.IndexByte(excluded, c) >= 0 {
			want = 0
		}
		if count != want {
			t.Errorf("character %q occurs %d times, want %d", c, count, want)
		}
	}
}

func TestClassesAreGroupedAsSpecified(t *testing.T) {
	want := strings.Repeat("D", 10) + strings.Repeat("U", 25) + strings.Repeat("L", 25) + strings.Repeat("S", 28)
	var got strings.Builder
	for i := range len(Alphabet) {
		got.WriteByte(ClassOf(Alphabet[i]))
	}
	if got.String() != want {
		t.Errorf("classes %s, want %s", got.String(), want)
	}
}

func TestEveryDrawSelectsItsCharacter(t *testing.T) {
	all := make([]int, Tokens)
	for i := range all {
		all[i] = i + 1
	}
	p, err := FromDraws(all)
	if err != nil {
		t.Fatal(err)
	}
	if p.Text != Alphabet {
		t.Errorf("draws 1..88 give %q, want the alphabet", p.Text)
	}
}

func TestVectors(t *testing.T) {
	set, err := vectors.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range set.Password {
		p, err := FromDraws(v.Draws)
		if err != nil || p.Text != v.Password || p.Classes != v.Classes {
			t.Errorf("%s: got %q %q %v, want %q %q", v.Name, p.Text, p.Classes, err, v.Password, v.Classes)
		}
	}
}

func TestRejectsInvalidInput(t *testing.T) {
	cases := map[string]struct {
		draws []int
		kind  draws.Kind
	}{
		"no draws": {nil, draws.MinimumCount},
		"draw 0":   {[]int{1, 0}, draws.OutOfRange},
		"draw 89":  {[]int{89}, draws.OutOfRange},
	}
	for name, c := range cases {
		_, err := FromDraws(c.draws)
		var e *draws.Error
		if !errors.As(err, &e) || e.Kind != c.kind {
			t.Errorf("%s: got %v, want kind %d", name, err, c.kind)
		}
	}
}

func TestBits(t *testing.T) {
	if got := Bits(RecommendedMinimum); got < 77.5 || got > 77.6 {
		t.Errorf("Bits(12) = %f, want about 77.5", got)
	}
}
