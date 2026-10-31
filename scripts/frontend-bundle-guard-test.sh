#!/usr/bin/env bash
# Self-test for frontend-bundle-check.sh: proves the guard fails on a bundle
# built WITHOUT VITE_CLERK_PUBLISHABLE_KEY and passes on one built WITH a
# (dummy) key. This is the regression that shipped a keyless landing page —
# the check must never silently accept that bundle again.
# Uses a dummy key only; never a real credential.
set -euo pipefail

cd "$(dirname "$0")/.."
NOKEY_DIR="frontend/dist-guard-nokey"
DUMMY_KEY="pk_test_ZHVtbXktZmFrZS1rZXktMTIzNDU2Nzg5MGFi"
trap 'rm -rf "$NOKEY_DIR" frontend/dist-guard-dummy' EXIT

# 1) Keyless build must produce a bundle the guard REJECTS.
(cd frontend && env -u VITE_CLERK_PUBLISHABLE_KEY npx vite build --outDir dist-guard-nokey --emptyOutDir >/dev/null 2>&1)
if ./scripts/frontend-bundle-check.sh "$NOKEY_DIR" >/dev/null 2>&1; then
  echo "FAIL: guard accepted a bundle with no inlined publishable key" >&2
  exit 1
fi
echo "OK: guard rejects keyless bundle"

# 2) Dummy-key build must produce a bundle the guard ACCEPTS.
(cd frontend && VITE_CLERK_PUBLISHABLE_KEY="$DUMMY_KEY" npx vite build --outDir dist-guard-dummy --emptyOutDir >/dev/null 2>&1)
./scripts/frontend-bundle-check.sh frontend/dist-guard-dummy
echo "OK: guard accepts keyed bundle"
