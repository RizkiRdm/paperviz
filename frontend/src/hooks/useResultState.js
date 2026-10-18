import { useEffect, useRef, useState, useCallback } from "react"
import { useDocumentPoll } from "@/hooks/use-document-poll"
import { capture } from "@/lib/analytics"
import { generateDocumentShare, updateDocumentVisibility, addDocumentToCollection, listCollections } from "@/lib/api"

const COPY_FEEDBACK_MS = 2000

// repository.NewID mints 16 chars from a 58-symbol alphabet with ambiguous
// chars removed; shape-checking /:documentId here rejects a typo before it
// polls into a 404.
const DOCUMENT_ID = /^[23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]{16}$/

// Pipeline stages are implementation detail (P06): map OnStage emissions to
// five user-facing labels. Unknown stages fall back to default loading copy —
// never the raw stage name or any LLM/provider detail.
const STAGE_LABELS = {
  simplifying: "Reading document...",
  extracting_structure: "Extracting structure...",
  verifying: "Preparing evidence...",
  generating_charts: "Rebuilding figures...",
  completing: "Completing...",
}

// Single source of truth for the result page: state machine
// (loading|partial|ready|error|retryable-error|empty), polling, analytics, handlers.
export function useResultState(documentId, navigate) {
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
  const [collectionsError, setCollectionsError] = useState(false)
  const [collectionsNonce, setCollectionsNonce] = useState(0)
  const [showAddToCollection, setShowAddToCollection] = useState(false)
  const [collectionAddError, setCollectionAddError] = useState(null)
  const [showResearchMap, setShowResearchMap] = useState(false)
  const [showAnnotations, setShowAnnotations] = useState(false)
  const copyTimerRef = useRef(null)
  const lastCollectionRef = useRef(null)

  // Route-shape + state machine: empty = unknown document (404 view),
  // retryable-error = poll failure or timeout (retry refetches), error =
  // terminal pipeline failure (start over), loading = processing,
  // partial = verification_failed content, ready = complete.
  const invalidId = !DOCUMENT_ID.test(documentId)
  const state = invalidId || notFound
    ? "empty"
    : error || timedOut
      ? "retryable-error"
      : !doc || doc.status === "processing"
        ? "loading"
        : doc.status === "failed"
          ? "error"
          : doc.status === "verification_failed"
            ? "partial"
            : "ready"

  // Stage label for the loading view — mapped, never raw.
  const stageLabel = (doc?.processing_stage && STAGE_LABELS[doc.processing_stage]) || null

  // Cleanup copy timer on unmount.
  useEffect(() => () => clearTimeout(copyTimerRef.current), [])

  // Fires paper_analysed once per document, not once per render: this hook
  // re-runs on every state change, and a capture in render would be noise.
  useEffect(() => {
    if (!doc || doc.status === "processing") return
    if (doc.status === "failed") return
    capture("paper_analysed", {
      document_id: doc.id,
      source_type: doc.source_type,
      reading_level: doc.reading_level,
      chart_count: doc.charts?.length ?? 0,
      verification_failed: doc.status === "verification_failed",
    })
    // Intentionally keyed on identity and status only. Including doc itself
    // would re-fire on every poll tick that returns an equal document.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [doc?.id, doc?.status])

  // Auto-select first chapter when document loads with multiple chapters.
  useEffect(() => {
    const hasChapters = doc?.chapters && doc.chapters.length > 1
    if (hasChapters && activeChapter === -1) setActiveChapter(0)
  }, [doc, activeChapter])

  // Fetch user's collections on mount; failure surfaces inline with Retry.
  useEffect(() => {
    let cancelled = false
    async function fetchCollections() {
      try {
        const data = await listCollections()
        if (!cancelled) {
          setCollections(data.collections || [])
          setCollectionsError(false)
        }
      } catch (err) {
        console.error("Failed to load collections", err)
        if (!cancelled) setCollectionsError(true)
      }
    }
    fetchCollections()
    return () => { cancelled = true }
  }, [collectionsNonce])

  // onBack navigates to upload page.
  const onBack = useCallback(() => navigate("/"), [navigate])

  // retryCollections re-runs the collections fetch after an inline failure.
  function retryCollections() {
    setCollectionsNonce((n) => n + 1)
  }

  // handleCopyText copies the visible text; failure keeps the view and offers inline Retry.
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

  // handleShare mints a share link then opens the dialog; failures surface
  // inline near the header button.
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

  // handleVisibilityChange applies optimistically, reverts if the PATCH fails.
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

  // handleAddToCollection adds the document to a collection; failure keeps
  // the menu open with an inline Retry.
  async function handleAddToCollection(colId) {
    lastCollectionRef.current = colId
    setCollectionAddError(null)
    try {
      await addDocumentToCollection(colId, documentId)
      setShowAddToCollection(false)
    } catch (err) {
      console.error("Failed to add to collection", err)
      setCollectionAddError("Couldn't add this document to the collection. Please try again.")
    }
  }

  // retryAddToCollection re-attempts the last failed collection add.
  function retryAddToCollection() {
    if (lastCollectionRef.current != null) handleAddToCollection(lastCollectionRef.current)
  }

  return {
    state, doc, error, timedOut, takingLong, stageLabel, retry, onBack,
    showOriginal, setShowOriginal,
    showShare, setShowShare, shareUrl, shareError,
    visibility, setVisibility, visibilityError,
    showClaims, setShowClaims,
    textCopied, textCopyError,
    activeChapter, setActiveChapter,
    collections, collectionsError, retryCollections,
    showAddToCollection, setShowAddToCollection,
    collectionAddError, retryAddToCollection,
    showResearchMap, setShowResearchMap,
    showAnnotations, setShowAnnotations,
    handleCopyText, handleShare, handleVisibilityChange, handleAddToCollection,
  }
}
