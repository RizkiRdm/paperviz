import { Link, useNavigate } from "react-router-dom"
import { useEffect, useRef } from "react"
import { createScope, animate } from "animejs"
import { Button } from "@/components/ui/button"
import { Ingest } from "@/components/ingest"
import { useCurrentUser } from "@/hooks/use-current-user"

export function UploadPage() {
  const navigate = useNavigate()
  const { user, checked } = useCurrentUser()
  const root = useRef(null)
  const scope = useRef(null)

  const onIngested = (documentId) => navigate(`/${documentId}`)

  useEffect(() => {
    scope.current = createScope({ root: root.current }).add(() => {
      animate(".hero", { opacity: [0, 1], translateY: [24, 0], duration: 700, ease: "out(3)" })
      animate(".product-frame", { opacity: [0, 1], translateY: [32, 0], duration: 900, delay: 150, ease: "out(4)" })
      animate(".feature-pill", { opacity: [0, 1], translateY: [12, 0], delay: (_el, i) => 120 + i * 90, duration: 500, ease: "out(2)" })
    })
    return () => scope.current?.revert()
  }, [])

  return (
    <div ref={root}>
      <header className="site-header">
        <div className="container site-header__inner">
          <Link className="brand" to="/"><span className="brand__mark">PV</span><span>PaperViz</span></Link>
          <nav className="site-nav" aria-label="Primary">
            <a href="#workflow">Workflow</a>
            <a href="#evidence-principle">Evidence</a>
            <a href="#agents-capability">Agents</a>
          </nav>
          <div className="cluster">
            {user ? (
              <Link to="/account"><Button variant="outline" size="sm">{user.email}</Button></Link>
            ) : checked ? (
              <>
                <Link to="/login"><Button variant="ghost" size="sm">Log in</Button></Link>
                <Link to="/signup"><Button size="sm">Sign up</Button></Link>
              </>
            ) : null}
          </div>
        </div>
      </header>

      <div className="dot-grid">
        <div className="container hero">
          <div className="pill-row pill-row--center">
            <span className="feature-pill feature-pill--green"><span>●</span> Source context</span>
            <span className="feature-pill feature-pill--violet"><span>●</span> Evidence graph</span>
            <span className="feature-pill feature-pill--orange"><span>●</span> Grounded figures</span>
          </div>
          <p className="eyebrow">RESEARCH CONTEXT, NOT ANOTHER CHAT WINDOW</p>
          <h1 className="display">Turn research papers into evidence you can inspect.</h1>
          <p className="hero__copy">PaperViz extracts claims, supporting evidence, and figures from research papers so you can understand, trace, and reuse the underlying research context.</p>
          <div className="cluster cluster--center">
            <a className="btn btn--dark btn--lg" href="#analyze-form">Analyze a paper</a>
            <a className="btn btn--outline btn--lg" href="#workflow">See how it works</a>
          </div>
        </div>

        <div className="container product-frame" id="analyze">
          <div className="product-frame__topbar">
            <div className="window-dots"><span></span><span></span><span></span></div>
            <span className="mono muted">paperviz / paper workspace</span>
            <span className="status status--source">source-backed</span>
          </div>
          <div className="mock-shell">
            <aside className="mock-sidebar">
              <div className="mock-sidebar__brand"><span className="brand__mark">PV</span><span>PaperViz</span></div>
              <div className="mock-sidebar__group">
                <span className="nav-label">Workspace</span>
                <span className="side-item is-active"><span>⌂</span>Home</span>
                <span className="side-item"><span>□</span>My Papers</span>
                <span className="side-item"><span>⌘</span>Agents</span>
              </div>
              <div className="mock-sidebar__group">
                <span className="nav-label nav-label--spaced">Account</span>
                <span className="side-item"><span>◎</span>Account</span>
                <span className="side-item"><span>⚙</span>Settings</span>
                <span className="side-item"><span>?</span>Help</span>
              </div>
            </aside>
            <div className="mock-main">
              <div className="mock-breadcrumb">Home <span>/</span> Paper workspace</div>
              <div className="mock-title-row">
                <div><span className="status status--source">SOURCE</span><h2>A practical framework for evaluating retrieval-augmented generation systems</h2><p className="muted">Example paper • 2025 • demo data</p></div>
                <button className="btn btn--outline">Share</button>
              </div>
              <div className="mock-tabs"><button className="is-active">Overview</button><button>Claims</button><button>Evidence</button><button>Figures</button><button>Notes</button></div>
              <div className="mock-grid">
                <article className="card card--reading"><div className="section-kicker">AI-DERIVED</div><h3>Paper summary</h3><p>This demonstration summary explains the paper in plain language. It is explicitly presented as interpretation, not as a replacement for the source.</p><div className="callout callout--source"><b>Trace:</b> compare this interpretation with the original source below.</div></article>
                <article className="card"><div className="section-kicker">SOURCE</div><h3>Key claim</h3><p>Retrieval quality changes the reliability of downstream answers.</p><div className="meta-line"><span>Evidence 03</span><span>Page 7</span></div></article>
                <article className="card"><div className="section-kicker">EVIDENCE</div><h3>Supporting text</h3><p>Source excerpt with section and page context shown beside the interpretation.</p><div className="source-ref">Results · p. 7</div></article>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="container section" id="analyze-form">
        <div className="section-heading section-heading--center">
          <p className="eyebrow">INPUT</p>
          <h2>Analyze a paper</h2>
          <p>Choose one supported source. The result opens as a paper workspace.</p>
        </div>
        <div className="input-card">
          <Ingest onIngested={onIngested} />
        </div>
        <p className="helper centered">Free to run. Add your own model key on the account page to analyse a paper.</p>
      </div>

      <div className="container section" id="workflow">
        <div className="section-heading section-heading--center"><p className="eyebrow">CORE WORKFLOW</p><h2>Understand → Inspect → Trace → Reuse</h2><p>Every surface answers a different research question without turning the paper into a generic dashboard.</p></div>
        <div className="four-up">
          <article className="workflow-card"><span className="workflow-card__num">01</span><h3>Understand</h3><p>Read a simplified interpretation without losing the boundary between source and generated content.</p></article>
          <article className="workflow-card"><span className="workflow-card__num">02</span><h3>Inspect</h3><p>Browse claims, evidence, methods, results, citations, tables, and figures that the current paper exposes.</p></article>
          <article className="workflow-card"><span className="workflow-card__num">03</span><h3>Trace</h3><p>Keep page, section, figure, and source context attached to evidence instead of hiding provenance in hover states.</p></article>
          <article className="workflow-card"><span className="workflow-card__num">04</span><h3>Reuse</h3><p>Add annotations, collections, export research data, or share a focused research view.</p></article>
        </div>
      </div>

      <div className="paper-band" id="evidence-principle">
        <div className="container section section--tight">
          <div className="section-heading section-heading--center"><p className="eyebrow">TRUST MODEL</p><h2>Paper → Claim → Evidence → Source</h2><p>Source material stays visually authoritative. AI interpretation stays clearly labeled. Unknown stays unknown.</p></div>
          <div className="three-up">
            <article className="trust-card"><span className="legend legend--source">SOURCE</span><h3>What the paper says</h3><p>Directly extracted material, original text, figure references, and source metadata.</p></article>
            <article className="trust-card"><span className="legend legend--evidence">EVIDENCE</span><h3>What supports the claim</h3><p>Text, tables, figures, sections, and other evidence linked back to a source location.</p></article>
            <article className="trust-card"><span className="legend legend--ai">AI-DERIVED</span><h3>What PaperViz interprets</h3><p>Simplified explanations and transformations that never get presented as independent truth.</p></article>
          </div>
        </div>
      </div>

      <div className="container section" id="agents-capability">
        <div className="agent-promo">
          <div><p className="eyebrow">AGENT CAPABILITY</p><h2>Research context that AI agents can inspect and reuse.</h2><p>The current repository ships a local stdio MCP server with five deterministic tools. This page shows setup, not a fictional hosted agent marketplace.</p></div>
          <div className="agent-promo__code"><code>claude mcp add paperviz -- /absolute/path/to/paperviz-mcp</code><Link to="/agents" className="btn btn--dark">Add to Claude Code</Link></div>
        </div>
      </div>

      <footer className="site-footer"><div className="container site-footer__inner"><div><div className="brand"><span className="brand__mark">PV</span><span>PaperViz</span></div><p className="muted">Structured research context for humans and agents.</p></div><div className="footer-links"><a href="#workflow">Workflow</a><a href="#evidence-principle">Evidence</a><a href="https://github.com" target="_blank" rel="noreferrer">GitHub</a></div></div></footer>
    </div>
  )
}
