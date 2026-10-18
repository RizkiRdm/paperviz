// Agent API key card — create/replace the MCP agent key and reveal it once.
import { useState } from "react"
import { Button } from "@/components/ui/button"

export function AgentKeyCard({ keyConfigured, revealedKey, busy, onIssue, onError }) {
  const [copied, setCopied] = useState(false)

  // copyToClipboard surfaces failure inline — a blocked clipboard API is a real state.
  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(revealedKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error("Copy to clipboard failed", err)
      onError("Could not copy. Select the key and copy it manually.")
    }
  }

  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
      <h2 className="text-sm font-medium text-[#0a0a0a] mb-4">Agent API key</h2>
      {revealedKey ? (
        <>
          <div className="relative">
            <pre className="rounded-[6px] bg-[#f5f5f5] border border-[#e5e5e5] p-3 text-xs text-[#171717] font-mono overflow-x-auto whitespace-pre-wrap break-all">
              {revealedKey}
            </pre>
            <button
              onClick={copyToClipboard}
              aria-label={copied ? "Copied" : "Copy API key"}
              className="absolute top-2 right-2 rounded-[6px] border border-[#e5e5e5] bg-white px-2 py-1 text-xs font-medium text-[#737373] hover:bg-[#f5f5f5] transition-colors"
            >
              {copied ? "Copied!" : "Copy"}
            </button>
          </div>
          <p className="mt-3 text-xs text-[#737373]">
            Copy it now. It is shown once and cannot be retrieved later — if you
            lose it, replace it.
          </p>
        </>
      ) : (
        <>
          <p className="text-xs text-[#737373] mb-4">
            {keyConfigured
              ? "A key is configured. It is shown only when created."
              : "No agent key yet. Create one to connect an MCP client."}
          </p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onIssue(false)}
            disabled={busy}
          >
            {busy ? "Working..." : keyConfigured ? "Replace key" : "Create key"}
          </Button>
        </>
      )}
    </div>
  )
}
