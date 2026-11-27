// NotFoundPage — 404 catch-all for invalid routes and expired document links.
import { Link } from "react-router-dom"
import { ArrowLeft } from "lucide-react"

export function NotFoundPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-white px-4">
      <div className="max-w-md w-full text-center py-16">
        <span className="font-mono text-xs font-semibold uppercase tracking-wider text-[#737373] bg-[#f5f5f5] px-2.5 py-1 rounded-[4px] border border-[#e5e5e5]">
          Error 404
        </span>
        <h1 className="mt-4 font-display text-4xl sm:text-5xl font-medium tracking-tight text-[#0a0a0a]">
          Page not found
        </h1>
        <p className="mt-3 text-sm text-[#737373] leading-relaxed">
          The paper or route you requested doesn&apos;t exist, was moved, or has expired.
        </p>
        <div className="mt-8 flex justify-center">
          <Link
            to="/"
            className="inline-flex items-center gap-2 rounded-[8px] bg-[#0a0a0a] px-4 py-2.5 text-sm font-medium text-white hover:bg-[#171717] transition-colors shadow-2xs"
          >
            <ArrowLeft className="h-4 w-4" />
            <span>Back to home</span>
          </Link>
        </div>
      </div>
    </div>
  )
}
