#!/usr/bin/env bash
# Guard against shipping a frontend bundle with no inlined Clerk publishable
# key. Vite inlines VITE_* at build time, so a bundle built without
# VITE_CLERK_PUBLISHABLE_KEY renders the App.jsx "key is not set" screen on
# every route — including the public landing page. This bit us twice, so the
# build now fails loud instead of shipping a silently broken SPA.
#
# Usage: frontend-bundle-check.sh [dist-dir]   (default: frontend/dist)
set -euo pipefail

DIST="${1:-frontend/dist}"
ASSETS="$DIST/assets"

if [ ! -d "$ASSETS" ]; then
  echo "ERROR: $ASSETS not found — run the frontend build before this check." >&2
  exit 1
fi

# Count real inlined keys (pk_live_/pk_test_ followed by a base64url instance
# identifier). The literal "pk_live_" prefix appears in the Clerk source inside
# every bundle, so require a non-empty value after it.
count="$(grep -rhoE 'pk_(live|test)_[A-Za-z0-9_-]{16,}' "$ASSETS" | wc -l | tr -d ' ')"

if [ "$count" -eq 0 ]; then
  echo "ERROR: no inlined VITE_CLERK_PUBLISHABLE_KEY in $ASSETS." >&2
  echo "The bundle was built without the key, so every route (including the" >&2
  echo "public landing page) would render 'VITE_CLERK_PUBLISHABLE_KEY is not set'." >&2
  echo "" >&2
  echo "Fix: set VITE_CLERK_PUBLISHABLE_KEY in the build environment, or use" >&2
  echo "'make frontend/dist' (local, exports $ENV_FILE) instead of a bare" >&2
  echo "'npm run build'. See .env.example." >&2
  exit 1
fi

echo "OK: $count inlined Clerk publishable key reference(s) in $ASSETS."
