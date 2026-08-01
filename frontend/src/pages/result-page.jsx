// ResultPage — PRD.md "User Flows > Primary Flow" steps 3-5: processing
// indicator while polling, then simplified text with an original-text
// toggle, charts inline with annotations, and a verification badge/warning
// depending on outcome. This is the single most state-heavy component in
// the app; read the polling effect first, then the render branches below
// it follow the same status values the backend can return.
//
// Note: the fetched document is stored as `doc`, not `document` — using
// `document` as a variable name would shadow the global window.document
// object for the rest of this file, which is a real footgun in any
// component that might later need DOM APIs.
import { useEffect, useRef, useState } from "react"
import { Button } from "@/components/ui/button"
import { VerificationBadge, WarningBanner, ErrorBanner } from "@/components/ui/status-banners"
import { ChartCard } from "@/components/chart-card"
import { getDocument } from "@/lib/api"

// POLL_INTERVAL_MS matches ARCHITECTURE.md Section E's client contract
// exactly: "Client polling interval for processing status: REQUIRED
// minimum 2s between polls (frontend-enforced)." Do not lower this without
// updating that doc first — it's a stated server-side assumption, not just
// a frontend performance choice.
const POLL_INTERVAL_MS = 2000

// POLL_TIMEOUT_MS bounds the processing wait: if a document is stuck (e.g.
// server restarted mid-pipeline), the user gets an exit instead of an
// infinite spinner. Assumption value — negotiable, not a contract.
const POLL_TIMEOUT_MS = 120000

// COPY_FEEDBACK_MS — how long the button label stays "Copied!" after a
// successful clipboard write before reverting to "Copy link".
const COPY_FEEDBACK_MS = 2000

// CopyLinkButton — PRD.md "User Flows > Primary Flow" share step: the user
// can copy the shareable link. Uses navigator.clipboard when available
// (secure context only); if the API is missing or the write is rejected
// (older browser, non-HTTPS, permission denied), falls back to a visible
// selectable <input> instead of failing silently. Confirmation is a local
// state swap for ~2s — no toast library (none in package.json).
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
        className={`transition active:scale-[0.98] ${
          copied ? "border-accent-verified/30 bg-accent-verified-soft text-accent-verified" : ""
        }`}
      >
        {copied ? "Copied!" : "Copy link"}
      </Button>
      {showUrlFallback && (
        <div className="flex flex-col gap-1">
          <span className="text-caption text-ink-secondary">Copy this link manually</span>
          <input
            readOnly
            value={window.location.href}
            onFocus={(e) => e.target.select()}
            className="w-72 max-w-full rounded-md border border-border-default bg-surface-raised px-3 py-2 text-body text-ink-primary"
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
              : "Something went wrong loading this document.",
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
      <main className="mx-auto flex min-h-screen max-w-2xl flex-col items-center justify-center px-6 text-center">
        <p className="text-body text-ink-secondary">This is taking longer than usual.</p>
        <button
          onClick={() => setRetryNonce((n) => n + 1)}
          className="mt-6 rounded-sm px-3 py-1.5 text-body font-medium bg-accent-verified-soft text-accent-verified"
        >
          Retry
        </button>
        <button onClick={onBack} className="mt-4 text-body text-accent-verified underline">
          Start over
        </button>
      </main>
    )
  }

  if (error) {
    return (
      <main className="mx-auto max-w-2xl px-6 py-16">
        <ErrorBanner message={error} />
        <button onClick={onBack} className="mt-4 text-body text-accent-verified underline">
          Start over
        </button>
      </main>
    )
  }

  if (!doc || doc.status === "processing") {
    return (
      <main className="mx-auto flex min-h-screen max-w-2xl flex-col items-center justify-center px-6 text-center">
        <p className="text-body text-ink-secondary">
          Reading your paper and rewriting it — this usually takes under a minute.
        </p>
      </main>
    )
  }

  if (doc.status === "failed") {
    return (
      <main className="mx-auto max-w-2xl px-6 py-16">
        <ErrorBanner message="We couldn't process this document. Please try again." />
        <button onClick={onBack} className="mt-4 text-body text-accent-verified underline">
          Start over
        </button>
      </main>
    )
  }

  // status is "complete" or "verification_failed" from here on — both
  // have simplified_text to show (ARCHITECTURE.md Acceptance Scenario 4:
  // a mismatch flags the result, it doesn't delete it).
  const displayedText = showOriginal ? doc.original_text : doc.simplified_text

  return (
    <main className="mx-auto max-w-2xl px-6 py-16">
      <div className="flex items-center justify-between gap-4">
        <button onClick={onBack} className="text-body text-ink-secondary hover:text-ink-primary">
          ← New document
        </button>
        <div className="flex items-center gap-3">
          <CopyLinkButton />
          {doc.status === "complete" && <VerificationBadge />}
        </div>
      </div>

      {doc.status === "verification_failed" && (
        <div className="mt-6">
          <WarningBanner />
        </div>
      )}

      <div className="mt-6 flex items-center gap-2">
        <button
          onClick={() => setShowOriginal(false)}
          className={`rounded-sm px-3 py-1.5 text-body font-medium ${!showOriginal ? "bg-accent-verified-soft text-accent-verified" : "text-ink-secondary"}`}
        >
          Simplified
        </button>
        <button
          onClick={() => setShowOriginal(true)}
          className={`rounded-sm px-3 py-1.5 text-body font-medium ${showOriginal ? "bg-accent-verified-soft text-accent-verified" : "text-ink-secondary"}`}
        >
          Original
        </button>
      </div>

      <article className="mt-6 max-w-prose text-reading font-reading text-ink-primary whitespace-pre-wrap">
        {displayedText}
      </article>

      <section className="mt-12">
        <h2 className="text-h2 text-ink-primary">Charts</h2>
        {doc.charts && doc.charts.length > 0 ? (
          <div className="mt-4 flex flex-col gap-4">
            {doc.charts.map((chart) => (
              <ChartCard key={chart.id} chart={chart} />
            ))}
          </div>
        ) : doc.chart_extraction_degraded ? (
          <p className="mt-4 text-body text-ink-secondary">
            Chart extraction couldn&apos;t complete for this document. The rest of the content is unaffected.
          </p>
        ) : (
          <p className="mt-4 text-body text-ink-secondary">No charts detected in this document.</p>
        )}
      </section>
    </main>
  )
}
