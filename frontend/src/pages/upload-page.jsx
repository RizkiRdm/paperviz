import { Link, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Ingest } from "@/components/ingest"
import { useCurrentUser } from "@/hooks/use-current-user"

// UploadPage: hero + CTAs above the single unified Ingest form (P05).
export function UploadPage() {
  const navigate = useNavigate()
  const { user, checked } = useCurrentUser()

  // onIngested routes to the result page once a document id exists.
  const onIngested = (documentId) => navigate(`/${documentId}`)

  return (
    <div className="min-h-screen bg-white bg-dotted-grid px-6 py-10">
      <div className="mx-auto w-full max-w-2xl">
        <div className="text-center mb-8">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">PV</div>
            <span className="font-mono text-sm font-semibold tracking-tight text-[#0a0a0a]">PaperViz</span>
          </Link>
          <h1 className="font-satoshi text-4xl sm:text-5xl font-medium tracking-tight text-[#0a0a0a] leading-tight">Papers, in plain language.</h1>
          <p className="mt-3 text-base text-[#737373] max-w-md mx-auto leading-relaxed">Transform dense academic PDFs into clear, verified summaries with interactive charts.</p>
          <div className="mt-5 flex flex-col sm:flex-row items-center justify-center gap-3">
            <Link to="/agents"><Button size="lg">Add to Claude Code</Button></Link>
            {user ? (
              <Link to="/account"><Button variant="outline" size="lg">{user.email}</Button></Link>
            ) : checked ? (
              <Link to="/login"><Button variant="outline" size="lg">Sign in</Button></Link>
            ) : null}
          </div>
        </div>

        <Ingest onIngested={onIngested} />

        <p className="mt-4 text-center text-xs text-[#737373]">Free to run. Add your own model key on the <Link to="/account">account page</Link> to analyse a paper.</p>
      </div>
    </div>
  )
}
