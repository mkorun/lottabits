package bip39

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/mkorun/lottabits/internal/draws"
	"github.com/mkorun/lottabits/vectors"
	"github.com/mkorun/lottabits/wordlists"
)

const wordCount = 2048

// SPEC.md 4.2: all 4096 pairs; each index exactly twice, once with extra bit 0 and once with 1.
func TestPairIndexIsExhaustivelyUniform(t *testing.T) {
	var seen [wordCount][2]int
	for first := 1; first <= Tokens; first++ {
		for second := 1; second <= Tokens; second++ {
			index, extra := PairIndex(first, second)
			if index < 0 || index >= wordCount || (extra != 0 && extra != 1) {
				t.Fatalf("PairIndex(%d, %d) = %d, %d out of range", first, second, index, extra)
			}
			seen[index][extra]++
		}
	}
	for index, counts := range seen {
		if counts != [2]int{1, 1} {
			t.Fatalf("index %d: preimages with extra bit 0 and 1 = %v, want [1 1]", index, counts)
		}
	}
}

func TestExtraBitBelongsToFirstDraw(t *testing.T) {
	for first := 1; first <= Tokens; first++ {
		_, extra := PairIndex(first, 1)
		want := 1
		if first <= 32 {
			want = 0
		}
		if extra != want {
			t.Errorf("first draw %d: extra bit %d, want %d", first, extra, want)
		}
	}
}

func TestBlockRanges(t *testing.T) {
	want := [][2]int{{1, 256}, {257, 512}, {513, 768}, {769, 1024}, {1025, 1280}, {1281, 1536}, {1537, 1792}, {1793, 2048}}
	for block := range Blocks {
		if first, last := BlockRange(block); first != want[block][0] || last != want[block][1] {
			t.Errorf("block %d: %d–%d, want %v", block, first, last, want[block])
		}
	}
}

func words(indexes [Words]int) string {
	list := wordlists.BIP39English()
	out := make([]string, Words)
	for i, index := range indexes {
		out[i] = list.Word(index)
	}
	return strings.Join(out, " ")
}

func TestMnemonicMatchesOfficialVectors(t *testing.T) {
	official, err := vectors.TrezorEnglish256()
	if err != nil {
		t.Fatal(err)
	}
	if len(official) != 8 {
		t.Fatalf("%d official 256-bit vectors, want 8", len(official))
	}
	for _, v := range official {
		var entropy [EntropyBytes]byte
		if _, err := hex.Decode(entropy[:], []byte(v.Entropy)); err != nil {
			t.Fatal(err)
		}
		if got := words(Mnemonic(entropy)); got != v.Mnemonic {
			t.Errorf("entropy %s:\n got %s\nwant %s", v.Entropy, got, v.Mnemonic)
		}
	}
}

func TestFromDrawsMatchesLottaBitsVectors(t *testing.T) {
	set, err := vectors.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range set.Seed {
		seed, err := FromDraws(v.Draws)
		if err != nil {
			t.Fatalf("%s: %v", v.Name, err)
		}
		checkSeed(t, v, seed)
	}
}

func checkSeed(t *testing.T, v vectors.Seed, seed Seed) {
	t.Helper()
	numbers := make([]int, Words)
	texts := make([]string, Words)
	for i, w := range seed.Words {
		numbers[i], texts[i] = w.Number, w.Text
	}
	extra := fmt.Sprintf("%d%d%d", seed.ExtraBits[0], seed.ExtraBits[1], seed.ExtraBits[2])
	if !slices.Equal(numbers, v.Numbers) || strings.Join(texts, " ") != v.Mnemonic {
		t.Errorf("%s: words %v, want %v", v.Name, numbers, v.Numbers)
	}
	if extra != v.ExtraBits || seed.Block != v.Block || hex.EncodeToString(seed.Entropy[:]) != v.Entropy ||
		int(seed.Checksum) != v.Checksum {
		t.Errorf("%s: extra %s block %d entropy %x checksum %d differ from the vector", v.Name, extra, seed.Block,
			seed.Entropy, seed.Checksum)
	}
}

// Words 1 to 23 come from the pairs directly; word 24 lies in the block chosen by the extra bits.
func TestFromDrawsInternalConsistency(t *testing.T) {
	set, err := vectors.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range set.Seed {
		seed, _ := FromDraws(v.Draws)
		for k := range pairs {
			index, _ := PairIndex(v.Draws[2*k], v.Draws[2*k+1])
			if seed.Words[k].Number != index+1 {
				t.Errorf("%s word %d: number %d, pair gives %d", v.Name, k+1, seed.Words[k].Number, index+1)
			}
		}
		if first, last := BlockRange(seed.Block); seed.Words[Words-1].Number < first || seed.Words[Words-1].Number > last {
			t.Errorf("%s: word 24 number %d outside block %d", v.Name, seed.Words[Words-1].Number, seed.Block)
		}
		if seed.Words[Words-1].Pair != (Pair{}) {
			t.Errorf("%s: word 24 must have no pair", v.Name)
		}
	}
}

// SPEC.md 4.5: after 23 words the eight valid final words lie in eight different blocks, and the candidate
// for extra bits b lies in block b. Checked for 4096 entropies derived deterministically from SHA-256.
func TestLastWordBlockProperty(t *testing.T) {
	for n := range 4096 {
		base := sha256.Sum256(fmt.Appendf(nil, "lottabits-last-word-%d", n))
		var blocks []int
		for b := range Blocks {
			entropy := base
			entropy[EntropyBytes-1] = entropy[EntropyBytes-1]&^0x07 | byte(b)
			last := Mnemonic(entropy)[Words-1]
			if last/BlockSize != b {
				t.Fatalf("entropy %x: candidate for bits %03b has index %d, outside block %d", entropy, b, last, b)
			}
			blocks = append(blocks, last/BlockSize)
		}
		if len(slices.Compact(blocks)) != Blocks {
			t.Fatalf("entropy %x: candidates do not cover all blocks: %v", base, blocks)
		}
	}
}

func TestFromDrawsRejectsInvalidInput(t *testing.T) {
	valid := slices.Repeat([]int{1}, Draws)
	cases := map[string]struct {
		draws []int
		kind  draws.Kind
	}{
		"45 draws": {valid[:45], draws.ExactCount},
		"47 draws": {append(slices.Clone(valid), 1), draws.ExactCount},
		"no draws": {nil, draws.ExactCount},
		"draw 0":   {append(slices.Clone(valid[:45]), 0), draws.OutOfRange},
		"draw 65":  {append([]int{65}, valid[1:]...), draws.OutOfRange},
	}
	for name, c := range cases {
		_, err := FromDraws(c.draws)
		var e *draws.Error
		if !errors.As(err, &e) || e.Kind != c.kind {
			t.Errorf("%s: got %v, want kind %d", name, err, c.kind)
		}
	}
}
