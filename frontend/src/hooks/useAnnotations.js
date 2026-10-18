// Annotation list hook — fetch + local cache of a document's annotations.
// Form open/close and collapse state stay in the panel component.
import { useState, useEffect, useCallback } from "react"
import { listAnnotations } from "@/lib/api"

export function useAnnotations(documentId) {
  const [annotations, setAnnotations] = useState([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(null)

  // fetchAnnotations loads (or reloads) the annotation list for this document.
  const fetchAnnotations = useCallback(async () => {
    try {
      const data = await listAnnotations(documentId)
      setAnnotations(data.annotations || [])
    } catch (err) {
      console.error("load annotations failed:", err)
      setLoadError("Couldn't load annotations")
    } finally {
      setLoading(false)
    }
  }, [documentId])

  useEffect(() => {
    fetchAnnotations()
  }, [fetchAnnotations])

  // retry clears the error and puts the panel back into its loading state.
  function retry() {
    setLoadError(null)
    setLoading(true)
    fetchAnnotations()
  }

  // handleSave appends a freshly created annotation.
  function handleSave(annotation) {
    setAnnotations((prev) => [...prev, annotation])
  }

  // handleUpdate patches one annotation's content after an inline edit.
  function handleUpdate(id, content) {
    setAnnotations((prev) =>
      prev.map((a) => (a.id === id ? { ...a, content } : a))
    )
  }

  // handleDelete drops a deleted annotation from the local cache.
  function handleDelete(id) {
    setAnnotations((prev) => prev.filter((a) => a.id !== id))
  }

  return { annotations, loading, loadError, retry, handleSave, handleUpdate, handleDelete }
}
