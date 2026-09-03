// api.js — thin wrapper around the backend endpoints from
// ARCHITECTURE.md Section E ("Internal Contracts"). Deliberately not a
// generic HTTP client abstraction: the endpoint surface stays small, so
// focused functions are simpler to read and maintain than a configurable
// client class would be here (YAGNI).
//
// Both functions throw an ApiError with the server's snake_case error code
// attached, so callers can show a specific message per ARCHITECTURE.md's
// error contract ({ "error": "<snake_case_code>" }) instead of a generic
// "something went wrong."

export class ApiError extends Error {
  constructor(code, status) {
    super(code)
    this.code = code
    this.status = status
  }
}

// Timeout constants (configurable per spec)
const CREATE_TIMEOUT_MS = 30000 // 30 seconds
const POLL_TIMEOUT_MS = 10000 // 10 seconds

async function parseErrorResponse(response) {
  try {
    const body = await response.json()
    return new ApiError(body.error || "unknown_error", response.status)
  } catch {
    return new ApiError("unknown_error", response.status)
  }
}

// createDocument uploads either a PDF file or pasted text, plus the chosen
// reading level. Exactly one of (file, text) must be provided — the
// backend enforces this too (ARCHITECTURE.md API Contract), but we also
// validate client-side in UploadPage so the person gets instant feedback
// instead of a round trip.
export async function createDocument({ file, text, readingLevel }) {
  const formData = new FormData()
  formData.append("reading_level", readingLevel)
  if (file) {
    formData.append("file", file)
  } else {
    formData.append("text", text)
  }

  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), CREATE_TIMEOUT_MS)
  try {
    const response = await fetch("/api/documents", {
      method: "POST",
      body: formData,
      signal: controller.signal,
    })

    if (!response.ok) {
      throw await parseErrorResponse(response)
    }

    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// importByDOI imports a paper by its DOI identifier.
export async function importByDOI(doi) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), CREATE_TIMEOUT_MS)
  try {
    const response = await fetch("/api/import/doi", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ doi }),
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// importByURL imports a paper from a URL (arxiv, pubmed, or direct PDF).
export async function importByURL(url) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), CREATE_TIMEOUT_MS)
  try {
    const response = await fetch("/api/import/url", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url }),
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// getDocument polls a document's current state. Per ARCHITECTURE.md
// Section E, the caller MUST wait at least 2 seconds between calls while
// status is "processing" — that interval is enforced in ResultPage.jsx's
// polling loop, not here, since this function is just the single-request
// primitive.
export async function getDocument(id) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), POLL_TIMEOUT_MS)
  try {
    const response = await fetch(`/api/documents/${id}`, {
      signal: controller.signal
    })
    clearTimeout(timeoutId)
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// generateDocumentShare creates (or refreshes) the document's share link
// and resolves to { share_url } pointing at /share/doc/:token.
export async function generateDocumentShare(documentId) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), POLL_TIMEOUT_MS)
  try {
    const response = await fetch(`/api/documents/${documentId}/share`, {
      method: "POST",
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// revokeDocumentShare deletes the document's active share link; the 200
// body is intentionally ignored.
export async function revokeDocumentShare(documentId) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), POLL_TIMEOUT_MS)
  try {
    const response = await fetch(`/api/documents/${documentId}/share`, {
      method: "DELETE",
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
  } finally {
    clearTimeout(timeoutId)
  }
}

// updateDocumentVisibility switches who may open the document (private,
// unlisted, or public) and resolves to { visibility }.
export async function updateDocumentVisibility(documentId, visibility) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), POLL_TIMEOUT_MS)
  try {
    const response = await fetch(`/api/documents/${documentId}/visibility`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ visibility }),
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// trackReferral attributes an upload to a shared-link referral token.
// Fire-and-forget by design: every failure is swallowed here so callers
// never see an error and the upload flow is never blocked or delayed.
export async function trackReferral(ref) {
  try {
    await fetch("/api/share-referrals", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ref }),
    })
  } catch {
    // intentionally ignored — referral tracking is best-effort
  }
}

// listAnnotations returns the current user's annotations for a document.
export async function listAnnotations(documentId) {
  const response = await fetch(`/api/documents/${documentId}/annotations`)
  if (!response.ok) throw await parseErrorResponse(response)
  return response.json()
}

// createAnnotation adds a new annotation to a document.
export async function createAnnotation(documentId, { targetType, targetId, content }) {
  const response = await fetch(`/api/documents/${documentId}/annotations`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ target_type: targetType, target_id: targetId, content }),
  })
  if (!response.ok) throw await parseErrorResponse(response)
  return response.json()
}

// updateAnnotation changes the content of an existing annotation.
export async function updateAnnotation(documentId, annotationId, content) {
  const response = await fetch(`/api/documents/${documentId}/annotations/${annotationId}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ content }),
  })
  if (!response.ok) throw await parseErrorResponse(response)
  return response.json()
}

// deleteAnnotation removes an annotation.
export async function deleteAnnotation(documentId, annotationId) {
  const response = await fetch(`/api/documents/${documentId}/annotations/${annotationId}`, {
    method: "DELETE",
  })
  if (!response.ok) throw await parseErrorResponse(response)
}

// listCollections returns all collections owned by current user.
export async function listCollections() {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), POLL_TIMEOUT_MS)
  try {
    const response = await fetch("/api/collections", {
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// createCollection creates a new collection with given name.
export async function createCollection(name) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), CREATE_TIMEOUT_MS)
  try {
    const response = await fetch("/api/collections", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// getCollection returns collection detail with documents.
export async function getCollection(id) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), POLL_TIMEOUT_MS)
  try {
    const response = await fetch(`/api/collections/${id}`, {
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// addDocumentToCollection adds a document to a collection.
export async function addDocumentToCollection(collectionId, documentId) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), CREATE_TIMEOUT_MS)
  try {
    const response = await fetch(`/api/collections/${collectionId}/documents`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ document_id: documentId }),
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// removeDocumentFromCollection removes a document from a collection.
export async function removeDocumentFromCollection(collectionId, docId) {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), CREATE_TIMEOUT_MS)
  try {
    const response = await fetch(`/api/collections/${collectionId}/documents/${docId}`, {
      method: "DELETE",
      signal: controller.signal,
    })
    if (!response.ok) {
      throw await parseErrorResponse(response)
    }
    return response.json()
  } finally {
    clearTimeout(timeoutId)
  }
}

// exportResearchContext downloads the full research context as JSON.
export async function exportResearchContext(documentId) {
  const response = await fetch(`/api/documents/${documentId}/export`)
  if (!response.ok) throw await parseErrorResponse(response)
  const data = await response.json()
  // Trigger download
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" })
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = `research-context-${documentId}.json`
  a.click()
  URL.revokeObjectURL(url)
  return data
}
