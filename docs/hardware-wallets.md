# Hardware wallets and word 24

After 23 words, a 24-word BIP39 seed has eight valid final words: one in each block of 256 BIP39 numbers. LottaBits'
three extra bits choose the block (SPEC.md 4.4), so a device can finish the seed without a computer if it either

- **lists the valid final words** (path A): take the word whose BIP39 number lies in the block of your extra bits.
  Exactly one does. Check the number in the booklet; never rely on the word's position in the device's list, or
- **accepts the final entropy bits** (path B): enter `extra_1`, `extra_2`, `extra_3` in this order; the device adds the
  checksum.

Otherwise calculate word 24 with `lottabits seed` on an offline device ([offline-use.md](offline-use.md)) and enter all 24
words.

## Devices

Status: research of October 2026 from vendor documentation and source code. **No entry has been tested on a real device by
this project yet.** Firmware changes; check your device's current documentation, and try the workflow with a test seed (for
example the `spec-mixed` vector in `vectors/vectors.json`) before using it for a real one.

| Device | Path | Notes | Source |
|---|---|---|---|
| SeedSigner (firmware 0.5.1 and later) | B | Tools → Calc 12th/24th word → Coin flip entropy: "Heads = 1", "Tails = 0", first flip is the most significant bit; enter the extra bits in order. Verified in the source code. | [`tools_screens.py`, `mnemonic_generation.py`](https://github.com/SeedSigner/seedsigner/tree/0736e066e2504f0383c23a51a46b58de387218f5/src/seedsigner) |
| Krux | B | "Binary Grid (manual)" entry; for the last word of a 24-word mnemonic only three bits are entered, Krux fills in the checksum. Accepts BIP39 word numbers. | [Krux documentation](https://selfcustody.github.io/krux/getting-started/usage/loading-a-mnemonic/) |
| COLDCARD Mk4, Q (firmware 4.0.0 and later) | A | After 23 words the device offers the eight valid final words; their order is not documented. | [firmware `seed.py`](https://github.com/Coldcard/firmware/blob/master/shared/seed.py) |
| BitBox02 (firmware 9.4.0 and later) | A | Shows all eight candidate words after 23 words; order not documented. **The recovery is started from the BitBoxApp**, so this path needs a computer or phone, although the words are entered only on the device. | [firmware changelog](https://github.com/BitBoxSwiss/bitbox02-firmware/blob/master/CHANGELOG.md), [BitBox dice guide](https://bitbox.swiss/bitbox02/BitBox_Diceware_HowTo.pdf) |
| Blockstream Jade, Jade Plus (firmware 0.1.43 and later) | A | Offers only valid final words; the starting letter and the first candidate shown are deliberately random, so position means nothing. | [Blockstream help](https://help.blockstream.com/blockstream-jade/add-more-security-functionality/calculate-the-final-word-from-a-provided-recovery-phrase-entry) |
| Specter DIY | (B) | A "fix" function keeps the entropy of an entered 24th word and recomputes its checksum: enter any word from the right block, then fix. Indirect; not documented as a final-word tool. | [`specter.py`](https://github.com/cryptoadvance/specter-diy/blob/master/src/gui/specter.py) |
| Foundation Passport, Passport Prime | – | Generate one valid final word; control over the three bits is not documented. Not suitable until clarified. | vendor release notes |
| Ledger, Trezor, KeepKey, Keystone 3 | – | Accept a complete 24-word phrase only; calculate word 24 with `lottabits` first. | vendor documentation |

A correction or a test report for any device is welcome as an issue (with firmware version), never with real seed words.
