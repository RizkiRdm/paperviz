// Collapsible collections panel for saving documents into named collections.
// List logic lives in hooks/useCollections; rows/forms in components/collections.
import { useState } from "react"
import { useCollections } from "@/hooks/useCollections"
import { CollectionRow } from "@/components/collections/collection-row"
import { AddCollectionForm } from "@/components/collections/add-collection-form"

// Folder SVG icon for the header.
function FolderIcon({ className = "h-4 w-4" }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden="true"
    >
      <path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z" />
    </svg>
  )
}

// Chevron SVG for collapse toggle.
function ChevronIcon({ open }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={`h-4 w-4 transition-transform ${open ? "rotate-180" : ""}`}
      aria-hidden="true"
    >
      <path d="m6 9 6 6 6-6" />
    </svg>
  )
}

// Main collections panel component.
export default function CollectionsPanel({ documentId }) {
  const [open, setOpen] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const {
    collections,
    selected,
    loading,
    loadError,
    actionError,
    adding,
    retry,
    handleSelect,
    handleCreate,
    handleAddCurrent,
    handleRemove,
  } = useCollections(documentId)

  // handleCreate appends the new collection and closes the form.
  function handleCreateAndClose(collection) {
    handleCreate(collection)
    setShowForm(false)
  }

  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center justify-between px-3 py-2.5 text-left cursor-pointer hover:bg-[#f5f5f5] transition-colors rounded-[12px]"
        aria-expanded={open}
        aria-label="Toggle collections panel"
      >
        <div className="flex items-center gap-2">
          <FolderIcon className="h-4 w-4 text-[#525252]" />
          <span className="text-[14px] font-medium text-[#262626]">
            My Collections
          </span>
          {!loading && collections.length > 0 && (
            <span className="inline-flex items-center justify-center rounded-full bg-[#f5f5f5] border border-[#e5e5e5] px-1.5 py-0.5 text-[11px] font-medium text-[#737373] min-w-[18px]">
              {collections.length}
            </span>
          )}
        </div>
        <ChevronIcon open={open} />
      </button>

      {open && (
        <div className="border-t border-[#e5e5e5] px-3 pb-3 pt-2.5">
          {loading && (
            <div className="flex items-center gap-2 py-4 justify-center">
              <div className="h-4 w-4 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent" />
              <span className="text-[12px] text-[#737373]">Loading collections...</span>
            </div>
          )}

          {!loading && loadError && (
            <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
              <p className="text-[12px] text-[#737373]">Couldn't load collections</p>
              <button
                type="button"
                onClick={retry}
                className="mt-2.5 inline-flex items-center rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-[13px] font-medium text-[#525252] hover:text-[#0a0a0a] transition-colors cursor-pointer"
              >
                Retry
              </button>
            </div>
          )}

          {!loading && !loadError && collections.length === 0 && !showForm && (
            <p className="py-4 text-center text-[13px] text-[#737373]">
              No collections yet
            </p>
          )}

          {!loading && !loadError && collections.length > 0 && (
            <div className="flex flex-col gap-2">
              {collections.map((c) => (
                <CollectionRow
                  key={c.id}
                  collection={c}
                  documentId={documentId}
                  selected={selected === c.id}
                  onSelect={handleSelect}
                  onRemove={handleRemove}
                />
              ))}
            </div>
          )}

          {actionError && (
            <p className="mt-2 text-[12px] text-[#ea580c]" role="alert">
              {actionError}
            </p>
          )}

          {selected && !loading && !loadError && (
            <button
              type="button"
              onClick={handleAddCurrent}
              disabled={adding}
              className="mt-2.5 inline-flex items-center gap-1.5 rounded-[8px] bg-[#0a0a0a] px-3 py-1.5 text-[13px] font-medium text-white hover:bg-[#262626] transition-colors cursor-pointer w-full justify-center disabled:opacity-50"
              aria-label="Add current document to selected collection"
            >
              {adding ? "Adding..." : "Add current document here"}
            </button>
          )}

          <div className="mt-2.5">
            {showForm ? (
              <AddCollectionForm
                onSave={handleCreateAndClose}
                onCancel={() => setShowForm(false)}
              />
            ) : (
              <button
                type="button"
                onClick={() => setShowForm(true)}
                className="inline-flex items-center gap-1.5 rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-[13px] font-medium text-[#525252] hover:text-[#0a0a0a] hover:border-[#d4d4d4] transition-colors cursor-pointer w-full justify-center"
                aria-label="Create new collection"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-3.5 w-3.5" aria-hidden="true">
                  <path d="M12 5v14" /><path d="M5 12h14" />
                </svg>
                New Collection
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
