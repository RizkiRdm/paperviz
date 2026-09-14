import { useState } from "react"
import { BookMarked, ChevronDown, FileText, Table, Image as ImageIcon } from "lucide-react"

// Render single evidence row with provenance — collapsible source text.
function EvidenceRow({ item }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-4">
      <div className="flex flex-wrap items-center gap-2">
        {item.source_reference && (
          <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">{item.source_reference}</span>
        )}
        {item.section && (
          <span className="inline-flex items-center gap-1 rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]"><FileText className="h-3 w-3" aria-hidden="true" />{item.section}</span>
        )}
        {item.page && (
          <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">Page {item.page}</span>
        )}
        {item.figure_id && (
          <span className="inline-flex items-center gap-1 rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]"><ImageIcon className="h-3 w-3" aria-hidden="true" />{item.figure_id}</span>
        )}
        {item.table_id && (
          <span className="inline-flex items-center gap-1 rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]"><Table className="h-3 w-3" aria-hidden="true" />{item.table_id}</span>
        )}
        {!item.source_reference && !item.section && !item.page && !item.figure_id && !item.table_id && (
          <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">Evidence</span>
        )}
      </div>
      <p className="mt-3 text-xs leading-relaxed text-[#171717] line-clamp-3">{item.source_text}</p>
      <button type="button" onClick={() => setOpen(v => !v)} aria-expanded={open} className="mt-2 inline-flex items-center gap-1 text-[11px] font-medium text-[#2563eb] hover:text-[#1e40af] transition-colors">
        {open ? "Hide source text" : "View source text"}<ChevronDown className={`h-3 w-3 transition-transform ${open ? "rotate-180" : ""}`} aria-hidden="true" />
      </button>
      {open && (
        <p className="mt-2 max-h-48 overflow-y-auto whitespace-pre-wrap rounded-[8px] border border-[#e5e5e5] bg-[#f5f5f5] p-3 text-[11px] leading-relaxed text-[#525252]">{item.source_text}</p>
      )}
    </div>
  )
}

// Render evidence section — claims/source text/page/section/figure refs with empty state.
export function EvidenceSection({ evidence = [], onRetry }) {
  if (!evidence || evidence.length === 0) {
    return (
      <section className="mt-8">
        <div className="flex items-center gap-2 mb-4">
          <BookMarked className="h-4 w-4 text-[#2563eb]" aria-hidden="true" />
          <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">Evidence</h2>
        </div>
        <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
          <p className="text-xs text-[#737373]">No evidence references extracted for this paper.</p>
          {onRetry && (
            <button type="button" onClick={onRetry} className="mt-2 text-[11px] font-medium text-[#2563eb] hover:text-[#1e40af] transition-colors">Retry</button>
          )}
        </div>
      </section>
    )
  }
  return (
    <section className="mt-8">
      <div className="flex items-center gap-2 mb-4">
        <BookMarked className="h-4 w-4 text-[#2563eb]" aria-hidden="true" />
        <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">Evidence</h2>
        <span className="inline-flex items-center rounded-full bg-[#f5f5f5] px-2 py-0.5 text-[11px] font-medium text-[#737373]">{evidence.length} {evidence.length === 1 ? "reference" : "references"}</span>
      </div>
      <div className="flex flex-col gap-3">
        {evidence.map((e) => (
          <EvidenceRow key={e.id} item={e} />
        ))}
      </div>
    </section>
  )
}
