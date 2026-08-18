import { useEffect, useState } from "react"
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip"

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

  const type = chartData.chart_type || "bar"
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
        <div className="flex items-center gap-2">
          {chartData.confidence && chartData.confidence !== "high" && (
            <Tooltip>
              <TooltipTrigger asChild>
                <span className={`rounded-full px-2 py-0.5 text-[10px] font-medium ${
                  chartData.confidence === "low"
                    ? "bg-[#fef3c7] text-[#92400e]"
                    : "bg-[#f5f5f5] text-[#737373]"
                }`}>
                  {chartData.confidence === "low" ? "Low Confidence" : "Interpreted"}
                </span>
              </TooltipTrigger>
              <TooltipContent>
                {chartData.confidence === "low"
                  ? "Low confidence — this figure may not be reliably interpretable from the original."
                  : "Interpreted from the original figure"}
              </TooltipContent>
            </Tooltip>
          )}
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="rounded-full bg-[#dbeaff] px-2.5 py-0.5 text-[11px] font-medium text-[#2563eb]">
                PaperViz AI Interpretation
              </span>
            </TooltipTrigger>
            <TooltipContent>AI-generated interpretation of the original figure</TooltipContent>
          </Tooltip>
        </div>
      </div>
      <div className="h-60 w-full pt-2">
        <recharts.ResponsiveContainer width="100%" height="100%">
          {renderChart()}
        </recharts.ResponsiveContainer>
      </div>
      {(chartData.x_axis || chartData.y_axis) && (
        <p className="mt-2 text-[10px] text-[#737373] text-center">
          {chartData.x_axis && chartData.y_axis
            ? `${chartData.y_axis} by ${chartData.x_axis}`
            : chartData.y_axis || chartData.x_axis}
        </p>
      )}
      {chartData.key_takeaway && (
        <div className="mt-3 rounded-lg bg-[#f0fdf4] border border-[#bbf7d0] px-3 py-2">
          <p className="text-xs text-[#166534] leading-relaxed">{chartData.key_takeaway}</p>
        </div>
      )}
      {chartData.limitations && (
        <p className="mt-2 text-[11px] text-[#737373] italic">{chartData.limitations}</p>
      )}
    </div>
  )
}
