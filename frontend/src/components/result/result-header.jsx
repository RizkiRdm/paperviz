import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { VerificationBadge } from "@/components/ui/status-banners"
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip"
import { ArrowLeft, Link2 } from "lucide-react"

// Render sticky top bar with nav, verification, visibility — DESIGN.md Dub tokens.
export function ResultHeader({ doc, visibility, visibilityError, shareError, showClaims, onBack, onVisibilityChange, onShare, onToggleClaims }) {
  return (
    <header className="border-b border-[#e5e5e5] bg-white/80 backdrop-blur-xs sticky top-0 z-10">
      <div className="mx-auto flex max-w-[1200px] items-center justify-between px-6 py-3.5">
        <div className="flex items-center gap-4">
          <button
            onClick={onBack}
            aria-label="Back to upload"
            className="flex items-center gap-1.5 text-xs font-medium text-[#737373] hover:text-[#0a0a0a] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20 rounded-[6px] px-1 -ml-1"
          >
            <ArrowLeft className="h-4 w-4" aria-hidden="true" /> New Document
          </button>
          <span className="h-4 w-px bg-[#e5e5e5]" aria-hidden="true" />
          <Link to="/account" aria-label="Go to account" className="text-xs text-[#737373] hover:text-[#0a0a0a] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2563eb]/20 rounded-[4px] px-1">
            Account
          </Link>
          <span className="h-4 w-px bg-[#e5e5e5]" aria-hidden="true" />
          <span className="font-mono text-xs font-semibold text-[#0a0a0a]">PaperViz</span>
        </div>
        <div className="flex flex-col items-end gap-1">
          <div className="flex items-center gap-3">
            {/* Gate badge on claim_diff; absent data renders disabled badge. */}
            {doc?.status === "complete" && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <VerificationBadge
                    onClick={onToggleClaims}
                    disabled={!doc.claim_diff}
                    title={doc.claim_diff ? "Claims checked against the original text. Click to compare." : "Verification data not available"}
                    aria-expanded={showClaims}
                    aria-controls="claim-comparison-panel"
                    aria-label={doc.claim_diff ? "Claims checked against the original text. Click to compare." : "Verification data not available"}
                  />
                </TooltipTrigger>
                <TooltipContent>Claims checked against the original text. Click to compare.</TooltipContent>
              </Tooltip>
            )}
            <select
              value={visibility}
              onChange={(e) => onVisibilityChange(e.target.value)}
              aria-label="Who can view this summary"
              className="h-9 cursor-pointer rounded-[8px] border border-[#e5e5e5] bg-white px-2 text-xs font-medium text-[#171717] transition-colors hover:bg-[#f5f5f5] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 disabled:text-[#a3a3a3]"
              disabled={doc?.status !== "complete"}
            >
              <option value="private">Private</option>
              <option value="unlisted">Unlisted</option>
              <option value="public">Public</option>
            </select>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="secondary" onClick={onShare} aria-label="Share document" className="h-9 px-3 text-xs gap-1.5 font-medium">
                  <Link2 className="h-3.5 w-3.5 text-[#737373]" aria-hidden="true" /> Share
                </Button>
              </TooltipTrigger>
              <TooltipContent>Share via a link that expires after 7 days of inactivity.</TooltipContent>
            </Tooltip>
          </div>
          {shareError && (
            <p className="text-[11px] text-[#ea580c]" role="alert">
              Couldn&apos;t create a share link. Please try again.
            </p>
          )}
          {visibilityError && (
            <p className="text-[11px] text-[#ea580c]" role="alert">
              Couldn&apos;t update visibility. Please try again.
            </p>
          )}
        </div>
      </div>
    </header>
  )
}
