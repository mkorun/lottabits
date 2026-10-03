# LottaBits specification

Status: draft for v1. This document is normative. The CLI, the printables and the tests implement exactly these rules.
The key words MUST, MUST NOT, SHOULD and MAY are used as described in RFC 2119.

LottaBits is not a new cryptographic primitive. It is a human interface to physical randomness: numbered tokens drawn from
one bag create secrets; software only calculates where mathematics requires it.

## 1. Versioned identifiers

Every algorithm and data set has an immutable identifier. Any change to a mapping, alphabet, word list or rule MUST get a new
identifier; an existing identifier never changes its meaning. Release versions (semantic versioning) are independent of these.

| Identifier | Section |
|---|---|
| `lottabits-seed-v1` | 4 |
| `lottabits-password-88-v1` | 5 |
| `lottabits-passphrase-7744-v1` | 6 |
| `bip39-english` | 7 |
| `lottabits-en-7744-v1`, `lottabits-de-7744-v1` | 7 |

## 2. Terms

- **Token:** a physical marker showing one number. The full set holds tokens `01` to `88`.
- **Set 64:** tokens `01`–`64`. **Set 88:** tokens `01`–`88`.
- **Draw:** taking one token blindly from the bag, recording its number, putting it back and mixing. A draw is an integer.
- **Number:** a 1-based position as printed for humans (BIP39 word number `1`–`2048`, password character `1`–`88`).
- **Index:** the 0-based position used in calculations. `index = number − 1`. Printables and CLI output show numbers, never indexes.
- **Pair:** two consecutive draws, the **first draw** `a` and the **second draw** `b`.

Draws are written with two digits (`01`, `09`, `64`, `88`).

## 3. Physical procedure

This procedure applies to all modes. Printables and documentation MUST describe it.

1. Use tokens that differ only in their printed number: same shape, size, weight, material and surface. Tokens `65`–`88`
   MAY be visually distinct (for example another colour) so the two sets can share one bag and be separated easily.
2. Before each session, lay all tokens of the set onto the inventory sheet and confirm that every number is present exactly once.
3. Seed mode uses set 64: tokens `65`–`88` MUST be removed from the bag. Password and passphrase modes use set 88.
4. For every draw: mix thoroughly, draw one token without looking, record its number, put it back. Draws are always
   **with replacement**.
5. Draw unobserved: no cameras, no onlookers, no recording devices in view.

Under these conditions every draw is independent and uniformly distributed over the set. LottaBits makes no claim beyond that;
it does not claim that tokens are more random than fair dice.

## 4. Seed: `lottabits-seed-v1`

Produces a 24-word BIP39 mnemonic (256 bits of entropy) from 46 draws of set 64. The software adds no randomness.

### 4.1 Input
46 draws `d1 … d46`, each in `1..64`. Pair `k` (`k = 1..23`) is `a_k = d(2k−1)`, `b_k = d(2k)`.

### 4.2 Words 1 to 23
For each pair:

```text
index_k = ((a_k − 1) mod 32) × 64 + (b_k − 1)        index_k ∈ 0..2047
number_k = index_k + 1                               number_k ∈ 1..2048
extra_k  = 0 if a_k ≤ 32, 1 if a_k ≥ 33
```

Word `k` is the BIP39 word with index `index_k`. First draws `a` and `a + 32` select the same group of 64 words; the extra bit
tells them apart. Each of the 4096 pairs maps to exactly one `(index, extra)` combination and each combination has exactly one
pair, so every word index has probability `2/4096 = 1/2048` and the extra bit is independent of the word. No rejection is needed.

### 4.3 Entropy
The 256-bit entropy is the concatenation, most significant bit first, of:

1. `index_1 … index_23`, each as 11 bits (253 bits),
2. `extra_1`, `extra_2`, `extra_3` (3 bits).

The 32 entropy bytes are this bit string read in big-endian order. The extra bits of pairs 4 to 23 are not used.
The 46 draws carry `46 × 6 = 276` bits; 256 are used.

