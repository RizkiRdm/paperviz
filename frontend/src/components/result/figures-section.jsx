import { BarChart2 } from "lucide-react"
import { ChartCard } from "@/components/chart-card"

// Map backend failure categories to user-facing copy — never imply trust.
function failureCopy(category) {
  switch (category) {
    case "EXTRACTION_ERROR": return "We couldn't extract chart data from the source."
    case "DATASET_ERROR": return "Chart source data was incomplete."
    case "CHART_SELECTION_ERROR": return "No suitable chart type for the extracted data."
    case "GROUNDING_ERROR": return "Chart failed grounding checks — not rendered to avoid misleading you."
    case "SCHEMA_ERROR": return "Chart specification was invalid."
    case "RENDER_ERROR": return "Chart couldn't be rendered."
    default: return null
  }
}

// Render figures section — chapter-aware ChartCard map with grounding badges and failure states.
export function FiguresSection({ charts = [], evidence = [], chapters = [], activeChapter = -1, documentId, chartExtractionDegraded = false, failureCategory }) {
  const hasChapters = chapters && chapters.length > 1
  const activeChapterData = hasChapters ? chapters[activeChapter] : null
  const evidenceFor = (chartId) => (evidence || []).filter((e) => e.figure_id === chartId)
  const degradedMsg = failureCopy(failureCategory)

  // Chapter mode: filter by chapter_id
  if (hasChapters) {
    const chapterCharts = (charts || []).filter((c) => c.chapter_id === activeChapterData?.id)
    return (
      <section className="mt-8">
        <div className="flex items-center gap-2 mb-4">
          <BarChart2 className="h-4 w-4 text-[#2563eb]" aria-hidden="true" />
          <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">{activeChapterData ? `Charts — ${activeChapterData.title}` : "Charts & Visualizations"}</h2>
          {chapterCharts.length > 0 && <span className="inline-flex items-center rounded-full bg-[#f5f5f5] px-2 py-0.5 text-[11px] font-medium text-[#737373]">{chapterCharts.length}</span>}
        </div>
        {chapterCharts.length > 0 ? (
          <div className="flex flex-col gap-4">
            {chapterCharts.map((chart) => (
              <ChartCard key={chart.id} chart={chart} chapterTitle={activeChapterData?.title} evidence={evidenceFor(chart.id)} documentId={documentId} />
            ))}
          </div>
        ) : chartExtractionDegraded || degradedMsg ? (
          <div className="rounded-[12px] border border-[#fef3c7] bg-[#fffbeb] p-5 text-center">
            <p className="text-xs text-[#92400e]">{degradedMsg || "We couldn't extract charts for this section. The summary is still available."}</p>
            <p className="mt-1 text-[11px] text-[#737373]">Unsupported charts are never rendered as trusted data.</p>
          </div>
        ) : (
          <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
            <p className="text-xs text-[#737373]">No charts for this chapter.</p>
          </div>
        )}
      </section>
    )
  }

  // Flat mode: all charts or empty/degraded
  if (charts && charts.length > 0) {
    return (
      <section className="mt-8">
        <div className="flex items-center gap-2 mb-4">
          <BarChart2 className="h-4 w-4 text-[#2563eb]" aria-hidden="true" />
          <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">Charts & Visualizations</h2>
        </div>
        <div className="flex flex-col gap-4">
          {charts.map((chart) => (
            <ChartCard key={chart.id} chart={chart} evidence={evidenceFor(chart.id)} documentId={documentId} />
          ))}
        </div>
      </section>
    )
  }

  if (chartExtractionDegraded || degradedMsg) {
    return (
      <section className="mt-8">
        <div className="flex items-center gap-2 mb-4">
          <BarChart2 className="h-4 w-4 text-[#2563eb]" aria-hidden="true" />
          <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">Charts & Visualizations</h2>
        </div>
        <div className="rounded-[12px] border border-[#fef3c7] bg-[#fffbeb] p-5 text-center">
          <p className="text-xs text-[#92400e]">{degradedMsg || "We couldn't extract charts from this paper. The summary is still available."}</p>
          <p className="mt-1 text-[11px] text-[#737373]">Unsupported charts are never rendered as trusted data.</p>
        </div>
      </section>
    )
  }

  return (
    <section className="mt-8">
      <div className="flex items-center gap-2 mb-4">
        <BarChart2 className="h-4 w-4 text-[#2563eb]" aria-hidden="true" />
        <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">Charts & Visualizations</h2>
      </div>
      <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
        <p className="text-xs text-[#737373]">No charts were detected in this paper.</p>
      </div>
    </section>
  )
}
