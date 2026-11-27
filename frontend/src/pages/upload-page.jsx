import { Link, useNavigate } from "react-router-dom"
import { useEffect, useRef, useState } from "react"
import { createScope, animate } from "animejs"
import { ArrowRight, Copy, Check, FileText, CheckCircle2, ShieldCheck, Terminal } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Ingest } from "@/components/ingest"
import { useCurrentUser } from "@/hooks/use-current-user"

export function UploadPage() {
  const navigate = useNavigate()
  const { user, checked } = useCurrentUser()
  const root = useRef(null)
  const scope = useRef(null)
  const [copiedCode, setCopiedCode] = useState(false)

  const mcpCommand = "claude mcp add paperviz -- /absolute/path/to/paperviz-mcp"

  const onIngested = (documentId) => navigate(`/${documentId}`)

  function handleCopyCommand() {
    navigator.clipboard.writeText(mcpCommand)
    setCopiedCode(true)
    setTimeout(() => setCopiedCode(false), 2000)
  }

  useEffect(() => {
    scope.current = createScope({ root: root.current }).add(() => {
      animate(".hero", { opacity: [0, 1], translateY: [24, 0], duration: 700, ease: "out(3)" })
      animate(".product-frame", { opacity: [0, 1], translateY: [32, 0], duration: 900, delay: 150, ease: "out(4)" })
      animate(".feature-pill", { opacity: [0, 1], translateY: [12, 0], delay: (_el, i) => 120 + i * 90, duration: 500, ease: "out(2)" })
    })
    return () => scope.current?.revert()
  }, [])

  return (
    <div ref={root} className="min-h-screen bg-[#ffffff] text-[#171717] selection:bg-[#2563eb]/10 selection:text-[#2563eb]">
      {/* Site Header */}
      <header className="sticky top-0 z-40 border-b border-[#e5e5e5] bg-white/90 backdrop-blur-md">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6">
          <Link className="flex items-center gap-2.5 group" to="/">
            <span className="flex h-7 w-7 items-center justify-center rounded-[6px] bg-[#0a0a0a] font-mono text-xs font-semibold text-white transition-transform group-hover:scale-105">
              PV
            </span>
            <span className="font-sans text-base font-semibold tracking-tight text-[#0a0a0a]">
              PaperViz
            </span>
          </Link>

          <nav className="hidden md:flex items-center gap-6" aria-label="Primary">
            <a href="#workflow" className="text-sm font-medium text-[#737373] hover:text-[#0a0a0a] transition-colors">
              Workflow
            </a>
            <a href="#evidence-principle" className="text-sm font-medium text-[#737373] hover:text-[#0a0a0a] transition-colors">
              Evidence
            </a>
            <a href="#agents-capability" className="text-sm font-medium text-[#737373] hover:text-[#0a0a0a] transition-colors">
              Agents
            </a>
          </nav>

          <div className="flex items-center gap-3">
            {user ? (
              <Link to="/account">
                <Button variant="outline" size="sm">
                  {user.email}
                </Button>
              </Link>
            ) : checked ? (
              <>
                <Link to="/login">
                  <Button variant="ghost" size="sm">
                    Log in
                  </Button>
                </Link>
                <Link to="/signup">
                  <Button size="sm">Sign up</Button>
                </Link>
              </>
            ) : null}
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="relative overflow-hidden border-b border-[#e5e5e5] bg-[#ffffff] pt-16 pb-20 md:pt-24 md:pb-28">
        <div className="mx-auto max-w-4xl px-4 text-center hero">
          {/* Feature Pills */}
          <div className="flex flex-wrap items-center justify-center gap-2 mb-8">
            <span className="feature-pill inline-flex items-center gap-1.5 rounded-full border border-[#bbf7d0] bg-[#dcfce7] px-3 py-1 text-xs font-medium text-[#16a34a]">
              <span className="h-1.5 w-1.5 rounded-full bg-[#16a34a]" /> Source context
            </span>
            <span className="feature-pill inline-flex items-center gap-1.5 rounded-full border border-[#ddd6fe] bg-[#f5f3ff] px-3 py-1 text-xs font-medium text-[#7c3aed]">
              <span className="h-1.5 w-1.5 rounded-full bg-[#7c3aed]" /> Evidence graph
            </span>
            <span className="feature-pill inline-flex items-center gap-1.5 rounded-full border border-[#fed7aa] bg-[#fff7ed] px-3 py-1 text-xs font-medium text-[#ea580c]">
              <span className="h-1.5 w-1.5 rounded-full bg-[#ea580c]" /> Grounded figures
            </span>
          </div>

          <h1 className="font-['Satoshi',sans-serif] text-4xl sm:text-5xl md:text-6xl font-medium tracking-tight text-[#0a0a0a] leading-[1.12] mb-6">
            Turn research papers into evidence you can inspect.
          </h1>

          <p className="text-lg sm:text-xl text-[#525252] max-w-2xl mx-auto leading-relaxed mb-10">
            PaperViz extracts claims, supporting evidence, and figures from academic papers so researchers and AI agents can understand, trace, and reuse findings with verified provenance.
          </p>

          <div className="flex flex-wrap items-center justify-center gap-3">
            <Link to="/sample">
              <Button size="lg" className="px-6 h-11 text-sm font-medium">
                See sample workspace <ArrowRight className="h-4 w-4 ml-1.5" />
              </Button>
            </Link>
            <a href="#analyze-form">
              <Button variant="outline" size="lg" className="px-6 h-11 text-sm font-medium">
                Analyze a paper
              </Button>
            </a>
          </div>

          <p className="mt-4 text-xs text-[#737373]">
            The sample is a real PaperViz document — a seeded paper with claims and evidence, not a mockup.
          </p>
        </div>

        {/* Product Workspace Preview Frame */}
        <div className="mx-auto max-w-5xl px-4 mt-14 product-frame">
          <div className="rounded-[16px] border border-[#e5e5e5] bg-[#ffffff] shadow-lg shadow-black/3 overflow-hidden">
            {/* Window bar */}
            <div className="flex items-center justify-between border-b border-[#e5e5e5] bg-[#f5f5f5] px-4 py-3">
              <div className="flex items-center gap-1.5">
                <span className="h-2.5 w-2.5 rounded-full bg-[#d4d4d4]" />
                <span className="h-2.5 w-2.5 rounded-full bg-[#d4d4d4]" />
                <span className="h-2.5 w-2.5 rounded-full bg-[#d4d4d4]" />
              </div>
              <span className="font-mono text-xs text-[#525252]">
                paperviz / workspace / attention-is-all-you-need
              </span>
              <span className="inline-flex items-center gap-1 rounded-full border border-[#bbf7d0] bg-[#dcfce7] px-2.5 py-0.5 text-[11px] font-medium text-[#16a34a]">
                <ShieldCheck className="h-3 w-3" /> verified
              </span>
            </div>

            {/* Teaser content inside workspace preview */}
            <div className="p-6 md:p-8 grid md:grid-cols-3 gap-6 bg-[#ffffff]">
              <div className="md:col-span-2 space-y-4">
                <div className="inline-flex items-center gap-2 rounded-full border border-[#e5e5e5] bg-[#f5f5f5] px-3 py-1 text-xs font-mono text-[#525252]">
                  <FileText className="h-3.5 w-3.5 text-[#2563eb]" /> Section 3: Model Architecture
                </div>
                <h3 className="text-xl font-medium text-[#0a0a0a]">
                  Multi-Head Attention replaces recurrence entirely
                </h3>
                <p className="text-sm text-[#525252] leading-relaxed">
                  Instead of performing a single attention function with keys, values and queries, the Transformer projects queries, keys and values h times with different, learned linear projections.
                </p>
                <div className="flex flex-wrap items-center gap-2 pt-2">
                  <span className="rounded-[6px] border border-[#e5e5e5] bg-[#f5f5f5] px-2.5 py-1 text-xs font-mono text-[#171717]">
                    BLEU: 28.4
                  </span>
                  <span className="rounded-[6px] border border-[#e5e5e5] bg-[#f5f5f5] px-2.5 py-1 text-xs font-mono text-[#171717]">
                    Training time: 3.5 days
                  </span>
                  <span className="rounded-[6px] border border-[#e5e5e5] bg-[#f5f5f5] px-2.5 py-1 text-xs font-mono text-[#171717]">
                    8 P100 GPUs
                  </span>
                </div>
              </div>

              <div className="rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-4 flex flex-col justify-between space-y-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs font-medium text-[#737373]">
                    <span>Grounding Evidence</span>
                    <span className="text-[#16a34a] flex items-center gap-1 font-mono">
                      <CheckCircle2 className="h-3.5 w-3.5" /> 10/10 checks
                    </span>
                  </div>
                  <p className="text-xs text-[#525252]">
                    Claims verified directly against Table 2 and Figure 1 of original PDF.
                  </p>
                </div>
                <Link to="/sample" className="w-full">
                  <Button variant="outline" size="sm" className="w-full justify-center">
                    Explore live document
                  </Button>
                </Link>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Input Ingest Section */}
      <section className="py-20 md:py-24 border-b border-[#e5e5e5] bg-[#f5f5f5]" id="analyze-form">
        <div className="mx-auto max-w-3xl px-4">
          <div className="text-center mb-10">
            <h2 className="text-2xl sm:text-3xl font-medium tracking-tight text-[#0a0a0a] mb-3">
              Analyze a research paper
            </h2>
            <p className="text-sm sm:text-base text-[#525252] max-w-xl mx-auto">
              Drop a PDF, paste text, or provide a DOI / arXiv URL. The result opens as a structured paper workspace.
            </p>
          </div>

          <Ingest onIngested={onIngested} />

          <p className="mt-4 text-center text-xs text-[#737373]">
            Bring your own key (BYOK): keys are encrypted at rest with AES-256-GCM and resolved per request.
          </p>
        </div>
      </section>

      {/* Core Workflow */}
      <section className="py-20 md:py-28 border-b border-[#e5e5e5] bg-[#ffffff]" id="workflow">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <div className="text-center mb-16">
            <h2 className="text-2xl sm:text-3xl font-medium tracking-tight text-[#0a0a0a] mb-3">
              Understand, inspect, trace, and reuse
            </h2>
            <p className="text-sm sm:text-base text-[#525252] max-w-xl mx-auto">
              Every surface answers a distinct research question without turning the paper into an ungrounded chat summary.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
            <article className="rounded-[12px] border border-[#e5e5e5] bg-white p-6 shadow-xs flex flex-col justify-between">
              <div>
                <span className="font-mono text-xs font-semibold text-[#2563eb] mb-3 block">
                  01
                </span>
                <h3 className="text-base font-medium text-[#0a0a0a] mb-2">Understand</h3>
                <p className="text-xs text-[#525252] leading-relaxed">
                  Read simplified interpretations without losing the boundary between original source and generated insights.
                </p>
              </div>
            </article>

            <article className="rounded-[12px] border border-[#e5e5e5] bg-white p-6 shadow-xs flex flex-col justify-between">
              <div>
                <span className="font-mono text-xs font-semibold text-[#2563eb] mb-3 block">
                  02
                </span>
                <h3 className="text-base font-medium text-[#0a0a0a] mb-2">Inspect</h3>
                <p className="text-xs text-[#525252] leading-relaxed">
                  Browse structured claims, extracted evidence, methods, results, citations, and figures exposed by the paper.
                </p>
              </div>
            </article>

            <article className="rounded-[12px] border border-[#e5e5e5] bg-white p-6 shadow-xs flex flex-col justify-between">
              <div>
                <span className="font-mono text-xs font-semibold text-[#2563eb] mb-3 block">
                  03
                </span>
                <h3 className="text-base font-medium text-[#0a0a0a] mb-2">Trace</h3>
                <p className="text-xs text-[#525252] leading-relaxed">
                  Keep page, section, figure, and source context attached to claims instead of hiding provenance in tooltips.
                </p>
              </div>
            </article>

            <article className="rounded-[12px] border border-[#e5e5e5] bg-white p-6 shadow-xs flex flex-col justify-between">
              <div>
                <span className="font-mono text-xs font-semibold text-[#2563eb] mb-3 block">
                  04
                </span>
                <h3 className="text-base font-medium text-[#0a0a0a] mb-2">Reuse</h3>
                <p className="text-xs text-[#525252] leading-relaxed">
                  Export structured research objects, add notes, or allow your AI agents to query the research graph via MCP.
                </p>
              </div>
            </article>
          </div>
        </div>
      </section>

      {/* Trust Model */}
      <section className="py-20 md:py-24 border-b border-[#e5e5e5] bg-[#f5f5f5]" id="evidence-principle">
        <div className="mx-auto max-w-6xl px-4 sm:px-6">
          <div className="text-center mb-16">
            <h2 className="text-2xl sm:text-3xl font-medium tracking-tight text-[#0a0a0a] mb-3">
              Visual authority for evidence
            </h2>
            <p className="text-sm sm:text-base text-[#525252] max-w-xl mx-auto">
              Source material stays authoritative. AI interpretation stays clearly labeled. Unknown stays unknown.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            <article className="rounded-[12px] border border-[#e5e5e5] bg-white p-6 shadow-xs">
              <span className="inline-flex items-center rounded-full border border-[#bbf7d0] bg-[#dcfce7] px-2.5 py-1 text-xs font-mono font-medium text-[#16a34a] mb-4">
                SOURCE
              </span>
              <h3 className="text-base font-medium text-[#0a0a0a] mb-2">What the paper says</h3>
              <p className="text-xs text-[#525252] leading-relaxed">
                Directly extracted text, numerical values, original figures, and author citations verbatim from the paper.
              </p>
            </article>

            <article className="rounded-[12px] border border-[#e5e5e5] bg-white p-6 shadow-xs">
              <span className="inline-flex items-center rounded-full border border-[#bfdbfe] bg-[#eff6ff] px-2.5 py-1 text-xs font-mono font-medium text-[#2563eb] mb-4">
                EVIDENCE
              </span>
              <h3 className="text-base font-medium text-[#0a0a0a] mb-2">What supports the claim</h3>
              <p className="text-xs text-[#525252] leading-relaxed">
                Deterministic linkages connecting claims to exact tables, metrics, sections, and figures with verified grounding.
              </p>
            </article>

            <article className="rounded-[12px] border border-[#e5e5e5] bg-white p-6 shadow-xs">
              <span className="inline-flex items-center rounded-full border border-[#ddd6fe] bg-[#f5f3ff] px-2.5 py-1 text-xs font-mono font-medium text-[#7c3aed] mb-4">
                AI-DERIVED
              </span>
              <h3 className="text-base font-medium text-[#0a0a0a] mb-2">What PaperViz interprets</h3>
              <p className="text-xs text-[#525252] leading-relaxed">
                Simplified breakdowns and structured syntheses that never get presented as unverified ground truth.
              </p>
            </article>
          </div>
        </div>
      </section>

      {/* Agents Capability */}
      <section className="py-20 md:py-24 border-b border-[#e5e5e5] bg-[#ffffff]" id="agents-capability">
        <div className="mx-auto max-w-4xl px-4 sm:px-6">
          <div className="rounded-[16px] border border-[#e5e5e5] bg-[#f5f5f5] p-8 md:p-10 shadow-xs">
            <div className="flex items-center gap-2 mb-4">
              <span className="flex h-7 w-7 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white">
                <Terminal className="h-4 w-4" />
              </span>
              <span className="font-mono text-xs font-medium text-[#525252]">
                Model Context Protocol (MCP)
              </span>
            </div>

            <h2 className="text-2xl sm:text-3xl font-medium tracking-tight text-[#0a0a0a] mb-4">
              Research context that AI agents can inspect and query
            </h2>
            <p className="text-sm sm:text-base text-[#525252] leading-relaxed mb-6">
              PaperViz ships a local stdio MCP server with five deterministic tools (`ingest_document`, `search_documents`, `get_document`, `get_evidence`, `get_figures`). Connect it directly to Claude Code, Cursor, Windsurf, or Codex.
            </p>

            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
              <div className="flex-1 rounded-[8px] border border-[#e5e5e5] bg-white px-4 py-2.5 font-mono text-xs text-[#171717] overflow-x-auto whitespace-pre">
                {mcpCommand}
              </div>
              <Button
                variant="outline"
                size="md"
                onClick={handleCopyCommand}
                className="shrink-0 flex items-center justify-center gap-1.5"
              >
                {copiedCode ? (
                  <>
                    <Check className="h-4 w-4 text-[#16a34a]" /> Copied
                  </>
                ) : (
                  <>
                    <Copy className="h-4 w-4 text-[#737373]" /> Copy
                  </>
                )}
              </Button>
              <Link to="/agents" className="shrink-0">
                <Button size="md">
                  View Setup Guide
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Site Footer */}
      <footer className="border-t border-[#e5e5e5] bg-white py-12">
        <div className="mx-auto flex max-w-6xl flex-col sm:flex-row items-center justify-between gap-6 px-4 sm:px-6">
          <div className="flex items-center gap-2.5">
            <span className="flex h-6 w-6 items-center justify-center rounded-[4px] bg-[#0a0a0a] font-mono text-[11px] font-semibold text-white">
              PV
            </span>
            <span className="text-xs font-medium text-[#525252]">
              PaperViz — Structured research context for humans and agents
            </span>
          </div>

          <div className="flex items-center gap-6 text-xs font-medium text-[#737373]">
            <a href="#workflow" className="hover:text-[#0a0a0a] transition-colors">
              Workflow
            </a>
            <a href="#evidence-principle" className="hover:text-[#0a0a0a] transition-colors">
              Evidence
            </a>
            <Link to="/agents" className="hover:text-[#0a0a0a] transition-colors">
              MCP Agents
            </Link>
            <a
              href="https://github.com/zendev-sh/paperviz"
              target="_blank"
              rel="noreferrer"
              className="hover:text-[#0a0a0a] transition-colors"
            >
              GitHub
            </a>
          </div>
        </div>
      </footer>
    </div>
  )
}
