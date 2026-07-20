// VerificationBadge — DESIGN.md "Specialty UI Elements > Verification
// Badge": small pill, radius-full, accent-verified-soft background,
// accent-verified text/icon. DESIGN.md Do's and Don'ts #4 is explicit:
// "Don't show a fake or placeholder verification badge before claim-diff
// actually passes" — so this component only ever renders when the caller
// already knows status === "complete" (see ResultPage.jsx), never as a
// default/loading state.
import { ShieldCheck } from "lucide-react"

export function VerificationBadge() {
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-accent-verified-soft px-3 py-1 text-caption text-accent-verified">
      <ShieldCheck className="h-3.5 w-3.5" aria-hidden="true" />
      Verified
    </span>
  )
}

// WarningBanner — the "explicit warning banner" required by ARCHITECTURE.md
// Acceptance Scenario 4 when claim-diff detects a mismatch: "does NOT
// silently serve unverified simplified text as if verified." Uses
// DESIGN.md's --state-warning tokens (amber), deliberately distinct from
// the error tokens below — a mismatch is not the same failure mode as a
// broken upload, and the color says so.
export function WarningBanner({ detail }) {
  return (
    <div
      role="alert"
      className="rounded-lg border border-state-warning/30 bg-state-warning-soft p-4 text-body text-ink-primary"
    >
      <p className="font-medium text-state-warning">
        This simplification could not be fully verified
      </p>
      <p className="mt-1 text-ink-secondary">
        {detail ||
          "Our automatic check found a possible difference between the original and simplified text. Please compare against the original before relying on this version."}
      </p>
    </div>
  )
}

// ErrorBanner — for hard failures (upload rejected, Gemini call failed).
// Uses DESIGN.md's --state-error tokens (red), distinct from the amber
// warning above.
export function ErrorBanner({ message }) {
  return (
    <div
      role="alert"
      className="rounded-lg border border-state-error/30 bg-state-error-soft p-4 text-body text-state-error"
    >
      {message}
    </div>
  )
}
