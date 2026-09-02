// Collapsible annotation panel for managing user notes on a document.
import { useState, useEffect, useCallback } from "react"
import { listAnnotations, createAnnotation, updateAnnotation, deleteAnnotation } from "@/lib/api"

// Small pen/note SVG icon for the header.
export function NoteIcon({ className = "h-4 w-4" }) {
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
      <path d="M17 3a2.85 2.85 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
      <path d="m15 5 4 4" />
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

// Trash icon for delete.
function TrashIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-3.5 w-3.5" aria-hidden="true">
      <path d="M3 6h18" /><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" /><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
    </svg>
  )
}

// Pencil icon for edit.
function PencilIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-3.5 w-3.5" aria-hidden="true">
      <path d="M17 3a2.85 2.85 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
    </svg>
  )
}

// Single annotation row with inline edit and delete.
function AnnotationRow({ annotation, documentId, onUpdate, onDelete }) {
  const [editing, setEditing] = useState(false)
  const [editContent, setEditContent] = useState(annotation.content)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [saving, setSaving] = useState(false)

  async function handleSave() {
    if (!editContent.trim() || saving) return
    setSaving(true)
    try {
      await updateAnnotation(documentId, annotation.id, editContent.trim())
      onUpdate(annotation.id, editContent.trim())
      setEditing(false)
    } catch {
      // error handled silently — row stays in edit state
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!confirmDelete) {
      setConfirmDelete(true)
      return
    }
    try {
      await deleteAnnotation(documentId, annotation.id)
      onDelete(annotation.id)
    } catch {
      // error handled silently
    }
  }

  const targetTypeLabel = annotation.target_type === "claim" ? "Claim" : "Paper"

  return (
    <div className="rounded-[8px] border border-[#e5e5e5] bg-[#f5f5f5] px-3 py-2.5">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <span className="inline-flex items-center rounded-full border border-[#e5e5e5] bg-white px-2 py-0.5 text-[11px] font-medium text-[#737373]">
            {targetTypeLabel}
          </span>
          {annotation.target_id && (
            <span className="ml-1.5 text-[11px] text-[#a3a3a3]">
              #{annotation.target_id}
            </span>
          )}
        </div>
        <div className="flex items-center gap-1 shrink-0">
          {!editing && (
            <button
              type="button"
              onClick={() => setEditing(true)}
              className="inline-flex items-center justify-center rounded-[6px] p-1 text-[#737373] hover:text-[#0a0a0a] hover:bg-[#e5e5e5] transition-colors cursor-pointer"
              aria-label="Edit annotation"
            >
              <PencilIcon />
            </button>
          )}
          <button
            type="button"
            onClick={handleDelete}
            className={`inline-flex items-center justify-center rounded-[6px] p-1 transition-colors cursor-pointer ${
              confirmDelete
                ? "text-[#ea580c] bg-[#fff7ed]"
                : "text-[#737373] hover:text-[#ea580c] hover:bg-[#fff7ed]"
            }`}
            aria-label={confirmDelete ? "Confirm delete" : "Delete annotation"}
          >
            <TrashIcon />
          </button>
        </div>
      </div>

      {editing ? (
        <div className="mt-2">
          <textarea
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            rows={3}
            className="w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-[14px] text-[#111827] leading-[1.43] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 resize-none"
          />
          <div className="mt-1.5 flex items-center gap-2">
            <button
              type="button"
              onClick={handleSave}
              disabled={saving || !editContent.trim()}
              className="inline-flex items-center rounded-[8px] bg-[#000000] px-3 py-1.5 text-[13px] font-medium text-white hover:bg-[#262626] transition-colors cursor-pointer disabled:opacity-50"
            >
              {saving ? "Saving..." : "Save"}
            </button>
            <button
              type="button"
              onClick={() => { setEditing(false); setEditContent(annotation.content) }}
              className="inline-flex items-center rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-[13px] font-medium text-[#525252] hover:text-[#0a0a0a] transition-colors cursor-pointer"
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <p className="mt-1.5 text-[14px] leading-[1.43] text-[#262626] whitespace-pre-wrap">
          {annotation.content}
        </p>
      )}

      {confirmDelete && !editing && (
        <p className="mt-1.5 text-[11px] text-[#ea580c]">
          Click delete again to confirm.
        </p>
      )}
    </div>
  )
}

