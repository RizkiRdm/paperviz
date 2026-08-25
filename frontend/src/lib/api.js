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
