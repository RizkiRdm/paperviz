import { useEffect, useState } from "react"
import { useParams, Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { ArrowLeft, ExternalLink } from "lucide-react"
import { authFetch } from "@/lib/auth"

const READING_LEVEL_LABELS = {
  simplified: "Simplified",
  eli5: "ELI5",
}

export function ShareFigurePage() {
  const { shareToken } = useParams()
  const [figure, setFigure] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    let cancelled = false
    async function fetchFigure() {
      try {
        const res = await authFetch(`/share/fig/${shareToken}`)
        if (!res.ok) {
          if (!cancelled) setError("not_found")
          return
        }
        const data = await res.json()
        if (!cancelled) setFigure(data)
      } catch (err) {
        console.error("Share figure fetch failed", err)
        if (!cancelled) setError("network")
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    fetchFigure()
    return () => { cancelled = true }
  }, [shareToken])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <p className="text-sm text-[#737373]">Loading figure…</p>
      </div>
    )
  }

  // Network failure gets its own state — a connection error must not read as
  // an expired link (mirrors SharePaperPage's offline branch).
  if (error === "network") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white px-4">
        <div className="max-w-md w-full text-center py-16">
          <span className="font-mono text-xs font-semibold uppercase tracking-wider text-[#ea580c] bg-[#fff7ed] px-2.5 py-1 rounded-[4px] border border-[#ffedd5]">
            Connection Error
          </span>
          <h1 className="mt-4 font-display text-3xl sm:text-4xl font-medium tracking-tight text-[#0a0a0a]">
            Unable to load figure
          </h1>
          <p className="mt-3 text-sm text-[#737373] leading-relaxed">
            Something went wrong connecting to PaperViz. Check your network connection and try again.
          </p>
          <div className="mt-8 flex justify-center">
            <Link
              to={`/?ref=${shareToken}`}
              className="inline-flex items-center gap-2 rounded-[8px] bg-[#0a0a0a] px-4 py-2.5 text-sm font-medium text-white hover:bg-[#171717] transition-colors shadow-2xs"
            >
              <span>Analyze your own paper</span>
            </Link>
          </div>
        </div>
      </div>
    )
  }

  if (error || !figure) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white px-4">
        <div className="max-w-md w-full text-center py-16">
          <span className="font-mono text-xs font-semibold uppercase tracking-wider text-[#737373] bg-[#f5f5f5] px-2.5 py-1 rounded-[4px] border border-[#e5e5e5]">
            Error 404
          </span>
          <h1 className="mt-4 font-display text-3xl sm:text-4xl font-medium tracking-tight text-[#0a0a0a]">
            Figure not found
          </h1>
          <p className="mt-3 text-sm text-[#737373] leading-relaxed">
            This research figure could not be found or its public share link has expired.
          </p>
          <div className="mt-8 flex justify-center">
            <Link
              to={`/?ref=${shareToken}`}
              className="inline-flex items-center gap-2 rounded-[8px] bg-[#0a0a0a] px-4 py-2.5 text-sm font-medium text-white hover:bg-[#171717] transition-colors shadow-2xs"
            >
              <span>Analyze your own paper</span>
            </Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white">
      <header className="border-b border-[#e5e5e5] bg-white sticky top-0 z-10">
        <div className="mx-auto flex max-w-3xl items-center justify-between px-4 py-3">
          <div className="flex items-center gap-2.5">
            <Link
              to={`/?ref=${shareToken}`}
              className="inline-flex items-center gap-1.5 text-xs font-medium text-[#737373] hover:text-[#0a0a0a] transition-colors"
            >
              <ArrowLeft className="h-3.5 w-3.5" />
              <span>PaperViz</span>
            </Link>
            <span className="text-[#d4d4d4]">/</span>
            <span className="text-xs text-[#737373]">Shared Figure</span>
          </div>
          <Link
            to={`/?ref=${shareToken}`}
            className="text-xs font-medium text-[#2563eb] hover:text-[#1e40af] transition-colors"
          >
            Analyze another paper &rarr;
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-4 py-8">
        <div className="mb-6">
          <h1 className="font-display text-xl sm:text-2xl font-medium tracking-tight text-[#0a0a0a]">
            {figure.source_paper_title || "Research Figure"}
          </h1>
          <div className="mt-2.5 flex flex-wrap items-center gap-2">
            {figure.page_number > 0 && (
              <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
                Page {figure.page_number}
              </span>
            )}
            <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
              {READING_LEVEL_LABELS[figure.reading_level] || figure.reading_level}
            </span>
            <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2.5 py-0.5 text-[11px] font-medium text-[#737373]">
              {figure.source_method === "data_extracted" ? "Data-extracted" : "Image capture"}
            </span>
          </div>
        </div>

        <div className="grid gap-6 md:grid-cols-2">
          {figure.original_image_url && (
            <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-4">
              <span className="mb-3 inline-flex items-center rounded-full bg-[#f0f0f0] px-2.5 py-0.5 text-[11px] font-semibold text-[#0a0a0a]">
                Original Figure
              </span>
              <img
                src={figure.original_image_url}
                alt="Original figure from the paper"
                className="mt-2 h-auto w-full rounded-[8px] border border-[#e5e5e5] bg-white"
              />
            </div>
          )}

          <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-4">
            <span className="mb-3 inline-flex items-center rounded-full bg-[#dbeaff] px-2.5 py-0.5 text-[11px] font-semibold text-[#2563eb]">
              PaperViz Interpretation
            </span>
            <div className="mt-2">
              {figure.explanation ? (
                <p className="text-sm leading-relaxed text-[#171717]">
                  {figure.explanation}
                </p>
              ) : figure.chart_data ? (
                <p className="text-xs text-[#737373]">
                  Interactive chart available in the full document view.
                </p>
              ) : (
                <p className="text-xs text-[#737373]">
                  No interpretation available for this figure.
                </p>
              )}
            </div>
          </div>
        </div>

        <div className="mt-10 rounded-[12px] border border-[#e5e5e5] bg-[#fafafa] p-6 text-center">
          <p className="text-sm text-[#525252]">
            This explanation was generated by PaperViz from the original paper.
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
