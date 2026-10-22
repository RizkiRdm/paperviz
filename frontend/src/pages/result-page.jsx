import { useNavigate, useParams } from "react-router-dom"
import { WarningBanner, ClaimComparisonPanel } from "@/components/ui/status-banners"
import { NotFoundPage } from "@/pages/not-found-page"
import { useResultState } from "@/hooks/useResultState"
import { ResultHeader } from "@/components/result/result-header"
import { UnderstandingSection } from "@/components/result/understanding-section"
import { EvidenceSection } from "@/components/result/evidence-section"
import { FiguresSection } from "@/components/result/figures-section"
import { SourceSection } from "@/components/result/source-section"
import { ResultActions, ResultOverlays, ShareDialog } from "@/components/result/result-actions"
import { ProcessingView, TimeoutView, FailureView } from "@/components/result/status-views"

// Thin orchestrator: every state decision, stage label, and handler lives in
// useResultState (single source of truth); sections own their own JSX.
export function ResultPage() {
  const { documentId } = useParams()
  const navigate = useNavigate()
  const {
    state, doc, error, stageLabel, takingLong, retry, onBack,
    showOriginal, setShowOriginal, showShare, setShowShare, shareUrl, shareError,
    visibility, visibilityError, showClaims, setShowClaims,
    textCopied, textCopyError, handleCopyText,
    activeChapter, setActiveChapter,
    collections, collectionsError, retryCollections,
    collectionAddError, retryAddToCollection,
    showAddToCollection, setShowAddToCollection, handleAddToCollection,
    showResearchMap, setShowResearchMap, showAnnotations, setShowAnnotations,
    handleShare, handleVisibilityChange,
  } = useResultState(documentId, navigate)

  if (state === "empty") return <NotFoundPage />
  if (state === "retryable-error" && !error) return <TimeoutView onRetry={retry} onBack={onBack} />
  if (state === "error" || state === "retryable-error") {
    return (
      <FailureView
        pollError={state === "retryable-error" ? error : null}
        failCode={doc?.error_message}
        onRetry={state === "retryable-error" ? retry : null}
        onBack={onBack}
      />
    )
  }
  if (state === "loading") return <ProcessingView stageLabel={stageLabel} takingLong={takingLong} />

  const hasChapters = doc.chapters && doc.chapters.length > 1

  return (
    <div className="workspace-shell bg-dotted-grid">
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

      <main className="container workspace-content">
        {state === "partial" && (
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

        {state === "ready" && (
          <ResultActions
            documentId={documentId}
            collections={collections}
            collectionsError={collectionsError}
            onRetryCollections={retryCollections}
            collectionAddError={collectionAddError}
            onRetryCollectionAdd={retryAddToCollection}
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
