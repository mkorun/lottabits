# Security

LottaBits handles material that can control money (BIP39 seeds) and accounts (passwords, passphrases). It is a small
open-source project, maintained by one person. **It has not been audited.** It comes without any warranty (see
[LICENSE](LICENSE)); nobody can recover funds or secrets lost through its use.

## Use it safely

- **Never type real seed words, draws or results into a website**, an online tool or a chat, and never photograph them.
- **An everyday computer with Wi-Fi switched off for a moment is not a safe place for a real seed.** A compromised
  operating system can record the input and send it later. Preferred, in this order:
  1. a hardware wallet that offers the paper workflow for word 24 (the seed never touches a computer; some devices need
     their companion app to start a recovery, see [docs/hardware-wallets.md](docs/hardware-wallets.md)),
  2. a dedicated offline device,
  3. a freshly booted, non-persistent live system with networking disabled.

  Details: [docs/offline-use.md](docs/offline-use.md).
- Verify the binary before use ([docs/verification.md](docs/verification.md)) and run `lottabits selftest` on the device
  you will use.
- Try the whole seed workflow first with a wallet that holds no funds, and restore it once from your written words.
- Draw unobserved, keep chips identical and mix thoroughly; see [docs/threat-model.md](docs/threat-model.md).

## Found a problem?

Thank you for looking! If it could put someone's secrets at risk, please tell me privately rather than in a public issue:
use *Report a vulnerability* on the repository's Security tab. Everything else is welcome as a normal issue.

Particularly interesting: a wrong result for some input, any way secrets could leak or be stored, a printable that could
lead to a wrong seed, password or passphrase, documentation that could mislead, and weak spots in the build or release.
I will reply as soon as I can; there is no bug bounty.

## Supported versions

Only the latest release receives fixes. Algorithm and data identifiers (for example `lottabits-seed-v1`) never change
their meaning, so results of older releases stay reproducible with newer ones.
