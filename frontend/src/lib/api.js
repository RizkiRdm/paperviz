// api.js — thin wrapper around the two backend endpoints from
// ARCHITECTURE.md Section E ("Internal Contracts"). Deliberately not a
// generic HTTP client abstraction: there are only two endpoints in this
// whole app (POST /api/documents, GET /api/documents/:id), so two focused
// functions are simpler to read and maintain than a configurable client
// class would be here (YAGNI).
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

  const response = await fetch("/api/documents", {
    method: "POST",
    body: formData,
  })

  if (!response.ok) {
    throw await parseErrorResponse(response)
  }

  return response.json() // { document_id, status }
}

// getDocument polls a document's current state. Per ARCHITECTURE.md
// Section E, the caller MUST wait at least 2 seconds between calls while
// status is "processing" — that interval is enforced in ResultPage.jsx's
// polling loop, not here, since this function is just the single-request
// primitive.
export async function getDocument(id) {
  const response = await fetch(`/api/documents/${id}`)

  if (!response.ok) {
    throw await parseErrorResponse(response)
  }

  return response.json()
}
