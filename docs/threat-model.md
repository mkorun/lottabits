# Threat model

What can go wrong when LottaBits is used, and what the project does about it. Each entry names the threat, its effect, the
mitigation and what remains. The software is small; most risks are physical or human.

## Assets and assumptions

- **Assets:** the draws, the resulting seed, password or passphrase, and the paper records.
- **Trusted:** the person drawing, the chips and bag they checked, the device that runs `lottabits` (only for the seed
  checksum), the printer for the printables.
- **Not trusted:** any network, any website, any device that was online and is not dedicated to this task, downloaded files
  until verified.
- **Out of scope:** what happens to a secret after it has been created (wallet security, password managers, backups of
  the written words beyond the advice below), and attacks by someone who controls the person drawing.

## Physical randomness

| Threat | Effect | Mitigation | Residual risk |
|---|---|---|---|
| Chips differ by touch, weight, size or shape (worn edges, sticker thickness, cut-out paper) | the drawer can feel numbers, draws are biased | identical chips with stickers of the same size; never look into the bag; cut-out chips documented as low assurance | small differences remain undetectable by eye |
| Incomplete set (lost or duplicated chip) | some numbers impossible or twice as likely | inventory sheet before every session: every circle covered exactly once | a chip lost during the session; check again afterwards |
| Chips 65–88 left in the bag for a seed | a draw above 64 | separate colours for 65–88; the record sheet and the command reject numbers above 64 | none: the draw is invalid, not biased (draw again) |
| Poor mixing | consecutive draws correlate | shake or stir the bag after putting the chip back, every time | depends on the person; not measurable by the software |
| Drawing without replacement | later draws depend on earlier ones, fewer possible results | every procedure says "put it back"; record sheets have one circle per draw | a human slip is not detectable afterwards |
| Choosing instead of drawing (redrawing a result one dislikes) | the result is no longer uniform | documentation never suggests redrawing a valid result | user behaviour |
| Observation (people, cameras, phones, smart speakers, windows) | the secret is known to someone else | draw alone, no devices in view; the printables say so | depends on the room |

The mapping itself adds no bias: all 4096 pairs map to each BIP39 word exactly twice and the extra bit is independent of
the word (tested exhaustively); password and passphrase mappings are one-to-one.

## Recording and transcription

| Threat | Effect | Mitigation | Residual risk |
|---|---|---|---|
| Misread number (6/9 rotated, 1/7, smudged) | wrong word or character | two-digit numbers, underlined on chips and printables; record sheets keep every draw | handwriting |
| Misread character (0/O, l/1, case) | wrong password | alphabet excludes six look-alike and quoting characters (SPEC 5.1); slashed-zero font; class letter D/U/L/S under every character; the draw number is the canonical record | look-alikes such as `S`/`5` remain |
| Wrong booklet page or row | wrong word | page number equals the first draw; every entry shows its second draw and word number | a careless lookup |
| Typing error in the command | wrong result | strict parsing, every entry validated and repeated on error, no defaults | a valid but wrong number |
| Wrong word 24 on a hardware wallet | wallet built on another seed than recorded | pick the candidate by its number in the block of the extra bits, never by position | none if the number is checked |

## Data and printables

| Threat | Effect | Mitigation | Residual risk |
|---|---|---|---|
| Manipulated or corrupted word list | wrong or predictable words | lists pinned by SHA-256, re-derived from pinned upstream files in tests, checked by `lottabits selftest` | an upstream list with weak words would be faithfully reproduced |
| Manipulated or misprinted booklet, map or reference | wrong mapping on paper | data hash on every data page; independent check of the generated files; spot checks ([verification](verification.md)) | a printer or someone with access altering one page after the check |
| Scaled print | chips do not fit the inventory sheet; small type | 50 mm calibration line on chip sheets; "print at 100 %" on every document | none once checked |

## Software and devices

| Threat | Effect | Mitigation | Residual risk |
|---|---|---|---|
| Bug in the mapping or BIP39 code | wrong seed, unrecoverable funds | exhaustive and vector tests, official BIP39 vectors, an independent Python implementation compared in CI and against the compiled binary, a third BIP39 implementation for the published vectors | bugs not covered by any test |
| Hidden randomness in the program | secret not from the chips, possibly predictable | randomness, network, subprocess and logging imports rejected by the linter; reproducible builds | the Go toolchain and standard library are trusted |
| Manipulated binary (download, mirror, supply chain) | arbitrary behaviour | checksums, build attestations and reproducible builds; build it yourself from source | a compromised toolchain on both sides |
| Compromised or online computer | seed recorded and exfiltrated | hardware-wallet path needs no computer; otherwise a dedicated offline device or a fresh live system ([offline use](offline-use.md)) | firmware-level compromise of that device |
| Secrets left behind (terminal scroll-back, swap, memory) | later recovery from the device | no files, no logs, input never in arguments; advice to close the terminal and power off | Go cannot reliably erase memory; swap on persistent storage; use a live system without swap |
| Dependency compromise | malicious code in the build | standard library only, checked by the gate; pinned toolchain and CI actions by commit hash | the Go toolchain itself |

## Records after creation

| Threat | Effect | Mitigation | Residual risk |
|---|---|---|---|
| Loss or destruction of the written words | funds or accounts lost | record sheet warns to keep it like the seed; test a restore before funding | storage is the user's responsibility |
| Drafts and scrap paper | secret leaks later | destroy drafts; keep the record sheet only | user behaviour |
| A seed backup that is stolen | funds lost | out of scope; consider a BIP39 passphrase or a multisignature setup with your wallet | — |

## What LottaBits does not protect against

Coercion of the user, a malicious printer firmware that changes printed content in a targeted way, a compromised hardware
wallet, and weaknesses of BIP39 itself. LottaBits is not audited; treat it as a carefully tested hobby project.
