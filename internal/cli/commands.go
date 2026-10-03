package cli

import (
	"encoding/hex"
	"flag"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"text/tabwriter"

	"github.com/mkorun/lottabits/internal/bip39"
	"github.com/mkorun/lottabits/internal/passphrase"
	"github.com/mkorun/lottabits/internal/password"
	"github.com/mkorun/lottabits/internal/selftest"
	"github.com/mkorun/lottabits/wordlists"
)

func (s *session) table() *tabwriter.Writer {
	return tabwriter.NewWriter(s.env.Stdout, 0, 0, 2, ' ', 0)
}

func (s *session) finishInteractive() {
	if s.env.Interactive {
		s.line(s.env.Stderr, "\n"+s.msg.t("after_interactive"))
	}
}

func (s *session) seed(args []string) int {
	var details bool
	if ok, code := s.parseFlags("seed", args, func(fs *flag.FlagSet) {
		fs.BoolVar(&details, "details", false, "")
	}); !ok {
		return code
	}
	d, ok := s.readDraws(entry{prompt: "prompt_seed", highest: bip39.Tokens, size: 2, count: bip39.Draws / 2})
	if !ok {
		return exitFailure
	}
	seed, err := bip39.FromDraws(d)
	if err != nil {
		s.drawError(err)
		return exitFailure
	}
	s.printSeed(seed, details)
	s.finishInteractive()
	return exitOK
}

func (s *session) printSeed(seed bip39.Seed, details bool) {
	out := s.env.Stdout
	s.line(out, s.msg.t("seed_title", bip39.ID)+"\n")
	tw := s.table()
	_, _ = fmt.Fprintf(tw, "#\t%s\t%s\t%s\n", s.msg.t("col_draws"), s.msg.t("col_number"), s.msg.t("col_word"))
	for i, w := range seed.Words {
		pair := ""
		if w.Pair != (bip39.Pair{}) {
			pair = fmt.Sprintf("%02d %02d", w.Pair.First, w.Pair.Second)
		}
		_, _ = fmt.Fprintf(tw, "%02d\t%s\t%04d\t%s\n", i+1, pair, w.Number, w.Text)
	}
	_ = tw.Flush()
	bits := fmt.Sprintf("%d%d%d", seed.ExtraBits[0], seed.ExtraBits[1], seed.ExtraBits[2])
	first, last := bip39.BlockRange(seed.Block)
	s.line(out, "\n"+s.msg.t("seed_extra", bits, seed.Block, first, last))
	s.line(out, s.msg.t("seed_wallet", first, last, bits))
	if details {
		s.line(out, s.msg.t("seed_entropy", hex.EncodeToString(seed.Entropy[:])))
		s.line(out, s.msg.t("seed_checksum", seed.Checksum))
	}
	s.line(out, s.msg.t("seed_source", bip39.Draws))
	s.line(out, s.msg.t("seed_rng"))
}

func (s *session) password(args []string) int {
	if ok, code := s.parseFlags("password", args, nil); !ok {
		return code
	}
	d, ok := s.readDraws(entry{prompt: "prompt_password", highest: password.Tokens, size: 1})
	if !ok {
		return exitFailure
	}
	p, err := password.FromDraws(d)
	if err != nil {
		s.drawError(err)
		return exitFailure
	}
	numbers := make([]string, len(p.Draws))
	for i, n := range p.Draws {
		numbers[i] = fmt.Sprintf("%02d", n)
	}
	tw := s.table()
	_, _ = fmt.Fprintf(tw, "%s\t%s\n", s.msg.t("label_password"), p.Text)
	_, _ = fmt.Fprintf(tw, "%s\t%s  %s\n", s.msg.t("label_classes"), p.Classes, s.msg.t("classes_legend"))
	_, _ = fmt.Fprintf(tw, "%s\t%s\n", s.msg.t("label_draws"), strings.Join(numbers, " "))
	strength := s.msg.t("strength_draws", len(d), s.msg.decimal(password.Bits(len(d))))
	_, _ = fmt.Fprintf(tw, "%s\t%s\n", s.msg.t("label_strength"), strength)
	_ = tw.Flush()
	if len(d) < password.RecommendedMinimum {
		s.line(s.env.Stderr, s.msg.t("warn_password_short", password.RecommendedMinimum))
	}
	s.finishInteractive()
	return exitOK
}

