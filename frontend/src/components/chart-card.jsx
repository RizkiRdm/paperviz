// ponytail: chart card redesign per DESIGN.md (hairline border, white canvas surface)
import { lazy, Suspense } from "react"
import { BarChart3, Image as ImageIcon } from "lucide-react"

const LazyDataChart = lazy(() => import("./data-chart"))

function ImageFallbackChart({ annotation, pageNumber }) {
  return (
    <div className="flex gap-4 items-start p-1">
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#f5f5f5] text-[#737373]">
        <ImageIcon className="h-5 w-5" />
      </div>
      <div>
        <h3 className="text-sm font-semibold text-[#0a0a0a]">Original Chart (Page {pageNumber})</h3>
        <p className="mt-1 text-xs text-[#737373] leading-relaxed">{annotation}</p>
      </div>
    </div>
  )
}

function OmittedChart({ annotation }) {
  return <p className="text-xs text-[#737373] italic">{annotation}</p>
}

export function ChartCard({ chart }) {
  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-5 transition-all hover:border-[#d4d4d4]">
      {chart.source_method === "data_extracted" && (
        <Suspense
          fallback={<p className="text-xs text-[#737373]">Loading interactive chart…</p>}
        >
          <LazyDataChart chartData={chart.chart_data} title={`Chart from Page ${chart.page_number}`} />
        </Suspense>
      )}
      {chart.source_method === "image_fallback" && (
        <ImageFallbackChart annotation={chart.annotation} pageNumber={chart.page_number} />
      )}
      {chart.source_method === "omitted" && <OmittedChart annotation={chart.annotation} />}
    </div>
  )
}

