#!/usr/bin/env python3
"""Write an SPDX 2.3 software bill of materials for a LottaBits release to standard output.

LottaBits uses only the Go standard library, so the SBOM lists the program, the Go toolchain and the embedded third-party
data (word lists and fonts) with their licenses and hashes. The creation time comes from SOURCE_DATE_EPOCH, so the output
is reproducible.

    python3 tools/sbom.py v1.0.0 > lottabits-v1.0.0.spdx.json
"""

import hashlib
import json
import os
import pathlib
import re
import sys
import time

ROOT = pathlib.Path(__file__).resolve().parent.parent

# (SPDX id, name, version, license, supplier, download location, file in this repository)
EMBEDDED = [
    ("bip39-english", "BIP39 English word list", "ce1862ac6bcffa1dd20aad858380e51e66e949ea", "MIT",
     "Organization: Bitcoin BIPs authors", "https://github.com/bitcoin/bips/blob/ce1862ac6bcffa1dd20aad858380e51e66e949ea/bip-0039/english.txt",
     "wordlists/bip39-english.txt"),
    ("eff-long-wordlist", "EFF Long Wordlist", "2016-07-18", "CC-BY-4.0",
     "Organization: Electronic Frontier Foundation", "https://www.eff.org/files/2016/07/18/eff_large_wordlist.txt",
     "wordlists/upstream/eff_large_wordlist.txt"),
    ("dys2p-wordlists-de", "dys2p German word list de-7776-v1", "6ef31b9aefb8735a7b066592393d12843ec502cd", "CC0-1.0",
     "Organization: dys2p", "https://github.com/dys2p/wordlists-de/blob/6ef31b9aefb8735a7b066592393d12843ec502cd/de-7776-v1.txt",
     "wordlists/upstream/dys2p-de-7776-v1.txt"),
    ("atkinson-hyperlegible-mono", "Atkinson Hyperlegible Mono", "95f4904fc8bcf26d3420fe315560c96417c6dec7", "OFL-1.1",
     "Organization: The Atkinson Hyperlegible Mono Project Authors", "https://github.com/google/fonts/tree/95f4904fc8bcf26d3420fe315560c96417c6dec7/ofl/atkinsonhyperlegiblemono",
     "printables/fonts/AtkinsonHyperlegibleMono-wght.ttf"),
    ("atkinson-hyperlegible-next", "Atkinson Hyperlegible Next", "95f4904fc8bcf26d3420fe315560c96417c6dec7", "OFL-1.1",
     "Organization: The Atkinson Hyperlegible Next Project Authors", "https://github.com/google/fonts/tree/95f4904fc8bcf26d3420fe315560c96417c6dec7/ofl/atkinsonhyperlegiblenext",
     "printables/fonts/AtkinsonHyperlegibleNext-wght.ttf"),
]


def package(spdx_id, name, version, license_id, supplier, location, path=None):
    entry = {
        "SPDXID": f"SPDXRef-{spdx_id}", "name": name, "versionInfo": version, "supplier": supplier,
        "downloadLocation": location, "licenseConcluded": license_id, "licenseDeclared": license_id,
        "copyrightText": "NOASSERTION", "filesAnalyzed": False,
    }
    if path:
        entry["checksums"] = [{"algorithm": "SHA256", "checksumValue": hashlib.sha256((ROOT / path).read_bytes()).hexdigest()}]
    return entry


def main():
    if len(sys.argv) != 2:
        raise SystemExit(__doc__)
    version = sys.argv[1]
    go = re.search(r'^go = "([^"]+)"', (ROOT / "mise.toml").read_text(), re.MULTILINE).group(1)
    created = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(int(os.environ.get("SOURCE_DATE_EPOCH", "0"))))
    packages = [
        package("lottabits", "lottabits", version, "MIT", "Person: mkorun", "https://github.com/mkorun/lottabits"),
        package("go-stdlib", "Go standard library", go, "BSD-3-Clause", "Organization: The Go Authors", "https://go.dev/dl/"),
    ] + [package(*e) for e in EMBEDDED]
    relationships = [{"spdxElementId": "SPDXRef-DOCUMENT", "relationshipType": "DESCRIBES", "relatedSpdxElement": "SPDXRef-lottabits"}]
    relationships += [{"spdxElementId": "SPDXRef-lottabits", "relationshipType": "CONTAINS", "relatedSpdxElement": p["SPDXID"]}
                      for p in packages[1:]]
    document = {
        "spdxVersion": "SPDX-2.3", "dataLicense": "CC0-1.0", "SPDXID": "SPDXRef-DOCUMENT",
        "name": f"lottabits-{version}",
        "documentNamespace": f"https://github.com/mkorun/lottabits/releases/{version}/sbom",
        "creationInfo": {"created": created, "creators": ["Tool: lottabits tools/sbom.py"]},
        "packages": packages, "relationships": relationships,
    }
    json.dump(document, sys.stdout, indent=1, sort_keys=True)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