func (s *session) passphrase(args []string) int {
	var name string
	if ok, code := s.parseFlags("passphrase", args, func(fs *flag.FlagSet) {
		fs.StringVar(&name, "wordlist", "", "")
	}); !ok {
		return code
	}
	if name == "" {
		s.errorLine(s.msg.t("err_missing_wordlist"))
		return exitUsage
	}
	list, err := wordlists.Passphrase(name)
	if err != nil {
		s.errorLine(s.msg.t("err_wordlist", name))
		return exitUsage
	}
	d, ok := s.readDraws(entry{prompt: "prompt_passphrase", highest: passphrase.Tokens, size: 2})
	if !ok {
		return exitFailure
	}
	words, err := passphrase.FromDraws(d, list)
	if err != nil {
		s.drawError(err)
		return exitFailure
	}
	s.printPassphrase(words, list)
	s.finishInteractive()
	return exitOK
}

func (s *session) printPassphrase(words []passphrase.Word, list wordlists.List) {
	texts := make([]string, len(words))
	for i, w := range words {
		texts[i] = w.Text
	}
	tw := s.table()
	_, _ = fmt.Fprintf(tw, "%s\t%s\n", s.msg.t("label_passphrase"), strings.Join(texts, " "))
	_, _ = fmt.Fprintf(tw, "%s\t%s\n", s.msg.t("label_wordlist"), list.ID)
	strength := s.msg.t("strength_words", len(words), s.msg.decimal(passphrase.Bits(len(words))))
	_, _ = fmt.Fprintf(tw, "%s\t%s\n\n", s.msg.t("label_strength"), strength)
	for i, w := range words {
		_, _ = fmt.Fprintf(tw, "%02d\t%s\t%s\n", i+1, w.Coordinate(), w.Text)
	}
	_ = tw.Flush()
	if len(words) < passphrase.RecommendedMinimum {
		s.line(s.env.Stderr, s.msg.t("warn_passphrase_short", passphrase.RecommendedMinimum))
	}
}

func (s *session) selftest(args []string) int {
	if ok, code := s.parseFlags("selftest", args, nil); !ok {
		return code
	}
	results := selftest.Run()
	failed := 0
	for _, r := range results {
		if r.Err != nil {
			failed++
			s.line(s.env.Stdout, s.msg.t("selftest_fail", r.Name, r.Err))
			continue
		}
		s.line(s.env.Stdout, s.msg.t("selftest_ok", r.Name))
	}
	if failed > 0 {
		s.line(s.env.Stdout, "\n"+s.msg.t("selftest_failed", failed, len(results)))
		return exitFailure
	}
	s.line(s.env.Stdout, "\n"+s.msg.t("selftest_passed", len(results)))
	return exitOK
}

func (s *session) version(args []string) int {
	if ok, code := s.parseFlags("version", args, nil); !ok {
		return code
	}
	tw := s.table()
	_, _ = fmt.Fprintf(tw, "lottabits\t%s\n", s.env.Version)
	_, _ = fmt.Fprintf(tw, "revision\t%s\n", revision())
	_, _ = fmt.Fprintf(tw, "go\t%s\n\n", runtime.Version())
	_, _ = fmt.Fprintf(tw, "%s\t\n", bip39.ID)
	_, _ = fmt.Fprintf(tw, "%s\talphabet SHA-256 %s\n", password.ID, password.AlphabetSHA256)
	_, _ = fmt.Fprintf(tw, "%s\t\n", passphrase.ID)
	for _, l := range wordlists.All() {
		_, _ = fmt.Fprintf(tw, "%s\tSHA-256 %s\n", l.ID, l.Expected)
	}
	_ = tw.Flush()
	return exitOK
}

// revision reports the source revision recorded by the Go toolchain, if any.
func revision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	rev, modified := "unknown", ""
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			rev = setting.Value
		case "vcs.modified":
			if setting.Value == "true" {
				modified = " (modified)"
			}
		}
	}
	return rev + modified
}
