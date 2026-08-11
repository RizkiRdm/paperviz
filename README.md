# PaperViz

Upload a PDF. Get a plain-language summary with re-visualized charts.

---

## Quick Start

```bash
# Prerequisites: Go 1.25+, Node.js 18+, npm 9+
# Get a Gemini API key: https://aistudio.google.com/apikey

# 1. Configure
cp .env.example .env
# Edit .env — set GEMINI_API_KEY=your-key

# 2. Install deps
go mod tidy
cd frontend && npm install && cd ..

# 3. Run (two terminals)
# Terminal 1 — backend:
export $(grep -v '^#' .env | xargs) && go run ./cmd/server
# Terminal 2 — frontend dev server:
cd frontend && npm run dev
```

Open `http://localhost:5173`.

## Build & Run (production)

```bash
make build    # builds frontend + Go binary
make run      # runs ./server with .env vars
```

Or manually:

```bash
cd frontend && npm run build && cd ..
CGO_ENABLED=0 go build -o server ./cmd/server
export $(grep -v '^#' .env | xargs) && ./server
```

Single binary serves API + built frontend on port 8080 (set via `PORT` in `.env`).

## Run with Podman

```bash
make container IMAGE=paperviz TAG=latest            # build image
make container-run IMAGE=paperviz GEMINI_API_KEY=... # run container
```

Volume `paperviz-data` persists the SQLite database across restarts.

## How It Works

1. Upload a PDF (text-layer only, no OCR) or paste text
2. Backend extracts text + chart images
3. Gemini simplifies content to chosen reading level (ELI5 / Simplified), then verifies claims against the original
4. Simplified text is split into chapters; charts are re-generated per chapter (or re-rendered from captured images)
5. Shareable link expires after 7 days of inactivity

Processing runs in the background — the result page polls until it completes, so a large paper (with rate-limit retries) can take a few minutes. The page shows a "still working" note past ~2 minutes instead of giving up.

## Tech Stack

- **Backend:** Go 1.25+, chi router, modernc.org/sqlite (pure Go, no CGO)
- **Frontend:** React 19, Vite 8, Tailwind CSS v4, Recharts
- **LLM:** Google Gemini API (direct HTTP client, no SDK) — single-slot serialization + exponential-backoff retries to stay inside free-tier rate limits
- **PDF:** pdfcpu + ledongthuc/pdf (in-memory, no disk writes)

## Deployment

PaperViz is a **single Go binary + SQLite** — no Docker required, but container-friendly.

### Netlify

**Not suitable** for the backend. Netlify hosts static sites and serverless functions, but PaperViz runs a long-lived Go server with persistent SQLite and background goroutines. The frontend (`frontend/dist`) *could* be deployed to Netlify, but the backend must run elsewhere.

### Recommended platforms

Any service that runs Docker containers or Go binaries:

| Platform | Notes |
|---|---|
| **Railway** / **Fly.io** | Container-native, easy DB volume setup |
| **Render** | Deploy from Git, supports Go natively |
| **Heroku** | Use Container Registry + podman push |
| **VPS** (DigitalOcean, Linux, etc.) | Run the binary or container directly |

All require `GEMINI_API_KEY` set as an environment variable.

## Documentation

| File | What it covers |
|---|---|
| `INSTALLATION_AND_SETUP.md` | Detailed setup, caveats, project structure |
| `ARCHITECTURE.md` | Design decisions, layers, data flow |
| `PRD.md` | Product requirements, acceptance scenarios |
| `DESIGN.md` | Design system, tokens, components |
| `PLAN.md` | Phase tracking, implementation checklist |
| `AGENTS.md` | Agent rules, known issues |

## License

MIT
