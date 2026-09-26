<!-- Context: project-intelligence/technical | Priority: critical | Version: 2.0 | Updated: 2026-09-26 -->

# Technical Domain — PaperViz

**Purpose**: Go 1.25 + chi + SQLite + React 19, BYOK-agent-first paper pipeline. Agents follow these patterns to generate code matching this repo.
**Last Updated**: 2026-09-26

## Quick Reference
**Update Triggers**: Stack change | New handler/service/app pattern | Naming shift | Security contract change
**Audience**: Developers, AI agents

## Layer Boundaries (non-negotiable)
```text
Web UI → REST handlers → internal/app (use cases + read models) → services/external → repository → SQLite
MCP client → internal/mcp/tools → internal/app + repository   (same service layer, never imports handlers)
```
- Handlers: transport only — parse, validate, call app layer, serialize. No SQL joins, no repo fan-out.
- Read-model aggregation lives in `internal/app/documents/readmodel.go`, not in a handler.
- No ORM, no job queue, no message broker, no microservice split.

## Primary Stack
| Layer | Technology | Version | Rationale |
|-------|------------|---------|-----------|
| Backend | Go + chi | 1.25.0 / v5.1.0 | Single binary serves API + static assets, raw `database/sql` |
| DB | modernc.org/sqlite | v1.34.4 | Pure Go (CGO-free); WAL + synchronous=NORMAL |
| Migrations | ordered SQL files | 21 (`001`–`021`) | Flat single-file runner; delete DB on schema change |
| Frontend | React + Vite + Tailwind + oxlint | 19.2.7 / 8.1.1 / 4.3.3 / 1.71.0 | SPA built to `frontend/dist`, served by Go |
| Charts | Recharts | 3.9.2 | Missing values excluded, never zero-filled |
| LLM | `zendev-sh/goai` + own `Transport` | v0.10.4 | Vendor protocol for gemini/anthropic/openai; retry, 90s budget, one global slot are OURS |
| Crypto | `golang.org/x/crypto` | v0.54.0 | AES-256-GCM credential encryption at rest |
| PDF | pdfcpu + ledongthuc/pdf | v0.9.1 / pinned | In-memory text+image, never writes PDF bytes to disk |
| Agent | MCP stdio (`cmd/mcp`) | go-sdk v1.7.0 | 5 stateless read/intake tools, shares service layer with REST |

## Code Patterns

### API Endpoint (chi handler → app service → repo)
```go
// Create handles POST /api/documents via Service.Create.
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
  r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+1<<20)
  if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
    writeError(w, http.StatusBadRequest, "file_too_large"); return
  }
  // ... validate reading_level, exactly one of file|text, isPDFContent ...
  svc := documents.New(h.db, h.provider) // provider = *credentials.Resolver
  docID, code, err := svc.Create(r.Context(), readingLevel, hasFile, pdfBytes, pastedText, userID)
  if err != nil {
    switch code {
    case "missing_credential", "credential_unreadable", "internal_error":
      writeCredentialError(w, err); return
    }
    slog.Error("document intake failed", "error", err)
    writeError(w, http.StatusBadRequest, code); return
  }
  writeJSON(w, http.StatusCreated, createDocumentResponse{DocumentID: docID, Status: repository.StatusProcessing})
}
```

Handler rules:
- Errors are snake_case codes, never raw Go errors: `writeError(w, status, "invalid_file_type")`.
- `readJSON(r, &req)` caps body at 1MB and drains it; `writeJSON`/`writeError` are the only serializers.
- Wrap document creation in `requireIngestionEnabled` — kill switch returns 503 without a deploy.
- Every model-backed path resolves the key per request: `h.provider.For(ctx, userID)`. There is no process-wide client and no fallback key.

### Component (React functional, DESIGN.md tokens only)
```jsx
// Functional, hooks, Tailwind tokens from DESIGN.md, no invented colors/spacing
export function ResultHeader({ doc, visibility, onVisibilityChange }) {
  return (
    <header className="border-b border-[#e5e5e5] bg-white/80 sticky top-0">
      <select value={visibility} onChange={(e) => onVisibilityChange(e.target.value)}
        className="h-9 rounded-[8px] border border-[#e5e5e5] bg-white text-xs">
        <option value="private">Private</option>
      </select>
    </header>
  )
}
```

## Naming Conventions
| Type | Convention | Example |
|------|------------|---------|
| Files (Go) | snake_case, flat per-package | `documents_create.go`, `readmodel.go` |
| Files (FE) | kebab-case | `result-header.jsx`, `upload-page.jsx` |
| Components/Types | PascalCase | `ResultHeader`, `CredentialHandler` |
| Funcs/Vars | camelCase, Go idiomatic | `UserIDFromContext`, `createDocument` |
| DB | snake_case, ordered numeric prefix | `user_credentials`, `021_drop_oauth_billing.sql` |
| Error codes | snake_case strings | `missing_credential`, `invalid_file_type` |

