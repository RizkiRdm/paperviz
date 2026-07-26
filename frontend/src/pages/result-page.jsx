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
import { VerificationBadge, WarningBanner, ErrorBanner } from "@/components/ui/status-banners"
import { ChartCard } from "@/components/chart-card"
import { getDocument } from "@/lib/api"

// POLL_INTERVAL_MS matches ARCHITECTURE.md Section E's client contract
// exactly: "Client polling interval for processing status: REQUIRED
// minimum 2s between polls (frontend-enforced)." Do not lower this without
// updating that doc first — it's a stated server-side assumption, not just
// a frontend performance choice.
const POLL_INTERVAL_MS = 2000

export function ResultPage({ documentId, onBack }) {
  const [doc, setDoc] = useState(null)
  const [error, setError] = useState(null)
  const [showOriginal, setShowOriginal] = useState(false)
  const pollTimer = useRef(null)

  useEffect(() => {
    let cancelled = false

    async function poll() {
      try {
        const fetched = await getDocument(documentId)
        if (cancelled) return
        setDoc(fetched)

        if (fetched.status === "processing") {
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
  }, [documentId])

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
        {doc.status === "complete" && <VerificationBadge />}
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
