import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"

export function UploadPage() {
  return (
    <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center px-6">
      <div className="w-full max-w-lg text-center">
        <div className="mb-8">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold tracking-tight text-[#0a0a0a]">PaperViz</span>
          </Link>
          <h1 className="font-satoshi text-4xl sm:text-5xl font-medium tracking-tight text-[#0a0a0a] leading-tight">
            Papers, in plain language.
          </h1>
          <p className="mt-4 text-base text-[#737373] max-w-md mx-auto leading-relaxed">
            Transform dense academic PDFs into clear, verified summaries with interactive charts.
          </p>
        </div>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
          <Link to="/agents">
            <Button size="lg" className="w-full sm:w-auto">
              Add to Claude Code
            </Button>
          </Link>
          <Link to="/login">
            <Button variant="outline" size="lg" className="w-full sm:w-auto">
              Sign in
            </Button>
          </Link>
        </div>

        <p className="mt-8 text-xs text-[#737373]">
          Free for researchers. No account required to try.
        </p>
      </div>
    </div>
  )
}

