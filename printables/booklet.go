package printables

import (
	"fmt"

	"github.com/mkorun/lottabits/internal/bip39"
	"github.com/mkorun/lottabits/internal/passphrase"
	"github.com/mkorun/lottabits/wordlists"
)

// Booklets are A5 pages imposed two per A4 landscape side for saddle stitching: print double-sided (flip on the short
// edge), stack, fold in the middle, staple. A cover sheet (front, back, inside front, inside back) wraps the content.

// Kinds of A5 pages.
const (
	coverFront    = "front"
	coverBack     = "back"
	insideFront   = "inside-front"
	insideBack    = "inside-back"
	content       = "content"
	pagesPerSheet = 4
)

// a5 is one A5 page of a booklet. Number is the content page number (0 for cover pages).
type a5 struct {
	Kind   string
	Number int
	Hash   string
	Data   any
}

// impose returns, for content pages 1..n (n a multiple of 4), the left and right page of every A4 side in printing
// order: sheet i carries (n−2i | 2i+1) on its front and (2i+2 | n−2i−1) on its back.
func impose(n int) [][2]int {
	sides := make([][2]int, 0, n/2)
	for k := range n / 2 {
		if k%2 == 0 {
			sides = append(sides, [2]int{n - k, k + 1})
		} else {
			sides = append(sides, [2]int{k + 1, n - k})
		}
	}
	return sides
}

// booklet imposes the content pages behind a cover sheet whose inside back page is given.
func booklet(pages []a5, back a5) ([]page, error) {
	if len(pages)%pagesPerSheet != 0 {
		return nil, fmt.Errorf("booklet needs a multiple of %d pages, got %d", pagesPerSheet, len(pages))
	}
	back.Kind = insideBack
	sides := []page{
		{Data: [2]a5{{Kind: coverBack}, {Kind: coverFront}}},
		sideOf(a5{Kind: insideFront}, back),
	}
	for _, s := range impose(len(pages)) {
		sides = append(sides, sideOf(pages[s[0]-1], pages[s[1]-1]))
	}
	return numberPages(sides), nil
}

func sideOf(left, right a5) page {
	p := page{Data: [2]a5{left, right}}
	for _, half := range []a5{left, right} {
		if half.Hash != "" {
			p.Hashes = append(p.Hashes, half.Hash)
		}
	}
	return p
}

// entry is one booklet entry: the first and second draw, the word number (seed only) and the word.
type entry struct {
	First, Second int
	Number        int
	Word          string
}

// columns arranges entries column by column into rows of up to n entries.
func columns(entries []entry, n int) [][]entry {
	rows := (len(entries) + n - 1) / n
	out := make([][]entry, rows)
	for i, e := range entries {
		out[i%rows] = append(out[i%rows], e)
	}
	return out
}

// seedGroup is one seed booklet page: first draws Low and Low+32 share 64 entries.
type seedGroup struct {
	Low, High int
	Rows      [][]entry
}

func seedEntry(first, second int) entry {
	index, _ := bip39.PairIndex(first, second)
	return entry{First: first, Second: second, Number: index + 1, Word: wordlists.BIP39English().Word(index)}
}

// seedBooklet has one page per group: page n lists first draws n and n+32 (SPEC.md 4.5 and 9).
func seedBooklet() ([]page, error) {
	const half = setSmall / 2
	pages := make([]a5, 0, half)
	for low := 1; low <= half; low++ {
		var entries []entry
		var lines []string
		for second := 1; second <= setSmall; second++ {
			e := seedEntry(low, second)
			entries = append(entries, e)
			lines = append(lines, fmt.Sprintf("%02d %02d %04d %s", e.First, e.Second, e.Number, e.Word))
		}
		group := seedGroup{Low: low, High: low + half, Rows: columns(entries, 2)}
		pages = append(pages, a5{Kind: content, Number: low, Hash: dataHash(lines), Data: group})
	}
	blockRows, blockLines := blocks()
	return booklet(pages, a5{Hash: dataHash(blockLines), Data: blockRows})
}

// passphrasePage is one passphrase booklet page: first draw First, 88 entries in three columns.
type passphrasePage struct {
	First int
	Rows  [][]entry
}

const passphraseColumns = 3

// passphraseBooklet has one page per first draw (SPEC.md 6 and 9).
func passphraseBooklet(list wordlists.List, strength []strengthRow) ([]page, error) {
	pages := make([]a5, 0, setLarge)
	for first := 1; first <= setLarge; first++ {
		var entries []entry
		var lines []string
		for second := 1; second <= setLarge; second++ {
			e := entry{First: first, Second: second, Word: list.Word(passphrase.Index(first, second))}
			entries = append(entries, e)
			lines = append(lines, fmt.Sprintf("%02d-%02d %s", first, second, e.Word))
		}
		p := passphrasePage{First: first, Rows: columns(entries, passphraseColumns)}
		pages = append(pages, a5{Kind: content, Number: first, Hash: dataHash(lines), Data: p})
	}
	return booklet(pages, a5{Data: strength})
}
