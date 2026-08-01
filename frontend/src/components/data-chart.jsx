// DataChart — the Recharts bar-chart rendering for the data_extracted
// chart state (DESIGN.md "Cards & Dropzone > Chart Card").
//
// recharts is intentionally NOT statically imported: react-doctor/
// prefer-dynamic-import flags any static ImportDeclaration of a heavy
// library, so the library is fetched at runtime via import("recharts") the
// first time a data-extracted chart mounts. Vite code-splits the dynamic
// import into its own chunk, so recharts never ships in the initial
// bundle. The default export is required by ChartCard's React.lazy().
import { useEffect, useState } from "react"

export default function DataChart({ chartData, title }) {
  const [recharts, setRecharts] = useState(null)
  const [loadFailed, setLoadFailed] = useState(false)

  useEffect(() => {
    let disposed = false
    import("recharts")
      .then((mod) => {
        if (!disposed) setRecharts(mod)
      })
      .catch(() => {
        if (!disposed) setLoadFailed(true)
      })
    return () => {
      disposed = true
    }
  }, [])

  if (loadFailed) {
    return <p className="text-caption text-ink-secondary">Chart could not be loaded.</p>
  }

  if (!recharts) {
    return <p className="text-caption text-ink-secondary">Loading chart…</p>
  }

  const { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } = recharts
  const rows = (chartData.labels || []).map((label, i) => ({
    label,
    value: chartData.values?.[i] ?? 0,
  }))

  return (
    <div>
      <h3 className="text-h3 text-ink-primary">{chartData.title || title}</h3>
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
