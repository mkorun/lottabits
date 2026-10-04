#!/usr/bin/env python3
"""Independent check of generated printables (SPEC.md sections 9 and 11).

Reads the visible entries of the HTML files written by `go run ./cmd/printables -out DIR`, recomputes every mapping
with tools/crosscheck.py's own word lists and alphabet, recomputes every page's data hash and compares it with the
printed hash and with manifest.json, and checks that both languages print identical data.

    python3 tools/check_printables.py DIR
"""

import hashlib
import html
import json
import pathlib
import re
import sys

from crosscheck import PASSWORD_ALPHABET, load_wordlist

SEED_ENTRY = re.compile(r'<td class="b"><u>(\d\d)</u></td><td class="n">(\d{4})</td><td class="w">([a-z]+)</td>')
SEED_HALF = re.compile(r'<span class="first"><u>(\d\d)</u> / <u>(\d\d)</u></span>')
PP_FIRST = re.compile(r'<u class="first">(\d\d)</u>')
PP_ENTRY = re.compile(r'<td class="b"><u>(\d\d)</u>-<u>(\d\d)</u></td><td class="w">([a-z-]+)</td>')
CELL = re.compile(r'<span class="n"><u>(\d\d)</u></span><span class="k">([DULS])</span><span class="glyph">(.*?)</span>')
BLOCK = re.compile(r'<tr class="block"><td class="bits">([01]{3})</td><td>(\d)</td><td class="range">(\d{4})–(\d{4})</td>'
                   r'<td class="w">([a-z]+)</td><td class="w">([a-z]+)</td></tr>')
TOKEN = re.compile(r'<div class="token( high)?"[^>]*><u>(\d\d)</u></div>')
HASH = re.compile(r'<span class="hash">([0-9a-f ]+)</span>')


