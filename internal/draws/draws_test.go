package draws

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestParseAcceptsSeparatorsAndLeadingZeros(t *testing.T) {
	got, err := Parse(" 01,2\t64\r\n07 ,, 9\n", 64)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{1, 2, 64, 7, 9}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseEmptyInput(t *testing.T) {
	got, err := Parse(" \n,\t", 88)
	if err != nil || len(got) != 0 {
		t.Errorf("got %v, %v; want no draws and no error", got, err)
	}
}

func TestParseRejectsInvalidFields(t *testing.T) {
	cases := []struct {
		input    string
		position int
		kind     Kind
	}{
		{"0", 1, OutOfRange},
		{"00", 1, OutOfRange},
		{"1 65", 2, OutOfRange},
		{"1 2 007", 3, NotANumber},
		{"1;2", 1, NotANumber},
		{"-1", 1, NotANumber},
		{"+1", 1, NotANumber},
		{"1.0", 1, NotANumber},
		{"x", 1, NotANumber},
		{"١", 1, NotANumber}, // Arabic-Indic digit one
	}
	for _, c := range cases {
		_, err := Parse(c.input, 64)
		var e *Error
		if !errors.As(err, &e) || e.Position != c.position || e.Kind != c.kind {
			t.Errorf("Parse(%q): got %v, want kind %d at position %d", c.input, err, c.kind, c.position)
		}
	}
}

func TestParseOneBoundaries(t *testing.T) {
	for _, highest := range []int{64, 88} {
		for v := 1; v <= highest; v++ {
			for _, field := range []string{fmt.Sprint(v), fmt.Sprintf("%02d", v)} {
				if got, err := ParseOne(1, field, highest); err != nil || got != v {
					t.Fatalf("ParseOne(%q, %d) = %d, %v", field, highest, got, err)
				}
			}
		}
		if _, err := ParseOne(1, fmt.Sprint(highest+1), highest); err == nil {
			t.Errorf("ParseOne(%d, %d) must fail", highest+1, highest)
		}
	}
}

func TestValidate(t *testing.T) {
	if err := Validate([]int{1, 64}, 64); err != nil {
		t.Error(err)
	}
	var e *Error
	if err := Validate([]int{1, 0}, 64); !errors.As(err, &e) || e.Position != 2 {
		t.Errorf("got %v, want an error at position 2", err)
	}
	if err := Validate([]int{89}, 88); !errors.As(err, &e) || e.Kind != OutOfRange {
		t.Errorf("got %v, want OutOfRange", err)
	}
}

func TestErrorMessages(t *testing.T) {
	cases := map[string]*Error{
		`draw 17: "65" is not a number from 01 to 64`:        {Kind: OutOfRange, Position: 17, Field: "65", Highest: 64},
		"46 draws needed, got 45":                            {Kind: ExactCount, Want: 46, Got: 45},
		"at least 1 draws needed, got 0":                     {Kind: MinimumCount, Want: 1, Got: 0},
		"an even number of draws (at least 2) needed, got 3": {Kind: EvenCount, Want: 2, Got: 3},
	}
	for want, e := range cases {
		if e.Error() != want {
			t.Errorf("got %q, want %q", e.Error(), want)
		}
	}
}

func FuzzParse(f *testing.F) {
	for _, seed := range []string{"01 02 64", "1,2,3", "", "65", "007", "\x00", "1\n\r\t,2", "٣"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		got, err := Parse(input, 88)
		if err != nil {
			var e *Error
			if !errors.As(err, &e) || e.Position < 1 {
				t.Fatalf("unexpected error %v", err)
			}
			return
		}
		if len(got) != len(Fields(input)) || Validate(got, 88) != nil {
			t.Fatalf("Parse(%q) = %v is inconsistent", input, got)
		}
		formatted := make([]string, len(got))
		for i, v := range got {
			formatted[i] = fmt.Sprintf("%02d", v)
		}
		again, err := Parse(strings.Join(formatted, " "), 88)
		if err != nil || !slices.Equal(again, got) {
			t.Fatalf("round trip of %v gave %v, %v", got, again, err)
		}
	})
}
