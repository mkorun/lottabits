#!/usr/bin/env python3
"""Check that no box of a printable overflows (clipped text, a header wrapping into the content, a table wider than its page).

The printables must be generated with `go run ./cmd/printables -layout-check -out DIR`: that build contains a small script
which measures every page, header, body, footer and cell after the fonts have loaded. This tool opens every file in
headless Chrome or Chromium (set CHROME to choose one), reads the result from the DOM and fails on any overflow.

    python3 tools/check_layout.py DIR
"""

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
    raise SystemExit("check_layout: no Chrome or Chromium found (set CHROME)")


def main():
    if len(sys.argv) != 2:
        raise SystemExit(__doc__)
    chrome = find_chrome()
    failures = 0
    for html in sorted(pathlib.Path(sys.argv[1]).resolve().glob("*/*.html")):
        dom = subprocess.run([chrome, "--headless=new", "--disable-gpu", "--no-sandbox", "--virtual-time-budget=20000",
                              "--window-size=1200,1600", "--dump-dom", html.as_uri()],
                             capture_output=True, text=True, check=True, timeout=300).stdout
        found = re.search(r'<body[^>]*data-layout="([^"]*)"', dom)
        result = found.group(1) if found else "no result (was the file generated with -layout-check?)"
        status = "ok" if result == "ok" else "FAIL"
        failures += status == "FAIL"
        print(f"{status:4}  {html.parent.name}/{html.name}" + ("" if status == "ok" else f": {result}"))
    print(f"\n{failures} failure(s)")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
