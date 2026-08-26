// Usage display — shows current tier, papers used, limit, and a color-coded progress bar.
import { useState, useEffect } from "react"

export default function UsageDisplay() {
  const [usage, setUsage] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    async function fetchUsage() {
      try {
        const res = await fetch("/api/usage")
        if (!res.ok) throw new Error("Failed to fetch usage")
        const data = await res.json()
        setUsage(data)
      } catch (err) {
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }
    fetchUsage()
  }, [])

  if (loading) {
    return (
      <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-4">
        <div className="flex items-center gap-2">
          <div className="h-4 w-4 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent" />
          <span className="text-xs text-[#737373]">Loading usage...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-4">
        <p className="text-xs text-[#ea580c]">Could not load usage data.</p>
      </div>
    )
  }

  if (!usage) return null

  const { tier, papers_used, limit } = usage
  const pct = limit > 0 ? Math.min((papers_used / limit) * 100, 100) : 0
  const barColor =
    pct >= 100 ? "bg-[#ea580c]" : pct >= 80 ? "bg-[#eab308]" : "bg-[#16a34a]"

  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-semibold text-[#0a0a0a]">
          {tier} Plan
        </span>
        <span className="text-[11px] text-[#737373]">
          {papers_used} / {limit} papers
        </span>
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-[#f5f5f5]">
        <div
          className={`h-full rounded-full transition-all ${barColor}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  )
}
