// App — top-level page switch between the two screens in PRD.md's primary
// flow: upload (choose input + reading level) and result (view output).
// There is no client-side router library here — with exactly two screens
// and no deep-linkable sub-routes beyond "/{documentId}" (handled below via
// plain window.location, not react-router), a full router is unnecessary
// weight for this MVP (YAGNI). If PaperViz later needs more than two
// screens, that's the point to introduce one.
import { useState, useEffect } from "react"
import { UploadPage } from "@/pages/upload-page"
import { ResultPage } from "@/pages/result-page"
import { ErrorBoundary } from "@/components/error-boundary"

// Reads a document ID directly out of the URL path (e.g. "/V1StGXR8_Z5jdHi6B")
// so a shared link opens straight to the result page on first load, per
// PRD.md: "User can copy/share the link." Without this, refreshing a
// shared link would just show the upload screen instead of the document.
function documentIdFromLocation() {
  const path = window.location.pathname.replace(/^\/+/, "")
  return path.length > 0 ? path : null
}

export default function App() {
  const [documentId, setDocumentId] = useState(documentIdFromLocation)

  // Keep the URL in sync with which document is showing, and support the
  // browser back/forward buttons — both are ordinary expectations for a
  // shareable-link product, not extra scope.
  useEffect(() => {
    function onPopState() {
      setDocumentId(documentIdFromLocation())
    }
    window.addEventListener("popstate", onPopState)
    return () => window.removeEventListener("popstate", onPopState)
  }, [])

  function handleCreated(newDocumentId) {
    window.history.pushState({}, "", `/${newDocumentId}`)
    setDocumentId(newDocumentId)
  }

  function handleBack() {
    window.history.pushState({}, "", "/")
    setDocumentId(null)
  }

  return (
    <ErrorBoundary>
      {documentId ? (
        <ResultPage documentId={documentId} onBack={handleBack} />
      ) : (
        <UploadPage onCreated={handleCreated} />
      )}
    </ErrorBoundary>
  )
}
