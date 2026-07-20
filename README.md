# PaperViz

Upload a PDF. Get a plain-language summary with re-visualized charts.

---

## Quick Start

```bash
# 1. Prerequisites: Go 1.22+, Node.js 18+, npm 9+
# 2. Get a Gemini API key: https://aistudio.google.com/apikey

# 3. Configure
cp .env.example .env
# Edit .env — set GEMINI_API_KEY=your-key

# 4. Install deps
go mod tidy
cd frontend && npm install && cd ..

# 5. Run (two terminals)
# Terminal 1 — backend:
export $(grep -v '^#' .env | xargs) && go run ./cmd/server

# Terminal 2 — frontend dev server:
cd frontend && npm run dev
```

Open `http://localhost:5173` (Vite default, proxies API to backend).

## Production Build

```bash
cd frontend && npm run build && cd ..
go build -o server ./cmd/server
export $(grep -v '^#' .env | xargs) && ./server
```

Single binary serves API + built frontend on `http://localhost:8080` (override via `PORT` in `.env`).

## How It Works

1. Upload a PDF (text-layer only, no OCR)
2. Backend extracts text + chart images
3. Gemini simplifies content to chosen reading level (ELI5 / Simplified / Original)
4. Charts are re-rendered server-side; plain-language annotation overlaid
5. Shareable link expires after 7 days of inactivity

## Tech Stack

- **Backend:** Go 1.22+, chi router, modernc.org/sqlite (pure Go, no CGO)
- **Frontend:** React 19, Vite 8, Tailwind CSS v4, Recharts
- **LLM:** Google Gemini API (direct HTTP client, no SDK)
- **PDF:** pdfcpu + ledongthuc/pdf (in-memory, no disk writes)

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
