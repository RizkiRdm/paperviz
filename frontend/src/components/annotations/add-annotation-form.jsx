// Inline form for adding a new annotation.
import { useState } from "react"
import { createAnnotation } from "@/lib/api"

// AddAnnotationForm creates one annotation; onSave hands it back for the cache.
export function AddAnnotationForm({ documentId, onSave, onCancel }) {
  const [targetType, setTargetType] = useState("paper")
  const [targetId, setTargetId] = useState("")
  const [content, setContent] = useState("")
  const [saving, setSaving] = useState(false)
  const [createError, setCreateError] = useState(null)

  // handleSubmit posts the new annotation; error keeps the raw code visible.
  async function handleSubmit(e) {
    e.preventDefault()
    if (!content.trim() || saving) return
    setSaving(true)
    setCreateError(null)
    try {
      const annotation = await createAnnotation(documentId, {
        targetType,
        targetId: targetType === "claim" ? targetId.trim() : null,
        content: content.trim(),
      })
      onSave(annotation)
    } catch (err) {
      console.error("create annotation failed:", err)
      setCreateError(err && err.code ? `Save failed (${err.code}), retry` : "Save failed, retry")
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
            onChange={(e) => { setContent(e.target.value); setCreateError(null) }}
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
      {createError && (
        <p className="mt-1.5 text-[11px] text-[#ea580c]">
          {createError}
        </p>
      )}
    </form>
  )
}
