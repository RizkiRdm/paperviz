<!-- Context: project-intelligence/navigation | Priority: high | Version: 1.1 | Updated: 2026-09-26 -->

# Project Intelligence — Navigation

**Purpose**: 30s map of project intelligence; agents pick the right context file in <10s.

## Quick Routes
| Need | File | Priority |
|------|------|----------|
| Tech stack, layer boundaries, code patterns, security contract | `technical-domain.md` | critical |
| This map | `navigation.md` | high |

## Deep Dives
| Area | Where to Look | When |
|------|---------------|------|
| Layer boundaries & internal contracts | `docs/ARCHITECTURE.md` + `AGENTS.md` | Before handler/service/repo changes |
| Design tokens | `DESIGN.md` | Before any UI work (read completely) |
| Pipeline stages & evidence trust model | `docs/DATA_PIPELINE.md` | Before pipeline/evidence/chart work |
| MCP ↔ REST parity & tool scope | `docs/mcp-parity.md` | Before MCP tools |
| Routes, files, config, commands | `docs/codebase-reference.md` | Before adding an endpoint or file |
| Current focus & known issues | `docs/PROJECT_STATE.md` + `AGENTS.md` | Session start |

## Hard Rules (do not violate)
- Layers: `handlers → internal/app → services/external → repository`. Handlers never join repos; read models live in `internal/app/documents/readmodel.go`.
- No ORM, job queue, message broker, or microservice split. No PDF bytes on disk.
- No server-side model credential. BYOK per request; `missing_credential` is correct.
- Never log/return plaintext model keys. `CREDENTIAL_ENCRYPTION_KEY` is the only required server secret.
- DB schema change → delete `paperviz.db*` and let the 21 migrations rebuild.

## Status
- `technical-domain.md` v2.0 (2026-09-26) — Go 1.25/chi/SQLite/React 19, goai BYOK, AES-256-GCM credentials, 21 migrations. Major bump from v1.0: OAuth/Stripe env rules removed (they no longer exist).
- Next: add `business-domain.md` via `/add-context --business` if product/GTM patterns are needed.

## 📂 Codebase References
**Intelligence**: `.opencode/context/project-intelligence/technical-domain.md` — single source for stack, patterns, security contract
**Index**: `graphify-out/` — run `graphify query "<question>"` before grep; `codegraph_explore` for symbol source + call paths
**Rules**: `AGENTS.md` (engineering rules), `DESIGN.md` (UI tokens)

## Related Files
- `technical-domain.md` — Primary stack, layer boundaries, patterns, security
- `CONTEXT_SYSTEM_GUIDE.md` — Context system (if present)
