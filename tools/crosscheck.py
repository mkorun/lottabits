#!/usr/bin/env python3
"""Independent second implementation of the LottaBits algorithms (SPEC.md sections 4 to 6).

Written separately from the Go code and in a different style (bit strings instead of bit arithmetic), using only the
Python standard library, so that anyone can recompute a result from the same draws.

    python3 tools/crosscheck.py seed                      < draws.txt
    python3 tools/crosscheck.py password                  < draws.txt
    python3 tools/crosscheck.py passphrase --wordlist de  < draws.txt
    python3 tools/crosscheck.py verify                    (checks vectors/ and the word-list hashes)
    python3 tools/crosscheck.py cli path/to/lottabits     (compares the compiled CLI with this implementation)

It never generates random numbers and never writes files.
"""

import argparse
import hashlib
import json
import math
import pathlib
import re
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent

WORDLISTS = {
    "bip39": ("wordlists/bip39-english.txt", "2f5eed53a4727b4bf8880d8f3f199efc90e58503646d9ff8eff3a2ed3b24dbda"),
    "en": ("wordlists/lottabits-en-7744-v1.txt", "5c4caefc140efbf20d30e481123fb4beadb1324c12c315971f2bfe600ddafcdb"),
    "de": ("wordlists/lottabits-de-7744-v1.txt", "8023bf123831341641b1e33a26abdfb4b1739a5e01b3fbad62f35e77704ea156"),
}

# SPEC.md 5.1, written out by class rather than derived from the excluded characters.
PASSWORD_ALPHABET = (
    "0123456789"
    "ABCDEFGHJKLMNOPQRSTUVWXYZ"
    "abcdefghijkmnopqrstuvwxyz"
    "!\"#$%&()*+,-./:;<=>?@[]^_{}~"
)
PASSWORD_ALPHABET_SHA256 = "a608b1e22ae80afcdbb5989a971da973d1632e38fca4ddcfbe57e919dab9badf"


def load_wordlist(name):
    path, expected = WORDLISTS[name]
    data = (ROOT / path).read_bytes()
    actual = hashlib.sha256(data).hexdigest()
    if actual != expected:
        raise SystemExit(f"{path}: SHA-256 {actual} does not match {expected}")
    return data.decode("utf-8").split("\n")[:-1]


def parse_draws(text, highest):
    draws = []
    for position, field in enumerate(re.split(r"[\s,]+", text.strip()), start=1):
        if field == "":
            continue
        if not re.fullmatch(r"[0-9]{1,2}", field) or not 1 <= int(field) <= highest:
            raise SystemExit(f"draw {position}: {field!r} is not a number from 01 to {highest:02d}")
        draws.append(int(field))
    return draws


def seed(draws):
    """SPEC.md section 4: 46 draws of 1..64 -> 24 BIP39 word numbers and details."""
    if len(draws) != 46 or not all(1 <= d <= 64 for d in draws):
        raise SystemExit("seed needs exactly 46 draws from 01 to 64")
    words = load_wordlist("bip39")
    pairs = [(draws[i], draws[i + 1]) for i in range(0, 46, 2)]
    indexes = [((a - 1) % 32) * 64 + (b - 1) for a, b in pairs]
    extra = ["0" if a <= 32 else "1" for a, _ in pairs[:3]]
    bits = "".join(format(index, "011b") for index in indexes) + "".join(extra)
    entropy = bytes(int(bits[i:i + 8], 2) for i in range(0, 256, 8))
    checksum = hashlib.sha256(entropy).digest()[0]
    all_bits = bits + format(checksum, "08b")
    numbers = [int(all_bits[i:i + 11], 2) + 1 for i in range(0, 264, 11)]
    return {
        "numbers": numbers,
        "mnemonic": " ".join(words[n - 1] for n in numbers),
        "extra_bits": "".join(extra),
        "block": int("".join(extra), 2),
        "entropy": entropy.hex(),
        "checksum": checksum,
    }


def mnemonic_from_entropy(entropy_hex):
    """Plain BIP39 for 256-bit entropy, used to check the official vectors."""
    words = load_wordlist("bip39")
    entropy = bytes.fromhex(entropy_hex)
    bits = "".join(format(byte, "08b") for byte in entropy) + format(hashlib.sha256(entropy).digest()[0], "08b")
    return " ".join(words[int(bits[i:i + 11], 2)] for i in range(0, 264, 11))


def password(draws):
    """SPEC.md section 5: one draw of 1..88 per character."""
    if not draws or not all(1 <= d <= 88 for d in draws):
        raise SystemExit("password needs at least one draw from 01 to 88")
    text = "".join(PASSWORD_ALPHABET[d - 1] for d in draws)
    classes = "".join("D" if c.isdigit() else "U" if c.isupper() else "L" if c.islower() else "S" for c in text)
    return {"password": text, "classes": classes, "bits": round(len(draws) * math.log2(88), 1)}


