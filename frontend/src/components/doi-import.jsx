// DOI import component — input field + import button for importing papers by DOI identifier
import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { ErrorBanner } from "@/components/ui/status-banners"
import { importByDOI } from "@/lib/api"
import { ArrowRight } from "lucide-react"

export function DOIImport() {
  const navigate = useNavigate()
  const [doi, setDoi] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  // Basic DOI validation — must start with "10." per DOI standard
  function isValidDOI(value) {
    return /^10\.\d{4,}\/\S+/.test(value.trim())
  }

  async function handleImport() {
    setError(null)
    const trimmed = doi.trim()

    if (!trimmed) {
      setError("Please enter a DOI.")
      return
    }
    if (!isValidDOI(trimmed)) {
      setError("Invalid DOI format. It should start with \"10.\" (e.g. 10.1038/s41586-020-2649-2).")
      return
    }

    setLoading(true)
    try {
      const result = await importByDOI(trimmed)
      navigate(`/${result.document_id}`)
    } catch (err) {
      setError(
        err.code === "fetch_failed"
          ? "Could not fetch the paper by this DOI. Please check the DOI and try again."
          : "Something went wrong. Please try again.",
      )
      setLoading(false)
    }
  }

  return (
    <div className="space-y-3">
      <label htmlFor="doi-input" className="block text-xs font-medium text-[#737373] uppercase tracking-wider">
        DOI
      </label>
      <div className="flex gap-2">
        <input
          id="doi-input"
          type="text"
          value={doi}
          onChange={(e) => {
            setDoi(e.target.value)
            setError(null)
          }}
          placeholder="10.1038/s41586-020-2649-2"
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
