import { Button } from "@/components/ui/button"
import { ErrorBanner } from "@/components/ui/status-banners"
import { RefreshCw } from "lucide-react"

// Fallback display copy for terminal pipeline failures — mapped error_message
// codes, never raw backend codes or provider/LLM detail.
const FAILURE_MESSAGES = {
  simplification_failed: "We couldn't simplify this paper. Please try again.",
  verification_failed_to_run: "We couldn't verify the summary against the original. Please try again.",
  extraction_failed: "We couldn't read this PDF. Please check that it has a text layer and try again.",
}

const GENERIC_FAILURE_MESSAGE = "We couldn't process this PDF. It may be image-only or corrupted. Please try a different file."

// ProcessingView — polite live region with mapped stage label (P06 labels only).
export function ProcessingView({ stageLabel, takingLong }) {
  return (
    <div
      role="status"
      aria-live="polite"
      className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6 text-center"
    >
      <div className="max-w-md rounded-[16px] border border-[#e5e5e5] bg-white p-8 shadow-xs flex flex-col items-center">
        <div className="h-10 w-10 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent mb-4" />
        <h2 className="font-satoshi text-lg font-medium text-[#0a0a0a]">
          {stageLabel || "Reading document..."}
        </h2>
        <p className="mt-2 text-xs text-[#737373] leading-relaxed">
          {takingLong
            ? "Still working — large papers with many figures can take a few minutes. This page refreshes automatically."
            : "Reading your paper, creating a plain-language summary, and checking it against the original."}
        </p>
      </div>
    </div>
  )
}

// TimeoutView — polling hit the hard timeout; retry re-checks status.
export function TimeoutView({ onRetry, onBack }) {
  return (
    <div className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6 text-center">
      <div className="max-w-md rounded-[16px] border border-[#e5e5e5] bg-white p-8 shadow-xs">
        <p className="text-sm text-[#737373]">Processing is taking longer than expected.</p>
        <div className="mt-6 flex flex-col gap-2">
          <Button onClick={onRetry} variant="primary">
            <RefreshCw className="h-4 w-4 mr-2" /> Check Status Again
          </Button>
          <Button onClick={onBack} variant="secondary">
            Upload Another Paper
          </Button>
        </div>
      </div>
    </div>
  )
}

// FailureView — poll failure renders with Retry (12.1); a terminal pipeline
// failure renders without one, since re-fetching cannot recover it.
export function FailureView({ pollError, failCode, onRetry, onBack }) {
  const message = pollError || FAILURE_MESSAGES[failCode] || GENERIC_FAILURE_MESSAGE
  return (
    <div className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6">
      <div className="max-w-md w-full">
        <ErrorBanner message={message} />
        {onRetry && (
          <Button onClick={onRetry} variant="secondary" className="mt-4 w-full">
            <RefreshCw className="h-4 w-4 mr-2" /> Retry
          </Button>
        )}
        <Button onClick={onBack} variant="secondary" className="mt-4 w-full">
          Start Over
        </Button>
      </div>
    </div>
  )
}
