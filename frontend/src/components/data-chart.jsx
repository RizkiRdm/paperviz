import { useEffect, useState } from "react"

const CHART_COLORS = ["#2563eb", "#16a34a", "#ea580c", "#7c3aed", "#0a0a0a", "#737373"]

export default function DataChart({ chartData, title }) {
  const [recharts, setRecharts] = useState(null)
  const [loadFailed, setLoadFailed] = useState(false)

  useEffect(() => {
    let disposed = false
    import("recharts")
      .then((mod) => { if (!disposed) setRecharts(mod) })
      .catch(() => { if (!disposed) setLoadFailed(true) })
    return () => { disposed = true }
  }, [])

  if (loadFailed) return <p className="text-xs text-[#dc2626]">Chart could not be loaded.</p>
  if (!recharts) return <p className="text-xs text-[#737373]">Loading chart component…</p>

  const type = chartData.type || "bar"
  const rows = (chartData.labels || []).map((label, i) => ({
    name: label,
    value: chartData.values?.[i] ?? 0,
  }))

  const gridAndAxes = (
    <>
      <recharts.CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" vertical={false} />
      <recharts.XAxis
        dataKey="name"
        tick={{ fontSize: 11, fill: "#737373" }}
        axisLine={{ stroke: "#e5e5e5" }}
        tickLine={false}
      />
      <recharts.YAxis
        tick={{ fontSize: 11, fill: "#737373" }}
        axisLine={false}
        tickLine={false}
      />
      <recharts.Tooltip
        contentStyle={{
          backgroundColor: "#ffffff",
          borderColor: "#e5e5e5",
          borderRadius: "8px",
          boxShadow: "rgba(0, 0, 0, 0.05) 0px 4px 6px -1px",
          fontSize: "12px",
          color: "#171717",
        }}
      />
    </>
  )

  function renderChart() {
    switch (type) {
      case "line":
        return (
          <recharts.LineChart data={rows} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            {gridAndAxes}
            <recharts.Line
              type="monotone"
              dataKey="value"
              stroke="#2563eb"
              strokeWidth={2}
              dot={{ fill: "#2563eb", r: 4 }}
              activeDot={{ r: 6 }}
            />
          </recharts.LineChart>
        )
      case "pie":
        return (
          <recharts.PieChart>
            <recharts.Pie
              data={rows}
              dataKey="value"
              nameKey="name"
              cx="50%"
              cy="50%"
              outerRadius={80}
              label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
              labelLine={false}
            >
              {rows.map((_, i) => (
                <recharts.Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
              ))}
            </recharts.Pie>
            <recharts.Tooltip
              contentStyle={{
                backgroundColor: "#ffffff",
                borderColor: "#e5e5e5",
                borderRadius: "8px",
                boxShadow: "rgba(0, 0, 0, 0.05) 0px 4px 6px -1px",
                fontSize: "12px",
                color: "#171717",
              }}
            />
          </recharts.PieChart>
        )
      case "scatter":
        return (
          <recharts.ScatterChart margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            <recharts.CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" />
            <recharts.XAxis
              type="category"
              dataKey="name"
              tick={{ fontSize: 11, fill: "#737373" }}
              axisLine={{ stroke: "#e5e5e5" }}
              tickLine={false}
            />
            <recharts.YAxis
              type="number"
              dataKey="value"
              tick={{ fontSize: 11, fill: "#737373" }}
              axisLine={false}
              tickLine={false}
            />
            <recharts.Tooltip
              contentStyle={{
                backgroundColor: "#ffffff",
                borderColor: "#e5e5e5",
                borderRadius: "8px",
                boxShadow: "rgba(0, 0, 0, 0.05) 0px 4px 6px -1px",
                fontSize: "12px",
                color: "#171717",
              }}
            />
            <recharts.Scatter data={rows} fill="#2563eb" />
          </recharts.ScatterChart>
        )
      default: // bar
        return (
          <recharts.BarChart data={rows} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            {gridAndAxes}
            <recharts.Bar dataKey="value" fill="#2563eb" radius={[6, 6, 0, 0]} barSize={32} />
          </recharts.BarChart>
        )
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-semibold text-[#0a0a0a]">{chartData.title || title}</h3>
        <span className="rounded-full bg-[#dbeaff] px-2.5 py-0.5 text-[11px] font-medium text-[#2563eb]">
          PaperViz AI Interpretation
        </span>
      </div>
      <div className="h-60 w-full pt-2">
        <recharts.ResponsiveContainer width="100%" height="100%">
          {renderChart()}
        </recharts.ResponsiveContainer>
      </div>
    </div>
  )
}
