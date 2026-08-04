// ponytail: recharts styling updated with Dub Electric Blue (#2563eb) and hairline grid
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
    return <p className="text-xs text-[#dc2626]">Chart could not be loaded.</p>
  }

  if (!recharts) {
    return <p className="text-xs text-[#737373]">Loading chart component…</p>
  }

  const { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } = recharts
  const rows = (chartData.labels || []).map((label, i) => ({
    label,
    value: chartData.values?.[i] ?? 0,
  }))

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-semibold text-[#0a0a0a]">{chartData.title || title}</h3>
        <span className="rounded-full bg-[#dbeaff] px-2.5 py-0.5 text-[11px] font-medium text-[#2563eb]">
          Re-visualized
        </span>
      </div>

      <div className="h-60 w-full pt-2">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={rows} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" vertical={false} />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 11, fill: "#737373" }}
              axisLine={{ stroke: "#e5e5e5" }}
              tickLine={false}
            />
            <YAxis
              tick={{ fontSize: 11, fill: "#737373" }}
              axisLine={false}
              tickLine={false}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "#ffffff",
                borderColor: "#e5e5e5",
                borderRadius: "8px",
                boxShadow: "rgba(0, 0, 0, 0.05) 0px 4px 6px -1px",
                fontSize: "12px",
                color: "#171717",
              }}
            />
            <Bar dataKey="value" fill="#2563eb" radius={[6, 6, 0, 0]} barSize={32} />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  )
}

