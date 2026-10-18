// Account data hook — loads user/apikey/usage/credentials and owns credential
// + agent-key mutations. Presentation stays in components/account/*.
import { useState, useEffect, useCallback } from "react"
import { useNavigate } from "react-router-dom"
import { capture } from "@/lib/analytics"
import { authFetch } from "@/lib/auth"

export function useAccount() {
  const navigate = useNavigate()
  const [user, setUser] = useState(null)
  const [keyConfigured, setKeyConfigured] = useState(false)
  const [revealedKey, setRevealedKey] = useState(null)
  const [usage, setUsage] = useState(null)
  const [credentials, setCredentials] = useState([])
  const [loading, setLoading] = useState(true)
  const [credError, setCredError] = useState(null)
  const [busy, setBusy] = useState(false)
  const [savingCred, setSavingCred] = useState(false)

  // loadCredentials refetches the credential list after a mutation.
  const loadCredentials = useCallback(async () => {
    const res = await authFetch("/api/credentials")
    if (!res.ok) throw new Error(`credentials ${res.status}`)
    setCredentials(await res.json())
  }, [])

  // Initial load: verify session, then fetch apikey/usage/credentials in parallel.
  useEffect(() => {
    let cancelled = false
    authFetch("/api/auth/me")
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
          authFetch("/api/auth/apikey"),
          authFetch("/api/usage"),
          authFetch("/api/credentials"),
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

  // The agent key is returned exactly once, at issue time. It is never
  // retrievable afterwards, because the server stores only a digest.
  async function issueKey(regenerate) {
    setBusy(true)
    setCredError(null)
    try {
      const res = await authFetch(
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

  // addCredential saves a model key; returns true on success so the form can clear its input.
  async function addCredential({ provider, model, apiKey }) {
    if (!apiKey.trim()) {
      setCredError("Enter your model API key.")
      return false
    }
    setSavingCred(true)
    setCredError(null)
    try {
      const res = await authFetch("/api/credentials", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          provider,
          model: model.trim(),
          api_key: apiKey.trim(),
        }),
      })
      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        throw new Error(body.error || `credentials ${res.status}`)
      }
      capture("api_key_added", { provider })
      await loadCredentials()
      return true
    } catch (err) {
      console.error("Save credential failed", err)
      setCredError(credentialErrorMessage(err))
      return false
    } finally {
      setSavingCred(false)
    }
  }

  // makeDefault promotes a stored credential to the active one.
  async function makeDefault(id) {
    setCredError(null)
    try {
      const res = await authFetch(`/api/credentials/${id}/default`, { method: "POST" })
      if (!res.ok) throw new Error(`default ${res.status}`)
      await loadCredentials()
    } catch (err) {
      console.error("Set default credential failed", err)
      setCredError("Could not change the default key. Retry.")
    }
  }

  // removeCredential deletes a stored credential.
  async function removeCredential(id) {
    setCredError(null)
    try {
      const res = await authFetch(`/api/credentials/${id}`, { method: "DELETE" })
      if (!res.ok) throw new Error(`delete ${res.status}`)
      await loadCredentials()
    } catch (err) {
      console.error("Delete credential failed", err)
      setCredError("Could not remove the key. Retry.")
    }
  }

  return {
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
  }
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
