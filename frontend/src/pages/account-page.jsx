// Account page — thin orchestrator: session gate, sign-out, usage card, and
// the shared error banner; per-section logic lives in hooks/components.
import { useState } from "react"
import { Link } from "react-router-dom"
import { useClerk } from "@clerk/clerk-react"
import { useAccount } from "@/hooks/useAccount"
import { CredentialsCard } from "@/components/account/credentials-card"
import { AgentKeyCard } from "@/components/account/agent-key-card"
import { AppShell } from "@/components/app-shell"

export function AccountPage() {
  const {
    user,
    loading,
    usage,
    credentials,
    keyConfigured,
    revealedKey,
    credError,
    setCredError,
    busy,
    savingCred,
    issueKey,
    addCredential,
    makeDefault,
    removeCredential,
  } = useAccount()

  const { signOut } = useClerk()
  const [isSigningOut, setIsSigningOut] = useState(false)
  const [signOutError, setSignOutError] = useState(null)

  // handleSignOut delegates the session teardown to Clerk.
  async function handleSignOut() {
    setIsSigningOut(true)
    try {
      await signOut({ redirectUrl: "/login" })
    } catch (err) {
      console.error("Sign out failed", err)
      setSignOutError("Could not sign out. Retry.")
      setIsSigningOut(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center">
        <p className="text-sm text-[#737373]">Loading...</p>
      </div>
    )
  }

  // No user yet: either the /login redirect is in flight or the load failed —
  // surface the load error inline instead of a blank screen (P47).
  if (!user) {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center px-6">
        <p className="text-sm text-[#737373]">{credError || "Loading..."}</p>
      </div>
    )
  }

  return (
    <AppShell active="account">
      <div className="page-intro">
        <p className="eyebrow">ACCOUNT</p>
        <h1>Account</h1>
        <p className="muted">{user.email}</p>
      </div>
      <div className="stack">
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
          <CredentialsCard
            credentials={credentials}
            savingCred={savingCred}
            onAdd={addCredential}
            onMakeDefault={makeDefault}
            onRemove={removeCredential}
          />

          <AgentKeyCard
            keyConfigured={keyConfigured}
            revealedKey={revealedKey}
            busy={busy}
            onIssue={issueKey}
            onError={setCredError}
          />

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

          <div className="text-center">
            {signOutError && (
              <div className="rounded-[6px] bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
                {signOutError}
              </div>
            )}
            <button
              type="button"
              onClick={handleSignOut}
              disabled={isSigningOut}
              className="text-xs text-[#737373] hover:text-[#171717] disabled:opacity-50"
            >
              {isSigningOut ? "Signing out..." : "Sign out"}
            </button>
          </div>
        </div>
    </AppShell>
  )
}
