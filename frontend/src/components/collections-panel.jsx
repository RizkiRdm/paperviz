// Collapsible collections panel for saving documents into named collections.
import { useState, useEffect, useCallback } from "react"
import { listCollections, createCollection, getCollection, addDocumentToCollection, removeDocumentFromCollection } from "@/lib/api"

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

// Trash icon for remove.
function TrashIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-3.5 w-3.5" aria-hidden="true">
      <path d="M3 6h18" /><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" /><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
    </svg>
  )
}

// Single collection row with expand and remove-current-document.
function CollectionRow({ collection, documentId, selected, onSelect, onRemove }) {
  const [confirmRemove, setConfirmRemove] = useState(false)
  const [removeError, setRemoveError] = useState(null)
  const containsCurrent = (collection.documents || []).some((d) => d.id === documentId || d === documentId)

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

// Inline form for creating a new collection.
function AddCollectionForm({ onSave, onCancel }) {
  const [name, setName] = useState("")
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState(null)

  async function handleSubmit(e) {
    e.preventDefault()
    if (!name.trim() || saving) return
    setSaving(true)
    setFormError(null)
    try {
      const collection = await createCollection(name.trim())
      onSave(collection)
    } catch (err) {
      console.error("create collection failed:", err)
      setFormError(err && err.code ? `Create failed (${err.code}), retry` : "Create failed, retry")
    } finally {
      setSaving(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="rounded-[8px] border border-[#e5e5e5] bg-[#f5f5f5] p-3">
      <label className="text-[12px] font-medium text-[#525252]">
        Collection name
        <input
          type="text"
          value={name}
          onChange={(e) => { setName(e.target.value); setFormError(null) }}
          placeholder="e.g. Thesis background"
          aria-label="Collection name"
          className="mt-1 block w-full rounded-[6px] border border-[#0a0a0a] bg-white px-3 py-2 text-[14px] text-[#0a0a0a] placeholder:text-[#737373] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20"
        />
      </label>
      {formError && (
        <p className="mt-1.5 text-[12px] text-[#ea580c]" role="alert">
          {formError}
        </p>
      )}
      <div className="mt-2.5 flex items-center gap-2">
        <button
          type="submit"
          disabled={saving || !name.trim()}
          className="inline-flex items-center rounded-[8px] bg-[#0a0a0a] px-3 py-1.5 text-[13px] font-medium text-white hover:bg-[#262626] transition-colors cursor-pointer disabled:opacity-50"
        >
          {saving ? "Saving..." : "Save"}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="inline-flex items-center rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-[13px] font-medium text-[#525252] hover:text-[#0a0a0a] transition-colors cursor-pointer"
        >
          Cancel
        </button>
      </div>
    </form>
  )
}

// Main collections panel component.
export default function CollectionsPanel({ documentId }) {
  const [open, setOpen] = useState(true)
  const [collections, setCollections] = useState([])
  const [selected, setSelected] = useState(null)
  const [selectedDetail, setSelectedDetail] = useState(null)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(null)
  const [showForm, setShowForm] = useState(false)
  const [actionError, setActionError] = useState(null)
  const [adding, setAdding] = useState(false)

  const fetchCollections = useCallback(async () => {
    try {
      const data = await listCollections()
      setCollections(data.collections || [])
    } catch (err) {
      console.error("load collections failed:", err)
      setLoadError("Couldn't load collections")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchCollections()
  }, [fetchCollections])

  async function handleSelect(id) {
    if (selected === id) {
      setSelected(null)
      setSelectedDetail(null)
      return
    }
    setSelected(id)
    try {
      const detail = await getCollection(id)
      setSelectedDetail(detail)
      setCollections((prev) =>
        prev.map((c) => (c.id === id ? { ...c, documents: detail.documents || [] } : c))
      )
    } catch (err) {
      console.error("load collection detail failed:", err)
      setActionError(err && err.code ? `Load failed (${err.code}), retry` : "Load failed, retry")
    }
  }

  function handleCreate(collection) {
    setCollections((prev) => [...prev, collection])
    setShowForm(false)
  }

  async function handleAddCurrent() {
    if (!selected || adding) return
    setAdding(true)
    setActionError(null)
    try {
      await addDocumentToCollection(selected, documentId)
      const detail = await getCollection(selected)
      setSelectedDetail(detail)
      setCollections((prev) =>
        prev.map((c) =>
          c.id === selected
            ? { ...c, documents: detail.documents || [], document_count: (detail.documents || []).length }
            : c
        )
      )
    } catch (err) {
      console.error("add document to collection failed:", err)
      setActionError(err && err.code ? `Add failed (${err.code}), retry` : "Add failed, retry")
    } finally {
      setAdding(false)
    }
  }

  function handleRemove(collectionId) {
    setCollections((prev) =>
      prev.map((c) =>
        c.id === collectionId
          ? {
              ...c,
              documents: (c.documents || []).filter((d) =>
                typeof d === "string" ? d !== documentId : d.id !== documentId
              ),
            }
          : c
      )
    )
    if (selectedDetail && selectedDetail.id === collectionId) {
      setSelectedDetail((prev) => ({
        ...prev,
        documents: (prev.documents || []).filter((d) =>
          typeof d === "string" ? d !== documentId : d.id !== documentId
        ),
      }))
    }
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
                onClick={() => { setLoadError(null); setLoading(true); fetchCollections() }}
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
                onSave={handleCreate}
                onCancel={() => setShowForm(false)}
              />
            ) : (
              <button
                type="button"
                onClick={() => setShowForm(true)}
                className="inline-flex items-center gap-1.5 rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-[13px] font-medium text-[#525252] hover:text-[#0a0a0a] transition-colors cursor-pointer w-full justify-center"
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
