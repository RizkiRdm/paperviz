import { useState } from "react"
import { ChevronDown, ChevronRight, ExternalLink } from "lucide-react"
import { Link } from "react-router-dom"

const STANCE_COLORS = {
  supporting: {
    bg: "bg-[#dcfce7]",
    text: "text-[#16a34a]",
    border: "border-[#bbf7d0]",
    label: "Supports",
  },
  contradicting: {
    bg: "bg-[#fff7ed]",
    text: "text-[#ea580c]",
    border: "border-[#fed7aa]",
    label: "Contradicts",
  },
  unclear: {
    bg: "bg-[#f5f5f5]",
    text: "text-[#737373]",
    border: "border-[#e5e5e5]",
    label: "Unclear",
  },
}

export function EvidenceComparisonPanel({ evidenceClaims, papers }) {
  const [expandedClaim, setExpandedClaim] = useState(null)

  if (!evidenceClaims || evidenceClaims.length === 0) {
    return null
  }

  return (
    <section className="mb-10">
      <h2 className="text-sm font-medium text-[#0a0a0a] mb-4">Evidence Claims</h2>
      <p className="text-xs text-[#737373] mb-4">
        Specific claims found across papers, with per-paper stance.
      </p>
      <div className="rounded-[12px] border border-[#e5e5e5] bg-white overflow-hidden">
        {evidenceClaims.map((claim, i) => (
          <div
            key={i}
            className={`${i > 0 ? "border-t border-[#e5e5e5]" : ""}`}
          >
            <button
              onClick={() => setExpandedClaim(expandedClaim === i ? null : i)}
              className="w-full px-5 py-4 text-left flex items-start justify-between gap-3 hover:bg-[#fafafa] transition-colors"
            >
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-[#0a0a0a]">
                  {claim.claim}
                </p>
                <div className="flex flex-wrap gap-1.5 mt-2">
                  {papers.map((paper) => {
                    const stance = claim.stances?.[paper.document_id] || "unclear"
                    const colors = STANCE_COLORS[stance] || STANCE_COLORS.unclear
                    return (
                      <span
                        key={paper.document_id}
                        className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium ${colors.bg} ${colors.text} border ${colors.border}`}
                      >
                        {paper.title?.slice(0, 20) || "Untitled"}: {colors.label}
                      </span>
                    )
                  })}
                </div>
              </div>
              {expandedClaim === i ? (
                <ChevronDown className="h-4 w-4 text-[#a3a3a3] shrink-0 mt-0.5" />
              ) : (
                <ChevronRight className="h-4 w-4 text-[#a3a3a3] shrink-0 mt-0.5" />
              )}
            </button>

            {expandedClaim === i && (
              <div className="px-5 pb-4 border-t border-[#f5f5f5]">
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 mt-3">
                  {papers.map((paper) => {
                    const stance = claim.stances?.[paper.document_id] || "unclear"
                    const sourceRef = claim.source_refs?.[paper.document_id]
                    const colors = STANCE_COLORS[stance] || STANCE_COLORS.unclear
                    return (
                      <div
                        key={paper.document_id}
                        className={`rounded-[8px] border p-3 ${colors.border} ${colors.bg}`}
                      >
                        <div className="flex items-center justify-between gap-2 mb-2">
                          <Link
                            to={`/${paper.document_id}`}
                            className="text-[11px] font-medium text-[#2563eb] hover:underline truncate"
                          >
                            {paper.title || "Untitled"}
                          </Link>
                          <ExternalLink className="h-3 w-3 text-[#a3a3a3] shrink-0" />
                        </div>
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium ${colors.text}`}>
                          {colors.label}
                        </span>
                        {sourceRef && (
                          <p className="mt-2 text-xs text-[#525252] italic leading-relaxed">
                            &ldquo;{sourceRef}&rdquo;
                          </p>
                        )}
                      </div>
                    )
                  })}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </section>
  )
}
