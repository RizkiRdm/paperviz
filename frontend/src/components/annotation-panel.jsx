// Collapsible annotation panel for managing user notes on a document.
// List logic lives in hooks/useAnnotations; rows/forms in components/annotations.
import { useState } from "react"
import { useAnnotations } from "@/hooks/useAnnotations"
import { AnnotationRow } from "@/components/annotations/annotation-row"
import { AddAnnotationForm } from "@/components/annotations/add-annotation-form"

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

// Main annotation panel component.
export default function AnnotationPanel({ documentId }) {
  const [open, setOpen] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const { annotations, loading, loadError, retry, handleSave, handleUpdate, handleDelete } =
    useAnnotations(documentId)

  // handleSave appends the new note and closes the form.
  function handleSaveAndClose(annotation) {
    handleSave(annotation)
    setShowForm(false)
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
          {!loading && loadError && (
            <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-5 text-center">
              <p className="text-[12px] text-[#737373]">Couldn't load annotations</p>
              <button
                type="button"
                onClick={retry}
                className="mt-2.5 inline-flex items-center rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-[13px] font-medium text-[#525252] hover:text-[#0a0a0a] transition-colors cursor-pointer"
              >
                Retry
              </button>
            </div>
          )}

          {/* Empty state */}
          {!loading && !loadError && annotations.length === 0 && !showForm && (
            <p className="py-4 text-center text-[13px] text-[#737373]">
              No annotations yet
            </p>
          )}

          {/* Annotation list */}
          {!loading && !loadError && annotations.length > 0 && (
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
                onSave={handleSaveAndClose}
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
