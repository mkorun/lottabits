# Security

LottaBits handles material that can control money (BIP39 seeds) and accounts (passwords, passphrases). It is a small
open-source project by one person. **It has not been audited.** It comes without any warranty (see [LICENSE](LICENSE));
nobody can recover funds or secrets lost through its use.

## Use it safely

- **Never type real seed words, draws or results into a website**, an online tool or a chat, and never photograph them.
- **An everyday computer with Wi-Fi switched off for a moment is not a safe place for a real seed.** A compromised
  operating system can record the input and send it later. Preferred, in this order:
  1. a hardware wallet that offers the paper workflow for word 24 (no computer at all, see
     [docs/hardware-wallets.md](docs/hardware-wallets.md)),
  2. a dedicated offline device,
  3. a freshly booted, non-persistent live system with networking disabled.

  Details: [docs/offline-use.md](docs/offline-use.md).
- Verify the binary before use ([docs/verification.md](docs/verification.md)) and run `lottabits selftest` on the device
  you will use.
- Try the whole seed workflow first with a wallet that holds no funds, and restore it once from your written words.
- Draw unobserved, keep chips identical and mix thoroughly; see [docs/threat-model.md](docs/threat-model.md).

## Reporting a vulnerability

Please report privately through GitHub: **Security → Report a vulnerability** in this repository
(private vulnerability reporting). Do not open a public issue for a security problem.

Useful reports include: a wrong result for some input, any way the program could leak or persist secrets, problems in
the printables that could produce a wrong seed, password or passphrase, misleading documentation, and supply-chain
weaknesses in the build or release process. You will get an answer when the maintainer is available; this is a hobby
project without a bug bounty and without guaranteed response times.

## Supported versions

Only the latest release receives fixes. Algorithm and data identifiers (for example `lottabits-seed-v1`) never change
their meaning, so results of older releases stay reproducible with newer ones.