### 4.4 Word 24
```text
checksum = first byte of SHA-256(entropy)                          (BIP39 checksum for 256-bit entropy)
block    = 4 × extra_1 + 2 × extra_2 + extra_3                     block ∈ 0..7
index_24 = block × 256 + checksum
```

Word 24 lies in block `block`, the BIP39 numbers `256 × block + 1` to `256 × block + 256`:

| extra bits | block | numbers |
|---|---:|---|
| `000` | 0 | 1–256 |
| `001` | 1 | 257–512 |
| `010` | 2 | 513–768 |
| `011` | 3 | 769–1024 |
| `100` | 4 | 1025–1280 |
| `101` | 5 | 1281–1536 |
| `110` | 6 | 1537–1792 |
| `111` | 7 | 1793–2048 |

### 4.5 Paper and hardware-wallet procedure
- Words 1–23: the BIP39 booklet has one page per group, labelled with both first draws (`01 / 33` … `32 / 64`), and on each
  page entries `01`–`64` for the second draw, each showing the word number and the word.
- The record sheet has 23 rows with the two draws, the word number, the word, and the extra bit for rows 1–3.
- Word 24 with a hardware wallet that, after 23 words, offers the eight valid final words: exactly one candidate has a word
  number in the block from 4.4. The user MUST identify it by its number in the booklet. The order in which a device lists the
  candidates MUST NOT be relied on.
- Informative: the English BIP39 list is sorted alphabetically, so a list of the eight candidates in alphabetical order has the
  correct word at position `block + 1`. This is a cross-check only, never the rule.
- Word 24 with a device that accepts the final entropy bits directly (for example as coin flips): the user enters
  `extra_1`, `extra_2`, `extra_3` in this order; they are the three most significant bits of word 24 and the device adds the
  checksum. No candidate list is involved.
- `docs/hardware-wallets.md` lists which devices support which of these two paths, with sources, and marks every entry that
  has not been tested on a real device. Some devices need a companion app to start a recovery; the documentation MUST say so
  and MUST NOT claim a computer-free workflow for them.
- Word 24 without such a device is calculated by the CLI (section 8) on an offline device.

## 5. Password: `lottabits-password-88-v1`

One draw of set 88 selects one character. The software adds no randomness.

### 5.1 Alphabet
The 88 printable ASCII characters `0x21`–`0x7E` without `'` `I` `\` `` ` `` `l` `|`, grouped in this order:

| Numbers | Class | Characters |
|---|---|---|
| 01–10 | `D` digit | `0123456789` |
| 11–35 | `U` upper case | `ABCDEFGHJKLMNOPQRSTUVWXYZ` |
| 36–60 | `L` lower case | `abcdefghijkmnopqrstuvwxyz` |
| 61–88 | `S` symbol | `` !"#$%&()*+,-./:;<=>?@[]^_{}~ `` |

Canonical string (88 bytes, ASCII, no trailing newline), SHA-256 `a608b1e22ae80afcdbb5989a971da973d1632e38fca4ddcfbe57e919dab9badf`:

```text
0123456789ABCDEFGHJKLMNOPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz!"#$%&()*+,-./:;<=>?@[]^_{}~
```

The six excluded characters remove quoting characters and the `1 l I |` look-alikes. The order groups characters for paper
lookup; it has no security effect.

### 5.2 Mapping
Draw `n` (`1..88`) selects character number `n`. A password of length `L` needs `L` draws.

### 5.3 Presentation
- The draw numbers are the canonical record. Record sheets MUST provide a column for the draw number next to the character.
- Every presentation of a password (CLI output, record sheet) MUST also show the class of each character (`D`, `U`, `L`, `S`)
  aligned under it, so that look-alikes such as `0`/`O`, `S`/`s` or `,`/`.` can be resolved.
- The character map MUST show `0` with a slash or dot and name every symbol (for example `, comma`, `. period`).

