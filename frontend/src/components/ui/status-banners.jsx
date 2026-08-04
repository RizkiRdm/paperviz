// ponytail: status indicators styled per DESIGN.md (Soft Mint badge, Tangerine warning, Red error)
import { ShieldCheck, AlertTriangle, AlertCircle } from "lucide-react"

export function VerificationBadge() {
  return (
    <span className="inline-flex items-center gap-1.5 rounded-full bg-[#dcfce7] px-3 py-1 text-xs font-medium text-[#16a34a] border border-[#bbf7d0]">
      <ShieldCheck className="h-3.5 w-3.5" aria-hidden="true" />
      Verified
    </span>
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