class Checker:
    def __init__(self):
        self.failures = []
        self.bip39 = load_wordlist("bip39")
        self.lists = {"en": load_wordlist("en"), "de": load_wordlist("de")}

    def expect(self, condition, message):
        if not condition:
            self.failures.append(message)

    @staticmethod
    def data_hash(lines):
        digest = hashlib.sha256(("\n".join(lines) + "\n").encode()).hexdigest()[:16]
        return " ".join(digest[i:i + 4] for i in range(0, 16, 4))

    def pages(self, text):
        return text.split('<section class="page">')[1:]

    def page_hash(self, label, page, lines, printed):
        found = HASH.search(page)
        self.expect(found is not None and found.group(1) == self.data_hash(lines), f"{label}: data hash differs")
        if found:
            printed.append(found.group(1))

    def seed_booklet(self, label, text, printed):
        numbers, data = set(), []
        for n, page in enumerate(self.pages(text), start=1):
            halves = SEED_HALF.findall(page)
            entries = SEED_ENTRY.findall(page)
            self.expect(len(halves) == 2 and len(entries) == 128, f"{label} page {n}: layout")
            lines = []
            for i, (second, number, word) in enumerate(entries):
                low, high = (int(x) for x in halves[i // 64])
                self.expect(high == low + 32, f"{label} page {n}: heading {low}/{high}")
                index = ((low - 1) % 32) * 64 + (int(second) - 1)
                self.expect(int(number) == index + 1 and word == self.bip39[index], f"{label}: {low} {second} {word}")
                numbers.add(int(number))
                lines.append((low, int(second), f"{low:02d} {second} {number} {word}"))
            self.page_hash(f"{label} page {n}", page, [line for *_, line in sorted(lines)], printed)
            data.extend(sorted(lines))
        self.expect(numbers == set(range(1, 2049)), f"{label}: not every word number 1–2048 exactly once")
        return data

    def passphrase_booklet(self, label, text, wordlist, printed):
        data = []
        for n, page in enumerate(self.pages(text), start=1):
            first = int(PP_FIRST.search(page).group(1))
            entries = PP_ENTRY.findall(page)
            self.expect(first == n and len(entries) == 88, f"{label} page {n}: layout")
            lines = []
            for a, b, word in entries:
                self.expect(int(a) == first and word == self.lists[wordlist][(first - 1) * 88 + int(b) - 1],
                            f"{label}: {a}-{b} {word}")
                lines.append((int(b), f"{a}-{b} {word}"))
            self.page_hash(f"{label} page {n}", page, [line for _, line in sorted(lines)], printed)
            data.extend(sorted(lines))
        return data

    def password_map(self, label, text, printed):
        cells = [(int(n), k, html.unescape(glyph)) for n, k, glyph in CELL.findall(text)]
        self.expect([n for n, _, _ in cells] == list(range(1, 89)), f"{label}: numbers")
        self.expect("".join(c for _, _, c in cells) == PASSWORD_ALPHABET, f"{label}: characters")
        for n, k, c in cells:
            want = "D" if c.isdigit() else "U" if c.isupper() else "L" if c.islower() else "S"
            self.expect(k == want, f"{label}: class of {n:02d}")
        self.page_hash(label, text, [f"{n:02d} {c} {k}" for n, k, c in cells], printed)
        return cells

    def seed_reference(self, label, text, printed):
        rows = BLOCK.findall(text)
        self.expect(len(rows) == 8, f"{label}: {len(rows)} blocks")
        for bits, block, start, end, first, last in rows:
            b = int(bits, 2)
            self.expect(int(block) == b and int(start) == 256 * b + 1 and int(end) == 256 * b + 256, f"{label}: block {b}")
            self.expect(first == self.bip39[256 * b] and last == self.bip39[256 * b + 255], f"{label}: words of block {b}")
        self.page_hash(label, text, [" ".join(f"{r[0]} {r[1]} {r[2]}-{r[3]} {r[4]} {r[5]}".split()) for r in rows], printed)
        return rows

    def tokens(self, label, text):
        found = TOKEN.findall(text)
        self.expect([int(n) for _, n in found] == list(range(1, 89)), f"{label}: tokens 01–88 once each")
        self.expect(all((high != "") == (int(n) > 64) for high, n in found), f"{label}: tokens 65–88 must be marked")
        return found


def check_language(checker, directory, manifest):
    """Check one language directory; returns its data for the comparison between languages."""
    data = {}
    for name in sorted(p.name for p in directory.glob("*.html")):
        path = directory / name
        raw = path.read_bytes()
        text = raw.decode("utf-8")
        label = f"{directory.name}/{name}"
        entry = manifest.get((directory.name, name))
        checker.expect(entry is not None and entry["sha256"] == hashlib.sha256(raw).hexdigest(), f"{label}: manifest SHA-256")
        printed = []
        if name == "seed-booklet.html":
            data[name] = checker.seed_booklet(label, text, printed)
        elif name.startswith("passphrase-booklet-"):
            data[name] = checker.passphrase_booklet(label, text, name[-7:-5], printed)
        elif name == "password-map.html":
            data[name] = checker.password_map(label, text, printed)
        elif name == "seed-reference.html":
            data[name] = checker.seed_reference(label, text, printed)
        elif name.startswith("token-"):
            data[name] = checker.tokens(label, text)
        if entry is not None:
            listed = [p["data_hash"] for p in entry["pages"] if "data_hash" in p]
            checker.expect(listed == printed, f"{label}: manifest page hashes differ from the printed ones")
        print(f"checked {label}")
    return data


def main():
    if len(sys.argv) != 2:
        raise SystemExit(__doc__)
    root = pathlib.Path(sys.argv[1])
    manifest_data = json.loads((root / "manifest.json").read_text(encoding="utf-8"))
    manifest = {(f["lang"], f["file"]): f for f in manifest_data["files"]}
    checker = Checker()
    results = {lang: check_language(checker, root / lang, manifest) for lang in ("en", "de")}
    checker.expect(set(results["en"]) == set(results["de"]) and len(results["en"]) == 7, "data files per language")
    for name in results["en"]:
        checker.expect(results["en"][name] == results["de"].get(name), f"{name}: data differs between en and de")
    for failure in checker.failures:
        print(f"FAIL  {failure}")
    print(f"\n{len(checker.failures)} failure(s)")
    return 1 if checker.failures else 0


if __name__ == "__main__":
    sys.exit(main())
