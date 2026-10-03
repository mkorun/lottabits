// Package cli implements the lottabits command line (SPEC.md section 8). The main package only connects the
// process to Run, so every behaviour is testable in-process.
package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"
)

// Env is everything Run needs from the process.
type Env struct {
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	Interactive bool   // standard input is a terminal
	Version     string // release version, set at build time
}

// Exit codes (SPEC.md section 8).
const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

type session struct {
	env Env
	msg messages
	in  *bufio.Reader
}

// Run executes one command and returns the process exit code.
func Run(args []string, env Env) int {
	s := &session{env: env, msg: messages{lang: "en"}, in: bufio.NewReader(env.Stdin)}
	lang, rest, err := extractLang(args)
	if err != nil {
		s.errorLine(err.Error())
		return exitUsage
	}
	s.msg.lang = lang
	if len(rest) == 0 {
		s.write(s.env.Stderr, s.msg.t("usage"))
		return exitUsage
	}
	switch name, args := rest[0], rest[1:]; name {
	case "seed":
		return s.seed(args)
	case "password":
		return s.password(args)
	case "passphrase":
		return s.passphrase(args)
	case "selftest":
		return s.selftest(args)
	case "version":
		return s.version(args)
	case "help", "-h", "-help", "--help":
		s.write(s.env.Stdout, s.msg.t("usage"))
		return exitOK
	default:
		s.errorLine(s.msg.t("err_unknown_command", name))
		s.write(s.env.Stderr, s.msg.t("usage"))
		return exitUsage
	}
}

// extractLang removes --lang (also -lang and --lang=x) from anywhere in args and validates it.
func extractLang(args []string) (string, []string, error) {
	lang := "en"
	var rest []string
	for i := 0; i < len(args); i++ {
		name, value, hasValue := strings.Cut(args[i], "=")
		if name != "--lang" && name != "-lang" {
			rest = append(rest, args[i])
			continue
		}
		if !hasValue {
			if i+1 >= len(args) {
				return "", nil, errors.New("--lang needs a value: en or de")
			}
			i++
			value = args[i]
		}
		if !slices.Contains(languages, value) {
			return "", nil, fmt.Errorf("unknown language %q for --lang (en or de)", value)
		}
		lang = value
	}
	return lang, rest, nil
}

// parseFlags parses the options of a command. It returns false and an exit code when the command must stop.
func (s *session) parseFlags(name string, args []string, define func(*flag.FlagSet)) (bool, int) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if define != nil {
		define(fs)
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			s.write(s.env.Stdout, s.msg.t("usage"))
			return false, exitOK
		}
		s.errorLine(s.msg.t("err_option", err))
		return false, exitUsage
	}
	if fs.NArg() > 0 {
		s.errorLine(s.msg.t("err_arguments"))
		return false, exitUsage
	}
	return true, exitOK
}

func (s *session) write(w io.Writer, text string) {
	_, _ = io.WriteString(w, text)
}

func (s *session) line(w io.Writer, text string) {
	s.write(w, text+"\n")
}

func (s *session) errorLine(text string) {
	s.line(s.env.Stderr, "lottabits: "+text)
}
