// Command lottabits turns physical chip draws into BIP39 seeds, passwords and passphrases (see SPEC.md).
// It never generates randomness itself.
package main

import (
	"os"

	"github.com/mkorun/lottabits/internal/cli"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	interactive := false
	if info, err := os.Stdin.Stat(); err == nil {
		interactive = info.Mode()&os.ModeCharDevice != 0
	}
	os.Exit(cli.Run(os.Args[1:], cli.Env{
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Interactive: interactive,
		Version:     version,
	}))
}
