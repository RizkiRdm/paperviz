import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { UploadDropzone } from "@/components/upload-dropzone"
import { ReadingLevelSelector } from "@/components/ui/reading-level-selector"
import { ErrorBanner } from "@/components/ui/status-banners"
import { createDocument, importByDOI, importByURL } from "@/lib/api"
import { useApi } from "@/hooks/useApi"
import { ArrowRight } from "lucide-react"

// tab config drives mode switch and explicit source_type badge
const TABS = [
  { id: "pdf", label: "PDF", source: "pdf" },
  { id: "paste", label: "Paste", source: "pasted_text" },
  { id: "doi", label: "DOI", source: "doi" },
  { id: "url", label: "URL", source: "url" },
]

export function UploadPage() {
  const navigate = useNavigate()
  const [mode, setMode] = useState("pdf")
  const [file, setFile] = useState(null)
  const [text, setText] = useState("")
  const [readingLevel, setReadingLevel] = useState("simplified")
  const [doi, setDoi] = useState("")
  const [url, setUrl] = useState("")

  // derive explicit source_type for badge and backend routing
  const activeTab = TABS.find((t) => t.id === mode)
  const sourceType = activeTab.source

  // useApi wraps submit with loading/error/retry — no manual setLoading needed.
  const { execute, loading, error, setError } = useApi(async () => {
    if (mode === "pdf" || mode === "paste") {
      const result = await createDocument({ file: mode === "pdf" ? file : null, text: mode === "paste" ? text : null, readingLevel })
      navigate(`/${result.document_id}`)
      return
    }
    if (mode === "doi") {
      const result = await importByDOI(doi.trim())
      navigate(`/${result.document_id}`)
      return
    }
    const result = await importByURL(url.trim())
    navigate(`/${result.document_id}`)
  })

  // validate current mode and return terse message or null
  function validate() {
    if (mode === "pdf" && !file) return "Select a PDF to upload."
    if (mode === "paste" && text.trim().length < 50) return "Paste at least 50 characters."
    if (mode === "doi" && !/^10\.\d{4,}\/\S+/.test(doi.trim())) return "Invalid DOI. Must start with 10. (e.g. 10.1038/s41586-020-2649-2)"
    if (mode === "url") {
      try {
        const u = new URL(url.trim())
        if (u.protocol !== "http:" && u.protocol !== "https:") throw new Error()
      } catch {
        return "Invalid URL. Use https://"
      }
      if (!url.trim()) return "Please enter a URL."
    }
    return null
  }

  // submit selected mode with explicit source_type handling
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

  // retry preserves inputs and re-runs last submit
  function handleRetry() {
    handleSubmit()
  }

  return (
    <div className="min-h-screen bg-white bg-dotted-grid px-6 py-10">
      <div className="mx-auto w-full max-w-2xl">
        <div className="text-center mb-8">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">PV</div>
            <span className="font-mono text-sm font-semibold tracking-tight text-[#0a0a0a]">PaperViz</span>
          </Link>
          <h1 className="font-satoshi text-4xl sm:text-5xl font-medium tracking-tight text-[#0a0a0a] leading-tight">Papers, in plain language.</h1>
          <p className="mt-3 text-base text-[#737373] max-w-md mx-auto leading-relaxed">Transform dense academic PDFs into clear, verified summaries with interactive charts.</p>
          <div className="mt-5 flex flex-col sm:flex-row items-center justify-center gap-3">
            <Link to="/agents"><Button size="lg">Add to Claude Code</Button></Link>
            <Link to="/login"><Button variant="outline" size="lg">Sign in</Button></Link>
          </div>
        </div>

        <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-4 sm:p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="flex gap-1 rounded-full border border-[#e5e5e5] bg-white p-1" role="tablist" aria-label="Ingest source">
              {TABS.map((t) => (
                <button key={t.id} role="tab" aria-selected={mode === t.id} onClick={() => { setMode(t.id); setError(null) }} className={`rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${mode === t.id ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717] hover:bg-[#f5f5f5]"}`}>{t.label}</button>
              ))}
            </div>
            <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-[#f5f5f5] px-2.5 py-1 text-[11px] font-mono font-medium text-[#525252]">source: {sourceType}</span>
          </div>

          <div role="tabpanel">
            {mode === "pdf" && <UploadDropzone file={file} text={text} mode="file" onFileChange={(f) => { setFile(f); setError(null) }} onTextChange={setText} />}
            {mode === "paste" && (
              <div>
                <UploadDropzone file={file} text={text} mode="text" onFileChange={setFile} onTextChange={(v) => { setText(v); setError(null) }} />
                <p className="mt-2 text-[11px] text-[#737373]">Charts derived from text only; no image extraction for pasted_text.</p>
              </div>
            )}
            {mode === "doi" && (
              <div className="space-y-3">
                <label htmlFor="doi-input" className="block text-xs font-medium text-[#737373] uppercase tracking-wider">DOI</label>
                <div className="flex gap-2">
                  <input id="doi-input" type="text" value={doi} onChange={(e) => { setDoi(e.target.value); setError(null) }} placeholder="10.1038/s41586-020-2649-2" className="flex-1 rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]" disabled={loading} onKeyDown={(e) => e.key === "Enter" && handleSubmit()} />
                  <Button onClick={handleSubmit} disabled={loading} variant="secondary"><span className="flex items-center gap-1.5">{loading ? "Importing…" : <>Import <ArrowRight className="h-4 w-4" /></>}</span></Button>
                </div>
              </div>
            )}
            {mode === "url" && (
              <div className="space-y-3">
                <label htmlFor="url-input" className="block text-xs font-medium text-[#737373] uppercase tracking-wider">URL</label>
                <div className="flex gap-2">
                  <input id="url-input" type="url" value={url} onChange={(e) => { setUrl(e.target.value); setError(null) }} placeholder="https://arxiv.org/abs/2301.00001" className="flex-1 rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]" disabled={loading} onKeyDown={(e) => e.key === "Enter" && handleSubmit()} />
                  <Button onClick={handleSubmit} disabled={loading} variant="secondary"><span className="flex items-center gap-1.5">{loading ? "Importing…" : <>Import <ArrowRight className="h-4 w-4" /></>}</span></Button>
                </div>
              </div>
            )}
          </div>

          {(mode === "pdf" || mode === "paste") && (
            <div className="mt-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 border-t border-[#e5e5e5] pt-4">
              <ReadingLevelSelector value={readingLevel} onChange={setReadingLevel} />
              <Button onClick={handleSubmit} disabled={loading} className="w-full sm:w-auto">{loading ? "Processing…" : "Simplify paper"}</Button>
            </div>
          )}

          {error && (
            <div className="mt-4 space-y-2">
              <ErrorBanner message={error} />
              <Button variant="outline" size="lg" onClick={handleRetry} disabled={loading} className="w-full">Retry</Button>
            </div>
          )}
        </div>

        <p className="mt-4 text-center text-xs text-[#737373]">Free for researchers. No account required to try.</p>
      </div>
    </div>
  )
}
