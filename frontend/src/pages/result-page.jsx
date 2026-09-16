import { useEffect, useRef, useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { WarningBanner, ErrorBanner, ClaimComparisonPanel } from "@/components/ui/status-banners"
import { useDocumentPoll } from "@/hooks/use-document-poll"
import { generateDocumentShare, updateDocumentVisibility } from "@/lib/api"
import { RefreshCw } from "lucide-react"
import { NotFoundPage } from "@/pages/not-found-page"
import { ResultHeader } from "@/components/result/result-header"
import { UnderstandingSection } from "@/components/result/understanding-section"
import { EvidenceSection } from "@/components/result/evidence-section"
import { FiguresSection } from "@/components/result/figures-section"
import { SourceSection } from "@/components/result/source-section"
import { ResultActions, ResultOverlays, ShareDialog } from "@/components/result/result-actions"

const COPY_FEEDBACK_MS = 2000

const ERROR_MESSAGES = {
  simplification_failed: "We couldn't simplify this paper. Please try again.",
  verification_failed_to_run: "We couldn't verify the summary against the original. Please try again.",
  extraction_failed: "We couldn't read this PDF. Please check that it has a text layer and try again.",
}

export function ResultPage() {
  const { documentId } = useParams()
  const navigate = useNavigate()
  const onBack = () => navigate("/")
  const { doc, error, notFound, timedOut, takingLong, retry } = useDocumentPoll(documentId)
  const [showOriginal, setShowOriginal] = useState(false)
  const [showShare, setShowShare] = useState(false)
  const [shareUrl, setShareUrl] = useState(null)
  const [shareError, setShareError] = useState(false)
  const [visibility, setVisibility] = useState("private")
  const [visibilityError, setVisibilityError] = useState(false)
  const [showClaims, setShowClaims] = useState(false)
  const [textCopied, setTextCopied] = useState(false)
  const [textCopyError, setTextCopyError] = useState(false)
  const [activeChapter, setActiveChapter] = useState(-1)
  const [collections, setCollections] = useState([])
  const [showAddToCollection, setShowAddToCollection] = useState(false)
  const [showResearchMap, setShowResearchMap] = useState(false)
  const [showAnnotations, setShowAnnotations] = useState(false)
  const copyTimerRef = useRef(null)

  useEffect(() => () => clearTimeout(copyTimerRef.current), [])
  useEffect(() => {
    const hasChapters = doc?.chapters && doc.chapters.length > 1
    if (hasChapters && activeChapter === -1) setActiveChapter(0)
  }, [doc, activeChapter])

  useEffect(() => {
    async function fetchCollections() {
      try {
        const res = await fetch("/api/collections")
        if (res.ok) {
          const data = await res.json()
          setCollections(data.collections || [])
        }
      } catch (err) { console.error("Failed to load collections", err) }
    }
    fetchCollections()
  }, [])

  async function handleCopyText() {
    const text = showOriginal ? doc.original_text : doc.simplified_text
    try {
      await navigator.clipboard.writeText(text)
      setTextCopied(true)
      setTextCopyError(false)
      clearTimeout(copyTimerRef.current)
      copyTimerRef.current = setTimeout(() => setTextCopied(false), COPY_FEEDBACK_MS)
    } catch (err) {
      console.error("Copy to clipboard failed", err)
      setTextCopied(false)
      setTextCopyError(true)
    }
  }

  // handleShare mints a real share link via the API, then opens the
  // dialog with the absolute URL; failures surface inline near the button.
  async function handleShare() {
    setShareError(false)
    try {
      const { share_url } = await generateDocumentShare(documentId)
      setShareUrl(window.location.origin + share_url)
      setShowShare(true)
    } catch (err) {
      console.error("Share link creation failed", err)
      setShareError(true)
    }
  }

  // handleVisibilityChange applies the new visibility optimistically and
  // reverts to the previous value if the PATCH fails.
  async function handleVisibilityChange(next) {
    const previous = visibility
    setVisibility(next)
    setVisibilityError(false)
    try {
      await updateDocumentVisibility(documentId, next)
    } catch (err) {
      console.error("Visibility update failed", err)
      setVisibility(previous)
      setVisibilityError(true)
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

const STAGE_LABELS = { simplifying: "Reading document...", verifying: "Preparing evidence...", generating_charts: "Rebuilding figures...", extracting_structure: "Extracting structure...", completing: "Completing..." }

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
            {stageLabel || "Reading document..."}
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

  const hasChapters = doc.chapters && doc.chapters.length > 1

  async function handleAddToCollection(colId) {
    try {
      const res = await fetch(`/api/collections/${colId}/documents`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ document_id: documentId }),
      })
      if (res.ok) {
        setShowAddToCollection(false)
      }
    } catch (err) { console.error("Failed to add to collection", err) }
  }

  return (
    <div className="min-h-screen bg-white text-[#171717] bg-dotted-grid">
      <ResultHeader
        doc={doc}
        visibility={visibility}
        visibilityError={visibilityError}
        shareError={shareError}
        showClaims={showClaims}
        onBack={onBack}
        onVisibilityChange={handleVisibilityChange}
        onShare={handleShare}
        onToggleClaims={() => setShowClaims(v => !v)}
      />

      <main className="mx-auto max-w-[900px] px-6 py-12">
        {doc.status === "verification_failed" && (
          <div className="mb-6">
            <WarningBanner detail={doc.claim_diff?.mismatch_detail} />
            {doc.claim_diff && (
              <div className="mt-3 flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() => setShowClaims(v => !v)}
                  aria-expanded={showClaims}
                  aria-controls="claim-comparison-panel"
                  className="inline-flex items-center rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-xs font-medium text-[#171717] hover:bg-[#f5f5f5] transition-colors"
                >
                  {showClaims ? "Hide checked claims" : "View checked claims"}
                </button>
                <button
                  type="button"
                  onClick={() => { setShowOriginal(true); document.querySelector("main article")?.scrollIntoView({ behavior: "smooth" }) }}
                  className="inline-flex items-center rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-xs font-medium text-[#171717] hover:bg-[#f5f5f5] transition-colors"
                >
                  Compare with Original
                </button>
              </div>
            )}
          </div>
        )}

        {showClaims && doc.claim_diff && (
          <div className="mb-6">
            <ClaimComparisonPanel claimDiff={doc.claim_diff} onClose={() => setShowClaims(false)} />
          </div>
        )}

        <ResultOverlays
          documentId={documentId}
          showResearchMap={showResearchMap}
          setShowResearchMap={setShowResearchMap}
          showAnnotations={showAnnotations}
          setShowAnnotations={setShowAnnotations}
        />

        <UnderstandingSection
          doc={doc}
          documentId={documentId}
          showOriginal={showOriginal}
          setShowOriginal={setShowOriginal}
          textCopied={textCopied}
          textCopyError={textCopyError}
          handleCopyText={handleCopyText}
          activeChapter={activeChapter}
          setActiveChapter={setActiveChapter}
          showResearchMap={showResearchMap}
          setShowResearchMap={setShowResearchMap}
          showAnnotations={showAnnotations}
          setShowAnnotations={setShowAnnotations}
        />

        <EvidenceSection evidence={doc.evidence} />

        {/* Figures: flat mode via FiguresSection, chapter mode in UnderstandingSection. */}
        {!hasChapters && (
          <FiguresSection
            charts={doc.charts}
            evidence={doc.evidence}
            chapters={doc.chapters}
            activeChapter={activeChapter}
            documentId={documentId}
            chartExtractionDegraded={doc.chart_extraction_degraded}
          />
        )}

        <SourceSection doc={doc} showOriginal={showOriginal} onToggleOriginal={setShowOriginal} />

        {doc.status === "complete" && (
          <ResultActions
            documentId={documentId}
            collections={collections}
            showAddToCollection={showAddToCollection}
            setShowAddToCollection={setShowAddToCollection}
            onAddToCollection={handleAddToCollection}
          />
        )}
      </main>

      {showShare && shareUrl && <ShareDialog url={shareUrl} onClose={() => setShowShare(false)} />}
    </div>
  )
}
