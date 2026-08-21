import { useState, useEffect } from "react"
import { useSearchParams, Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { ArrowRight, FileText, Calendar, ExternalLink } from "lucide-react"
import { EvidenceComparisonPanel } from "@/components/evidence-comparison"

const DIMENSION_LABELS = {
  research_question: "Research Question",
  methodology: "Methodology",
  dataset: "Dataset",
  sample_size: "Sample Size",
  findings: "Findings",
  limitations: "Limitations",
  evidence: "Evidence",
  conclusions: "Conclusions",
}

export function ComparePage() {
  const [searchParams] = useSearchParams()
  const [comparison, setComparison] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const documentIds = searchParams.get("ids")?.split(",") || []

  useEffect(() => {
    if (documentIds.length < 2) {
      setError("Select at least 2 papers to compare")
      setLoading(false)
      return
    }

    async function fetchComparison() {
      try {
        const response = await fetch("/api/documents/compare", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ document_ids: documentIds }),
        })

        if (!response.ok) {
          const data = await response.json()
          throw new Error(data.error || "Failed to compare papers")
        }

        const data = await response.json()
        setComparison(data)
      } catch (err) {
        setError(err.message)
      } finally {
        setLoading(false)
      }
    }

    fetchComparison()
  }, [documentIds.join(",")])

  if (loading) {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center">
        <div className="text-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent mx-auto mb-4" />
          <p className="text-sm text-[#737373]">Comparing papers...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center">
        <div className="text-center">
          <FileText className="h-10 w-10 text-[#a3a3a3] mx-auto mb-4" />
          <p className="text-sm text-[#0a0a0a] mb-4">{error}</p>
          <Link to="/dashboard">
            <Button variant="secondary" className="gap-1.5">
              Back to Dashboard <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white bg-dotted-grid">
      <header className="border-b border-[#e5e5e5] bg-white/80 backdrop-blur-xs">
        <div className="mx-auto flex max-w-[1200px] items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold text-[#0a0a0a]">PaperViz</span>
          </div>
          <Link to="/dashboard">
            <Button variant="secondary" className="h-8 px-3 text-xs gap-1.5">
              Dashboard
            </Button>
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-[1200px] px-6 py-10">
        <div className="mb-8">
          <h1 className="font-satoshi text-2xl font-medium text-[#0a0a0a] mb-2">
            Paper Comparison
          </h1>
          <p className="text-sm text-[#737373]">
            Comparing {comparison.papers.length} papers side by side.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-10">
          {comparison.papers.map((paper) => (
            <div
              key={paper.document_id}
              className="rounded-[12px] border border-[#e5e5e5] bg-white p-5"
            >
              <div className="flex items-start justify-between gap-2 mb-3">
                <h3 className="text-sm font-medium text-[#0a0a0a] line-clamp-2">
                  {paper.title || "Untitled paper"}
                </h3>
                <Link
                  to={`/${paper.document_id}`}
                  className="shrink-0 text-[#a3a3a3] hover:text-[#2563eb]"
                >
                  <ExternalLink className="h-4 w-4" />
                </Link>
              </div>
              <p className="text-xs text-[#737373] line-clamp-3 mb-3">
                {paper.research_question}
              </p>
              <div className="flex items-center gap-2 text-[11px] text-[#a3a3a3]">
                <Calendar className="h-3 w-3" />
                <span>{paper.methodology}</span>
              </div>
            </div>
          ))}
        </div>

        {comparison.agreement.length > 0 && (
          <section className="mb-10">
            <h2 className="text-sm font-medium text-[#16a34a] mb-3">Areas of Agreement</h2>
            <div className="flex flex-wrap gap-2">
              {comparison.agreement.map((item, i) => (
                <span
                  key={i}
                  className="inline-flex items-center rounded-full bg-[#dcfce7] px-3 py-1 text-xs text-[#16a34a]"
                >
                  {item}
                </span>
              ))}
            </div>
          </section>
        )}

        {comparison.disagreement.length > 0 && (
          <section className="mb-10">
            <h2 className="text-sm font-medium text-[#ea580c] mb-3">Areas of Disagreement</h2>
            <div className="flex flex-wrap gap-2">
              {comparison.disagreement.map((item, i) => (
                <span
                  key={i}
                  className="inline-flex items-center rounded-full bg-[#fff7ed] px-3 py-1 text-xs text-[#ea580c]"
                >
                  {item}
                </span>
              ))}
            </div>
          </section>
        )}

        <EvidenceComparisonPanel
          evidenceClaims={comparison.evidence_claims}
          papers={comparison.papers}
        />

        <section className="mb-10">
          <h2 className="text-sm font-medium text-[#0a0a0a] mb-4">Detailed Comparison</h2>
          <div className="rounded-[12px] border border-[#e5e5e5] bg-white overflow-hidden">
            {comparison.dimensions.map((dim, i) => (
              <div
                key={dim.dimension}
                className={`${i > 0 ? "border-t border-[#e5e5e5]" : ""}`}
              >
                <div className="px-5 py-4">
                  <h3 className="text-xs font-medium text-[#737373] mb-3 font-mono uppercase tracking-wider">
                    {DIMENSION_LABELS[dim.dimension] || dim.dimension}
                  </h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {comparison.papers.map((paper) => (
                      <div key={paper.document_id}>
                        <Link
                          to={`/${paper.document_id}`}
                          className="text-[11px] font-medium text-[#2563eb] hover:underline mb-1 block"
                        >
                          {paper.title || "Untitled"}
                        </Link>
                        <p className="text-sm text-[#0a0a0a]">
                          {dim.values[paper.document_id] || "\u2014"}
                        </p>
                      </div>
                    ))}
                  </div>
                  {dim.notes && (
                    <p className="mt-3 text-xs text-[#737373] italic border-t border-[#f5f5f5] pt-3">
                      {dim.notes}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </section>

        <div className="flex items-center gap-4">
          <Link to="/dashboard">
            <Button variant="secondary" className="gap-1.5">
              Back to Dashboard
            </Button>
          </Link>
        </div>
      </main>
    </div>
  )
}
