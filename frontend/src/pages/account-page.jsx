import { useState, useEffect, useCallback } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"

const PROVIDERS = [
  { value: "gemini", label: "Google Gemini" },
  { value: "anthropic", label: "Anthropic Claude" },
  { value: "openai", label: "OpenAI" },
]

export function AccountPage() {
  const navigate = useNavigate()
  const [user, setUser] = useState(null)
  const [keyConfigured, setKeyConfigured] = useState(false)
  const [revealedKey, setRevealedKey] = useState(null)
  const [usage, setUsage] = useState(null)
  const [credentials, setCredentials] = useState([])
  const [loading, setLoading] = useState(true)

  const [provider, setProvider] = useState("gemini")
  const [model, setModel] = useState("")
  const [apiKeyInput, setApiKeyInput] = useState("")
  const [credError, setCredError] = useState(null)
  const [savingCred, setSavingCred] = useState(false)

  const [copied, setCopied] = useState(false)
  const [busy, setBusy] = useState(false)

  const loadCredentials = useCallback(async () => {
    const res = await fetch("/api/credentials")
    if (!res.ok) throw new Error(`credentials ${res.status}`)
    setCredentials(await res.json())
  }, [])

  useEffect(() => {
    let cancelled = false
    fetch("/api/auth/me")
      .then((res) => {
        if (!res.ok) {
          navigate("/login")
          return null
        }
        return res.json()
      })
      .then(async (userData) => {
        if (!userData || cancelled) return
        setUser(userData)
        const [keyRes, usageRes, credRes] = await Promise.all([
          fetch("/api/auth/apikey"),
          fetch("/api/usage"),
          fetch("/api/credentials"),
        ])
        if (cancelled) return
        if (keyRes.ok) setKeyConfigured((await keyRes.json()).configured === true)
        if (usageRes.ok) setUsage(await usageRes.json())
        if (credRes.ok) setCredentials(await credRes.json())
      })
      .catch((err) => {
        console.error("Failed to load account data", err)
        setCredError("Could not load your account. Retry.")
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [navigate])

  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(revealedKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error("Copy to clipboard failed", err)
      setCredError("Could not copy. Select the key and copy it manually.")
    }
  }

  // The key is returned exactly once, at issue time. It is never retrievable
  // afterwards, because the server stores only a digest.
  async function issueKey(regenerate) {
    setBusy(true)
    setCredError(null)
    try {
      const res = await fetch(
        regenerate ? "/api/auth/apikey/regenerate" : "/api/auth/apikey",
        { method: "POST" },
      )
      if (!res.ok) throw new Error(`apikey ${res.status}`)
      const data = await res.json()
      setRevealedKey(data.key)
      setKeyConfigured(true)
    } catch (err) {
      console.error("Issue API key failed", err)
      setCredError(
        regenerate
          ? "Could not replace the key. Retry."
          : "Could not create a key. Retry.",
      )
    } finally {
      setBusy(false)
    }
  }

  async function addCredential(e) {
    e.preventDefault()
    if (!apiKeyInput.trim()) {
      setCredError("Enter your model API key.")
      return
    }
    setSavingCred(true)
    setCredError(null)
    try {
      const res = await fetch("/api/credentials", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          provider,
          model: model.trim(),
          api_key: apiKeyInput.trim(),
        }),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        throw new Error(body.error || `credentials ${res.status}`)
      }
      setApiKeyInput("")
      await loadCredentials()
    } catch (err) {
      console.error("Save credential failed", err)
      setCredError(credentialErrorMessage(err))
    } finally {
      setSavingCred(false)
    }
  }

  async function makeDefault(id) {
    setCredError(null)
    try {
      const res = await fetch(`/api/credentials/${id}/default`, { method: "POST" })
      if (!res.ok) throw new Error(`default ${res.status}`)
      await loadCredentials()
    } catch (err) {
      console.error("Set default credential failed", err)
      setCredError("Could not change the default key. Retry.")
    }
  }

  async function removeCredential(id) {
    setCredError(null)
    try {
      const res = await fetch(`/api/credentials/${id}`, { method: "DELETE" })
      if (!res.ok) throw new Error(`delete ${res.status}`)
      await loadCredentials()
    } catch (err) {
      console.error("Delete credential failed", err)
      setCredError("Could not remove the key. Retry.")
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center">
        <p className="text-sm text-[#737373]">Loading...</p>
      </div>
    )
  }

  if (!user) {
    return null
  }

  const hasCredential = credentials.length > 0

  return (
    <div className="min-h-screen bg-white bg-dotted-grid">
      <div className="mx-auto max-w-lg px-6 py-12">
        <div className="mb-8">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold tracking-tight text-[#0a0a0a]">PaperViz</span>
          </Link>
          <h1 className="font-satoshi text-2xl font-medium text-[#0a0a0a]">Account</h1>
          <p className="text-sm text-[#737373]">{user.email}</p>
        </div>

        <div className="space-y-6">
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
                          onClick={() => makeDefault(c.id)}
                        >
                          Make default
                        </Button>
                      )}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => removeCredential(c.id)}
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

            <form onSubmit={addCredential} className="mt-4 space-y-3">
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
                  onClick={() => issueKey(false)}
                  disabled={busy}
                >
                  {busy ? "Working..." : keyConfigured ? "Replace key" : "Create key"}
                </Button>
              </>
            )}
          </div>

          <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
            <h2 className="text-sm font-medium text-[#0a0a0a] mb-4">Usage This Month</h2>
            <p className="text-3xl font-mono font-medium text-[#0a0a0a]">
              {usage?.count ?? 0}
            </p>
            <p className="text-xs text-[#737373] mt-1">papers analyzed</p>
          </div>

          {credError && (
            <div className="rounded-[6px] bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700 flex items-center justify-between gap-3">
              <span>{credError}</span>
              <button
                onClick={() => setCredError(null)}
                className="shrink-0 underline"
                aria-label="Dismiss error"
              >
                Dismiss
              </button>
            </div>
          )}

          <div className="text-center">
            <Link to="/agents" className="text-sm text-[#2563eb] hover:underline">
              Get started with PaperViz →
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}

// credentialErrorMessage turns a server error code into something a user can act
// on, falling back to a retry prompt rather than exposing the raw code.
function credentialErrorMessage(err) {
  const code = String(err.message || "")
  if (code.includes("unsupported_provider")) {
    return "That provider is not supported yet."
  }
  if (code.includes("missing_api_key")) {
    return "Enter your model API key."
  }
  return "Could not save the key. Retry."
}
