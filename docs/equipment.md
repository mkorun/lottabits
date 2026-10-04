# Equipment

LottaBits needs very little: 88 numbered chips, an opaque container and the printables. This page explains what to look for.
The chips and the container are not secret; only the draws are.

## Chips

- **Identical except for the number.** Same size, shape, weight, material and surface. Plastic counting chips from school
  supplies work well; 25 mm is the default size of the inventory sheet (other sizes: `go run ./cmd/printables -chip-mm 30`).
- **Two sets that can share one container.** Chips 01–64 in one colour and 65–88 in another make it easy to take out 65–88
  for a seed. Colour does not matter for randomness as long as you never look into the container.
- **Numbers that cannot be felt.** Stickers of the same size and paper on every chip; no embossed or engraved numbers.
  Test it: mix the chips, close your eyes and try to tell a few apart by touch. If you can, the set is not good enough.
- **Underlined numbers.** `06`/`90`, `16`/`91`, `18`/`81`, `19`/`61` and `68`/`89` look alike upside down. Underline the
  numbers on the stickers, as the printables do.
- **Cut-out chips** from the printables, on card stock of at least 300 g/m², are a stopgap: edges and folds make them easy
  to feel. Use them to learn the method, not for a seed that protects real money.
- **Bingo balls** are a good alternative: common sets number 1 to 90. Remove 89 and 90 (and 65 to 90 for a seed). Use a
  set that marks `6` and `9` (underline or dot).

Check the set on the inventory sheet before every session and again afterwards: every circle covered exactly once.

## Containers

| Container | How to draw | Watch out for |
|---|---|---|
| **Bag** (standard) | close it, shake for about five seconds, reach into the middle without looking | opaque fabric, a drawstring, room for the chips to move (at least three times their volume) |
| **Box with a hand opening** (as for a raffle) | lid closed, shake in several directions, reach in through the opening | a fabric sleeve over the opening so that you cannot see inside; four to five times the chips' volume, otherwise flat chips only slide as a stack |
| **Bingo cage** with numbered balls | turn it several times, let it release one ball | turn at least five times before each release, so that balls near the gate are not favoured |
| Box with a slot that releases one chip | shake, then tip | chips next to the slot are favoured unless you shake well right before tipping; not recommended |

Never open a container and pick a chip by sight: choosing, even unconsciously, is the one failure that destroys
randomness completely.

## Every draw

1. Close the container and shake it (or turn the cage).
2. Reach into the middle without looking; take one chip.
3. Write down its number.
4. Put it back. Always.

## Why small imperfections are fine

Chips or dice are never perfectly equal, and they do not need to be. A chip that is drawn 20 % less often than it should
costs a seed about 0.02 bit of its 256. What matters are gross failures: chips that can be felt, mixing that leaves the last
chip on top, drawing without putting back, and looking. Details and numbers: [threat-model.md](threat-model.md).
