import { useState } from "react"
import { createDocument, importByDOI, importByURL, listCredentials } from "@/lib/api"
import { useApi } from "@/hooks/useApi"
import { useCurrentUser } from "@/hooks/use-current-user"
import { loadDraft, saveDraft, clearDraft } from "@/lib/draft"

// Tab config drives the mode switch and the explicit source_type badge (P05).
export const INGEST_TABS = [
  { id: "pdf", label: "PDF", source: "pdf" },
  { id: "paste", label: "Paste", source: "pasted_text" },
  { id: "doi", label: "DOI", source: "doi" },
  { id: "url", label: "URL", source: "url" },
]

// useIngest owns mode state, per-mode validation and submit for all four
// ingestion sources — one form, explicit source_type routing, and no
// PDF-only operations leaking into pasted_text (file stays null there).
// State initialises from a sessionStorage draft saved by the pre-submit
// auth gate, so a guest returns from sign-in to their input. Navigation
// stays router-agnostic: guests are handed to onAuthRequired (wired by
// the component) instead of navigating inside the hook.
export function useIngest({ onIngested, onAuthRequired }) {
  const [draft] = useState(loadDraft)
  const [mode, setMode] = useState(draft?.mode || "pdf")
  const [file, setFile] = useState(null)
  const [text, setText] = useState(draft?.text || "")
  const [readingLevel, setReadingLevel] = useState(draft?.readingLevel || "simplified")
  const [doi, setDoi] = useState(draft?.doi || "")
  const [url, setUrl] = useState(draft?.url || "")

  // restored flags a draft was present at mount; PDF bytes are never
  // restorable, so the component prompts for a re-pick when mode is pdf.
  const restored = draft !== null

  // Auth status gates submit before any API call: guests save a draft and
  // are sent to sign-in rather than hitting a missing_credential surprise.
  const { user, checked } = useCurrentUser()

  // needsCredential flags the signed-in-but-no-key wall so the component can
  // link to /account instead of offering a Retry that cannot succeed.
  const [needsCredential, setNeedsCredential] = useState(false)

  // Active tab defines the source_type sent for badge display and routing.
  const sourceType = INGEST_TABS.find((t) => t.id === mode).source

  // useApi wraps submit with loading/error state; navigation happens in the
  // onIngested callback so the hook stays router-agnostic.
  const { execute, loading, error, setError } = useApi(async () => {
    if (mode === "pdf" || mode === "paste") {
      const result = await createDocument({
        file: mode === "pdf" ? file : null,
        text: mode === "paste" ? text : null,
        readingLevel,
      })
      onIngested(result.document_id)
      return
    }
    if (mode === "doi") {
      const result = await importByDOI(doi.trim())
      onIngested(result.document_id)
      return
    }
    const result = await importByURL(url.trim())
    onIngested(result.document_id)
  })

  // validate returns a terse message for the active mode, or null.
  function validate() {
    if (mode === "pdf" && !file) return "Select a PDF to upload."
    if (mode === "paste" && text.trim().length < 50) return "Paste at least 50 characters."
    if (mode === "doi" && !/^10\.\d{4,}\/\S+/.test(doi.trim())) return "Invalid DOI. Must start with 10. (e.g. 10.1038/s41586-020-2649-2)"
    if (mode === "url") {
      if (!url.trim()) return "Please enter a URL."
      try {
        const u = new URL(url.trim())
        if (u.protocol !== "http:" && u.protocol !== "https:") return "Invalid URL. Use https://"
      } catch {
        return "Invalid URL. Use https://"
      }
    }
    return null
  }

  // handleSubmit validates, then runs the pre-submit auth gate, then
  // submits. Guests' draft is saved before onAuthRequired navigates away.
  async function handleSubmit() {
    const msg = validate()
    if (msg) {
      setError(msg)
      return
    }
    if (!checked) {
      setError("Checking sign-in…")
      return
    }
    if (!user) {
      saveDraft({ mode, text, doi, url, readingLevel })
      setError("Sign in to analyse — your input is saved")
      onAuthRequired?.()
      return
    }
    setNeedsCredential(false)
    // Credential wall: a signed-in user with no model key cannot analyse
    // anything (BYOK), so fail here with a path forward instead of a 400
    // from the document endpoint. Input is preserved untouched.
    try {
      const creds = await listCredentials()
      if (!Array.isArray(creds) || creds.length === 0) {
        setNeedsCredential(true)
        setError("Add a model key to analyse — PaperViz runs on your own key.")
        return
      }
    } catch (err) {
      console.error("Credential check failed", err)
      setError("Could not check your model key. Retry.")
      return
    }
    try {
      await execute()
      clearDraft()
    } catch (err) {
      console.error(`ingest failed [${sourceType}]`, err)
      const code = err.code
      if (code === "missing_credential") {
        setNeedsCredential(true)
        setError("Add a model key to analyse — PaperViz runs on your own key.")
      }
      else if (code === "fetch_failed") setError(mode === "doi" ? "Could not fetch that DOI. Check ID and retry." : "Could not fetch that URL. Check link and retry.")
      else if (code === "rate_limited") setError("Too many requests. Wait a moment and retry.")
      else if (code) setError(code.replaceAll("_", " "))
      else setError("Something went wrong. Please retry.")
    }
  }

  // handleRetry re-runs submit with every input preserved (12.1).
  function handleRetry() {
    handleSubmit()
  }

  // switchMode clears stale errors when the user changes source type.
  function switchMode(nextMode) {
    setMode(nextMode)
    setError(null)
    setNeedsCredential(false)
  }

  return {
    mode, switchMode, sourceType,
    file, setFile, text, setText,
    readingLevel, setReadingLevel,
    doi, setDoi, url, setUrl,
    loading, error, setError,
    needsCredential,
    restored,
    handleSubmit, handleRetry,
  }
}
