# Test vectors

- `vectors.json`: LottaBits vectors (SPEC.md section 10). Draws are 1-based token numbers. Besides the hand-chosen vectors,
  `sha256-*` vectors use draws derived deterministically from SHA-256 of a fixed label (no randomness), and `block-*`
  vectors cover each of the eight word-24 blocks.
- `trezor-vectors.json`: the official BIP39 vectors from
  <https://github.com/trezor/python-mnemonic/blob/b57a5ad77a981e743f4167ab2f7927a55c1e82a8/vectors.json>, unchanged,
  SHA-256 `fa3b937b7cff9c9b8ecd3aa011faeb8d6dd67993174b72326e83f4de8fdb30f8`, MIT License, Copyright (c) 2013-2016 Pavol Rusnak.
  Only the English entries with 256-bit entropy are used.

## How the LottaBits vectors are verified

1. The Go tests (`go test ./...`) and `lottabits selftest` recompute every vector.
2. `tools/crosscheck.py verify`, an independent Python implementation, recomputes every vector and checks the official BIP39
   vectors.
3. Before the vectors were committed, the entropy and mnemonic of every seed vector were also checked with a third, external
   BIP39 implementation (`bitcoinjs/bip39` 3.1.0, `entropyToMnemonic` and `validateMnemonic`); all matched.
