import { useState } from "react"
import { Copy, Check } from "lucide-react"
import { AppShell } from "@/components/app-shell"

// PaperViz ships a stdio MCP server (cmd/mcp). It runs on the user's own
// machine, against their own SQLite file, and calls no model provider.
// A StreamableHTTP remote exists only when cmd/mcp is run with MCP_HTTP_ADDR
// set (self-hosted); there is no hosted MCP endpoint by default, so every
// client gets the same stdio command block; only the destination file differs.
const STDIO_CONFIG = `{
  "mcpServers": {
    "paperviz": {
      "command": "/absolute/path/to/paperviz-mcp",
      "env": {
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
    destination: "ChatGPT connects to remote MCP servers only. PaperViz has no hosted MCP endpoint by default (a StreamableHTTP remote exists only if you self-host cmd/mcp with MCP_HTTP_ADDR), so there is nothing to paste here.",
  },
]

const TEST_PROMPT = "Search my papers for anything about image classification accuracy"

export function AgentsPage() {
  const [activeTab, setActiveTab] = useState("claude-code")
  const [copied, setCopied] = useState(false)
  const [copyError, setCopyError] = useState(false)

  const activeClient = CLIENTS.find((c) => c.id === activeTab) ?? CLIENTS[0]

  async function copyToClipboard() {
    setCopyError(false)
    try {
      await navigator.clipboard.writeText(STDIO_CONFIG)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      // A blocked clipboard API is a real state — surface it inline (12.1).
      console.error("Copy to clipboard failed", err)
      setCopyError(true)
    }
  }

  return (
    <AppShell active="agents">
      <div className="mb-8">
        <h1 className="font-['Satoshi',sans-serif] text-2xl md:text-3xl font-medium tracking-tight text-[#0a0a0a]">
          Add PaperViz to your agent
        </h1>
        <p className="mt-2 text-sm text-[#525252] leading-relaxed max-w-2xl">
          PaperViz ships a local stdio MCP server with five deterministic tools (`ingest_document`, `search_documents`, `get_document`, `get_evidence`, `get_figures`). Setup runs directly on your machine against your local research database.
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
                  <pre className="rounded-[6px] bg-[#f5f5f5] border border-[#e5e5e5] p-4 text-xs font-mono text-[#171717] overflow-x-auto whitespace-pre-wrap pr-24">
                    {STDIO_CONFIG}
                  </pre>
                  <button
                    onClick={copyToClipboard}
                    aria-label={copied ? "Copied configuration" : "Copy configuration"}
                    className="absolute top-2.5 right-2.5 inline-flex items-center gap-1.5 rounded-[6px] border border-[#e5e5e5] bg-white px-2.5 py-1 text-xs font-medium text-[#171717] hover:bg-[#f5f5f5] transition-colors shadow-2xs"
                  >
                    {copied ? (
                      <>
                        <Check className="h-3.5 w-3.5 text-[#16a34a]" />
                        <span>Copied</span>
                      </>
                    ) : (
                      <>
                        <Copy className="h-3.5 w-3.5 text-[#737373]" />
                        <span>Copy</span>
                      </>
                    )}
                  </button>
                </div>
                {copyError && (
                  <p className="mt-2 text-[11px] text-[#ea580c]">
                    Couldn&apos;t copy automatically. Select the block and copy it manually.
                  </p>
                )}
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
          Need help? Check the documentation.
        </p>
    </AppShell>
  )
}
