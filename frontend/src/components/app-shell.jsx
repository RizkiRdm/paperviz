import { useState } from "react"
import { Link } from "react-router-dom"

export function AppShell({ active, children }) {
  const [open, setOpen] = useState(false)
  const nav = [
    { id: "home", label: "Home", to: "/" },
    { id: "papers", label: "My Papers", to: "/papers" },
    { id: "agents", label: "Agents", to: "/agents" },
  ]
  const account = [
    { id: "account", label: "Account", to: "/account" },
    { id: "settings", label: "Settings", to: "/settings" },
    { id: "help", label: "Help", to: "/help" },
  ]

  return (
    <div className="app-shell">
      <aside className={`sidebar ${open ? "is-open" : ""}`}>
        <div className="sidebar__top">
          <Link to="/" className="brand"><span className="brand__mark">PV</span><span>PaperViz</span></Link>
        </div>
        <nav className="sidebar__nav" aria-label="Application">
          <span className="nav-label">Workspace</span>
          {nav.map((item) => (
            <Link key={item.id} to={item.to} className={`side-item ${active === item.id ? "is-active" : ""}`} onClick={() => setOpen(false)}>
              <span>{item.id === "home" ? "⌂" : item.id === "papers" ? "□" : "⌘"}</span>{item.label}
            </Link>
          ))}
          <span className="nav-label nav-label--spaced">Account</span>
          {account.map((item) => (
            <Link key={item.id} to={item.to} className={`side-item ${active === item.id ? "is-active" : ""}`} onClick={() => setOpen(false)}>
              <span>{item.id === "account" ? "◎" : item.id === "settings" ? "⚙" : "?"}</span>{item.label}
            </Link>
          ))}
        </nav>
        <div className="sidebar__bottom"><span className="mono muted">v0.x · demo shell</span></div>
      </aside>
      <div className="app-main">
        <header className="mobile-header">
          <button className="icon-btn" type="button" aria-label="Open navigation" onClick={() => setOpen((v) => !v)}>☰</button>
          <span className="brand brand--compact"><span className="brand__mark">PV</span><span>PaperViz</span></span>
        </header>
        <main className="container app-content">{children}</main>
      </div>
    </div>
  )
}