// Inline form for adding a new annotation.
function AddAnnotationForm({ documentId, onSave, onCancel }) {
  const [targetType, setTargetType] = useState("paper")
  const [targetId, setTargetId] = useState("")
  const [content, setContent] = useState("")
  const [saving, setSaving] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    if (!content.trim() || saving) return
    setSaving(true)
    try {
      const annotation = await createAnnotation(documentId, {
        targetType,
        targetId: targetType === "claim" ? targetId.trim() : null,
        content: content.trim(),
      })
      onSave(annotation)
    } catch {
      // error handled silently — form stays open
    } finally {
      setSaving(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="rounded-[8px] border border-[#e5e5e5] bg-[#f5f5f5] p-3">
      <div className="flex flex-col gap-2">
        <label className="text-[12px] font-medium text-[#525252]">
          Target type
          <select
            value={targetType}
            onChange={(e) => setTargetType(e.target.value)}
            className="mt-1 block w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-[14px] text-[#111827] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 cursor-pointer"
          >
            <option value="paper">Paper-level note</option>
            <option value="claim">Claim note</option>
          </select>
        </label>

        {targetType === "claim" && (
          <label className="text-[12px] font-medium text-[#525252]">
            Claim ID
            <input
              type="text"
              value={targetId}
              onChange={(e) => setTargetId(e.target.value)}
              placeholder="e.g. claim-3"
              className="mt-1 block w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-[14px] text-[#111827] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20"
            />
          </label>
        )}

        <label className="text-[12px] font-medium text-[#525252]">
          Content
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={3}
            placeholder="Write your annotation..."
            className="mt-1 block w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-[14px] text-[#111827] placeholder:text-[#a3a3a3] leading-[1.43] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 resize-none"
          />
        </label>
      </div>

      <div className="mt-2.5 flex items-center gap-2">
        <button
          type="submit"
          disabled={saving || !content.trim()}
          className="inline-flex items-center rounded-[8px] bg-[#000000] px-3 py-1.5 text-[13px] font-medium text-white hover:bg-[#262626] transition-colors cursor-pointer disabled:opacity-50"
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

// Main annotation panel component.
export default function AnnotationPanel({ documentId }) {
  const [open, setOpen] = useState(true)
  const [annotations, setAnnotations] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [showForm, setShowForm] = useState(false)

  const fetchAnnotations = useCallback(async () => {
    try {
      const data = await listAnnotations(documentId)
      setAnnotations(data.annotations || [])
    } catch (err) {
      setError(err.message || "Failed to load annotations")
    } finally {
      setLoading(false)
    }
  }, [documentId])

  useEffect(() => {
    fetchAnnotations()
  }, [fetchAnnotations])

  function handleSave(annotation) {
    setAnnotations((prev) => [...prev, annotation])
    setShowForm(false)
  }

  function handleUpdate(id, content) {
    setAnnotations((prev) =>
      prev.map((a) => (a.id === id ? { ...a, content } : a))
    )
  }

  function handleDelete(id) {
    setAnnotations((prev) => prev.filter((a) => a.id !== id))
  }

  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white">
      {/* Header — always visible, toggles collapse */}
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center justify-between px-3 py-2.5 text-left cursor-pointer hover:bg-[#f5f5f5] transition-colors rounded-[12px]"
        aria-expanded={open}
      >
        <div className="flex items-center gap-2">
          <NoteIcon className="h-4 w-4 text-[#525252]" />
          <span className="text-[14px] font-medium text-[#262626]">
            My Annotations
          </span>
          {!loading && annotations.length > 0 && (
            <span className="inline-flex items-center justify-center rounded-full bg-[#f5f5f5] border border-[#e5e5e5] px-1.5 py-0.5 text-[11px] font-medium text-[#737373] min-w-[18px]">
              {annotations.length}
            </span>
          )}
        </div>
        <ChevronIcon open={open} />
      </button>

      {/* Collapsible body */}
      {open && (
        <div className="border-t border-[#e5e5e5] px-3 pb-3 pt-2.5">
          {/* Loading state */}
          {loading && (
            <div className="flex items-center gap-2 py-4 justify-center">
              <div className="h-4 w-4 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent" />
              <span className="text-[12px] text-[#737373]">Loading annotations...</span>
            </div>
          )}

          {/* Error state */}
          {!loading && error && (
            <p className="py-4 text-center text-[12px] text-[#ea580c]">
              {error}
            </p>
          )}

          {/* Empty state */}
          {!loading && !error && annotations.length === 0 && !showForm && (
            <p className="py-4 text-center text-[13px] text-[#737373]">
              No annotations yet
            </p>
          )}

          {/* Annotation list */}
          {!loading && !error && annotations.length > 0 && (
            <div className="flex flex-col gap-2">
              {annotations.map((a) => (
                <AnnotationRow
                  key={a.id}
                  annotation={a}
                  documentId={documentId}
                  onUpdate={handleUpdate}
                  onDelete={handleDelete}
                />
              ))}
            </div>
          )}

          {/* Add note form or button */}
          <div className="mt-2.5">
            {showForm ? (
              <AddAnnotationForm
                documentId={documentId}
                onSave={handleSave}
                onCancel={() => setShowForm(false)}
              />
            ) : (
              <button
                type="button"
                onClick={() => setShowForm(true)}
                className="inline-flex items-center gap-1.5 rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-[13px] font-medium text-[#525252] hover:text-[#0a0a0a] hover:border-[#d4d4d4] transition-colors cursor-pointer w-full justify-center"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-3.5 w-3.5" aria-hidden="true">
                  <path d="M12 5v14" /><path d="M5 12h14" />
                </svg>
                Add Note
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
