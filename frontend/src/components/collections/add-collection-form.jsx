// Inline form for creating a new collection.
import { useState } from "react"
import { createCollection } from "@/lib/api"

// AddCollectionForm creates one collection; onSave hands it back for the cache.
export function AddCollectionForm({ onSave, onCancel }) {
  const [name, setName] = useState("")
  const [saving, setSaving] = useState(false)
  const [formError, setFormError] = useState(null)

  // handleSubmit posts the new collection; error keeps the raw code visible.
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