def passphrase(draws, wordlist):
    """SPEC.md section 6: two draws of 1..88 per word."""
    if not draws or len(draws) % 2 or not all(1 <= d <= 88 for d in draws):
        raise SystemExit("passphrase needs an even number of draws from 01 to 88")
    words = load_wordlist(wordlist)
    pairs = [(draws[i], draws[i + 1]) for i in range(0, len(draws), 2)]
    return {
        "passphrase": " ".join(words[(a - 1) * 88 + (b - 1)] for a, b in pairs),
        "coordinates": [f"{a:02d}-{b:02d}" for a, b in pairs],
        "bits": round(len(pairs) * math.log2(7744), 1),
    }


def check(label, actual, expected, failures):
    status = "ok" if actual == expected else "FAIL"
    if status == "FAIL":
        failures.append(label)
    print(f"{status:4}  {label}")


def verify():
    failures = []
    check("password alphabet SHA-256", hashlib.sha256(PASSWORD_ALPHABET.encode()).hexdigest(),
          PASSWORD_ALPHABET_SHA256, failures)
    for name in WORDLISTS:
        load_wordlist(name)
        print(f"ok    word list {name} SHA-256")
    trezor = json.loads((ROOT / "vectors/trezor-vectors.json").read_text(encoding="utf-8"))
    for entropy_hex, mnemonic, *_ in trezor["english"]:
        if len(entropy_hex) == 64:
            check(f"trezor {entropy_hex[:16]}…", mnemonic_from_entropy(entropy_hex), mnemonic, failures)
    vectors = json.loads((ROOT / "vectors/vectors.json").read_text(encoding="utf-8"))
    for vector in vectors["seed"]:
        result = seed(vector["draws"])
        expected = {key: vector[key] for key in result}
        check(f"seed {vector['name']}", result, expected, failures)
    for vector in vectors["password"]:
        result = password(vector["draws"])
        check(f"password {vector['name']}", (result["password"], result["classes"]),
              (vector["password"], vector["classes"]), failures)
    for vector in vectors["passphrase"]:
        result = passphrase(vector["draws"], vector["wordlist"])
        check(f"passphrase {vector['name']}", (result["passphrase"], result["coordinates"]),
              (vector["passphrase"], vector["coordinates"]), failures)
    print(f"\n{len(failures)} failure(s)")
    return 1 if failures else 0


def run_cli(binary, arguments, draws):
    text = " ".join(f"{d:02d}" for d in draws)
    completed = subprocess.run([binary, *arguments], input=text, capture_output=True, text=True, check=False)
    if completed.returncode != 0:
        raise SystemExit(f"{binary} {' '.join(arguments)} failed: {completed.stderr}")
    return completed.stdout


def compare_cli(binary):
    """Feed every vector's draws to the compiled CLI and compare its output with this implementation."""
    failures = []
    vectors = json.loads((ROOT / "vectors/vectors.json").read_text(encoding="utf-8"))
    row = re.compile(r"^\d\d\s+(?:\d\d \d\d)?\s+(\d{4})\s+(\S+)$", re.MULTILINE)
    for vector in vectors["seed"]:
        output = run_cli(binary, ["seed", "--details"], vector["draws"])
        expected = seed(vector["draws"])
        rows = row.findall(output)
        actual = {
            "numbers": [int(number) for number, _ in rows],
            "mnemonic": " ".join(word for _, word in rows),
            "entropy": re.search(r"^Entropy: ([0-9a-f]{64})$", output, re.MULTILINE).group(1),
        }
        check(f"cli seed {vector['name']}", actual, {key: expected[key] for key in actual}, failures)
    for vector in vectors["password"]:
        output = run_cli(binary, ["password"], vector["draws"])
        expected = password(vector["draws"])
        actual = re.search(r"^Password:\s+(\S+)\nClasses:\s+([DULS]+)", output, re.MULTILINE).groups()
        check(f"cli password {vector['name']}", actual, (expected["password"], expected["classes"]), failures)
    for vector in vectors["passphrase"]:
        output = run_cli(binary, ["passphrase", "--wordlist", vector["wordlist"]], vector["draws"])
        expected = passphrase(vector["draws"], vector["wordlist"])
        actual = re.search(r"^Passphrase:\s+(.+)$", output, re.MULTILINE).group(1)
        check(f"cli passphrase {vector['name']}", actual, expected["passphrase"], failures)
    print(f"\n{len(failures)} failure(s)")
    return 1 if failures else 0


def main():
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("mode", choices=["seed", "password", "passphrase", "verify", "cli"])
    parser.add_argument("binary", nargs="?", help="path of the compiled lottabits binary (mode cli)")
    parser.add_argument("--wordlist", choices=["en", "de"], help="passphrase word list (required for passphrase)")
    args = parser.parse_args()
    if args.mode == "verify":
        return verify()
    if args.mode == "cli":
        if args.binary is None:
            parser.error("cli needs the path of the compiled binary")
        return compare_cli(args.binary)
    if args.mode == "passphrase" and args.wordlist is None:
        parser.error("passphrase needs --wordlist en|de")
    highest = 64 if args.mode == "seed" else 88
    draws = parse_draws(sys.stdin.read(), highest)
    if args.mode == "seed":
        result = seed(draws)
    elif args.mode == "password":
        result = password(draws)
    else:
        result = passphrase(draws, args.wordlist)
    print(json.dumps(result, indent=2, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
