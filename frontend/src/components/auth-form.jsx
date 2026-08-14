import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"

const INPUT_CLASS =
  "w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 focus:border-[#2563eb]"

const LABEL_CLASS = "block text-xs font-medium text-[#737373] mb-1.5"

export function AuthForm({
  title,
  subtitle,
  submitLabel,
  submittingLabel,
  footerQuestion,
  footerLinkText,
  footerLinkTo,
  error,
  isSubmitting,
  onSubmit,
  autoCompletePassword,
}) {
  return (
    <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center px-6">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold tracking-tight text-[#0a0a0a]">PaperViz</span>
          </Link>
          <h1 className="font-satoshi text-2xl font-medium text-[#0a0a0a]">{title}</h1>
          <p className="mt-1 text-sm text-[#737373]">{subtitle}</p>
        </div>

        <form onSubmit={onSubmit} className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
          <div className="space-y-4">
            <div>
              <label htmlFor="email" className={LABEL_CLASS}>Email</label>
              <input
                id="email"
                name="email"
                type="email"
                className={INPUT_CLASS}
                placeholder="you@example.com"
                autoComplete="email"
                disabled={isSubmitting}
              />
            </div>
            <div>
              <label htmlFor="password" className={LABEL_CLASS}>Password</label>
              <input
                id="password"
                name="password"
                type="password"
                className={INPUT_CLASS}
                placeholder="At least 8 characters"
                autoComplete={autoCompletePassword}
                disabled={isSubmitting}
              />
            </div>
          </div>

          {error && (
            <div className="mt-4 rounded-[6px] bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {error}
            </div>
          )}

          <Button type="submit" disabled={isSubmitting} className="mt-4 w-full">
            {isSubmitting ? submittingLabel : submitLabel}
          </Button>

          <p className="mt-4 text-center text-xs text-[#737373]">
            {footerQuestion}{" "}
            <Link to={footerLinkTo} className="text-[#2563eb] hover:underline">{footerLinkText}</Link>
          </p>
        </form>
      </div>
    </div>
  )
}