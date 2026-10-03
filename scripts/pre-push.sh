#!/bin/sh
# Local pre-push gate: the same checks as CI, including the secret scan, so a red CI never starts from a push.
# Install once: ln -sf ../../scripts/pre-push.sh .git/hooks/pre-push
set -eu
cd "$(dirname "$0")/.."
mise exec -- sh scripts/check.sh
mise exec -- gitleaks git --redact --no-banner --config .gitleaks.toml .
