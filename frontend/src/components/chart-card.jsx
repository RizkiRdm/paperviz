// ponytail: chart card redesign per DESIGN.md (hairline border, white canvas surface)
import { lazy, Suspense, useState } from "react"
import { Image as ImageIcon, BookMarked, ChevronDown } from "lucide-react"
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip"

const LazyDataChart = lazy(() => import("./data-chart"))

// Evidence provenance strip — ties the AI interpretation back to the original
// source page (Rule 4: explanations must carry provenance). Only rendered for
// charts that have evidence rows (image-origin charts).
function EvidenceBlock({ evidence }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="mt-5 border-t border-[#e5e5e5] pt-4">
      <div className="flex flex-wrap items-center gap-2">
        <span className="inline-flex items-center gap-1.5 text-[11px] font-semibold text-[#0a0a0a]">
          <BookMarked className="h-3.5 w-3.5 text-[#2563eb]" aria-hidden="true" />
          Source
        </span>
        {evidence.source_reference && (
          <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
            {evidence.source_reference}
          </span>
        )}
        {evidence.section && (
          <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
            {evidence.section}
          </span>
        )}
        {!evidence.source_reference && evidence.page > 0 && (
          <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
            Page {evidence.page}
          </span>
        )}
      </div>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        className="mt-2 inline-flex items-center gap-1 text-[11px] font-medium text-[#2563eb] hover:text-[#1e40af] transition-colors cursor-pointer"
      >
        {open ? "Hide source text" : "View source text"}
        <ChevronDown className={`h-3 w-3 transition-transform ${open ? "rotate-180" : ""}`} aria-hidden="true" />
      </button>
      {open && (
        <p className="mt-2 max-h-48 overflow-y-auto whitespace-pre-wrap rounded-[8px] border border-[#e5e5e5] bg-[#f5f5f5] p-3 text-[11px] leading-relaxed text-[#525252]">
          {evidence.source_text}
        </p>
      )}
    </div>
  )
}

export function ChartCard({ chart, chapterTitle, evidence = [] }) {
  const pageNumber = chart.page_number || 0
  const isDataExtracted = chart.source_method === "data_extracted"
  const isImageFallback = chart.source_method === "image_fallback"
  const isOmitted = chart.source_method === "omitted"

  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-5">
      <div className="grid gap-5 md:grid-cols-2">
        <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-4">
          <div className="flex flex-wrap items-center gap-2 mb-3">
            <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
              {pageNumber > 0 ? `Original Figure · Page ${pageNumber}` : "Original Source"}
            </span>
            {chapterTitle && (
              <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
                {chapterTitle}
              </span>
            )}
          </div>

          {isImageFallback && chart.image_url ? (
            <img
              src={chart.image_url}
              alt={`Original figure${pageNumber > 0 ? ` on page ${pageNumber}` : ""}`}
              className="h-auto w-full rounded-[8px] border border-[#e5e5e5] bg-white"
            />
          ) : (
            <div className="flex items-start gap-3">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-[#e5e5e5] bg-white text-[#737373]">
                <ImageIcon className="h-4 w-4" />
              </div>
              <p className="text-xs leading-relaxed text-[#737373]">
                {isOmitted
                  ? chart.annotation
                  : "No original figure captured for this figure — PaperViz reconstructed it from the paper's text."}
              </p>
            </div>
          )}
        </div>

        <div>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-flex items-center rounded-full bg-[#dbeaff] px-2.5 py-0.5 text-[11px] font-medium text-[#2563eb]">
                PaperViz AI Interpretation
              </span>
            </TooltipTrigger>
            <TooltipContent>AI-generated interpretation of the original figure</TooltipContent>
          </Tooltip>
          <div className="mt-3">
            {isDataExtracted && (
              <Suspense
                fallback={<p className="text-xs text-[#737373]">Loading interactive chart…</p>}
              >
                <LazyDataChart chartData={chart.chart_data} />
              </Suspense>
            )}
            {isImageFallback && chart.annotation && (
              <p className="text-xs leading-relaxed text-[#171717]">{chart.annotation}</p>
            )}
            {isOmitted && <p className="text-xs italic leading-relaxed text-[#737373]">{chart.annotation}</p>}
          </div>
        </div>
      </div>

      {evidence.length > 0 && <EvidenceBlock evidence={evidence[0]} />}
    </div>
  )
}