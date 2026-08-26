// ExplainPage — render one published /explain/{slug} explanation from the static SEO registry.
import { useEffect, useState } from "react"
import { useParams, Link } from "react-router-dom"
import { ArrowLeft, ExternalLink } from "lucide-react"

export function ExplainPage() {
  const { slug } = useParams()
  const [entry, setEntry] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // Fetch the minimal publish registry once, then find the current slug.
  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const res = await fetch("/seo/explain-pages.json")
        if (!res.ok) {
          if (!cancelled) setError("not_found")
          return
        }
        const data = await res.json()
        const match = (data?.pages || []).find((p) => p.slug === slug && p.status === "published")
        if (!match) {
          if (!cancelled) setError("not_found")
          return
        }
        if (!cancelled) setEntry(match)
      } catch {
        if (!cancelled) setError("network")
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [slug])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <p className="text-sm text-[#737373]">Loading explanation…</p>
      </div>
    )
  }

  if (error === "network") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="text-center">
          <h1 className="text-4xl font-semibold text-[#0a0a0a]">Offline</h1>
          <p className="mt-2 text-sm text-[#737373]">Couldn’t load this explanation. Check your connection and try again.</p>
          <Link to="/" className="mt-6 inline-block text-sm text-[#2563eb] hover:underline">Analyze your own paper</Link>
        </div>
      </div>
    )
  }

  if (!entry) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="text-center">
          <h1 className="text-4xl font-semibold text-[#0a0a0a]">404</h1>
          <p className="mt-2 text-sm text-[#737373]">Explanation not found</p>
          <Link to="/" className="mt-6 inline-block text-sm text-[#2563eb] hover:underline">Analyze your own paper</Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white">
      <header className="border-b border-[#e5e5e5]">
        <div className="mx-auto flex max-w-3xl items-center gap-3 px-4 py-3">
          <Link to="/" className="inline-flex items-center gap-1.5 text-xs font-medium text-[#737373] hover:text-[#0a0a0a] transition-colors">
            <ArrowLeft className="h-3.5 w-3.5" />
            PaperViz
          </Link>
          <span className="text-[#e5e5e5]">·</span>
          <span className="text-xs text-[#737373]">Published explanation</span>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-4 py-8">
        <article className="rounded-[16px] border border-[#e5e5e5] bg-white p-6 sm:p-8 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]">
          <h1 className="font-satoshi text-xl sm:text-2xl font-medium leading-tight text-[#0a0a0a]">
            {entry.title}
          </h1>
          <p className="mt-3 text-sm leading-relaxed text-[#737373]">
            {entry.description}
          </p>

          {entry.source && (
            <div className="mt-6 rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-4 text-sm text-[#404040]">
              <p className="font-medium text-[#171717]">Source paper</p>
              <p className="mt-1">{entry.source.paperTitle}</p>
              {entry.source.url && (
                <a
                  href={entry.source.url}
                  target="_blank"
                  rel="noreferrer"
                  className="mt-2 inline-flex items-center gap-1 text-[#2563eb] hover:underline"
                >
                  View source <ExternalLink className="h-3 w-3" />
                </a>
              )}
            </div>
          )}

          <p className="mt-6 text-xs text-[#737373]">
            This explanation is a derived interpretation from the source paper, not a replacement for the original research.
          </p>
        </article>

        <div className="mt-10 rounded-[12px] border border-[#e5e5e5] bg-[#fafafa] p-6 text-center">
          <p className="text-sm text-[#525252]">Understand your own paper in seconds.</p>
          <Link to="/" className="mt-4 inline-flex items-center gap-1.5 rounded-[8px] bg-[#0a0a0a] px-4 py-2 text-sm font-medium text-white hover:bg-[#171717] transition-colors">
            Analyze a paper
          </Link>
        </div>
      </main>
    </div>
  )
}
