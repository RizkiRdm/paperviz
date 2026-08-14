import { useEffect, useRef, useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { VerificationBadge, WarningBanner, ErrorBanner, ClaimComparisonPanel } from "@/components/ui/status-banners"
import { ChartCard } from "@/components/chart-card"
import { useDocumentPoll } from "@/hooks/use-document-poll"
import { ArrowLeft, Copy, Check, Sparkles, RefreshCw, BarChart2, Link2, X } from "lucide-react"
import { NotFoundPage } from "@/pages/not-found-page"

const COPY_FEEDBACK_MS = 2000

const READING_LEVEL_LABELS = {
  simplified: "Simplified",
  eli5: "ELI5",
}

const ERROR_MESSAGES = {
  simplification_failed: "We couldn't simplify this paper. Please try again.",
  verification_failed_to_run: "We couldn't verify the summary against the original. Please try again.",
  extraction_failed: "We couldn't read this PDF. Please check that it has a text layer and try again.",
}

// Share dialog modal — DESIGN.md Elevated Feature Card (16px radius, subtle-2 shadow)
function ShareDialog({ url, onClose }) {
  const [copied, setCopied] = useState(false)
  const [copyError, setCopyError] = useState(false)
  const timerRef = useRef(null)
  const inputRef = useRef(null)

  useEffect(() => {
    inputRef.current?.focus()
    return () => clearTimeout(timerRef.current)
  }, [])

  useEffect(() => {
    function handleKeyDown(e) {
      if (e.key === "Escape") onClose()
    }
    document.addEventListener("keydown", handleKeyDown)
    return () => document.removeEventListener("keydown", handleKeyDown)
  }, [onClose])

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(url)
      setCopied(true)
      setCopyError(false)
      clearTimeout(timerRef.current)
      timerRef.current = setTimeout(() => setCopied(false), COPY_FEEDBACK_MS)
    } catch {
      // Fallback: select the input so user can Cmd+C / Ctrl+C manually
      setCopyError(true)
      inputRef.current?.select()
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 backdrop-blur-[2px]">
      <div className="w-full max-w-md rounded-[16px] border border-[#e5e5e5] bg-white p-6 shadow-[rgba(0,0,0,0.1)_0px_0px_0px_4px]">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Link2 className="h-4 w-4 text-[#2563eb]" />
            <h3 className="text-sm font-semibold text-[#0a0a0a]">Share this summary</h3>
          </div>
          <button onClick={onClose} className="rounded-full p-1 text-[#737373] hover:text-[#0a0a0a] transition-colors">
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="flex gap-2">
          <input
            ref={inputRef}
            readOnly
            value={url}
            onFocus={(e) => e.target.select()}
            aria-label="Shareable link"
            className="flex-1 rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-xs text-[#171717] font-mono"
          />
          <Button onClick={handleCopy} variant="secondary" className="shrink-0">
            {copied ? <Check className="h-3.5 w-3.5 text-[#16a34a]" /> : <Copy className="h-3.5 w-3.5 text-[#737373]" />}
          </Button>
        </div>
        {copyError && (
          <p className="mt-2 text-[11px] text-[#ea580c]">Couldn't copy automatically. Select the link and press Ctrl+C / Cmd+C.</p>
        )}
        <p className="mt-2 text-[11px] text-[#737373]">Link expires after 7 days of inactivity.</p>
      </div>
    </div>
  )
}

export function ResultPage() {
  const { documentId } = useParams()
  const navigate = useNavigate()
  const onBack = () => navigate("/")
  const { doc, error, notFound, timedOut, takingLong, retry } = useDocumentPoll(documentId)
  const [showOriginal, setShowOriginal] = useState(false)
  const [showShare, setShowShare] = useState(false)
  const [showClaims, setShowClaims] = useState(false)
  const [textCopied, setTextCopied] = useState(false)
  const copyTimerRef = useRef(null)

  useEffect(() => () => clearTimeout(copyTimerRef.current), [])

  async function handleCopyText() {
    const text = showOriginal ? doc.original_text : doc.simplified_text
    try {
      await navigator.clipboard.writeText(text)
      setTextCopied(true)
      clearTimeout(copyTimerRef.current)
      copyTimerRef.current = setTimeout(() => setTextCopied(false), COPY_FEEDBACK_MS)
    } catch {
      setTextCopied(false)
    }
  }

  if (notFound) {
    return <NotFoundPage />
  }

  if (timedOut) {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6 text-center">
        <div className="max-w-md rounded-[16px] border border-[#e5e5e5] bg-white p-8 shadow-xs">
          <p className="text-sm text-[#737373]">Processing is taking longer than expected.</p>
          <div className="mt-6 flex flex-col gap-2">
            <Button onClick={retry} variant="primary">
              <RefreshCw className="h-4 w-4 mr-2" /> Check Status Again
            </Button>
            <Button onClick={onBack} variant="secondary">
              Upload Another Paper
            </Button>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6">
        <div className="max-w-md w-full">
          <ErrorBanner message={error} />
          <Button onClick={onBack} variant="secondary" className="mt-4 w-full">
            Start Over
          </Button>
        </div>
      </div>
    )
  }

