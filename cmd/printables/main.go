// Command printables writes the LottaBits printables (SPEC.md section 9) as HTML into a directory, one subdirectory per
// language, plus manifest.json with the SHA-256 of every file and the data hash of every page.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mkorun/lottabits/printables"
)

type manifestFile struct {
	Lang   string                `json:"lang"`
	File   string                `json:"file"`
	SHA256 string                `json:"sha256"`
	Pages  []printables.PageInfo `json:"pages"`
}

type manifest struct {
	Version string         `json:"version"`
	TokenMM float64        `json:"token_mm"`
	Files   []manifestFile `json:"files"`
}

func main() {
	out := flag.String("out", "dist/printables", "output directory")
	version := flag.String("version", "dev", "release version shown on every page")
	tokenMM := flag.Float64("token-mm", printables.DefaultTokenMM, "token diameter in millimetres")
	flag.Parse()
	if err := write(*out, printables.Options{Version: *version, TokenMM: *tokenMM}); err != nil {
		fmt.Fprintln(os.Stderr, "printables:", err)
		os.Exit(1)
	}
}

func write(out string, opts printables.Options) error {
	m := manifest{Version: opts.Version, TokenMM: opts.TokenMM}
	for _, lang := range printables.Languages() {
		files, err := printables.Render(lang, opts)
		if err != nil {
			return err
		}
		dir := filepath.Join(out, lang)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(dir, f.Name), f.HTML, 0o600); err != nil {
				return err
			}
			sum := sha256.Sum256(f.HTML)
			m.Files = append(m.Files, manifestFile{Lang: lang, File: f.Name, SHA256: hex.EncodeToString(sum[:]), Pages: f.Pages})
		}
	}
	data, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, "manifest.json"), append(data, '\n'), 0o600)
}
