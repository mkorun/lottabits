# Prior art

LottaBits is not a new cryptographic primitive. It explores a simpler human interface to physical randomness, and it builds
on ideas that others published before. This page acknowledges the ones known to the author. It is **not** a worldwide
novelty or patent search; if you know of earlier or closer work, please open an issue.

All sources were accessed in October 2026.

## Passphrases from physical randomness

- **Diceware** by Arnold G. Reinhold, first published on `sci.crypt.research` on 1 August 1995: five dice rolls select one
  of 7776 words. LottaBits' passphrase mode is the same idea with two draws of 88 chips and a 7744-word list.
  <https://theworld.com/~reinhold/diceware.html>
- **EFF dice-generated passphrases** (2016): improved English Diceware lists; LottaBits' English list is derived from the
  EFF long list. <https://www.eff.org/dice>
- **German Diceware lists by dys2p**: the source of LottaBits' German list. <https://github.com/dys2p/wordlists-de>
- **Card-based methods**, among them **Pokerware** by Chris Wellons (2017) and **Deckware** by Aaron Toponce (2021):
  passphrases or raw entropy from shuffled playing cards. <https://github.com/skeeto/pokerware>,
  <https://github.com/atoponce/deckware>

## BIP39 seeds from physical randomness

- **Dice, coins and cards in wallets and tools**: hardware wallets and tools such as COLDCARD, SeedSigner and Krux accept
  dice rolls or coin flips as entropy, and SeedSigner and Krux calculate the final word from user-supplied bits.
  <https://github.com/SeedSigner/seedsigner>, <https://selfcustody.github.io/krux/>
- **BitBox02 "Seed generation with dice"**: dice select the first 23 words from a lookup table; the BitBox02 shows the eight
  valid final words. <https://bitbox.swiss/bitbox02/BitBox_Diceware_HowTo.pdf>
- **SeedPicker** by merland: paper slips like raffle tickets and a die select the first 23 words; an offline calculator
  computes the last word. <https://github.com/merland/seedpicker>
- **SeedSticks**: wooden sticks engraved with BIP39 words, drawn blindly from a bag and thrown back; the final checksum word
  is calculated by the signing device. <https://seedsticks.org/>
- **Two bingo machines with 32 and 64 balls**, proposed by the forum user philipma1957 on bitcointalk.org on
  30 October 2022 in the thread "Are dices for generating seed words fair?": `32 × 64 = 2048` covers the BIP39 list.
  <https://bitcointalk.org/index.php?topic=5395587.80>
- **SeedGrid**: two bags of identical numbered tokens, 64 and 32, give a coordinate in a printed grid of 2048 words
  (`index = (A − 1) × 32 + B`); the final word is calculated on an air-gapped device. <https://seedgrid.org/>

## What LottaBits adds

The combination below was not found in the sources above. That is no proof that nobody has published it.

- **One bag, one set of chips.** Seed, password and passphrase use the same 88 chips; a seed uses chips 01–64.
- **Folding instead of a second bag.** Two draws from the same 64 chips give 4096 pairs; folding the first draw onto
  32 groups maps exactly two pairs to every BIP39 word, with no rejection.
- **The folded bit is not thrown away.** The extra bits of the first three pairs are the three missing entropy bits of a
  256-bit seed, so no further random source is needed, and they select one of eight blocks for word 24, which makes the
  last word choosable on paper with a suitable hardware wallet.
- **Paper first**: booklets, record sheets and references in English and German generated from the same data as the
  program, each data page with a verifiable hash.

The project's own 2024/25 prototype already used one bag and the 64 × 64 folding, but drew the three missing bits from the
computer's random number generator; v1 removes that.