### 5.4 Strength
Each draw contributes `log2(88) ≈ 6.459` bits. 12 draws ≈ 77.5 bits, 16 draws ≈ 103.4 bits, 20 draws ≈ 129.2 bits.
The recommended minimum is 12 characters.

## 6. Passphrase: `lottabits-passphrase-7744-v1`

Two draws of set 88 select one word from a 7744-entry word list. The software adds no randomness.

### 6.1 Mapping
```text
a, b ∈ 1..88
index = (a − 1) × 88 + (b − 1)       index ∈ 0..7743
```

The coordinate of a word is written `AA-BB` (for example `07-42`). The booklet lists all words in coordinate order.
Words of a passphrase are joined with a single space.

### 6.2 Strength
Each word contributes `log2(7744) ≈ 12.919` bits (two draws). 6 words ≈ 77.5 bits (comparable to six words of a 7776-entry
Diceware list, 77.55 bits), 7 words ≈ 90.4 bits. The recommended minimum is 6 words.
Password and passphrase have the same strength per draw: 12 draws give ≈ 77.5 bits in either mode.

### 6.3 Word-list selection
The word list is chosen explicitly and independently of the user-interface language. Both lists are equally valid; the choice
only affects which words the user memorises.

## 7. Data

### 7.1 File format
Word lists are UTF-8 text without byte-order mark, one entry per line, each line terminated by `LF`, no empty lines.

### 7.2 Word lists

| Identifier | Entries | Derivation | SHA-256 of the derived file |
|---|---:|---|---|
| `bip39-english` | 2048 | unchanged upstream file | `2f5eed53a4727b4bf8880d8f3f199efc90e58503646d9ff8eff3a2ed3b24dbda` |
| `lottabits-en-7744-v1` | 7744 | first 7744 words of the upstream list, upstream order, dice numbers removed | `5c4caefc140efbf20d30e481123fb4beadb1324c12c315971f2bfe600ddafcdb` |
| `lottabits-de-7744-v1` | 7744 | first 7744 lines of the upstream list, upstream order | `8023bf123831341641b1e33a26abdfb4b1739a5e01b3fbad62f35e77704ea156` |

| Identifier | Upstream | Immutable reference | Upstream SHA-256 | License |
|---|---|---|---|---|
| `bip39-english` | `github.com/bitcoin/bips`, `bip-0039/english.txt` | commit `ce1862ac6bcffa1dd20aad858380e51e66e949ea` | as above | MIT (BIP 39 header) |
| `lottabits-en-7744-v1` | EFF Long Wordlist, `www.eff.org/files/2016/07/18/eff_large_wordlist.txt` | file of 2016-07-18 (by hash) | `addd35536511597a02fa0a9ff1e5284677b8883b83e986e43f15a3db996b903e` | CC BY 4.0, © Electronic Frontier Foundation |
| `lottabits-de-7744-v1` | `github.com/dys2p/wordlists-de`, `de-7776-v1.txt` | commit `6ef31b9aefb8735a7b066592393d12843ec502cd` | `440fa02c65591328d6351435d3824c27b483a049f4eca0b13456d8c5090442e7` | Unlicense, CC0 or BSD-3-Clause (choice) |

The upstream files are stored unchanged next to the derived lists together with their license texts. A test re-derives each
list from its upstream file and compares both hashes. Known properties, verified by tests: all entries unique; `lottabits-de`
contains only `a`–`z`; `lottabits-en` contains `a`–`z` and the three hyphenated words `drop-down`, `felt-tip`, `t-shirt`;
in both derived lists no word is a prefix of another.

## 8. Command-line interface

Common rules for all commands:

- **Language:** `--lang en|de` (anywhere on the command line) selects the language of messages and labels only (default `en`).
  It never selects a word list. Numbers use the language's decimal separator (`77.5` / `77,5`).
- **Input:** draws are read from standard input, never from command-line arguments (which end up in shell history and process
  lists). A draw is one or two decimal digits (`7`, `07`); draws are separated by spaces, tabs, commas or line breaks.
