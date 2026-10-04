# Word-list provenance

The normative definitions are in `SPEC.md` section 7. This file records where every list comes from, under which license it is
redistributed, and how to reproduce the derived lists. `wordlists_test.go` checks all of it.

## `bip39-english.txt` – `bip39-english`

- Source: <https://github.com/bitcoin/bips/blob/ce1862ac6bcffa1dd20aad858380e51e66e949ea/bip-0039/english.txt>
- Unchanged. SHA-256 `2f5eed53a4727b4bf8880d8f3f199efc90e58503646d9ff8eff3a2ed3b24dbda`.
- License: MIT, as declared in the header of BIP 39 (Marek Palatinus, Pavol Rusnak, Aaron Voisine, Sean Bowe).

## `lottabits-en-7744-v1.txt` – derived from the EFF Long Wordlist

- Source: <https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt> (file dated 2016-07-18), stored unchanged as
  `upstream/eff_large_wordlist.txt`, SHA-256 `addd35536511597a02fa0a9ff1e5284677b8883b83e986e43f15a3db996b903e`.
- License: [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) (EFF copyright policy, <https://www.eff.org/copyright>).
  Full license text: `../LICENSE-DOCS`.
- Attribution: "EFF Long Wordlist" by the Electronic Frontier Foundation, CC BY 4.0.
  Changes: the dice numbers were removed and only the first 7744 of 7776 entries are kept (in upstream order).
- Derived file SHA-256 `5c4caefc140efbf20d30e481123fb4beadb1324c12c315971f2bfe600ddafcdb`.

## `lottabits-de-7744-v1.txt` – derived from dys2p `de-7776-v1`

- Source: <https://github.com/dys2p/wordlists-de/blob/6ef31b9aefb8735a7b066592393d12843ec502cd/de-7776-v1.txt>, stored unchanged
  as `upstream/dys2p-de-7776-v1.txt`, SHA-256 `440fa02c65591328d6351435d3824c27b483a049f4eca0b13456d8c5090442e7`.
- License: multi-licensed under Unlicense, CC0 1.0 and BSD-3-Clause; LottaBits uses it under CC0 1.0. Upstream license file:
  `upstream/dys2p-LICENSE.txt`.
- Changes: only the first 7744 of 7776 entries are kept (in upstream order).
- Derived file SHA-256 `8023bf123831341641b1e33a26abdfb4b1739a5e01b3fbad62f35e77704ea156`.

## Reproduce

```sh
sha256sum bip39-english.txt upstream/eff_large_wordlist.txt upstream/dys2p-de-7776-v1.txt
head -n 7744 upstream/eff_large_wordlist.txt | cut -f 2 | sha256sum   # lottabits-en-7744-v1
head -n 7744 upstream/dys2p-de-7776-v1.txt | sha256sum                # lottabits-de-7744-v1
```

## Why the first 7744 entries

Two draws of 88 chips give `88 × 88 = 7744` coordinates. Taking a prefix in upstream order is the simplest rule that anyone can
reproduce without judgement; no word was chosen or removed by preference. It drops the last 32 upstream entries
(`yield` … `zoom` in English, `zweckgebunden` … `zypressen` in German).
