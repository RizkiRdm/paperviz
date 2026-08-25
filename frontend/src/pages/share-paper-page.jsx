// SharePaperPage — public, token-based view of a simplified paper summary
// plus its charts (Chunk 4.2). Mirrors ShareFigurePage's flow: fetch by
// token, centered 404 on expiry, quiet editorial layout per DESIGN.md.
import { useEffect, useState } from "react"
import { useParams, Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { ArrowLeft, ExternalLink, BarChart2 } from "lucide-react"
import ReactMarkdown from "react-markdown"

const READING_LEVEL_LABELS = {
  simplified: "Simplified",
  eli5: "ELI5",
}

// Parse structured summary sections delimited by "## " headers — local
// copy of result-page's parseResearchSections so this page stays
// self-contained (no cross-imports between entry pages).
function parseResearchSections(text) {
  if (!text) return null
  const parts = text.split(/^## /m).filter(Boolean)
  if (parts.length < 2) return null
  return parts.map((part) => {
    const lines = part.split("\n")
    const title = lines[0].trim()
    const content = lines.slice(1).join("\n").trim()
    return { title, content }
  })
}

export function SharePaperPage() {
  const { shareToken } = useParams()
  const [paper, setPaper] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // fetchPaper loads the public share payload once per token; 404 maps to
  // not_found, everything else to network.
  useEffect(() => {
    let cancelled = false
    async function fetchPaper() {
      try {
        const res = await fetch(`/share/doc/${shareToken}`)
        if (!res.ok) {
          if (!cancelled) setError("not_found")
          return
        }
        const data = await res.json()
        if (!cancelled) setPaper(data)
      } catch {
        if (!cancelled) setError("network")
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    fetchPaper()
    return () => { cancelled = true }
  }, [shareToken])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <p className="text-sm text-[#737373]">Loading paper…</p>
      </div>
    )
  }

  if (error === "network") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="text-center">
          <h1 className="text-4xl font-semibold text-[#0a0a0a]">Offline</h1>
          <p className="mt-2 text-sm text-[#737373]">
            Something went wrong loading this summary. Check your connection and try again.
          </p>
          <Link
            to={`/?ref=${shareToken}`}
            className="mt-6 inline-block text-sm text-[#2563eb] hover:underline"
          >
            Analyze your own paper
          </Link>
        </div>
      </div>
    )
  }

  if (error || !paper) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="text-center">
          <h1 className="text-4xl font-semibold text-[#0a0a0a]">404</h1>
          <p className="mt-2 text-sm text-[#737373]">
            Summary not found or link expired
          </p>
          <Link
            to={`/?ref=${shareToken}`}
            className="mt-6 inline-block text-sm text-[#2563eb] hover:underline"
          >
            Analyze your own paper
          </Link>
        </div>
      </div>
    )
  }

  const sections = parseResearchSections(paper.simplified_text)
  const levelLabel = READING_LEVEL_LABELS[paper.reading_level] || paper.reading_level
  const charts = Array.isArray(paper.charts) ? paper.charts : []

  return (
    <div className="min-h-screen bg-white">
      <header className="border-b border-[#e5e5e5]">
        <div className="mx-auto flex max-w-3xl items-center gap-3 px-4 py-3">
          <Link
            to={`/?ref=${shareToken}`}
            className="inline-flex items-center gap-1.5 text-xs font-medium text-[#737373] hover:text-[#0a0a0a] transition-colors"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            PaperViz
          </Link>
          <span className="text-[#e5e5e5]">·</span>
          <span className="text-xs text-[#737373]">Shared paper</span>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-4 py-8">
        <div className="mb-8">
          <h1 className="font-satoshi text-xl sm:text-2xl font-medium leading-tight text-[#0a0a0a]">
            {paper.title || "Research Paper Summary"}
          </h1>
          <div className="mt-3 flex flex-wrap items-center gap-2">
            <span className="inline-flex items-center rounded-full bg-[#dbeaff] px-2.5 py-0.5 text-[11px] font-medium text-[#2563eb]">
              {levelLabel}
            </span>
            {charts.length > 0 && (
              <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
                {charts.length} {charts.length === 1 ? "chart" : "charts"}
              </span>
            )}
          </div>
          <p className="mt-3 text-xs leading-relaxed text-[#737373]">
            Simplified interpretation generated by PaperViz — see the original
            paper for the research itself.
          </p>
        </div>

        {sections ? (
          <div className="flex flex-col gap-4">
            {sections.map((section) => (
              <article key={section.title} className="rounded-[16px] border border-[#e5e5e5] bg-white p-6 sm:p-8 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]">
                <h2 className="font-satoshi text-base font-medium text-[#0a0a0a] mb-3">
                  {section.title}
                </h2>
                <div className="prose prose-neutral max-w-none break-words text-sm text-[#404040] leading-relaxed font-inter">
                  <ReactMarkdown>{section.content}</ReactMarkdown>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <article className="rounded-[16px] border border-[#e5e5e5] bg-white p-6 sm:p-8 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]">
            <div className="prose prose-neutral max-w-none break-words text-sm text-[#404040] leading-relaxed font-inter">
              <ReactMarkdown>{paper.simplified_text}</ReactMarkdown>
            </div>
          </article>
        )}

        <section className="mt-10">
          <div className="flex items-center gap-2 mb-4">
            <BarChart2 className="h-4 w-4 text-[#2563eb]" />
            <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">
              Charts & Visualizations
            </h2>
          </div>

          {charts.length > 0 ? (
            <div className="flex flex-col gap-4">
              {charts.map((chart, i) => (
                <article key={chart.chart_id || i} className="rounded-[12px] border border-[#e5e5e5] bg-white p-4">
                  {chart.image_url ? (
                    <img
                      src={chart.image_url}
                      alt={chart.annotation || "Re-visualized chart from the paper"}
                      className="mb-3 h-auto w-full rounded-[8px] border border-[#e5e5e5] bg-white"
                    />
                  ) : chart.chart_data ? (
                    <p className="mb-3 rounded-[8px] border border-dashed border-[#d4d4d4] bg-[#f5f5f5] px-3 py-6 text-center text-xs text-[#737373]">
                      Interactive chart available in the full document view.
                    </p>
                  ) : null}
                  {chart.annotation && (
                    <p className="text-sm leading-relaxed text-[#171717]">
                      {chart.annotation}
                    </p>
                  )}
                  <div className="mt-3 flex flex-wrap items-center gap-2">
                    {chart.source_method && (
                      <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
                        {chart.source_method === "data_extracted" ? "Data-extracted" : "Image capture"}
                      </span>
                    )}
                    {chart.page_number > 0 && (
                      <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
                        Page {chart.page_number}
                      </span>
                    )}
                  </div>
                </article>
              ))}
            </div>
          ) : (
            <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
              <p className="text-xs text-[#737373]">No charts were detected in this paper.</p>
            </div>
          )}
        </section>

        <div className="mt-10 rounded-[12px] border border-[#e5e5e5] bg-[#fafafa] p-6 text-center">
          <p className="text-sm text-[#525252]">
            This summary was generated by PaperViz from the original paper.
          </p>
          <Button asChild className="mt-4" size="sm">
            <Link to={`/?ref=${shareToken}`}>
              Analyze your own paper
              <ExternalLink className="ml-1.5 h-3.5 w-3.5" />
            </Link>
          </Button>
        </div>
      </main>
    </div>
  )
}
