// ChartCard — DESIGN.md "Cards & Dropzone > Chart Card": radius-lg,
// shadow-card, surface-raised fill, space-4 padding.
//
// Renders one of three states per chart, matching ARCHITECTURE.md Section 6
// Failure Scenarios 3-4 exactly:
//   - data_extracted: re-rendered as a Recharts bar chart (the "improved
//     chart" from PRD.md Success Metrics).
//   - image_fallback: the original chart image, with a generated
//     plain-language annotation alongside it.
//   - omitted: an inline note only — "rest of document unaffected."
import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts"

function DataChart({ chartData, title }) {
  let parsed
  try {
    parsed = JSON.parse(chartData)
  } catch {
    return null // malformed chart_data shouldn't crash the result page
  }

  const rows = (parsed.labels || []).map((label, i) => ({
    label,
    value: parsed.values?.[i] ?? 0,
  }))

  return (
    <div>
      <h3 className="text-h3 text-ink-primary">{parsed.title || title}</h3>
      <div className="mt-3 h-64">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={rows}>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border-default)" />
            <XAxis dataKey="label" tick={{ fontSize: 12, fill: "var(--color-ink-secondary)" }} />
            <YAxis tick={{ fontSize: 12, fill: "var(--color-ink-secondary)" }} />
            <Tooltip />
            <Bar dataKey="value" fill="var(--color-accent-verified)" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}

function ImageFallbackChart({ annotation, pageNumber }) {
  // Note: the backend does not currently expose a raw image download
  // endpoint (image_blob is stored but not served over HTTP in this MVP —
  // see ARCHITECTURE.md's Get Document contract, which returns chart_data
  // and annotation but no image URL). Until that endpoint exists, the
  // image-fallback case shows the annotation with a page reference instead
  // of a broken <img>, which is still useful and never misleading.
  return (
    <div>
      <h3 className="text-h3 text-ink-primary">Chart from page {pageNumber}</h3>
      <p className="mt-2 text-body text-ink-secondary">{annotation}</p>
    </div>
  )
}

function OmittedChart({ annotation }) {
  return <p className="text-body text-ink-secondary">{annotation}</p>
}

export function ChartCard({ chart }) {
  return (
    <div className="rounded-lg bg-surface-raised p-4 shadow-[0_1px_2px_rgba(20,23,31,0.04),0_4px_12px_rgba(20,23,31,0.06)]">
      {chart.source_method === "data_extracted" && (
        <DataChart chartData={chart.chart_data} title={`Chart, page ${chart.page_number}`} />
      )}
      {chart.source_method === "image_fallback" && (
        <ImageFallbackChart annotation={chart.annotation} pageNumber={chart.page_number} />
      )}
      {chart.source_method === "omitted" && <OmittedChart annotation={chart.annotation} />}
    </div>
  )
}
