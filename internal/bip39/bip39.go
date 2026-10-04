// Package bip39 implements lottabits-seed-v1 (SPEC.md section 4) on top of plain BIP39 for 256-bit entropy.
// It contains no randomness: every result is a deterministic function of the draws.
package bip39

import (
	"crypto/sha256"

	"github.com/mkorun/lottabits/internal/draws"
	"github.com/mkorun/lottabits/wordlists"
)

// Parameters of lottabits-seed-v1.
const (
	ID           = "lottabits-seed-v1"
	Chips        = 64 // set 64
	Draws        = 46 // 23 pairs
	Words        = 24
	EntropyBytes = 32
	Blocks       = 8
	BlockSize    = 256

	pairs       = Words - 1
	groups      = Chips / 2
	bitsPerWord = 11
	extraBits   = 3
	bitsPerByte = 8
)

// Pair holds the two draws of one word; it is zero for word 24.
type Pair struct {
	First, Second int
}

// Word is one mnemonic word with its 1-based BIP39 number (1..2048).
type Word struct {
	Number int
	Text   string
	Pair   Pair
}

// Seed is the result of 46 draws.
type Seed struct {
	Words     [Words]Word
	ExtraBits [extraBits]int // extra bits of pairs 1 to 3
	Block     int            // 0..7, the block of word 24
	Entropy   [EntropyBytes]byte
	Checksum  byte
}

// PairIndex maps a valid pair (both draws in 1..64) to the 0-based BIP39 index and the extra bit (SPEC.md 4.2).
func PairIndex(first, second int) (index, extra int) {
	index = ((first-1)%groups)*Chips + (second - 1)
	if first > groups {
		extra = 1
	}
	return index, extra
}

// BlockRange returns the first and last 1-based word number of a block (SPEC.md 4.4).
func BlockRange(block int) (first, last int) {
	return block*BlockSize + 1, (block + 1) * BlockSize
}

// FromDraws computes the seed from exactly 46 draws in 1..64.
func FromDraws(d []int) (Seed, error) {
	if len(d) != Draws {
		return Seed{}, &draws.Error{Kind: draws.ExactCount, Want: Draws, Got: len(d)}
	}
	if err := draws.Validate(d, Chips); err != nil {
		return Seed{}, err
	}
	var seed Seed
	var indexes [pairs]int
	for k := range pairs {
		first, second := d[2*k], d[2*k+1]
		index, extra := PairIndex(first, second)
		indexes[k] = index
		if k < extraBits {
			seed.ExtraBits[k] = extra
		}
		seed.Words[k].Pair = Pair{First: first, Second: second}
	}
	seed.Entropy = packEntropy(indexes, seed.ExtraBits)
	seed.Checksum = checksum(seed.Entropy)
	seed.Block = seed.ExtraBits[0]<<2 | seed.ExtraBits[1]<<1 | seed.ExtraBits[2]
	list := wordlists.BIP39English()
	for k, index := range Mnemonic(seed.Entropy) {
		seed.Words[k].Number = index + 1
		seed.Words[k].Text = list.Word(index)
	}
	return seed, nil
}

// Mnemonic returns the 24 0-based BIP39 word indexes for 256-bit entropy (plain BIP39).
func Mnemonic(entropy [EntropyBytes]byte) [Words]int {
	data := append(entropy[:], checksum(entropy)) // 264 bits
	var out [Words]int
	for i := range Words {
		out[i] = readBits(data, i*bitsPerWord, bitsPerWord)
	}
	return out
}

func checksum(entropy [EntropyBytes]byte) byte {
	sum := sha256.Sum256(entropy[:])
	return sum[0]
}

// packEntropy concatenates the 23 indexes (11 bits each) and the 3 extra bits, most significant bit first.
func packEntropy(indexes [pairs]int, extra [extraBits]int) [EntropyBytes]byte {
	var out [EntropyBytes]byte
	offset := 0
	for _, index := range indexes {
		writeBits(out[:], offset, index, bitsPerWord)
		offset += bitsPerWord
	}
	for _, bit := range extra {
		writeBits(out[:], offset, bit, 1)
		offset++
	}
	return out
}

func writeBits(dst []byte, offset, value, n int) {
	for i := range n {
		if value>>(n-1-i)&1 == 1 {
			pos := offset + i
			dst[pos/bitsPerByte] |= 1 << (bitsPerByte - 1 - pos%bitsPerByte)
		}
	}
}

func readBits(src []byte, offset, n int) int {
	value := 0
	for i := range n {
		pos := offset + i
		value = value<<1 | int(src[pos/bitsPerByte]>>(bitsPerByte-1-pos%bitsPerByte)&1)
	}
	return value
}