- **Interactive mode** (standard input is a terminal): the CLI prompts for one pair (seed, passphrase) or one draw (password) at a
  time, validates it immediately and asks again after an invalid entry. An empty entry ends input for password and passphrase.
- **Batch mode** (standard input is not a terminal): all of standard input (at most 64 KiB) is read and validated; any invalid
  draw is an error. In interactive mode, end of input before all seed draws are entered is an error.
- **Echo:** input is visible. Hiding it would not protect the secret, which is displayed as the result in any case.
  Users clear the screen and the terminal scroll-back afterwards (documented in the offline guide).
- **Output:** results go to standard output, prompts and diagnostics to standard error. Nothing is written to files, logged,
  copied to the clipboard or sent over a network.
- **Exit codes:** `0` success, `1` invalid input or failed self-test, `2` usage error (unknown command or flag).
- **Errors** name the position and the rule, for example `draw 17: "65" is not a number from 01 to 64`.
- **Trust boundary:** standard input is untrusted (typing errors). Parsing is strict; there are no defaults for missing draws and
  no correction of invalid ones.

### 8.1 `lottabits seed`
- **Input:** exactly 46 draws in `1..64`.
- **Output:** for each of the 24 words its position, the pair (none for word 24), the word number (four digits) and the word;
  then the extra bits with the block and its number range, a hint for hardware wallets (section 4.5), and the lines
  `Entropy source: 46 token draws` and `Software-generated randomness: none`.
- **Flag** `--details`: additionally prints the entropy (hexadecimal) and the checksum byte, for cross-verification.
- **Errors:** fewer or more than 46 draws, a draw outside `1..64`.

Example (batch input of 46 draws `01`; rows 03 to 22 omitted):

```text
$ lottabits seed < draws.txt
LottaBits seed (lottabits-seed-v1)

#   Draws  Number  Word
01  01 01  0001    abandon
02  01 01  0001    abandon
23  01 01  0001    abandon
24         0103    art

Extra bits: 000 -> block 0, word numbers 1–256
Hardware wallet: pick the final word numbered 1–256, or enter the extra bits 000 in this order.
Entropy source: 46 token draws
Software-generated randomness: none
```

### 8.2 `lottabits password`
- **Input:** 1 or more draws in `1..88`.
- **Output:** the password, the class row under it, the draw numbers, and the strength (`N draws ≈ X bits`).
  Below 12 draws a warning goes to standard error.
- **Errors:** no draws, a draw outside `1..88`.

```text
$ printf '1 11 36 61 88 10' | lottabits password
Password:  0Aa!~9
Classes:   DULSSD  (D digit, U upper case, L lower case, S symbol)
Draws:     01 11 36 61 88 10
Strength:  6 draws ≈ 38.8 bits
```

Standard error: `Warning: fewer than the recommended 12 characters.`

### 8.3 `lottabits passphrase --wordlist en|de`
- **Flag** `--wordlist` is required: the CLI never chooses a word list implicitly.
- **Input:** an even number of draws (2 or more) in `1..88`.
- **Output:** the passphrase, each word with its coordinate, the word-list identifier and the strength. Below 6 words a warning
  goes to standard error.
- **Errors:** missing `--wordlist`, an odd number of draws, a draw outside `1..88`.

### 8.4 `lottabits selftest`
Runs the built-in test vectors (section 10) and checks the SHA-256 of every embedded word list and of the password alphabet.
Prints one line per check and exits with `1` if any check fails. Meant to be run on the offline device before real use.

### 8.5 `lottabits version`
Prints the release version, the source revision, the Go version, and every identifier from section 1 with its SHA-256 where one
exists.

## 9. Printables

Printables are a primary interface. They MUST be generated from the same data as the CLI (the embedded word lists and alphabet)
and from one set of templates for all languages; a translation changes text only, never numbers, words or layout of data.

Every printable MUST:

- fit A4 portrait or landscape and print legibly in black and white,
- use no external fonts, scripts, style sheets or network resources,
- show the release version and the identifiers of the data it contains on every page,
- show token numbers with two digits and underlined (so `06`/`90`, `16`/`91`, `18`/`81`, `19`/`61`, `68`/`89` cannot be
  confused when rotated) and mark tokens `65`–`88` with a distinct black-and-white feature (double ring).

