// URL import component — input field + import button for importing papers from a URL
import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { ErrorBanner } from "@/components/ui/status-banners"
import { importByURL } from "@/lib/api"
import { ArrowRight } from "lucide-react"

export function URLImport() {
  const navigate = useNavigate()
  const [url, setUrl] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  // Basic URL validation — must be a valid URL with http(s) scheme
  function isValidURL(value) {
    try {
      const parsed = new URL(value)
      return parsed.protocol === "http:" || parsed.protocol === "https:"
    } catch {
      return false
    }
  }

  async function handleImport() {
    setError(null)
    const trimmed = url.trim()

    if (!trimmed) {
      setError("Please enter a URL.")
      return
    }
    if (!isValidURL(trimmed)) {
      setError("Invalid URL. Please enter a valid URL starting with https://")
      return
    }

    setLoading(true)
    try {
      const result = await importByURL(trimmed)
      navigate(`/${result.document_id}`)
    } catch (err) {
      setError(
        err.code === "fetch_failed"
          ? "Could not fetch the paper from this URL. Please check the link and try again."
          : "Something went wrong. Please try again.",
      )
      setLoading(false)
    }
  }

  return (
    <div className="space-y-3">
      <label htmlFor="url-input" className="block text-xs font-medium text-[#737373] uppercase tracking-wider">
        URL
      </label>
      <div className="flex gap-2">
        <input
          id="url-input"
          type="url"
          value={url}
          onChange={(e) => {
            setUrl(e.target.value)
            setError(null)
          }}
          placeholder="https://arxiv.org/abs/2301.00001"
          className="flex-1 rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]"
          disabled={loading}
          onKeyDown={(e) => {
            if (e.key === "Enter") handleImport()
          }}
        />
        <Button onClick={handleImport} disabled={loading} variant="secondary">
          {loading ? (
            "Importing…"
          ) : (
            <span className="flex items-center justify-center gap-1.5">
              Import <ArrowRight className="h-4 w-4" />
            </span>
          )}
        </Button>
      </div>

      {error && (
        <ErrorBanner message={error} />
      )}
    </div>
  )
}
