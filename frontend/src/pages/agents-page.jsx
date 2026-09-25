import { useState } from "react"
import { Link } from "react-router-dom"

const CLIENTS = [
  {
    id: "claude-code",
    name: "Claude Code",
    config: (key) => `{
  "mcpServers": {
    "paperviz": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-remote", "https://paperviz.com/api/mcp?key=${key}"]
    }
  }
}`,
  },
  {
    id: "claude-desktop",
    name: "Claude Desktop",
    config: (key) => `{
  "mcpServers": {
    "paperviz": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-remote", "https://paperviz.com/api/mcp?key=${key}"]
    }
  }
}`,
  },
  {
    id: "cursor",
    name: "Cursor",
    config: (key) => `{
  "mcpServers": {
    "paperviz": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-remote", "https://paperviz.com/api/mcp?key=${key}"]
    }
  }
}`,
  },
  {
    id: "chatgpt",
    name: "ChatGPT",
    config: (key) => `{
  "mcpServers": {
    "paperviz": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-remote", "https://paperviz.com/api/mcp?key=${key}"]
    }
  }
}`,
  },
]

const TEST_PROMPT = "Analyze this paper and give me a simplified summary with key findings"

export function AgentsPage() {
  const [activeTab, setActiveTab] = useState("claude-code")
  const [copied, setCopied] = useState(false)

  const activeClient = CLIENTS.find((c) => c.id === activeTab)
  // The service key cannot be fetched: the server returns it exactly once, at
  // issue time, and stores only a digest afterwards. The snippet carries a
  // placeholder the user substitutes with the key they revealed in /account.
  const config = activeClient?.config("YOUR_PAPERVIZ_API_KEY")

  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(config)
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
            One config block. Paste it in your client's MCP settings and you're done.
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
            <div className="relative">
              <pre className="rounded-[6px] bg-[#f5f5f5] border border-[#e5e5e5] p-4 text-xs font-mono text-[#171717] overflow-x-auto whitespace-pre-wrap">
                {config}
              </pre>
              <button
                onClick={copyToClipboard}
                className="absolute top-2 right-2 rounded-[6px] border border-[#e5e5e5] bg-white px-3 py-1 text-xs font-medium text-[#737373] hover:bg-[#f5f5f5] transition-colors"
              >
                {copied ? "Copied!" : "Copy"}
              </button>
            </div>

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
