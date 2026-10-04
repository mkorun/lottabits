package printables

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"github.com/mkorun/lottabits/internal/bip39"
	"github.com/mkorun/lottabits/internal/password"
	"github.com/mkorun/lottabits/wordlists"
)

// Page geometry in millimetres (see style.css: 6 mm padding, 7 mm header, 5 mm footer, 2 mm body padding).
const (
	a4Short    = 210.0
	a4Long     = 297.0
	pagePad    = 6.0
	headHeight = 7.0
	footHeight = 5.0
	bodyPad    = 2.0
	chipGap    = 1.5
	setSmall   = 64
	setLarge   = 88
)

// page is one printed A4 page (one side of a sheet). Hashes are the data hashes (SPEC.md 9.1) printed on it, in
// document order; booklet sides carry one per A5 page with data.
type page struct {
	Number, Total int
	Hashes        []string
	Data          any
}

// numberPages sets page numbers and totals.
func numberPages(pages []page) []page {
	for i := range pages {
		pages[i].Number, pages[i].Total = i+1, len(pages)
	}
	return pages
}

// dataHash is the first 16 hex digits of SHA-256 over the page's data lines, each terminated by LF, in groups of four.
func dataHash(lines []string) string {
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n") + "\n"))
	h := hex.EncodeToString(sum[:8])
	return h[0:4] + " " + h[4:8] + " " + h[8:12] + " " + h[12:16]
}

// chip is one chip circle; High marks chips 65–88.
type chip struct {
	N    int
	High bool
}

// chipGrid lays chips of the given diameter on landscape pages, keeping 01–64 and 65–88 on separate pages.
type chipGrid struct {
	Columns  int
	Diameter float64
}

func newChipGrid(diameter float64) (chipGrid, int) {
	width := a4Long - 2*pagePad
	height := a4Short - 2*pagePad - headHeight - footHeight - bodyPad
	columns := int(math.Floor((width + chipGap) / (diameter + chipGap)))
	rows := int(math.Floor((height + chipGap) / (diameter + chipGap)))
	return chipGrid{Columns: columns, Diameter: diameter}, columns * rows
}

func chipPages(diameter float64) ([]page, chipGrid, error) {
	grid, capacity := newChipGrid(diameter)
	if capacity < 1 {
		return nil, grid, fmt.Errorf("chip diameter %.1f mm does not fit on A4", diameter)
	}
	var pages []page
	for _, set := range [][2]int{{1, setSmall}, {setSmall + 1, setLarge}} {
		for first := set[0]; first <= set[1]; first += capacity {
			var chips []chip
			for n := first; n <= set[1] && n < first+capacity; n++ {
				chips = append(chips, chip{N: n, High: n > setSmall})
			}
			pages = append(pages, page{Data: chips})
		}
	}
	return numberPages(pages), grid, nil
}

// mapCell is one character of the password map. NameKey is the locale key of a symbol's name, or empty.
type mapCell struct {
	N       int
	Char    string
	Class   string
	NameKey string
}

func passwordMapPage() []page {
	var cells []mapCell
	var lines []string
	for i := range len(password.Alphabet) {
		c := password.Alphabet[i]
		cell := mapCell{N: i + 1, Char: string(c), Class: string(password.ClassOf(c))}
		if cell.Class == string(password.Symbol) {
			cell.NameKey = fmt.Sprintf("sym_%02x", c)
		}
		cells = append(cells, cell)
		lines = append(lines, fmt.Sprintf("%02d %s %s", cell.N, cell.Char, cell.Class))
	}
	return numberPages([]page{{Hashes: []string{dataHash(lines)}, Data: cells}})
}

// block is one row of the word-24 block table.
type block struct {
	Bits                string
	Block               int
	From, To            int
	FirstWord, LastWord string
}

func blocks() ([]block, []string) {
	list := wordlists.BIP39English()
	var out []block
	var lines []string
	for b := range bip39.Blocks {
		from, to := bip39.BlockRange(b)
		row := block{Bits: fmt.Sprintf("%03b", b), Block: b, From: from, To: to,
			FirstWord: list.Word(from - 1), LastWord: list.Word(to - 1)}
		out = append(out, row)
		lines = append(lines, fmt.Sprintf("%s %d %04d-%04d %s %s", row.Bits, row.Block, row.From, row.To, row.FirstWord, row.LastWord))
	}
	return out, lines
}

// strengthRow is one row of a strength table on a record sheet.
type strengthRow struct {
	Length int
	Bits   float64
}

func strengthRows(lengths []int, bits func(int) float64) []strengthRow {
	rows := make([]strengthRow, len(lengths))
	for i, n := range lengths {
		rows[i] = strengthRow{Length: n, Bits: bits(n)}
	}
	return rows
}

// rows returns 1..n for record sheets.
func rows(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i + 1
	}
	return out
}
