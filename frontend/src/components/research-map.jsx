import { useEffect, useState } from "react"
import { Network, ChevronDown, ChevronRight, ArrowRight } from "lucide-react"

const RELATIONSHIP_TYPE_CONFIG = {
  supporting: {
    label: "Supporting",
    bg: "bg-[#dcfce7]",
    text: "text-[#16a34a]",
    border: "border-[#bbf7d0]",
    icon: "\u2191",
  },
  contradicting: {
    label: "Contradicting",
    bg: "bg-[#fff7ed]",
    text: "text-[#ea580c]",
    border: "border-[#fed7aa]",
    icon: "\u2193",
  },
  citing: {
    label: "Citing",
    bg: "bg-[#dbeaff]",
    text: "text-[#2563eb]",
    border: "border-[#bfdbfe]",
    icon: "\u2192",
  },
  similar_methodology: {
    label: "Similar Methodology",
    bg: "bg-[#f3e8ff]",
    text: "text-[#7c3aed]",
    border: "border-[#ddd6fe]",
    icon: "\u2248",
  },
  different_findings: {
    label: "Different Findings",
    bg: "bg-[#f5f5f5]",
    text: "text-[#737373]",
    border: "border-[#e5e5e5]",
    icon: "\u2260",
  },
}

const RELATIONSHIP_ORDER = [
  "supporting",
  "contradicting",
  "citing",
  "similar_methodology",
  "different_findings",
]

export function ResearchMap({ documentId }) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [expandedTypes, setExpandedTypes] = useState(new Set())

  useEffect(() => {
    let cancelled = false
    async function fetchResearchMap() {
      try {
        const res = await fetch(`/api/documents/${documentId}/research-map`)
        if (!res.ok) throw new Error("failed")
        const json = await res.json()
        if (!cancelled) {
          setData(json)
          setLoading(false)
        }
      } catch {
        if (!cancelled) {
          setError(true)
          setLoading(false)
        }
      }
    }
    fetchResearchMap()
    return () => { cancelled = true }
  }, [documentId])

  function toggleType(type) {
    setExpandedTypes((prev) => {
      const next = new Set(prev)
      if (next.has(type)) next.delete(type)
      else next.add(type)
      return next
    })
  }

  if (loading) {
    return (
      <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-6 text-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent mx-auto mb-3" />
        <p className="text-xs text-[#737373]">Loading research map...</p>
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
        <p className="text-xs text-[#737373]">Couldn't load the research map.</p>
      </div>
    )
  }

  if (data.total_count === 0) {
    return (
      <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
        <Network className="h-5 w-5 text-[#a3a3a3] mx-auto mb-2" />
        <p className="text-xs text-[#737373]">
          No paper relationships found. Add relationships to see how papers connect.
        </p>
      </div>
    )
  }

  const activeTypes = RELATIONSHIP_ORDER.filter(
    (type) => data.relationships[type]?.length > 0
  )

  return (
    <div className="space-y-3">
      {activeTypes.map((type) => {
        const config = RELATIONSHIP_TYPE_CONFIG[type]
        const items = data.relationships[type]
        const isExpanded = expandedTypes.has(type)

        return (
          <div
            key={type}
            className="rounded-[12px] border border-[#e5e5e5] bg-white overflow-hidden"
          >
            <button
              onClick={() => toggleType(type)}
              className="w-full px-5 py-3.5 text-left flex items-center justify-between gap-3 hover:bg-[#fafafa] transition-colors"
            >
              <div className="flex items-center gap-3">
                <span
                  className={`inline-flex items-center justify-center h-6 w-6 rounded-full text-xs font-medium ${config.bg} ${config.text}`}
                >
                  {config.icon}
                </span>
                <div>
                  <span className="text-sm font-medium text-[#0a0a0a]">
                    {config.label}
                  </span>
                  <span className="ml-2 text-xs text-[#737373]">
                    {items.length} {items.length === 1 ? "paper" : "papers"}
                  </span>
                </div>
              </div>
              {isExpanded ? (
                <ChevronDown className="h-4 w-4 text-[#a3a3a3] shrink-0" />
              ) : (
                <ChevronRight className="h-4 w-4 text-[#a3a3a3] shrink-0" />
              )}
            </button>

            {isExpanded && (
              <div className="px-5 pb-4 border-t border-[#f5f5f5]">
                <div className="flex flex-col gap-2 mt-3">
                  {items.map((rel) => (
                    <div
                      key={rel.id}
                      className={`rounded-[8px] border p-3 ${config.border} ${config.bg}`}
                    >
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-xs font-medium text-[#0a0a0a] truncate">
                          {rel.target_paper_title || "Untitled Paper"}
                        </span>
                        <ArrowRight className="h-3 w-3 text-[#a3a3a3] shrink-0" />
                        <span className={`text-[10px] font-medium ${config.text}`}>
                          {config.label}
                        </span>
                      </div>
                      {rel.evidence_text && (
                        <p className="text-xs text-[#525252] italic leading-relaxed mt-1">
                          &ldquo;{rel.evidence_text}&rdquo;
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
