import { useState } from "react"
import { Link } from "react-router-dom"

// PaperViz ships a stdio MCP server (cmd/mcp). It runs on the user's own
// machine, against their own SQLite file, calling the provider with their own
// key. There is no hosted MCP endpoint, so every client gets the same stdio
// command block; only the destination file differs.
const STDIO_CONFIG = `{
  "mcpServers": {
    "paperviz": {
      "command": "/absolute/path/to/paperviz-mcp",
      "env": {
        "GEMINI_API_KEY": "your-gemini-key",
        "GEMINI_MODEL": "gemini-3.1-flash-lite",
        "DATABASE_PATH": "/absolute/path/to/paperviz.db",
        "MIGRATIONS_DIR": "/absolute/path/to/migrations",
        "PAPERVIZ_API_KEY": "a-local-key-of-your-choice"
      }
    }
  }
}`

const CLIENTS = [
  {
    id: "claude-code",
    name: "Claude Code",
    supported: true,
    destination: 'Run `claude mcp add paperviz -- /absolute/path/to/paperviz-mcp`, or paste the block into `.mcp.json` in your project.',
  },
  {
    id: "claude-desktop",
    name: "Claude Desktop",
    supported: true,
    destination: "Paste the block into claude_desktop_config.json (Settings → Developer → Edit Config).",
  },
  {
    id: "cursor",
    name: "Cursor",
    supported: true,
    destination: "Paste the block into ~/.cursor/mcp.json.",
  },
  {
    id: "chatgpt",
    name: "ChatGPT",
    supported: false,
    destination: "ChatGPT connects to remote MCP servers only. PaperViz does not expose a hosted MCP endpoint, so there is nothing to paste here.",
  },
]

const TEST_PROMPT = "Search my papers for anything about image classification accuracy"

export function AgentsPage() {
  const [activeTab, setActiveTab] = useState("claude-code")
  const [copied, setCopied] = useState(false)

  const activeClient = CLIENTS.find((c) => c.id === activeTab) ?? CLIENTS[0]

  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(STDIO_CONFIG)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) { console.error("Copy to clipboard failed", err) }
  }

  return (
    <div className="min-h-screen bg-white bg-dotted-grid">
      <div className="mx-auto max-w-3xl px-6 py-12">
        <div className="mb-8">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold tracking-tight text-[#0a0a0a]">PaperViz</span>
          </Link>
          <h1 className="font-satoshi text-3xl font-medium text-[#0a0a0a]">Add PaperViz to your agent</h1>
          <p className="mt-2 text-sm text-[#737373]">
            PaperViz ships an MCP server that runs on your machine. It reads your
            local PaperViz database and calls the model provider with your own key.
          </p>
        </div>

        <div className="rounded-[12px] border border-[#e5e5e5] bg-white">
          <div className="flex border-b border-[#e5e5e5]" role="tablist" aria-label="MCP client">
            {CLIENTS.map((client) => (
              <button
                key={client.id}
                role="tab"
                id={`tab-${client.id}`}
                aria-selected={activeTab === client.id}
                aria-controls={`panel-${client.id}`}
                onClick={() => setActiveTab(client.id)}
                className={`flex-1 px-4 py-3 text-sm font-medium transition-colors ${activeTab === client.id
                    ? "border-b-2 border-[#0a0a0a] text-[#0a0a0a]"
                    : "text-[#737373] hover:text-[#0a0a0a]"
                  }`}
              >
                {client.name}
              </button>
            ))}
          </div>

          <div className="p-6" role="tabpanel" id={`panel-${activeTab}`} aria-labelledby={`tab-${activeTab}`}>
            {activeClient.supported ? (
              <>
                <p className="mb-3 text-xs text-[#737373]">{activeClient.destination}</p>
                <div className="relative">
                  <pre className="rounded-[6px] bg-[#f5f5f5] border border-[#e5e5e5] p-4 text-xs font-mono text-[#171717] overflow-x-auto whitespace-pre-wrap">
                    {STDIO_CONFIG}
                  </pre>
                  <button
                    onClick={copyToClipboard}
                    className="absolute top-2 right-2 rounded-[6px] border border-[#e5e5e5] bg-white px-3 py-1 text-xs font-medium text-[#737373] hover:bg-[#f5f5f5] transition-colors"
              >
                {copied ? "Copied!" : "Copy"}
                  </button>
                </div>
                <p className="mt-3 text-xs text-[#737373]">
                  Build the binary first with <code className="font-mono">go build -o paperviz-mcp ./cmd/mcp</code>.
                </p>
              </>
            ) : (
              <p className="rounded-[6px] border border-[#e5e5e5] bg-[#fafafa] p-4 text-sm text-[#737373]">
                {activeClient.destination}
              </p>
            )}

            <div className="mt-6 rounded-[6px] border border-[#e5e5e5] bg-[#fafafa] p-4">
              <p className="text-xs font-medium text-[#737373] mb-2">Try it:</p>
              <code className="text-xs font-mono text-[#171717]">{TEST_PROMPT}</code>
            </div>
          </div>
        </div>

        <p className="mt-6 text-xs text-[#737373] text-center">
          Need help? Check the{" "}
          <a href="https://github.com/rizki/paperviz#readme" className="text-[#2563eb] hover:underline">
            documentation
          </a>
        </p>
      </div>
    </div>
  )
}
