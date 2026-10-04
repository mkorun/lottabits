#!/usr/bin/env python3
"""Pack a directory into a reproducible ZIP archive: sorted entries, fixed timestamps (SOURCE_DATE_EPOCH, at least
1980-01-01), fixed permissions and no compression (deflate output differs between zlib versions, so a compressed archive
would not be byte-identical on every system).

    python3 tools/release_archive.py SOURCE_DIR ARCHIVE.zip
"""

import os
import pathlib
import sys
import time
import zipfile

ZIP_EPOCH = 315532800  # 1980-01-01, the earliest time a ZIP entry can carry


def main():
    if len(sys.argv) != 3:
        raise SystemExit(__doc__)
    source, archive = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
    epoch = max(int(os.environ.get("SOURCE_DATE_EPOCH", ZIP_EPOCH)), ZIP_EPOCH)
    stamp = time.gmtime(epoch)[:6]
    root = source.name
    with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_STORED) as zf:
        for path in sorted(p for p in source.rglob("*") if p.is_file()):
            info = zipfile.ZipInfo(f"{root}/{path.relative_to(source).as_posix()}", date_time=stamp)
            info.compress_type = zipfile.ZIP_STORED
            info.external_attr = 0o644 << 16
            info.create_system = 3  # Unix, so the permissions above apply everywhere
            zf.writestr(info, path.read_bytes())


if __name__ == "__main__":
    main()
