import { useEffect, useState } from "react"
import { Link } from "react-router-dom"

export function AppShell({ active, children }) {
  const [open, setOpen] = useState(false)

  useEffect(() => {
    if (!open) return
    const handleKeyDown = (event) => {
      if (event.key === "Escape") setOpen(false)
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [open])

  const nav = [
    { id: "home", label: "Home", to: "/", icon: "⌂" },
    { id: "agents", label: "Agents", to: "/agents", icon: "⌘" },
  ]
  const account = [
    { id: "account", label: "Account", to: "/account", icon: "◎" },
  ]

  return (
    <div className="app-shell">
      <aside id="app-sidebar" inert={!open} className={`sidebar ${open ? "is-open" : ""}`}>
        <div className="sidebar__top">
          <Link to="/" className="brand"><span className="brand__mark">PV</span><span>PaperViz</span></Link>
        </div>
        <nav id="app-sidebar-nav" className="sidebar__nav" aria-label="Application">
          <span className="nav-label">Workspace</span>
          {nav.map((item) => (
            <Link key={item.id} to={item.to} className={`side-item ${active === item.id ? "is-active" : ""}`} aria-current={active === item.id ? "page" : undefined} onClick={() => setOpen(false)}>
              <span aria-hidden="true">{item.icon}</span>{item.label}
            </Link>
          ))}
          <span className="nav-label nav-label--spaced">Account</span>
          {account.map((item) => (
            <Link key={item.id} to={item.to} className={`side-item ${active === item.id ? "is-active" : ""}`} aria-current={active === item.id ? "page" : undefined} onClick={() => setOpen(false)}>
              <span aria-hidden="true">{item.icon}</span>{item.label}
            </Link>
          ))}
        </nav>
        <div className="sidebar__bottom"><span className="mono muted">v0.x · demo shell</span></div>
      </aside>
      <div className="app-main">
        <header className="mobile-header">
          <button className="icon-btn" type="button" aria-label={open ? "Close navigation" : "Open navigation"} aria-expanded={open} aria-controls="app-sidebar-nav" onClick={() => setOpen((v) => !v)}>
            <span aria-hidden="true">☰</span>
          </button>
          <span className="brand brand--compact"><span className="brand__mark">PV</span><span>PaperViz</span></span>
        </header>
        <main className="container app-content">{children}</main>
      </div>
    </div>
  )
}
