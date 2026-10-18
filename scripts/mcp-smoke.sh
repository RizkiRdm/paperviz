#!/usr/bin/env bash
# Smoke all 5 MCP tools over real stdio: ingest → search → get_document →
# get_evidence → get_figures, plus a not-found check.
#
# Complements (does not replace) the Go contract tests: those call handlers
# in-process; this proves the full wire path — registration, schema
# validation, dispatch, JSON-RPC framing — against a freshly built binary
# with no LLM credential, so a hidden provider dependency on the
# ingest/search/get paths fails here even when the suite is green.
set -uo pipefail

WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT

export DATABASE_PATH="$WORKDIR/mcp-smoke.db"
export MIGRATIONS_DIR=migrations
# Deliberately fake. No tool on the ingest/search/get path may reach a
# provider; if one did, this smoke fails on a bad key instead of passing.
export GEMINI_API_KEY=smoke-not-a-real-key
export PAPERVIZ_API_KEY=smoke-service-key

BIN="$WORKDIR/mcp"
go build -o "$BIN" ./cmd/mcp || {
  echo "FAIL: could not build cmd/mcp" >&2
  exit 1
}

# Drive the session: initialize, initialized, then the call sequence.
# Sleeps are generous because the server processes stdin sequentially.
OUTPUT=$(
  {
    printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"ci-smoke","version":"1"}}}\n'
    sleep 2
    printf '{"jsonrpc":"2.0","method":"notifications/initialized"}\n'
    sleep 1
    printf '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"ingest_document","arguments":{"text":"Smoke test paper: sleep improves memory consolidation. 42% of studies confirm this."}}}\n'
    sleep 4
    printf '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_documents","arguments":{"query":"memory","limit":5}}}\n'
    sleep 2
    printf '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_document","arguments":{"document_id":"nonexistent-id-000"}}}\n'
    sleep 2
    printf '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_evidence","arguments":{"document_id":"nonexistent-id-000"}}}\n'
    sleep 1
    printf '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_figures","arguments":{"document_id":"nonexistent-id-000"}}}\n'
    sleep 2
  } | timeout 60 "$BIN" 2>"$WORKDIR/stderr.log"
)

if [ -z "$OUTPUT" ]; then
  echo "FAIL: cmd/mcp produced no response; it likely panicked or exited." >&2
  cat "$WORKDIR/stderr.log" >&2
  exit 1
fi

# Parse each reply by id and assert its shape. Exit codes are cumulative:
# report every failed check, not just the first.
python3 - "$OUTPUT" <<'PY'
import json, sys

msgs = {}
for line in sys.argv[1].splitlines():
    line = line.strip()
    if not line:
        continue
    try:
        m = json.loads(line)
    except json.JSONDecodeError:
        continue
    if "id" in m:
        msgs[m["id"]] = m

failures = []

def reply(i, label):
    m = msgs.get(i)
    if m is None:
        failures.append(f"{label}: no reply for id {i}")
        return None
    if "error" in m:
        failures.append(f"{label}: JSON-RPC error {m['error']}")
        return None
    r = m.get("result", {})
    if r.get("isError"):
        failures.append(f"{label}: tool returned isError: {r.get('content')}")
        return None
    return r

def payload(r, label):
    # MCP tool results wrap JSON in a text content block.
    try:
        block = r["content"][0]["text"]
        return json.loads(block)
    except Exception as e:
        failures.append(f"{label}: cannot parse content text ({e}): {r}")
        return None

# id 2: ingest returns a document_id (proves ingest is LLM-free on a fake key).
r = reply(2, "ingest_document")
if r is not None:
    p = payload(r, "ingest_document")
    if p is not None:
        if not p.get("document_id"):
            failures.append(f"ingest_document: missing document_id: {p}")
        if p.get("source_type") != "pasted_text":
            failures.append(f"ingest_document: source_type={p.get('source_type')!r}, want pasted_text")

# id 3: search finds the ingested doc deterministically.
r = reply(3, "search_documents")
if r is not None:
    p = payload(r, "search_documents")
    if p is not None:
        docs = p.get("documents", [])
        if not isinstance(docs, list):
            failures.append(f"search_documents: 'documents' not a list: {p}")
        elif len(docs) < 1:
            failures.append(f"search_documents: ingested doc not found for query 'memory'")

# ids 4-6: not-found must be a structured error, not a crash or empty success.
for i, name in ((4, "get_document"), (5, "get_evidence"), (6, "get_figures")):
    m = msgs.get(i)
    if m is None:
        failures.append(f"{name}: no reply for id {i}")
    elif "error" not in m and not m.get("result", {}).get("isError"):
        failures.append(f"{name}: nonexistent id did not error: {m.get('result')}")

if failures:
    print("FAIL: MCP smoke found problems:")
    for f in failures:
        print(f"  - {f}")
    sys.exit(1)

print("MCP smoke OK: 5 tools exercised over stdio (ingest, search, get_document, get_evidence, get_figures).")
PY
RC=$?
if [ $RC -ne 0 ]; then
  echo "--- stderr ---" >&2
  cat "$WORKDIR/stderr.log" >&2
  exit $RC
fi
