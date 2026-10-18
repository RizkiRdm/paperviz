// Single collection row with expand and two-step remove-current-document.
import { useState } from "react"
import { removeDocumentFromCollection } from "@/lib/api"

// Trash icon for remove.
function TrashIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-3.5 w-3.5" aria-hidden="true">
      <path d="M3 6h18" /><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" /><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
    </svg>
  )
}

// CollectionRow renders one collection; onRemove updates the parent's cache.
export function CollectionRow({ collection, documentId, selected, onSelect, onRemove }) {
  const [confirmRemove, setConfirmRemove] = useState(false)
  const [removeError, setRemoveError] = useState(null)
  const containsCurrent = (collection.documents || []).some((d) => d.id === documentId || d === documentId)

  // handleRemove is two-step: first click arms, second click executes.
  async function handleRemove() {
    if (!confirmRemove) {
      setConfirmRemove(true)
      return
    }
    try {
      await removeDocumentFromCollection(collection.id, documentId)
      onRemove(collection.id)
      setConfirmRemove(false)
      setRemoveError(null)
    } catch (err) {
      console.error("remove document from collection failed:", err)
      setRemoveError(err && err.code ? `Remove failed (${err.code}), retry` : "Remove failed, retry")
      setConfirmRemove(false)
    }
  }

  return (
    <div className="rounded-[8px] border border-[#e5e5e5] bg-[#f5f5f5] px-3 py-2.5">
      <div className="flex items-start justify-between gap-2">
        <button
          type="button"
          onClick={() => onSelect(collection.id)}
          className="min-w-0 flex-1 text-left cursor-pointer"
          aria-label={selected ? `Collapse ${collection.name}` : `Expand ${collection.name}`}
          aria-expanded={selected}
        >
          <span className="block truncate text-[14px] font-medium text-[#0a0a0a]">
            {collection.name}
          </span>
          <span className="mt-0.5 inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2 py-0.5 text-[11px] font-medium text-[#737373]">
            {collection.document_count ?? (collection.documents ? collection.documents.length : 0)} saved
          </span>
        </button>
        {containsCurrent && (
          <button
            type="button"
            onClick={handleRemove}
            className={`inline-flex items-center justify-center rounded-[6px] p-1 transition-colors cursor-pointer shrink-0 ${
              confirmRemove
                ? "text-[#ea580c] bg-white border border-[#e5e5e5]"
                : "text-[#737373] hover:text-[#ea580c] hover:bg-white"
            }`}
            aria-label={confirmRemove ? "Confirm remove from collection" : "Remove from collection"}
          >
            <TrashIcon />
          </button>
        )}
      </div>

      {confirmRemove && containsCurrent && (
        <p className="mt-1.5 text-[11px] text-[#ea580c]">
          Click remove again to confirm.
        </p>
      )}

      {removeError && (
        <p className="mt-1.5 text-[11px] text-[#ea580c]" role="alert">
          {removeError}
        </p>
      )}

      {selected && collection.documents && collection.documents.length > 0 && (
        <ul className="mt-2 flex flex-col gap-1">
          {collection.documents.map((doc) => (
            <li key={typeof doc === "string" ? doc : doc.id} className="truncate text-[12px] text-[#737373]">
              {typeof doc === "string" ? doc : doc.title || doc.id}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
