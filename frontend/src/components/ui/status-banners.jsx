// ponytail: status indicators styled per DESIGN.md (Soft Mint badge, Tangerine warning, Red error)
import { ShieldCheck, AlertTriangle, AlertCircle } from "lucide-react"

export function VerificationBadge({ onClick, ...props }) {
  return (
    <button
      type="button"
      onClick={onClick}
      {...props}
      className="inline-flex items-center gap-1.5 rounded-full bg-[#dcfce7] px-3 py-1 text-xs font-medium text-[#16a34a] border border-[#bbf7d0] hover:bg-[#d1fae5] transition-colors cursor-pointer"
    >
      <ShieldCheck className="h-3.5 w-3.5" aria-hidden="true" />
      Verified
    </button>
  )
}

export function ClaimComparisonPanel({ claimDiff, onClose }) {
  const original = JSON.parse(claimDiff.original_claims || "[]")
  const simplified = JSON.parse(claimDiff.simplified_claims || "[]")
  return (
    <div className="mt-3 rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-4">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-semibold text-[#0a0a0a]">Claims checked</p>
        <button onClick={onClose} className="text-[11px] text-[#737373] hover:text-[#0a0a0a]">Hide</button>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
        <div>
          <p className="font-medium text-[#737373] mb-1.5">Original</p>
          <ul className="space-y-1 text-[#171717]">
            {original.map((c, i) => <li key={i}>• {c}</li>)}
          </ul>
        </div>
        <div>
          <p className="font-medium text-[#737373] mb-1.5">Simplified</p>
          <ul className="space-y-1 text-[#171717]">
            {simplified.map((c, i) => <li key={i}>• {c}</li>)}
          </ul>
        </div>
      </div>
    </div>
  )
}

export function WarningBanner({ detail }) {
  return (
    <div
      role="alert"
      className="flex gap-3 rounded-[12px] border border-[#fed7aa] bg-[#fff7ed] p-4 text-sm text-[#171717]"
    >
      <AlertTriangle className="h-5 w-5 shrink-0 text-[#ea580c]" />
      <div>
        <p className="font-medium text-[#ea580c]">
          This simplification could not be fully verified
        </p>
        <p className="mt-1 text-[#737373] text-xs">
          {detail ||
            "Our automatic check found a possible difference between the original and simplified text. Please compare against the original before relying on this version."}
        </p>
      </div>
    </div>
  )
}

export function ErrorBanner({ message }) {
  return (
    <div
      role="alert"
      className="flex gap-3 rounded-[12px] border border-[#fecaca] bg-[#fef2f2] p-4 text-sm text-[#dc2626]"
    >
      <AlertCircle className="h-5 w-5 shrink-0 text-[#dc2626]" />
      <span className="font-medium">{message}</span>
    </div>
  )
}

