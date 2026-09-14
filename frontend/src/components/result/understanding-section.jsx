import { Button } from "@/components/ui/button"
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip"
import { Copy, Check, Network, StickyNote, BarChart2 } from "lucide-react"
import { ChartCard } from "@/components/chart-card"
import ReactMarkdown from "react-markdown"

// Parse ## delimited sections for non-chapter fallback — returns null when unstructured.
function parseResearchSections(text) {
  if (!text) return null
  const parts = text.split(/^## /m).filter(Boolean)
  if (parts.length < 2) return null
  return parts.map((part) => {
    const lines = part.split("\n")
    const title = lines[0].trim()
    const content = lines.slice(1).join("\n").trim()
    return { title, content }
  })
}

// Render understanding block: toggle, copy, chapter tabs, article — DESIGN.md compact.
export function UnderstandingSection({
  doc,
  documentId,
  showOriginal,
  setShowOriginal,
  textCopied,
  textCopyError,
  handleCopyText,
  activeChapter,
  setActiveChapter,
  showResearchMap,
  setShowResearchMap,
  showAnnotations,
  setShowAnnotations,
}) {
  const displayedText = showOriginal ? doc.original_text : doc.simplified_text
  const levelLabel = doc.reading_level === "eli5" ? "ELI5" : doc.reading_level === "simplified" ? "Simplified" : doc.reading_level
  const hasChapters = doc.chapters && doc.chapters.length > 1
  const activeChapterData = hasChapters ? doc.chapters[activeChapter] : null
  const chapterContent = activeChapterData?.content || displayedText
  const evidenceFor = (chartId) => (doc.evidence || []).filter((e) => e.figure_id === chartId)
  const chapterCharts = hasChapters ? (doc.charts || []).filter((c) => c.chapter_id === activeChapterData?.id) : doc.charts || []

  return (
    <>
      {/* Action bar — reading level badge + view toggle + copy text */}
      <div className="mb-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-6 border-b border-[#e5e5e5]">
        <div className="flex items-center gap-3">
          <h1 className="font-satoshi text-2xl sm:text-3xl font-medium text-[#0a0a0a]">Paper Summary</h1>
          <span className="inline-flex items-center rounded-full bg-[#dbeaff] px-2.5 py-0.5 text-[11px] font-medium text-[#2563eb]">{levelLabel}</span>
        </div>
        <div className="flex items-center gap-2">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="secondary"
                onClick={() => setShowResearchMap((v) => !v)}
                aria-label={showResearchMap ? "Hide research map" : "Show research map"}
                aria-pressed={showResearchMap}
                className={`h-8 px-2.5 text-xs gap-1.5 font-medium ${showResearchMap ? "bg-[#dbeaff] text-[#2563eb] border-[#bfdbfe]" : ""}`}
              >
                <Network className="h-3.5 w-3.5" aria-hidden="true" />
                Research Map
              </Button>
            </TooltipTrigger>
            <TooltipContent>View how this paper relates to others</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="secondary"
                onClick={() => setShowAnnotations((v) => !v)}
                aria-label={showAnnotations ? "Hide notes" : "Show notes"}
                aria-pressed={showAnnotations}
                className={`h-8 px-2.5 text-xs gap-1.5 font-medium ${showAnnotations ? "bg-[#fef3c7] text-[#92400e] border-[#fde68a]" : ""}`}
              >
                <StickyNote className="h-3.5 w-3.5" aria-hidden="true" />
                Notes
              </Button>
            </TooltipTrigger>
            <TooltipContent>Add personal notes to this paper</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="secondary" onClick={handleCopyText} aria-label={textCopied ? "Copied" : "Copy text"} className="h-8 px-2.5 text-xs gap-1.5 font-medium">
                {textCopied ? <Check className="h-3.5 w-3.5 text-[#16a34a]" aria-hidden="true" /> : <Copy className="h-3.5 w-3.5 text-[#737373]" aria-hidden="true" />}
                {textCopied ? "Copied" : "Copy Text"}
              </Button>
            </TooltipTrigger>
            <TooltipContent>Copy the {showOriginal ? "original" : "simplified"} text</TooltipContent>
          </Tooltip>
          <div className="inline-flex gap-1 rounded-full border border-[#e5e5e5] bg-white p-1 shadow-2xs" role="group" aria-label="View toggle">
            <button
              onClick={() => setShowOriginal(false)}
              aria-pressed={!showOriginal}
              aria-label="Show simplified"
              className={`rounded-full px-3.5 py-1 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20 ${!showOriginal ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717]"}`}
            >
              Simplified
            </button>
            <button
              onClick={() => setShowOriginal(true)}
              aria-pressed={showOriginal}
              aria-label="Show original"
              className={`rounded-full px-3.5 py-1 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20 ${showOriginal ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717]"}`}
            >
              Original
            </button>
          </div>
        </div>
      </div>

      {textCopyError && (
        <p className="mb-4 -mt-2 text-[11px] text-[#ea580c]" role="alert">
          Couldn&apos;t copy automatically. Select the text and press Ctrl+C / Cmd+C.
        </p>
      )}

      {/* Main layout container — flex row with sidebar on desktop when chapters exist */}
      <div className={hasChapters ? "md:flex md:gap-8 md:items-start" : ""}>
        {/* Mobile chapter tabs — horizontal scrollable strip */}
        {hasChapters && (
          <div role="tablist" aria-label="Paper sections" className="md:hidden mb-6 flex gap-1 overflow-x-auto pb-2 scrollbar-hide">
            {doc.chapters.map((ch, i) => (
              <button
                key={ch.id || i}
                role="tab"
                id={`tab-mobile-${i}`}
                aria-selected={activeChapter === i}
                aria-controls={`panel-${i}`}
                tabIndex={activeChapter === i ? 0 : -1}
                onClick={() => setActiveChapter(i)}
                onKeyDown={(e) => {
                  if (e.key === "ArrowRight" || e.key === "ArrowDown") setActiveChapter(Math.min(i + 1, doc.chapters.length - 1))
                  if (e.key === "ArrowLeft" || e.key === "ArrowUp") setActiveChapter(Math.max(i - 1, 0))
                }}
                title={ch.title}
                className={`shrink-0 rounded-full px-4 py-2 text-xs font-medium transition-colors whitespace-nowrap max-w-[200px] truncate focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20 ${activeChapter === i ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717] hover:bg-[#f5f5f5]"}`}
              >
                {ch.title}
              </button>
            ))}
          </div>
        )}

        {/* Desktop chapter sidebar — vertical list */}
        {hasChapters && (
          <aside className="hidden md:block md:w-64 md:shrink-0 sticky top-20">
            <div
              role="tablist"
              aria-orientation="vertical"
              aria-label="Paper sections"
              className="flex flex-col gap-1 rounded-[16px] border border-[#e5e5e5] bg-white p-3 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]"
            >
              <h2 className="px-3 py-1.5 font-satoshi text-xs font-semibold text-[#737373] uppercase tracking-wider">Chapters</h2>
              {doc.chapters.map((ch, i) => (
                <button
                  key={ch.id || i}
                  role="tab"
                  id={`tab-desktop-${i}`}
                  aria-selected={activeChapter === i}
                  aria-controls={`panel-${i}`}
                  tabIndex={activeChapter === i ? 0 : -1}
                  onClick={() => setActiveChapter(i)}
                  onKeyDown={(e) => {
                    if (e.key === "ArrowDown" || e.key === "ArrowRight") setActiveChapter(Math.min(i + 1, doc.chapters.length - 1))
                    if (e.key === "ArrowUp" || e.key === "ArrowLeft") setActiveChapter(Math.max(i - 1, 0))
                  }}
                  title={ch.title}
                  className={`w-full text-left rounded-lg px-3 py-2 text-xs font-medium transition-colors truncate focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20 ${activeChapter === i ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717] hover:bg-[#f5f5f5]"}`}
                >
                  {ch.title}
                </button>
              ))}
            </div>
          </aside>
        )}

        {/* Chapter content panel */}
        <div role="tabpanel" id={`panel-${activeChapter}`} aria-labelledby={hasChapters ? `tab-desktop-${activeChapter}` : undefined} tabIndex={0} className="flex-1 min-w-0">
          {activeChapterData && <p className="mb-4 text-sm text-[#737373] italic">{activeChapterData.summary}</p>}

          {showOriginal || hasChapters ? (
            <article className="rounded-[16px] border border-[#e5e5e5] bg-white p-6 sm:p-8 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]">
              <div className="prose prose-neutral max-w-none break-words text-[#171717] text-base leading-relaxed font-inter">
                <ReactMarkdown>{showOriginal ? displayedText : chapterContent}</ReactMarkdown>
              </div>
            </article>
          ) : (() => {
            const sections = parseResearchSections(displayedText)
            if (!sections) {
              return (
                <article className="rounded-[16px] border border-[#e5e5e5] bg-white p-6 sm:p-8 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]">
                  <div className="prose prose-neutral max-w-none break-words text-[#171717] text-base leading-relaxed font-inter">
                    <ReactMarkdown>{displayedText}</ReactMarkdown>
                  </div>
                </article>
              )
            }
            return (
              <div className="flex flex-col gap-4">
                {sections.map((section) => (
                  <article key={section.title} className="rounded-[16px] border border-[#e5e5e5] bg-white p-6 sm:p-8 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]">
                    <h3 className="font-satoshi text-base font-medium text-[#0a0a0a] mb-3">{section.title}</h3>
                    <div className="prose prose-neutral max-w-none break-words text-sm text-[#404040] leading-relaxed font-inter">
                      <ReactMarkdown>{section.content}</ReactMarkdown>
                    </div>
                  </article>
                ))}
              </div>
            )
          })()}

          {/* Charts for this chapter — only in tabbed mode */}
          {hasChapters && chapterCharts.length > 0 && (
            <section className="mt-8">
              <div className="flex items-center gap-2 mb-4">
                <BarChart2 className="h-4 w-4 text-[#2563eb]" aria-hidden="true" />
                <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">{activeChapterData ? `Charts — ${activeChapterData.title}` : "Charts & Visualizations"}</h2>
              </div>
              <div className="flex flex-col gap-4">
                {chapterCharts.map((chart) => (
                  <ChartCard key={chart.id} chart={chart} chapterTitle={activeChapterData?.title} evidence={evidenceFor(chart.id)} documentId={documentId} />
                ))}
              </div>
            </section>
          )}
        </div>
      </div>
    </>
  )
}
