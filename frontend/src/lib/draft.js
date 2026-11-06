// draft.js — session-scoped ingest draft + allowlisted post-login destinations
// for the guest journey. File bytes are never stored: a PDF draft restores
// only mode/readingLevel and asks for re-pick.

// Session-storage key for the unsaved ingest draft (survives the login trip).
const DRAFT_KEY = "paperviz_ingest_draft"

// Internal-only destinations a guest may be sent back to after sign-in.
const ALLOWED_RETURN_TO = ["/", "/sample"]

// loadDraft reads and parses the saved draft, or null when absent/corrupt.
export function loadDraft() {
  try {
    const raw = sessionStorage.getItem(DRAFT_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    if (!parsed || typeof parsed !== "object") return null
    // Unvalidated mode would crash INGEST_TABS.find(...).source in useIngest;
    // a poisoned draft must be dropped, not rendered into a white screen.
    if (!["pdf", "paste", "doi", "url"].includes(parsed.mode)) return null
    return parsed
  } catch (err) {
    console.error("loadDraft failed", err)
    return null
  }
}

// saveDraft persists draft fields (never a File) for post-login restore.
export function saveDraft(draft) {
  try {
    sessionStorage.setItem(DRAFT_KEY, JSON.stringify(draft))
  } catch (err) {
    console.error("saveDraft failed", err)
  }
}

// clearDraft drops the draft after a successful submit (clear-on-success only).
export function clearDraft() {
  try {
    sessionStorage.removeItem(DRAFT_KEY)
  } catch (err) {
    console.error("clearDraft failed", err)
  }
}

// resolveReturnTo allowlists a post-login destination; anything else (incl.
// absolute or protocol-relative URLs) falls back, so no open redirect.
export function resolveReturnTo(value, fallback = "/account") {
  return ALLOWED_RETURN_TO.includes(value) ? value : fallback
}
