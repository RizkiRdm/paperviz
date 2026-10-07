#!/usr/bin/env bash
# Verify the MCP server actually boots and serves its tools over stdio.
#
# This check exists because the server once shipped unable to start: a bad
# jsonschema struct tag made AddTool panic before a single tool was served,
# while the whole test suite stayed green because no test ran the
# registration path. A grep over tool names cannot catch a panic, so this
# performs the real handshake.
set -uo pipefail

EXPECTED_TOOLS='["get_document","get_evidence","get_figures","ingest_document","search_documents"]'
WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT

export DATABASE_PATH="$WORKDIR/mcp-handshake.db"
export MIGRATIONS_DIR=migrations
# Placeholders. No tool in the registration path calls a provider, and no call
# is made here, so nothing reaches the network.
export GEMINI_API_KEY=handshake-not-a-real-key
export PAPERVIZ_API_KEY=handshake-service-key

BIN="$WORKDIR/mcp"
go build -o "$BIN" ./cmd/mcp || {
  echo "FAIL: could not build cmd/mcp" >&2
  exit 1
}

# Drive initialize, then the initialized notification, then tools/list. The
# sleeps are generous because the server processes stdin sequentially.
OUTPUT=$(
  {
    printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"ci-handshake","version":"1"}}}\n'
    sleep 2
    printf '{"jsonrpc":"2.0","method":"notifications/initialized"}\n'
    sleep 2
    printf '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}\n'
    sleep 3
  } | timeout 30 "$BIN" 2>"$WORKDIR/stderr.log"
)

if [ -z "$OUTPUT" ]; then
  echo "FAIL: cmd/mcp produced no response; it likely panicked or exited." >&2
  echo "--- stderr ---" >&2
  cat "$WORKDIR/stderr.log" >&2
  exit 1
fi

if ! printf '%s' "$OUTPUT" | grep -q '"serverInfo"'; then
  echo "FAIL: no initialize result in MCP response." >&2
  echo "$OUTPUT" >&2
  exit 1
fi

if ! printf '%s' "$OUTPUT" | grep -q '"tools"'; then
  echo "FAIL: tools/list returned no tools array." >&2
  echo "$OUTPUT" >&2
  cat "$WORKDIR/stderr.log" >&2
  exit 1
fi

ACTUAL=$(printf '%s' "$OUTPUT" | python3 -c '
import json, sys

# The server may emit several JSON-RPC lines; find the tools/list reply.
for line in sys.stdin.read().splitlines():
    line = line.strip()
    if not line:
        continue
    try:
        msg = json.loads(line)
    except json.JSONDecodeError:
        continue
    if msg.get("id") == 2 and "result" in msg:
        # Sorted so tool registration order is not part of the contract.
        print(json.dumps(sorted(t["name"] for t in msg["result"]["tools"]), separators=(",", ":")))
        break
')

if [ -z "$ACTUAL" ]; then
  echo "FAIL: could not parse the tools/list response." >&2
  echo "$OUTPUT" >&2
  exit 1
fi

if [ "$ACTUAL" != "$EXPECTED_TOOLS" ]; then
  echo "FAIL: MCP tool set changed." >&2
  echo "  expected: $EXPECTED_TOOLS" >&2
  echo "  actual:   $ACTUAL" >&2
  echo "If this is intentional, update EXPECTED_TOOLS in this script and" >&2
  echo "docs/mcp-parity.md in the same change." >&2
  exit 1
fi

echo "MCP handshake OK: initialize succeeded and 5 tools advertised."