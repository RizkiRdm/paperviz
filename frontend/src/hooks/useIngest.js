import { useState } from "react"
import { createDocument, importByDOI, importByURL } from "@/lib/api"
import { useApi } from "@/hooks/useApi"

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
export function useIngest({ onIngested }) {
  const [mode, setMode] = useState("pdf")
  const [file, setFile] = useState(null)
  const [text, setText] = useState("")
  const [readingLevel, setReadingLevel] = useState("simplified")
  const [doi, setDoi] = useState("")
  const [url, setUrl] = useState("")

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

  // handleSubmit validates first, then maps backend codes to inline copy.
  async function handleSubmit() {
    const msg = validate()
    if (msg) {
      setError(msg)
      return
    }
    try {
      await execute()
    } catch (err) {
      console.error(`ingest failed [${sourceType}]`, err)
      const code = err.code
      if (code === "fetch_failed") setError(mode === "doi" ? "Could not fetch that DOI. Check ID and retry." : "Could not fetch that URL. Check link and retry.")
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
  }

  return {
    mode, switchMode, sourceType,
    file, setFile, text, setText,
    readingLevel, setReadingLevel,
    doi, setDoi, url, setUrl,
    loading, error, setError,
    handleSubmit, handleRetry,
  }
}
