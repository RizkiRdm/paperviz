import { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { Home, Bot, User, Menu, X } from "lucide-react"

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
    { id: "home", label: "Home", to: "/", icon: Home },
    { id: "agents", label: "Agents", to: "/agents", icon: Bot },
  ]
  const account = [
    { id: "account", label: "Account", to: "/account", icon: User },
  ]

  return (
    <div className="app-shell">
      <aside id="app-sidebar" inert={!open} className={`sidebar ${open ? "is-open" : ""}`}>
        <div className="sidebar__top">
          <Link to="/" className="brand"><span className="brand__mark">PV</span><span>PaperViz</span></Link>
        </div>
        <nav id="app-sidebar-nav" className="sidebar__nav" aria-label="Application">
          <span className="nav-label">Workspace</span>
          {nav.map((item) => {
            const Icon = item.icon
            return (
              <Link key={item.id} to={item.to} className={`side-item ${active === item.id ? "is-active" : ""}`} aria-current={active === item.id ? "page" : undefined} onClick={() => setOpen(false)}>
                <Icon className="h-4 w-4 shrink-0 text-[#737373]" aria-hidden="true" />
                <span>{item.label}</span>
              </Link>
            )
          })}
          <span className="nav-label nav-label--spaced">Account</span>
          {account.map((item) => {
            const Icon = item.icon
            return (
              <Link key={item.id} to={item.to} className={`side-item ${active === item.id ? "is-active" : ""}`} aria-current={active === item.id ? "page" : undefined} onClick={() => setOpen(false)}>
                <Icon className="h-4 w-4 shrink-0 text-[#737373]" aria-hidden="true" />
                <span>{item.label}</span>
              </Link>
            )
          })}
        </nav>
        <div className="sidebar__bottom"><span className="mono muted">v0.x · PaperViz</span></div>
      </aside>
      <div className="app-main">
        <header className="mobile-header">
          <button className="icon-btn" type="button" aria-label={open ? "Close navigation" : "Open navigation"} aria-expanded={open} aria-controls="app-sidebar-nav" onClick={() => setOpen((v) => !v)}>
            {open ? <X className="h-5 w-5 text-[#171717]" aria-hidden="true" /> : <Menu className="h-5 w-5 text-[#171717]" aria-hidden="true" />}
          </button>
          <span className="brand brand--compact"><span className="brand__mark">PV</span><span>PaperViz</span></span>
        </header>
        <main className="container app-content">{children}</main>
      </div>
    </div>
  )
}
