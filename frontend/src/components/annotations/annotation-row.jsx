// Single annotation row with inline edit and two-step delete.
import { useState } from "react"
import { updateAnnotation, deleteAnnotation } from "@/lib/api"

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

// AnnotationRow renders one note; onUpdate/onDelete update the parent's cache.
export function AnnotationRow({ annotation, documentId, onUpdate, onDelete }) {
  const [editing, setEditing] = useState(false)
  const [editContent, setEditContent] = useState(annotation.content)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState(null)
  const [deleteError, setDeleteError] = useState(null)

  // handleSave persists the inline edit; error text keeps the raw code visible.
  async function handleSave() {
    if (!editContent.trim() || saving) return
    setSaving(true)
    setSaveError(null)
    try {
      await updateAnnotation(documentId, annotation.id, editContent.trim())
      onUpdate(annotation.id, editContent.trim())
      setEditing(false)
    } catch (err) {
      console.error("save annotation failed:", err)
      setSaveError(err && err.code ? `Save failed (${err.code}), retry` : "Save failed, retry")
    } finally {
      setSaving(false)
    }
  }

  // handleDelete is two-step: first click arms, second click executes.
  async function handleDelete() {
    if (!confirmDelete) {
      setConfirmDelete(true)
      return
    }
    setDeleteError(null)
    try {
      await deleteAnnotation(documentId, annotation.id)
      onDelete(annotation.id)
    } catch (err) {
      console.error("delete annotation failed:", err)
      setDeleteError(err && err.code ? `Delete failed (${err.code}), retry` : "Delete failed, retry")
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
            onChange={(e) => { setEditContent(e.target.value); setSaveError(null) }}
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
          {saveError && (
            <p className="mt-1.5 text-[11px] text-[#ea580c]">
              {saveError}
            </p>
          )}
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
      {deleteError && !editing && (
        <p className="mt-1.5 text-[11px] text-[#ea580c]">
          {deleteError}
        </p>
      )}
    </div>
  )
}
