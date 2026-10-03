package cli

import (
	"errors"
	"io"
	"strings"

	"github.com/mkorun/lottabits/internal/draws"
)

// maxInput bounds batch input; 46 draws need about 140 bytes.
const maxInput = 64 << 10

// entry describes how draws are asked for in interactive mode.
type entry struct {
	prompt  string // message key, formatted with the entry number
	highest int    // highest token number
	size    int    // draws per entry (1 or 2)
	count   int    // number of entries; 0 means until an empty line
}

// readDraws reads all draws: entry by entry on a terminal, all at once otherwise (SPEC.md section 8).
func (s *session) readDraws(e entry) ([]int, bool) {
	if s.env.Interactive {
		return s.promptDraws(e)
	}
	data, err := io.ReadAll(io.LimitReader(s.in, maxInput+1))
	if err != nil {
		s.errorLine(s.msg.t("err_read", err))
		return nil, false
	}
	if len(data) > maxInput {
		s.errorLine(s.msg.t("err_input_too_large", maxInput))
		return nil, false
	}
	d, err := draws.Parse(string(data), e.highest)
	if err != nil {
		s.drawError(err)
		return nil, false
	}
	return d, true
}

func (s *session) promptDraws(e entry) ([]int, bool) {
	var out []int
	for n := 1; e.count == 0 || n <= e.count; {
		s.write(s.env.Stderr, s.msg.t(e.prompt, n))
		text, err := s.in.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" && err != nil {
			s.write(s.env.Stderr, "\n")
			if e.count == 0 {
				return out, true
			}
			s.errorLine(s.msg.t("err_aborted"))
			return nil, false
		}
		if text == "" {
			if e.count == 0 {
				return out, true
			}
			continue
		}
		if values, ok := s.parseEntry(text, e); ok {
			out = append(out, values...)
			n++
		}
	}
	return out, true
}

// parseEntry validates one interactive entry and explains the problem if it is invalid.
func (s *session) parseEntry(text string, e entry) ([]int, bool) {
	fields := draws.Fields(text)
	if len(fields) != e.size {
		s.write(s.env.Stderr, s.msg.t("entry_count", e.size))
		return nil, false
	}
	values := make([]int, 0, e.size)
	for i, field := range fields {
		v, err := draws.ParseOne(i+1, field, e.highest)
		if err != nil {
			s.write(s.env.Stderr, s.msg.t("entry_invalid", field, e.highest))
			return nil, false
		}
		values = append(values, v)
	}
	return values, true
}

// drawError prints a localised message for an invalid draw or draw count.
func (s *session) drawError(err error) {
	var e *draws.Error
	if !errors.As(err, &e) {
		s.errorLine(err.Error())
		return
	}
	switch e.Kind {
	case draws.NotANumber, draws.OutOfRange:
		s.errorLine(s.msg.t("err_not_a_number", e.Position, e.Field, e.Highest))
	case draws.ExactCount:
		s.errorLine(s.msg.t("err_exact_count", e.Want, e.Got))
	case draws.MinimumCount:
		s.errorLine(s.msg.t("err_min_count", e.Want, e.Got))
	case draws.EvenCount:
		s.errorLine(s.msg.t("err_even_count", e.Want, e.Got))
	}
}
