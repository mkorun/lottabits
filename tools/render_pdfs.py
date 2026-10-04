#!/usr/bin/env python3
"""Render generated printables to PDF with headless Chrome or Chromium and check the page counts.

A PDF with more pages than manifest.json lists means that content overflowed a page, so the check fails.
Requires Chrome or Chromium (set CHROME to choose one) and `pdfinfo` from poppler-utils. Used for test prints and by the
release job; HTML stays the canonical form (PDFs carry creation dates and are not byte-reproducible).

    python3 tools/render_pdfs.py PRINTABLES_DIR PDF_DIR
"""

import json
import os
import pathlib
import re
import shutil
import subprocess
import sys


def find_chrome():
    for candidate in (os.environ.get("CHROME"), "google-chrome", "chromium", "chromium-browser"):
        if candidate and shutil.which(candidate):
            return candidate
    raise SystemExit("render_pdfs: no Chrome or Chromium found (set CHROME)")


def page_count(pdf):
    info = subprocess.run(["pdfinfo", str(pdf)], capture_output=True, text=True, check=True).stdout
    return int(re.search(r"^Pages:\s+(\d+)$", info, re.MULTILINE).group(1))


def main():
    if len(sys.argv) != 3:
        raise SystemExit(__doc__)
    source, target = pathlib.Path(sys.argv[1]).resolve(), pathlib.Path(sys.argv[2]).resolve()
    chrome = find_chrome()
    manifest = json.loads((source / "manifest.json").read_text(encoding="utf-8"))
    failures = 0
    for entry in manifest["files"]:
        html = source / entry["lang"] / entry["file"]
        pdf = target / entry["lang"] / entry["file"].replace(".html", ".pdf")
        pdf.parent.mkdir(parents=True, exist_ok=True)
        subprocess.run([chrome, "--headless=new", "--disable-gpu", "--no-sandbox", "--no-pdf-header-footer",
                        f"--print-to-pdf={pdf}", html.as_uri()], capture_output=True, check=True, timeout=300)
        pages, expected = page_count(pdf), len(entry["pages"])
        status = "ok" if pages == expected else "FAIL"
        failures += status == "FAIL"
        print(f"{status:4}  {entry['lang']}/{pdf.name}: {pages} page(s), expected {expected}")
    print(f"\n{failures} failure(s)")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
