// Collections list hook — fetch, expand/detail, and add/remove-document
// mutations for a document's collections. Collapse/form state stays in the panel.
import { useState, useEffect, useCallback } from "react"
import { listCollections, getCollection, addDocumentToCollection } from "@/lib/api"

export function useCollections(documentId) {
  const [collections, setCollections] = useState([])
  const [selected, setSelected] = useState(null)
  const [selectedDetail, setSelectedDetail] = useState(null)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(null)
  const [actionError, setActionError] = useState(null)
  const [adding, setAdding] = useState(false)

  // fetchCollections loads the summary list (documents detail arrives on expand).
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

  // retry clears the error and puts the panel back into its loading state.
  function retry() {
    setLoadError(null)
    setLoading(true)
    fetchCollections()
  }

  // handleSelect toggles expansion and fetches the document detail on open.
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

  // handleCreate appends a newly created collection.
  function handleCreate(collection) {
    setCollections((prev) => [...prev, collection])
  }

  // handleAddCurrent saves the open document into the selected collection.
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

  // handleRemove drops the document from one collection's local cache.
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

  return {
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
  }
}
