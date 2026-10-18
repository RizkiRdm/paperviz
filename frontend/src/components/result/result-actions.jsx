import { useEffect, useRef, useState } from "react"
import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { ArrowRight, Copy, Check, Link2, FolderPlus, LayoutDashboard, Network, StickyNote } from "lucide-react"
import { exportResearchContext } from "@/lib/api"
import { ResearchMap } from "@/components/research-map"
import AnnotationPanel from "@/components/annotation-panel"

const COPY_FEEDBACK_MS = 2000

// Share dialog — Esc dismiss, focus trap, hairline border, subtle-2 ring styling.
export function ShareDialog({ url, onClose }) {
  const [copied, setCopied] = useState(false)
  const [copyError, setCopyError] = useState(false)
  const timerRef = useRef(null)
  const inputRef = useRef(null)

  useEffect(() => {
    inputRef.current?.focus()
    return () => clearTimeout(timerRef.current)
  }, [])

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(url)
      setCopied(true)
      setCopyError(false)
      clearTimeout(timerRef.current)
      timerRef.current = setTimeout(() => setCopied(false), COPY_FEEDBACK_MS)
    } catch (err) {
      console.error("Copy share link failed", err)
      setCopyError(true)
      inputRef.current?.select()
    }
  }

  return (
    <Dialog open onOpenChange={(open) => { if (!open) onClose() }}>
      <DialogContent>
        <div className="flex items-center gap-2 mb-4">
          <Link2 className="h-4 w-4 text-[#2563eb]" />
          <DialogTitle className="text-sm font-semibold text-[#0a0a0a]">Share this summary</DialogTitle>
        </div>
        <DialogDescription className="mb-3 text-[11px] text-[#737373]">
          Anyone with the link can view it for 7 days of inactivity.
        </DialogDescription>
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
          <p className="mt-2 text-[11px] text-[#ea580c]">Couldn&apos;t copy automatically. Select the link and press Ctrl+C / Cmd+C.</p>
        )}
      </DialogContent>
    </Dialog>
  )
}

// Render research-map + annotation overlays — toggled from UnderstandingSection, housed here to thin page.
export function ResultOverlays({ documentId, showResearchMap, setShowResearchMap, showAnnotations, setShowAnnotations }) {
  return (
    <>
      {showResearchMap && (
        <div className="mb-6">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <Network className="h-4 w-4 text-[#2563eb]" />
              <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">Research Map</h2>
            </div>
            <button onClick={() => setShowResearchMap(false)} className="text-xs text-[#737373] hover:text-[#0a0a0a] transition-colors">Close</button>
          </div>
          <ResearchMap documentId={documentId} />
        </div>
      )}
      {showAnnotations && (
        <div className="mb-6 rounded-[12px] border border-[#e5e5e5] bg-white p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <StickyNote className="h-4 w-4 text-[#92400e]" />
              <h3 className="text-sm font-medium text-[#0a0a0a]">Personal Notes</h3>
            </div>
            <button onClick={() => setShowAnnotations(false)} className="text-xs text-[#737373] hover:text-[#0a0a0a]">Close</button>
          </div>
          <AnnotationPanel documentId={documentId} />
        </div>
      )}
    </>
  )
}

// Render management actions — collections, export, navigation, DESIGN.md compact density.
export function ResultActions({
  documentId,
  collections = [],
  collectionsError = false,
  onRetryCollections,
  collectionAddError = null,
  onRetryCollectionAdd,
  showAddToCollection,
  setShowAddToCollection,
  onAddToCollection,
}) {
  const [exportError, setExportError] = useState(false)

  // handleExport downloads research context; failure surfaces inline with Retry.
  async function handleExport() {
    setExportError(false)
    try {
      await exportResearchContext(documentId)
    } catch (err) {
      console.error("export failed", err)
      setExportError(true)
    }
  }

  return (
    <div className="mt-12 rounded-[12px] border border-[#e5e5e5] bg-[#fafafa] p-6">
      <h3 className="text-sm font-medium text-[#0a0a0a] mb-3">What&apos;s next?</h3>
      <div className="flex flex-wrap gap-3">
        <div
          className="relative"
          onKeyDown={(e) => { if (e.key === "Escape") setShowAddToCollection(false) }}
        >
          <button
            onClick={() => setShowAddToCollection(!showAddToCollection)}
            aria-expanded={showAddToCollection}
            aria-haspopup="menu"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-[#737373] bg-white border border-[#e5e5e5] rounded-full hover:bg-[#f5f5f5] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20"
          >
            <FolderPlus className="h-3.5 w-3.5" /> Add to Collection
          </button>
          {showAddToCollection && (
            <div role="menu" className="absolute left-0 top-full mt-1 z-10 bg-white border border-[#e5e5e5] rounded-[8px] shadow-lg py-1 w-48">
              {collectionsError ? (
                <div className="px-3 py-2 text-xs text-[#737373]">
                  <p>Couldn&apos;t load collections.</p>
                  <button
                    type="button"
                    onClick={onRetryCollections}
                    className="mt-1 font-medium text-[#2563eb] hover:text-[#1e40af] transition-colors"
                  >
                    Retry
                  </button>
                </div>
              ) : collections.length === 0 ? (
                <div className="px-3 py-2 text-xs text-[#737373]">No collections yet</div>
              ) : (
                collections.map((col) => (
                  <button
                    key={col.id}
                    role="menuitem"
                    onClick={() => onAddToCollection(col.id)}
                    className="w-full text-left px-3 py-1.5 text-xs hover:bg-[#f5f5f5] flex items-center gap-2 focus-visible:outline-none focus-visible:bg-[#f5f5f5]"
                  >
                    {col.name}
                  </button>
                ))
              )}
            </div>
          )}
        </div>
        <Link
          to="/"
          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-[#0a0a0a] rounded-full hover:bg-[#262626] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20"
        >
          Upload Another <ArrowRight className="h-3.5 w-3.5" />
        </Link>
        <Link
          to="/account"
          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-[#737373] bg-white border border-[#e5e5e5] rounded-full hover:bg-[#f5f5f5] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20"
        >
          <LayoutDashboard className="h-3.5 w-3.5" /> View Account
        </Link>
        <button
          onClick={handleExport}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-[#737373] bg-white border border-[#e5e5e5] rounded-full hover:bg-[#f5f5f5] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20"
        >
          Export Research Data
        </button>
      </div>
      {collectionAddError && (
        <p role="alert" className="mt-3 text-[11px] text-[#ea580c]">
          {collectionAddError}{" "}
          <button type="button" onClick={onRetryCollectionAdd} className="font-medium text-[#2563eb] hover:text-[#1e40af] transition-colors">
            Retry
          </button>
        </p>
      )}
      {exportError && (
        <p role="alert" className="mt-3 text-[11px] text-[#ea580c]">
          Couldn&apos;t export research data.{" "}
          <button type="button" onClick={handleExport} className="font-medium text-[#2563eb] hover:text-[#1e40af] transition-colors">
            Retry
          </button>
        </p>
      )}
    </div>
  )
}