const STAGE_LABELS = {
  simplifying: "Simplifying language...",
  verifying: "Checking accuracy against original...",
  generating_charts: "Generating visualizations...",
}

  if (!doc || doc.status === "processing") {
    const stageLabel = doc?.processing_stage ? STAGE_LABELS[doc.processing_stage] || doc.processing_stage : null
    return (
      <div
        role="status"
        aria-live="polite"
        className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6 text-center"
      >
        <div className="max-w-md rounded-[16px] border border-[#e5e5e5] bg-white p-8 shadow-xs flex flex-col items-center">
          <div className="h-10 w-10 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent mb-4" />
          <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">
            {stageLabel || "Simplifying & Verifying..."}
          </h2>
          <p className="mt-2 text-xs text-[#737373] leading-relaxed">
            {takingLong
              ? "Still working — large papers with many figures can take a few minutes. This page refreshes automatically."
              : "Reading your paper, creating a plain-language summary, and checking it against the original."}
          </p>
        </div>
      </div>
    )
  }

  if (doc.status === "failed") {
    const msg = doc.error_message && ERROR_MESSAGES[doc.error_message]
      ? ERROR_MESSAGES[doc.error_message]
      : "We couldn't process this PDF. It may be image-only or corrupted. Please try a different file."
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6">
        <div className="max-w-md w-full">
          <ErrorBanner message={msg} />
          <Button onClick={onBack} variant="secondary" className="mt-4 w-full">
            Start Over
          </Button>
        </div>
      </div>
    )
  }

  const displayedText = showOriginal ? doc.original_text : doc.simplified_text
  const levelLabel = READING_LEVEL_LABELS[doc.reading_level] || doc.reading_level
  const hasChapters = doc.chapters && doc.chapters.length > 1
  const [activeChapter, setActiveChapter] = useState(hasChapters ? 0 : -1)

  const activeChapterData = hasChapters ? doc.chapters[activeChapter] : null
  const chapterContent = activeChapterData?.content || displayedText
  const chapterCharts = hasChapters
    ? (doc.charts || []).filter(c => c.chapter_id === activeChapterData?.id)
    : (doc.charts || [])

  return (
    <div className="min-h-screen bg-white text-[#171717] bg-dotted-grid">
      <header className="border-b border-[#e5e5e5] bg-white/80 backdrop-blur-xs sticky top-0 z-10">
        <div className="mx-auto flex max-w-[1200px] items-center justify-between px-6 py-3.5">
          <div className="flex items-center gap-4">
            <button
              onClick={onBack}
              className="flex items-center gap-1.5 text-xs font-medium text-[#737373] hover:text-[#0a0a0a] transition-colors"
            >
              <ArrowLeft className="h-4 w-4" /> New Document
            </button>
            <span className="h-4 w-px bg-[#e5e5e5]" />
            <span className="font-mono text-xs font-semibold text-[#0a0a0a]">PaperViz</span>
          </div>
          <div className="flex items-center gap-3">
            {doc.status === "complete" && (
              <VerificationBadge onClick={() => setShowClaims(v => !v)} />
            )}
            <Button variant="secondary" onClick={() => setShowShare(true)} className="h-9 px-3 text-xs gap-1.5 font-medium">
              <Link2 className="h-3.5 w-3.5 text-[#737373]" /> Share
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-[900px] px-6 py-12">
        {doc.status === "verification_failed" && (
          <div className="mb-6"><WarningBanner /></div>
        )}

        {showClaims && doc.claim_diff && (
          <div className="mb-6">
            <ClaimComparisonPanel claimDiff={doc.claim_diff} onClose={() => setShowClaims(false)} />
          </div>
        )}

        {/* Action bar — reading level badge + view toggle + copy text */}
        <div className="mb-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-6 border-b border-[#e5e5e5]">
          <div className="flex items-center gap-3">
            <h1 className="font-satoshi text-2xl sm:text-3xl font-medium text-[#0a0a0a]">
              Paper Summary
            </h1>
            <span className="inline-flex items-center rounded-full bg-[#dbeaff] px-2.5 py-0.5 text-[11px] font-medium text-[#2563eb]">
              {levelLabel}
            </span>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="secondary"
              onClick={handleCopyText}
              className="h-8 px-2.5 text-xs gap-1.5 font-medium"
            >
              {textCopied ? <Check className="h-3.5 w-3.5 text-[#16a34a]" /> : <Copy className="h-3.5 w-3.5 text-[#737373]" />}
              {textCopied ? "Copied" : "Copy Text"}
            </Button>

            <div className="inline-flex gap-1 rounded-full border border-[#e5e5e5] bg-white p-1 shadow-2xs">
              <button
                onClick={() => setShowOriginal(false)}
                className={`rounded-full px-3.5 py-1 text-xs font-medium transition-colors ${
                  !showOriginal ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717]"
                }`}
              >
                Simplified
              </button>
              <button
                onClick={() => setShowOriginal(true)}
                className={`rounded-full px-3.5 py-1 text-xs font-medium transition-colors ${
                  showOriginal ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717]"
                }`}
              >
                Original
              </button>
            </div>
          </div>
        </div>

        {/* Chapter tabs — horizontal scrollable, only when 2+ chapters */}
        {hasChapters && (
          <div
            role="tablist"
            aria-label="Paper sections"
            className="mb-6 flex gap-1 overflow-x-auto pb-2 scrollbar-hide"
          >
            {doc.chapters.map((ch, i) => (
              <button
                key={ch.id || i}
                role="tab"
                id={`tab-${i}`}
                aria-selected={activeChapter === i}
                aria-controls={`panel-${i}`}
                tabIndex={activeChapter === i ? 0 : -1}
                onClick={() => setActiveChapter(i)}
                onKeyDown={(e) => {
                  if (e.key === "ArrowRight") setActiveChapter(Math.min(i + 1, doc.chapters.length - 1))
                  if (e.key === "ArrowLeft") setActiveChapter(Math.max(i - 1, 0))
                }}
                className={`shrink-0 rounded-full px-4 py-2 text-xs font-medium transition-colors whitespace-nowrap ${
                  activeChapter === i
                    ? "bg-[#0a0a0a] text-white"
                    : "text-[#737373] hover:text-[#171717] hover:bg-[#f5f5f5]"
                }`}
              >
                {ch.title}
              </button>
            ))}
          </div>
        )}

        {/* Chapter content panel */}
        <div
          role="tabpanel"
          id={`panel-${activeChapter}`}
          aria-labelledby={`tab-${activeChapter}`}
          tabIndex={0}
        >
          {/* Chapter summary — only in tabbed mode */}
          {activeChapterData && (
            <p className="mb-4 text-sm text-[#737373] italic">{activeChapterData.summary}</p>
          )}

          <article className="rounded-[16px] border border-[#e5e5e5] bg-white p-6 sm:p-8 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]">
            <div className="prose prose-neutral max-w-none text-[#171717] text-base leading-relaxed whitespace-pre-wrap font-inter">
              {showOriginal ? displayedText : chapterContent}
            </div>
          </article>

          {/* Charts for this chapter */}
          {chapterCharts.length > 0 && (
            <section className="mt-8">
              <div className="flex items-center gap-2 mb-4">
                <BarChart2 className="h-4 w-4 text-[#2563eb]" />
                <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">
                  {activeChapterData ? `Charts — ${activeChapterData.title}` : "Charts & Visualizations"}
                </h2>
              </div>
              <div className="flex flex-col gap-4">
                {chapterCharts.map((chart) => (
                  <ChartCard key={chart.id} chart={chart} chapterTitle={activeChapterData?.title} />
                ))}
              </div>
            </section>
          )}
        </div>

        {/* Empty states — only when no chapters */}
        {!hasChapters && (
          <>
            {doc.charts && doc.charts.length > 0 ? (
              <section className="mt-8">
                <div className="flex items-center gap-2 mb-4">
                  <BarChart2 className="h-4 w-4 text-[#2563eb]" />
                  <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">
                    Charts & Visualizations
                  </h2>
                </div>
                <div className="flex flex-col gap-4">
                  {doc.charts.map((chart) => (
                    <ChartCard key={chart.id} chart={chart} />
                  ))}
                </div>
              </section>
            ) : doc.chart_extraction_degraded ? (
              <div className="mt-8 rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
                <p className="text-xs text-[#737373]">
                  We couldn't extract charts from this paper. The summary is still available.
                </p>
              </div>
            ) : (
              <div className="mt-8 rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
                <p className="text-xs text-[#737373]">No charts were detected in this paper.</p>
              </div>
            )}
          </>
        )}
      </main>

      {showShare && <ShareDialog url={window.location.href} onClose={() => setShowShare(false)} />}
    </div>
  )
}
