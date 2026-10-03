// Package draws parses and validates token draws typed by a person (SPEC.md sections 2 and 8).
package draws

import (
	"fmt"
	"strings"
)

// Kind classifies an invalid input.
type Kind int

// Kinds of invalid input. Count kinds describe the number of draws, the others a single draw.
const (
	NotANumber   Kind = iota + 1 // not one or two decimal digits
	OutOfRange                   // a number outside 1..Highest
	ExactCount                   // exactly Want draws are needed
	MinimumCount                 // at least Want draws are needed
	EvenCount                    // an even number of draws, at least Want, is needed
)

// Error describes invalid input. Position is the 1-based number of the draw, or 0 for count errors.
type Error struct {
	Kind     Kind
	Position int
	Field    string
	Highest  int
	Want     int
	Got      int
}

func (e *Error) Error() string {
	switch e.Kind {
	case NotANumber, OutOfRange:
		return fmt.Sprintf("draw %d: %q is not a number from 01 to %02d", e.Position, e.Field, e.Highest)
	case ExactCount:
		return fmt.Sprintf("%d draws needed, got %d", e.Want, e.Got)
	case MinimumCount:
		return fmt.Sprintf("at least %d draws needed, got %d", e.Want, e.Got)
	default:
		return fmt.Sprintf("an even number of draws (at least %d) needed, got %d", e.Want, e.Got)
	}
}

const maxDigits = 2

// ParseOne validates a single field such as "7" or "07" against 1..highest.
func ParseOne(position int, field string, highest int) (int, error) {
	if field == "" || len(field) > maxDigits {
		return 0, &Error{Kind: NotANumber, Position: position, Field: field, Highest: highest}
	}
	value := 0
	for _, c := range []byte(field) {
		if c < '0' || c > '9' {
			return 0, &Error{Kind: NotANumber, Position: position, Field: field, Highest: highest}
		}
		value = value*10 + int(c-'0')
	}
	if value < 1 || value > highest {
		return 0, &Error{Kind: OutOfRange, Position: position, Field: field, Highest: highest}
	}
	return value, nil
}

func isSeparator(r rune) bool {
	return r == ' ' || r == '\t' || r == ',' || r == '\n' || r == '\r'
}

// Fields splits input at spaces, tabs, commas and line breaks.
func Fields(input string) []string { return strings.FieldsFunc(input, isSeparator) }

// Parse validates every field of input against 1..highest and stops at the first invalid one.
func Parse(input string, highest int) ([]int, error) {
	fields := Fields(input)
	out := make([]int, 0, len(fields))
	for i, field := range fields {
		value, err := ParseOne(i+1, field, highest)
		if err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, nil
}

// Validate checks already parsed draws against 1..highest. Core functions call it on every input.
func Validate(draws []int, highest int) error {
	for i, d := range draws {
		if d < 1 || d > highest {
			return &Error{Kind: OutOfRange, Position: i + 1, Field: fmt.Sprint(d), Highest: highest}
		}
	}
	return nil
}
