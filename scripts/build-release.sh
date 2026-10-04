#!/bin/sh
# Reproducible release build: the same commit and the Go version pinned in mise.toml give byte-identical binaries
# and printables archive. Writes dist/release/ with the binaries, the printables (HTML), an SPDX SBOM and SHA256SUMS.
# usage: scripts/build-release.sh <version, e.g. v1.0.0>
set -eu
export PYTHONDONTWRITEBYTECODE=1
cd "$(dirname "$0")/.."
version=${1:?usage: scripts/build-release.sh <version>}
case $version in v[0-9]*) ;; *) echo "version must look like v1.2.3" >&2; exit 2 ;; esac

pinned=$(sed -nE 's/^go = "([^"]+)"/go\1/p' mise.toml)
actual=$(go env GOVERSION)
if [ "$actual" != "$pinned" ]; then
  echo "Go $actual is not the pinned $pinned (mise.toml); the build would not be reproducible" >&2
  exit 1
fi
if [ -n "$(git status --porcelain)" ]; then
  echo "the working tree is not clean; commit or stash first" >&2
  exit 1
fi

out=dist/release
rm -rf "$out"
mkdir -p "$out"
SOURCE_DATE_EPOCH=$(git log -1 --format=%ct)
export SOURCE_DATE_EPOCH CGO_ENABLED=0

# GOOS GOARCH GOARM suffix
targets="linux amd64 - linux-amd64
linux arm64 - linux-arm64
linux arm 6 linux-armv6
windows amd64 - windows-amd64.exe
windows arm64 - windows-arm64.exe
darwin amd64 - macos-amd64
darwin arm64 - macos-arm64"

printf '%s\n' "$targets" | while read -r goos goarch goarm suffix; do
  if [ "$goarm" = "-" ]; then goarm=""; fi
  GOOS=$goos GOARCH=$goarch GOARM=$goarm go build -trimpath -buildvcs=true \
    -ldflags "-s -w -buildid= -X main.version=$version" \
    -o "$out/lottabits-$version-$suffix" ./cmd/lottabits
done

go run ./cmd/printables -version "$version" -out "$out/printables-$version"
python3 tools/release_archive.py "$out/printables-$version" "$out/lottabits-$version-printables.zip"
rm -rf "$out/printables-$version"
python3 tools/sbom.py "$version" > "$out/lottabits-$version.spdx.json"

(cd "$out" && sha256sum lottabits-* > SHA256SUMS)
cat "$out/SHA256SUMS"