Required printables, each in English and German:

| Printable | Content |
|---|---|
| Inventory sheet | one circle per token `01`–`88` at token size (default 25 mm, configurable) for the completeness check |
| Cut-out tokens | tokens `01`–`88` to cut from card stock; documented as a low-assurance option |
| Seed booklet | section 4.5, 32 pages of 64 entries, imposed as an A5 booklet on A4 |
| Seed record sheet | 23 rows: draws, word number, word, extra bit (rows 1–3); row 24 with the block |
| Seed quick reference | extra bits → block → number range (table of 4.4), first and last word of each block |
| Password character map | section 5.1 with numbers, classes, unambiguous glyphs and symbol names |
| Password record sheet | rows with draw number, character and class |
| Passphrase booklet | 7744 words in coordinate order, one per word-list identifier |
| Passphrase record sheet | rows with both draws, coordinate and word, plus the word-list identifier |
| Quick reference | one page: procedure of section 3 and the three modes |

## 10. Test vectors

Normative vectors live in `vectors/vectors.json`; the official BIP39 vectors used for the plain BIP39 part are in `vectors/trezor-vectors.json`. The following are part of this specification.

**Seed, all draws `01`:** words 1–23 `abandon` (number 1), extra bits `000`, entropy all zero, checksum `0x66`,
word 24 `art` (number 103).

**Seed, all draws `64`:** words 1–23 `zoo` (number 2048), extra bits `111`, entropy all `0xff`, checksum `0xaf`,
word 24 `vote` (number 1968).

**Seed, mixed:** draws

```text
53 08  12 47  33 01  64 64  07 19  40 22  01 63  28 35  17 50  61 02  45 30  09 58
36 11  24 44  57 13  03 39  49 26  62 05  20 31  15 54  42 60  29 10  31 64
```

give word numbers `1288 751 1 2048 403 470 63 1763 1074 1794 798 570 203 1516 1549 167 1050 1861 1247 950 636 1802 1984`,
extra bits `101` (block 5), entropy `a0ebb8007ff3247541f6e2863c058ea391957af060a6833d126f3b54f7c27dfd`, checksum `0xba`,
word 24 number 1467, mnemonic:

```text
path fruit abandon zoo crane deny amazing sword mail then glove elbow bone runway screen below lobster trigger orange issue exhaust thrive wave resemble
```

**Password:** draws `01 11 36 61 88 10` give `0Aa!~9`, classes `DULSSD`.

**Passphrase coordinates:**

| Coordinate | `lottabits-en-7744-v1` | `lottabits-de-7744-v1` |
|---|---|---|
| `01-01` | abacus | aalen |
| `01-88` | aghast | abladen |
| `02-01` | agile | ablagen |
| `44-44` | mandatory | kaltgestellt |
| `88-88` | yiddish | zwanzig |

## 11. Verification

- **Independent implementation:** `tools/crosscheck.py` is a second implementation of sections 4 to 6 in Python using only its
  standard library, written independently of the Go code. Users can run it to recompute any result from the same draws.
  CI runs it against `vectors/`.
- **Self-test:** `lottabits selftest` (8.4) on the device that will be used.
- **Releases** are reproducible (the same source and pinned toolchain give byte-identical binaries), signed without long-lived
  keys through the build platform's attestation (build provenance), and published with `SHA256SUMS` and an SBOM.
  `docs/verification.md` describes how to check a download and how to rebuild it.
- **Printables:** every data page carries a short hash of its data; `docs/verification.md` describes spot checks of booklet
  entries against an independent copy of the word list.

## 12. Non-goals for v1

- No random number generation of any kind, no network access, no persistence, no clipboard.
- No BIP39 lengths other than 24 words, no BIP39 word lists other than English.
- No passphrase word lists other than English and German, no further password alphabets.
- No graphical or browser user interface.
