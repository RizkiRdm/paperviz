// Model API key card — stored-credential list plus the add-key form.
import { useState } from "react"
import { Button } from "@/components/ui/button"

const PROVIDERS = [
  { value: "gemini", label: "Google Gemini" },
  { value: "anthropic", label: "Anthropic Claude" },
  { value: "openai", label: "OpenAI" },
]

export function CredentialsCard({ credentials, savingCred, onAdd, onMakeDefault, onRemove }) {
  const [provider, setProvider] = useState("gemini")
  const [model, setModel] = useState("")
  const [apiKeyInput, setApiKeyInput] = useState("")

  // handleSubmit validates at the boundary, then clears the key input only on success.
  async function handleSubmit(e) {
    e.preventDefault()
    const saved = await onAdd({ provider, model, apiKey: apiKeyInput })
    if (saved) setApiKeyInput("")
  }

  const hasCredential = credentials.length > 0

  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
      <h2 className="text-sm font-medium text-[#0a0a0a] mb-1">Model API key</h2>
      <p className="text-xs text-[#737373] mb-4">
        PaperViz runs on your own key, so nothing is billed to us. Add one to
        analyse papers.
      </p>

      {hasCredential ? (
        <ul className="space-y-3">
          {credentials.map((c) => (
            <li
              key={c.id}
              className="flex items-center justify-between rounded-[6px] border border-[#e5e5e5] px-3 py-2"
            >
              <div>
                <p className="text-sm text-[#171717] capitalize">{c.provider}</p>
                <p className="text-xs text-[#737373] font-mono">
                  {c.model} ····{c.key_hint}
                </p>
              </div>
              <div className="flex items-center gap-2">
                {c.is_default ? (
                  <span className="rounded-full bg-[#dcfce7] px-2 py-0.5 text-xs text-[#0a0a0a]">
                    Default
                  </span>
                ) : (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => onMakeDefault(c.id)}
                  >
                    Make default
                  </Button>
                )}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onRemove(c.id)}
                  aria-label={`Remove ${c.provider} key`}
                >
                  Remove
                </Button>
              </div>
            </li>
          ))}
        </ul>
      ) : (
        <p className="rounded-[6px] border border-[#e5e5e5] bg-[#f5f5f5] px-3 py-2 text-xs text-[#737373]">
          No key yet. PaperViz cannot analyse a paper until you add one.
        </p>
      )}

      <form onSubmit={handleSubmit} className="mt-4 space-y-3">
        <div>
          <label htmlFor="provider" className="block text-xs font-medium text-[#737373] mb-1.5">
            Provider
          </label>
          <select
            id="provider"
            value={provider}
            onChange={(e) => setProvider(e.target.value)}
            disabled={savingCred}
            className="w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20"
          >
            {PROVIDERS.map((p) => (
              <option key={p.value} value={p.value}>
                {p.label}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label htmlFor="model" className="block text-xs font-medium text-[#737373] mb-1.5">
            Model <span className="font-normal">(optional)</span>
          </label>
          <input
            id="model"
            type="text"
            value={model}
            onChange={(e) => setModel(e.target.value)}
            placeholder="Provider default"
            disabled={savingCred}
            className="w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20"
          />
        </div>
        <div>
          <label htmlFor="modelKey" className="block text-xs font-medium text-[#737373] mb-1.5">
            API key
          </label>
          <input
            id="modelKey"
            type="password"
            autoComplete="off"
            value={apiKeyInput}
            onChange={(e) => setApiKeyInput(e.target.value)}
            placeholder="Paste your provider key"
            disabled={savingCred}
            className="w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20"
          />
        </div>
        <Button type="submit" disabled={savingCred}>
          {savingCred ? "Saving..." : "Add key"}
        </Button>
      </form>
    </div>
  )
}
