// Package vectors embeds the normative test vectors (SPEC.md section 10) so that the tests and
// `tokenware selftest` check the same data. It is data for the tokenware command, not a stable API.
package vectors

import (
	"bytes"
	_ "embed" // vectors are compiled into the binary
	"encoding/json"
	"fmt"
)

var (
	//go:embed vectors.json
	vectorsFile []byte
	//go:embed trezor-vectors.json
	trezorFile []byte
)

// Seed is one tokenware-seed-v1 vector.
type Seed struct {
	Name      string `json:"name"`
	Draws     []int  `json:"draws"`
	Numbers   []int  `json:"numbers"`
	Mnemonic  string `json:"mnemonic"`
	ExtraBits string `json:"extra_bits"`
	Block     int    `json:"block"`
	Entropy   string `json:"entropy"`
	Checksum  int    `json:"checksum"`
}

// Password is one tokenware-password-88-v1 vector.
type Password struct {
	Name     string `json:"name"`
	Draws    []int  `json:"draws"`
	Password string `json:"password"`
	Classes  string `json:"classes"`
}

// Passphrase is one tokenware-passphrase-7744-v1 vector.
type Passphrase struct {
	Name        string   `json:"name"`
	Wordlist    string   `json:"wordlist"`
	Draws       []int    `json:"draws"`
	Passphrase  string   `json:"passphrase"`
	Coordinates []string `json:"coordinates"`
}

// Set holds all Tokenware vectors.
type Set struct {
	Comment    string       `json:"comment"`
	Seed       []Seed       `json:"seed"`
	Password   []Password   `json:"password"`
	Passphrase []Passphrase `json:"passphrase"`
}

// BIP39 is one official BIP39 vector (entropy and mnemonic, English).
type BIP39 struct {
	Entropy  string
	Mnemonic string
}

// Load decodes the Tokenware vectors strictly.
func Load() (Set, error) {
	var set Set
	decoder := json.NewDecoder(bytes.NewReader(vectorsFile))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&set); err != nil {
		return Set{}, fmt.Errorf("vectors.json: %w", err)
	}
	return set, nil
}

const entropy256HexLen = 64

// TrezorEnglish256 returns the official English BIP39 vectors with 256-bit entropy
// (python-mnemonic vectors.json, entries are [entropy, mnemonic, seed, xprv]).
func TrezorEnglish256() ([]BIP39, error) {
	var file struct {
		English [][]string `json:"english"`
	}
	if err := json.Unmarshal(trezorFile, &file); err != nil {
		return nil, fmt.Errorf("trezor-vectors.json: %w", err)
	}
	var out []BIP39
	for _, entry := range file.English {
		if len(entry) < 2 {
			return nil, fmt.Errorf("trezor-vectors.json: short entry %q", entry)
		}
		if len(entry[0]) == entropy256HexLen {
			out = append(out, BIP39{Entropy: entry[0], Mnemonic: entry[1]})
		}
	}
	return out, nil
}
