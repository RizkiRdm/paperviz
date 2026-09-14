import { useState } from "react"
import { ChevronDown, FileText } from "lucide-react"

// Render source section — collapsible original_text + metadata grid, DESIGN.md Dub tokens.
export function SourceSection({ doc, showOriginal, onToggleOriginal }) {
  const [expanded, setExpanded] = useState(showOriginal)

  // Keep local expanded in sync when page toggles showOriginal externally.
  const open = showOriginal || expanded
  const originalText = doc?.original_text || ""
  const hasOriginal = originalText.trim().length > 0

  function handleToggle() {
    const next = !open
    setExpanded(next)
    onToggleOriginal?.(next)
  }

  return (
    <section className="mt-8">
      <div className="flex items-center gap-2 mb-4">
        <FileText className="h-4 w-4 text-[#2563eb]" aria-hidden="true" />
        <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">Source</h2>
        {doc?.reading_level && (
          <span className="inline-flex items-center rounded-full bg-[#f5f5f5] px-2 py-0.5 text-[11px] font-medium text-[#737373]">
            {doc.reading_level === "eli5" ? "ELI5" : doc.reading_level}
          </span>
        )}
      </div>

      {/* Metadata — hairline border, 12px radius, compact grid */}
      <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-4">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
          <div>
            <p className="text-[11px] font-medium text-[#737373] uppercase tracking-wider">Status</p>
            <p className="mt-1 font-medium text-[#171717]">{doc?.status || "—"}</p>
          </div>
          <div>
            <p className="text-[11px] font-medium text-[#737373] uppercase tracking-wider">Chapters</p>
            <p className="mt-1 font-medium text-[#171717]">{doc?.chapters?.length ?? 0}</p>
          </div>
          <div>
            <p className="text-[11px] font-medium text-[#737373] uppercase tracking-wider">Charts</p>
            <p className="mt-1 font-medium text-[#171717]">{doc?.charts?.length ?? 0}</p>
          </div>
          <div>
            <p className="text-[11px] font-medium text-[#737373] uppercase tracking-wider">Evidence</p>
            <p className="mt-1 font-medium text-[#171717]">{doc?.evidence?.length ?? 0}</p>
          </div>
        </div>
        {doc?.id && (
          <p className="mt-3 text-[11px] text-[#737373] font-mono break-all">ID {doc.id}</p>
        )}
      </div>

      {/* Original text collapsible — preserve input on toggle, no silent catch */}
      <div className="mt-3 rounded-[12px] border border-[#e5e5e5] bg-white p-4">
        <button
          type="button"
          onClick={handleToggle}
          aria-expanded={open}
          aria-controls="source-original-text"
          className="inline-flex items-center gap-1.5 text-xs font-medium text-[#171717] hover:text-[#0a0a0a] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20 rounded-[6px] px-1 -ml-1"
        >
          {open ? "Hide original text" : "Show original text"}
          <ChevronDown className={`h-3.5 w-3.5 text-[#737373] transition-transform ${open ? "rotate-180" : ""}`} aria-hidden="true" />
        </button>
        {!hasOriginal && open && (
          <p className="mt-3 text-xs text-[#737373]">No original text available for this document.</p>
        )}
        {hasOriginal && open && (
          <div
            id="source-original-text"
            className="mt-3 max-h-[420px] overflow-y-auto whitespace-pre-wrap break-words rounded-[8px] border border-[#e5e5e5] bg-[#f5f5f5] p-4 text-xs leading-relaxed text-[#404040]"
          >
            {originalText}
          </div>
        )}
      </div>
    </section>
  )
}