## Code Standards
- One-line `// comment` per function. `gofmt` clean. No panics for expected errors.
- Services pure with explicit args `(text string, level string) (string, error)`; pass config, never read globals.
- `log/slog` only. Never log document text, claims, or key material.
- App layer owns use cases + read-model aggregation; handlers never join repositories.
- Frontend: every `catch` MUST show inline error + Retry + `console.error` + preserve inputs (`research-map.jsx` is canonical). No silent catches.
- Ceiling comments only: `// ponytail: <simplification> — ceiling: <limit> ; upgrade: <path>`; never logic inside the marking.
- Trust model: deterministic code extracts evidence/datasets; the model may transform but never manufacture values; grounding validator rejects unsupported data.
- Gate before done: `go build ./cmd/server ./cmd/mcp ./cmd/admin`, `go test ./...`, `go vet ./...`, `gofmt -l .`, `npm --prefix frontend run lint && npm --prefix frontend run build`.

## Security Requirements
- **The server holds no model vendor credential.** Every model call runs on a key the user supplied. No process-wide client, no fallback — `missing_credential` is the correct answer, not a bug to paper over.
- `CREDENTIAL_ENCRYPTION_KEY` (base64, env-only, never persisted) is the ONE required server secret. Fails loud at startup; a truncated/short key is rejected rather than silently downgraded to AES-128.
- Model keys encrypted at rest with AES-256-GCM. `external.CredentialAAD(userID, provider, model)` on **both** encrypt and decrypt — a mismatch makes credentials permanently unreadable.
- Never log, return, or store a plaintext model key. Credential API returns `key_hint` (last 4) only.
- Service API keys: `users.api_key_hash` is a SHA-256 hex digest. Plaintext returned exactly once at issue. A lost key is replaced, never recovered.
- No self-service password reset by design (no email provider account). `go run ./cmd/admin reset-password <email>` also revokes live sessions.
- `INGESTION_ENABLED=false` → 503 on document creation; read/share/library paths keep working.
- `PIPELINE_MAX_CONCURRENCY` (default 2) bounds concurrent pipelines. Slot taken inside `RunPipelineAndPersist`, not at the request boundary — a handler-side cap would bound submissions, not CPU.
- Ownership: annotations/collections enforce `userID` → 403 on mismatch. Export excludes `OriginalText`/`SimplifiedText`.
- Input: `maxUploadBytes` (20 MiB) + `isPDFContent` + `readJSON` 1MB cap. SSRF: URL import https-only + private-IP block. Cookies `HttpOnly+Secure+SameSite=Lax`.
- Rate limit: `POST /api/documents` 1/30s burst 2; `POST /api/auth/*` 5/60s burst 3.

## 📂 Codebase References
**Wiring**: `internal/handlers/router.go:57` `NewRouter(db, provider *credentials.Resolver, staticDir)`; `internal/handlers/respond.go:16,27,33` `writeJSON`/`writeError`/`readJSON`; `internal/handlers/ingest.go:17` `requireIngestionEnabled`; `internal/handlers/documents_create.go:25` `Create` (canonical handler).
**App layer**: `internal/app/documents/service.go:22` `New(db, provider)`, `:31` `Create` → `(docID, code, err)`; `:88-131` read-model getters; `internal/app/documents/readmodel.go`; `internal/app/credentials/resolver.go:41` `NewResolver`, `:54` `For(ctx, userID)`.
**LLM + crypto**: `internal/external/llm.go:88` `NewTransport` (retry/90s/slot); `internal/external/llm_goai.go` provider switch; `internal/external/crypto.go` `Cipher`, `CredentialAAD`, `KeyHint`.
**Domain + limits**: `internal/services/` pipeline, `chapters.go`, `grounding.go`; `internal/services/concurrency.go:19` default cap 2, `:81` `IngestionEnabled`; `internal/apperr/errors.go`; `internal/infra/`.
**Data**: `internal/repository/db.go:16` `Open` (WAL); `internal/repository/migrations.go:42` migration runner; `migrations/001…021`.
**MCP**: `internal/mcp/tools.go` — 5 tools (`ingest_document`, `search_documents`, `get_document`, `get_figures`, `get_evidence`); parity in `docs/mcp-parity.md`.
**Entrypoints**: `cmd/server/main.go:83` `CREDENTIAL_ENCRYPTION_KEY` fail-loud; `cmd/mcp/main.go:21` `GEMINI_API_KEY` (MCP only); `cmd/admin/` recovery CLI.
**Frontend**: `frontend/src/App.jsx` routes; `frontend/src/pages/upload-page.jsx` ingest tabs; `frontend/src/components/result/` 6-component split; `frontend/src/components/research-map.jsx` canonical catch.
**Config**: `go.mod` (pinned), `frontend/package.json`, `Makefile`, `.github/workflows/*.yml`.

## Related Files
- `navigation.md` — Quick Routes & Deep Dives
- `AGENTS.md` — repo rules, security rounds, known issues
- `DESIGN.md` — design tokens (absolute source for UI)
- `docs/ARCHITECTURE.md` — layer boundaries, contracts, logging policy
- `docs/DATA_PIPELINE.md` — pipeline stages
- `docs/mcp-parity.md` — MCP ↔ REST parity
- `docs/PROJECT_STATE.md` — current engineering state
