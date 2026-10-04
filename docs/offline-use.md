# Offline use

Passwords and passphrases need no computer: use the character map or the passphrase booklet. A seed needs one calculation,
the BIP39 checksum of word 24. This guide is about where to do that calculation when the seed protects real money.

**The chips provide all randomness. The device only calculates, but it sees the whole seed.** Choose it accordingly.

## Option A: a hardware wallet (no computer)

If your hardware wallet lets you enter 23 words and then either lists the valid final words or accepts the final entropy
bits, you need no computer: pick the final word by its number in the block of your extra bits, or enter the bits. Which
devices do this, and how: [hardware-wallets.md](hardware-wallets.md). Some devices need a companion app to start the
recovery; the words themselves are still entered only on the device.

## Option B: a dedicated offline device

A device used only for this purpose and never connected to a network, for example a Raspberry Pi with a freshly written
minimal system:

1. Write a trusted, minimal operating system image to a new or wiped memory card, on a separate computer.
2. Copy a verified `lottabits` binary ([verification.md](verification.md)) onto the card before the first boot.
3. Never connect Ethernet. On models with radios, disable them before the first boot (Raspberry Pi: `dtoverlay=disable-wifi`
   and `dtoverlay=disable-bt` in `config.txt`). Software switches can be undone by software; a model without radios (such as
   the Raspberry Pi Zero v1.3) is a stronger boundary, but none is required.
4. Avoid swap and shell history on the card; do not save anything.
5. Run `lottabits selftest`, then `lottabits seed`, write the words down, power off.
6. Keep the card only for this purpose, or wipe it.

## Option C: a fresh live system

A computer you already own, booted from a read-only live system that keeps nothing (for example Tails with networking
disabled at startup):

1. Prepare the live medium and a USB stick with the verified binary on another computer.
2. Disconnect the network cable; boot with networking disabled; do not unlock persistent storage or internal disks.
3. Run `lottabits selftest` and `lottabits seed` from the USB stick.
4. Write the words down, close the terminal, shut down.

## Not recommended: an everyday computer

Switching Wi-Fi off for a moment on your normal Windows, macOS or Linux installation is **not** a safe way to create a real
seed. Malware already present can record what you type and send it later. Use such a computer only with test draws.

## In every case

- Run `lottabits selftest` on the device you use.
- Enter the draws when prompted; never as command-line arguments, never from a file you keep.
- The terminal shows the result; afterwards close it and clear its scroll-back, then power off.
- Check the result on paper: words 1 to 23 must match your booklet lookups, word 24 must lie in the block of your extra bits.
- Before you fund the wallet, restore it once from your written words.
