import { Button } from "@/components/ui/button"
import { UploadDropzone } from "@/components/upload-dropzone"
import { ReadingLevelSelector } from "@/components/ui/reading-level-selector"
import { ErrorBanner } from "@/components/ui/status-banners"
import { ArrowRight } from "lucide-react"
import { useIngest, INGEST_TABS } from "@/hooks/useIngest"

// Ingest unifies PDF / paste / DOI / URL into one form: mode switch, explicit
// source_type badge, per-mode validation, and inline error + Retry (12.1).
export function Ingest({ onIngested }) {
  const {
    mode, switchMode, sourceType,
    file, setFile, text, setText,
    readingLevel, setReadingLevel,
    doi, setDoi, url, setUrl,
    loading, error, setError,
    handleSubmit, handleRetry,
  } = useIngest({ onIngested })

  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-4 sm:p-6">
      <div className="flex items-center justify-between mb-4">
        <div className="flex gap-1 rounded-full border border-[#e5e5e5] bg-white p-1" role="tablist" aria-label="Ingest source">
          {INGEST_TABS.map((t) => (
            <button
              key={t.id}
              type="button"
              role="tab"
              aria-selected={mode === t.id}
              onClick={() => switchMode(t.id)}
              className={`rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${mode === t.id ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717] hover:bg-[#f5f5f5]"}`}
            >
              {t.label}
            </button>
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
  )
}
