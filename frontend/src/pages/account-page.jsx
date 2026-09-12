import { useState, useEffect } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"

export function AccountPage() {
  const navigate = useNavigate()
  const [user, setUser] = useState(null)
  const [apiKey, setApiKey] = useState(null)
  const [usage, setUsage] = useState(null)
  const [loading, setLoading] = useState(true)
  const [copied, setCopied] = useState(false)
  const [regenerating, setRegenerating] = useState(false)

  useEffect(() => {
    fetch("/api/auth/me")
      .then((res) => {
        if (!res.ok) {
          navigate("/login")
          return null
        }
        return res.json()
      })
      .then((userData) => {
        if (userData) {
          setUser(userData)
          return Promise.all([
            fetch("/api/auth/apikey").then((r) => (r.ok ? r.json() : null)),
            fetch("/api/usage").then((r) => (r.ok ? r.json() : null)),
          ])
        }
        return [null, null]
      })
      .then(([keyData, usageData]) => {
        if (keyData?.key) setApiKey(keyData.key)
        if (usageData) setUsage(usageData)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [navigate])

  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(apiKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {}
  }

  async function regenerateKey() {
    setRegenerating(true)
    try {
      const res = await fetch("/api/auth/apikey/regenerate", { method: "POST" })
      if (res.ok) {
        const data = await res.json()
        setApiKey(data.key)
      }
    } catch {}
    setRegenerating(false)
  }

  async function openBilling() {
    try {
      const res = await fetch("/api/billing/portal", { method: "POST" })
      if (res.ok) {
        const data = await res.json()
        window.location.href = data.url
      }
    } catch {}
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
          <p className="mt-1 text-sm text-[#737373]">{user.email}</p>
        </div>

        <div className="space-y-6">
          <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
            <h2 className="text-sm font-medium text-[#0a0a0a] mb-4">API Key</h2>
            {apiKey ? (
              <>
                <div className="relative">
                  <pre className="rounded-[6px] bg-[#f5f5f5] border border-[#e5e5e5] p-3 text-xs font-mono text-[#171717] overflow-x-auto whitespace-pre-wrap break-all">
                    {apiKey}
                  </pre>
                  <button
                    onClick={copyToClipboard}
                    className="absolute top-2 right-2 rounded-[6px] border border-[#e5e5e5] bg-white px-2 py-1 text-xs font-medium text-[#737373] hover:bg-[#f5f5f5] transition-colors"
                  >
                    {copied ? "Copied!" : "Copy"}
                  </button>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={regenerateKey}
                  disabled={regenerating}
                  className="mt-3"
                >
                  {regenerating ? "Regenerating..." : "Regenerate Key"}
                </Button>
              </>
            ) : (
              <p className="text-sm text-[#737373]">No API key yet. Visit /agents to generate one.</p>
            )}
          </div>

          <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
            <h2 className="text-sm font-medium text-[#0a0a0a] mb-2">Usage This Month</h2>
            <p className="text-3xl font-mono font-medium text-[#0a0a0a]">
              {usage?.count ?? 0}
            </p>
            <p className="text-xs text-[#737373] mt-1">papers analyzed</p>
          </div>

          <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
            <h2 className="text-sm font-medium text-[#0a0a0a] mb-4">Subscription</h2>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-[#0a0a0a] capitalize">
                  {user.subscription_tier || "free"} Plan
                </p>
                <p className="text-xs text-[#737373]">
                  {user.subscription_status === "active" ? "Active" : "Inactive"}
                </p>
              </div>
              <Button variant="outline" size="sm" onClick={openBilling}>
                Manage Billing
              </Button>
            </div>
          </div>

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
