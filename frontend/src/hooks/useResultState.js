import { useEffect, useRef, useState, useCallback } from "react"
import { useDocumentPoll } from "@/hooks/use-document-poll"
import { generateDocumentShare, updateDocumentVisibility, addDocumentToCollection, listCollections } from "@/lib/api"

const COPY_FEEDBACK_MS = 2000

// useResultState centralizes all result-page state and handlers — single
// source for doc polling, UI toggles, share/visibility, collections, chapters.
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
  const [showAddToCollection, setShowAddToCollection] = useState(false)
  const [showResearchMap, setShowResearchMap] = useState(false)
  const [showAnnotations, setShowAnnotations] = useState(false)
  const copyTimerRef = useRef(null)

  // Cleanup copy timer on unmount.
  useEffect(() => () => clearTimeout(copyTimerRef.current), [])

  // Auto-select first chapter when document loads with multiple chapters.
  useEffect(() => {
    const hasChapters = doc?.chapters && doc.chapters.length > 1
    if (hasChapters && activeChapter === -1) setActiveChapter(0)
  }, [doc, activeChapter])

  // Fetch user's collections on mount.
  useEffect(() => {
    async function fetchCollections() {
      try {
        const data = await listCollections()
        setCollections(data.collections || [])
      } catch {
        // collections fetch is non-critical, fail silently
      }
    }
    fetchCollections()
  }, [])

  // onBack navigates to upload page.
  const onBack = useCallback(() => navigate("/"), [navigate])

  // handleCopyText copies simplified or original text to clipboard.
  async function handleCopyText() {
    const text = showOriginal ? doc.original_text : doc.simplified_text
    try {
      await navigator.clipboard.writeText(text)
      setTextCopied(true)
      setTextCopyError(false)
      clearTimeout(copyTimerRef.current)
      copyTimerRef.current = setTimeout(() => setTextCopied(false), COPY_FEEDBACK_MS)
    } catch {
      setTextCopied(false)
      setTextCopyError(true)
    }
  }

  // handleShare mints a share link via API, then opens the dialog.
  async function handleShare() {
    setShareError(false)
    try {
      const { share_url } = await generateDocumentShare(documentId)
      setShareUrl(window.location.origin + share_url)
      setShowShare(true)
    } catch {
      setShareError(true)
    }
  }

  // handleVisibilityChange applies visibility optimistically, reverts on failure.
  async function handleVisibilityChange(next) {
    const previous = visibility
    setVisibility(next)
    setVisibilityError(false)
    try {
      await updateDocumentVisibility(documentId, next)
    } catch {
      setVisibility(previous)
      setVisibilityError(true)
    }
  }

  // handleAddToCollection adds document to the selected collection.
  async function handleAddToCollection(colId) {
    try {
      await addDocumentToCollection(colId, documentId)
      setShowAddToCollection(false)
    } catch {
      // collection add is non-critical, fail silently
    }
  }

  return {
    doc, error, notFound, timedOut, takingLong, retry, onBack,
    showOriginal, setShowOriginal,
    showShare, setShowShare, shareUrl, shareError,
    visibility, setVisibility, visibilityError,
    showClaims, setShowClaims,
    textCopied, textCopyError,
    activeChapter, setActiveChapter,
    collections, showAddToCollection, setShowAddToCollection,
    showResearchMap, setShowResearchMap,
    showAnnotations, setShowAnnotations,
    handleCopyText, handleShare, handleVisibilityChange, handleAddToCollection,
  }
}
