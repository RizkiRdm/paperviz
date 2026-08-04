// ponytail: result page redesign with Dub top bar, card containers, and toggle pills
import { useEffect, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { VerificationBadge, WarningBanner, ErrorBanner } from "@/components/ui/status-banners"
import { ChartCard } from "@/components/chart-card"
import { getDocument } from "@/lib/api"
import { ArrowLeft, Copy, Check, Sparkles, RefreshCw, BarChart2 } from "lucide-react"

const POLL_INTERVAL_MS = 2000
const POLL_TIMEOUT_MS = 120000
const COPY_FEEDBACK_MS = 2000

function CopyLinkButton() {
  const [copied, setCopied] = useState(false)
  const [showUrlFallback, setShowUrlFallback] = useState(false)
  const copyTimer = useRef(null)

  useEffect(() => () => clearTimeout(copyTimer.current), [])

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(window.location.href)
      setShowUrlFallback(false)
      setCopied(true)
      clearTimeout(copyTimer.current)
      copyTimer.current = setTimeout(() => setCopied(false), COPY_FEEDBACK_MS)
    } catch {
      setShowUrlFallback(true)
    }
  }

  return (
    <div aria-live="polite" className="flex flex-col items-end gap-2">
      <Button
        variant="secondary"
        onClick={handleCopy}
        className="h-9 px-3 text-xs gap-1.5 font-medium"
      >
        {copied ? (
          <>
            <Check className="h-3.5 w-3.5 text-[#16a34a]" /> Copied Link
          </>
        ) : (
          <>
            <Copy className="h-3.5 w-3.5 text-[#737373]" /> Share Link
          </>
        )}
      </Button>
      {showUrlFallback && (
        <div className="flex flex-col gap-1">
          <span className="text-xs text-[#737373]">Copy manually:</span>
          <input
            readOnly
            value={window.location.href}
            onFocus={(e) => e.target.select()}
            className="w-64 rounded-[6px] border border-[#000000] bg-white px-2.5 py-1 text-xs text-[#171717]"
          />
        </div>
      )}
    </div>
  )
}

export function ResultPage({ documentId, onBack }) {
  const [doc, setDoc] = useState(null)
  const [error, setError] = useState(null)
  const [showOriginal, setShowOriginal] = useState(false)
  const [timedOut, setTimedOut] = useState(false)
  const [retryNonce, setRetryNonce] = useState(0)
  const pollTimer = useRef(null)
  const pollStartRef = useRef(Date.now())

  useEffect(() => {
    let cancelled = false
    setTimedOut(false)
    pollStartRef.current = Date.now()

    async function poll() {
      try {
        const fetched = await getDocument(documentId)
        if (cancelled) return
        setDoc(fetched)

        if (fetched.status === "processing") {
          if (Date.now() - pollStartRef.current > POLL_TIMEOUT_MS) {
            setTimedOut(true)
            return
          }
          pollTimer.current = setTimeout(poll, POLL_INTERVAL_MS)
        }
      } catch (err) {
        if (!cancelled) {
          setError(
            err.code === "not_found"
              ? "This link has expired or doesn't exist."
              : err.code === "network_timeout"
                ? "The request timed out. Please check your connection."
                : "Something went wrong loading this document."
          )
        }
      }
    }

    poll()
    return () => {
      cancelled = true
      clearTimeout(pollTimer.current)
    }
  }, [documentId, retryNonce])

  if (timedOut) {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6 text-center">
        <div className="max-w-md rounded-[16px] border border-[#e5e5e5] bg-white p-8 shadow-xs">
          <p className="text-sm text-[#737373]">Processing is taking longer than expected.</p>
          <div className="mt-6 flex flex-col gap-2">
            <Button onClick={() => setRetryNonce((n) => n + 1)} variant="primary">
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

  if (!doc || doc.status === "processing") {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6 text-center">
        <div className="max-w-md rounded-[16px] border border-[#e5e5e5] bg-white p-8 shadow-xs flex flex-col items-center">
          <div className="h-10 w-10 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent mb-4" />
          <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">Simplifying & Verifying...</h2>
          <p className="mt-2 text-xs text-[#737373] leading-relaxed">
            Reading paper, generating plain language summary, and verifying key statements against source.
          </p>
        </div>
      </div>
    )
  }

  if (doc.status === "failed") {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6">
        <div className="max-w-md w-full">
          <ErrorBanner message="We couldn't process this document. Please check the PDF format and try again." />
          <Button onClick={onBack} variant="secondary" className="mt-4 w-full">
            Start Over
          </Button>
        </div>
      </div>
    )
  }

  const displayedText = showOriginal ? doc.original_text : doc.simplified_text

  return (
    <div className="min-h-screen bg-white text-[#171717] bg-dotted-grid">
      {/* Top Bar Navigation */}
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
            {doc.status === "complete" && <VerificationBadge />}
            <CopyLinkButton />
          </div>
        </div>
      </header>

      {/* Content Container */}
      <main className="mx-auto max-w-[900px] px-6 py-12">
        {doc.status === "verification_failed" && (
          <div className="mb-6">
            <WarningBanner />
          </div>
        )}

        {/* Action Bar & Mode Switcher */}
        <div className="mb-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-6 border-b border-[#e5e5e5]">
          <div>
            <h1 className="font-satoshi text-2xl sm:text-3xl font-medium text-[#0a0a0a]">
              Paper Summary
            </h1>
            <p className="text-xs text-[#737373] mt-1">
              7-day share link ready · {showOriginal ? "Showing Original Text" : "Showing Plain Language Version"}
            </p>
          </div>

          {/* View Toggle Pill */}
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

        {/* Simplified / Original Text Container */}
        <article className="rounded-[16px] border border-[#e5e5e5] bg-white p-6 sm:p-8 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]">
          <div className="prose prose-neutral max-w-none text-[#171717] text-base leading-relaxed whitespace-pre-wrap font-inter">
            {displayedText}
          </div>
        </article>

        {/* Charts Section */}
        <section className="mt-12">
          <div className="flex items-center gap-2 mb-6">
            <BarChart2 className="h-5 w-5 text-[#2563eb]" />
            <h2 className="font-satoshi text-xl font-medium text-[#0a0a0a]">
              Charts & Visualizations
            </h2>
          </div>

          {doc.charts && doc.charts.length > 0 ? (
            <div className="flex flex-col gap-6">
              {doc.charts.map((chart) => (
                <ChartCard key={chart.id} chart={chart} />
              ))}
            </div>
          ) : doc.chart_extraction_degraded ? (
            <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
              <p className="text-xs text-[#737373]">
                Chart extraction could not complete for this document. Summary content is unaffected.
              </p>
            </div>
          ) : (
            <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
              <p className="text-xs text-[#737373]">No charts were detected in this paper.</p>
            </div>
          )}
        </section>
      </main>
    </div>
  )
}

