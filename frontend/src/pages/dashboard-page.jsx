import { useEffect, useState } from "react"
import { useNavigate, Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { ArrowRight, FileText, Plus, Calendar, Star, Pencil, Trash2, X, Check } from "lucide-react"

const STATUS_STYLES = {
  complete: "bg-[#dcfce7] text-[#16a34a]",
  processing: "bg-blue-50 text-[#2563eb]",
  failed: "bg-red-50 text-red-600",
  verification_failed: "bg-yellow-50 text-yellow-700",
}

const STATUS_LABELS = {
  complete: "Complete",
  processing: "Processing",
  failed: "Failed",
  verification_failed: "Verification issue",
}

export function DashboardPage() {
  const navigate = useNavigate()
  const [user, setUser] = useState(null)
  const [docs, setDocs] = useState([])
  const [loading, setLoading] = useState(true)
  const [authChecked, setAuthChecked] = useState(false)
  const [filter, setFilter] = useState("all") // "all" | "saved"
  const [editingId, setEditingId] = useState(null)
  const [editTitle, setEditTitle] = useState("")
  const [deleteConfirmId, setDeleteConfirmId] = useState(null)

  useEffect(() => {
    async function checkAuth() {
      try {
        const res = await fetch("/api/auth/me")
        if (!res.ok) {
          navigate("/login")
          return
        }
        const data = await res.json()
        setUser(data)
        setAuthChecked(true)
      } catch {
        navigate("/login")
      }
    }
    checkAuth()
  }, [navigate])

  useEffect(() => {
    if (!authChecked) return

    async function fetchDocs() {
      try {
        const res = await fetch("/api/documents")
        if (res.ok) {
          const data = await res.json()
          setDocs(data.documents || [])
        }
      } catch {
        // silent — empty state shown
      } finally {
        setLoading(false)
      }
    }
    fetchDocs()
  }, [authChecked])

  async function handleToggleSaved(e, doc) {
    e.preventDefault()
    e.stopPropagation()
    try {
      const res = await fetch(`/api/documents/${doc.id}/save`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ saved: !doc.saved }),
      })
      if (res.ok) {
        setDocs(docs.map(d => d.id === doc.id ? { ...d, saved: !d.saved } : d))
      }
    } catch {}
  }

  async function handleRename(e, docId) {
    e.preventDefault()
    e.stopPropagation()
    if (!editTitle.trim()) return
    try {
      const res = await fetch(`/api/documents/${docId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title: editTitle.trim() }),
      })
      if (res.ok) {
        setDocs(docs.map(d => d.id === docId ? { ...d, title: editTitle.trim() } : d))
        setEditingId(null)
        setEditTitle("")
      }
    } catch {}
  }

  async function handleDelete(e, docId) {
    e.preventDefault()
    e.stopPropagation()
    try {
      const res = await fetch(`/api/documents/${docId}`, { method: "DELETE" })
      if (res.ok) {
        setDocs(docs.filter(d => d.id !== docId))
        setDeleteConfirmId(null)
      }
    } catch {}
  }

  if (!authChecked) {
    return (
      <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white bg-dotted-grid">
      <header className="border-b border-[#e5e5e5] bg-white/80 backdrop-blur-xs">
        <div className="mx-auto flex max-w-[1200px] items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold text-[#0a0a0a]">PaperViz</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-xs text-[#737373]">{user?.email}</span>
            <Link to="/">
              <Button variant="secondary" className="h-8 px-3 text-xs gap-1.5">
                <Plus className="h-3.5 w-3.5" /> New Upload
              </Button>
            </Link>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-[900px] px-6 py-10">
        <h1 className="font-satoshi text-2xl font-medium text-[#0a0a0a] mb-6">Your Documents</h1>

        {docs.length > 0 && (
          <div className="flex items-center gap-1 mb-4">
            <button
              onClick={() => setFilter("all")}
              className={`px-3 py-1.5 text-xs font-medium rounded-full transition-colors ${
                filter === "all"
                  ? "bg-[#0a0a0a] text-white"
                  : "bg-[#f5f5f5] text-[#737373] hover:bg-[#e5e5e5]"
              }`}
            >
              All
            </button>
            <button
              onClick={() => setFilter("saved")}
              className={`px-3 py-1.5 text-xs font-medium rounded-full transition-colors ${
                filter === "saved"
                  ? "bg-[#0a0a0a] text-white"
                  : "bg-[#f5f5f5] text-[#737373] hover:bg-[#e5e5e5]"
              }`}
            >
              Saved
            </button>
          </div>
        )}

        {loading ? (
          <div className="flex justify-center py-12">
            <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent" />
          </div>
        ) : docs.length === 0 ? (
          <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-12 text-center">
            <FileText className="h-10 w-10 text-[#a3a3a3] mx-auto mb-4" />
            <h2 className="text-sm font-medium text-[#0a0a0a]">
              {filter === "saved" ? "No saved papers yet." : "No documents yet."}
            </h2>
            {filter !== "saved" && (
              <>
                <p className="mt-1 text-xs text-[#737373]">Upload your first PDF to get started.</p>
                <Link to="/">
                  <Button className="mt-6 gap-1.5">
                    Upload your first PDF <ArrowRight className="h-4 w-4" />
                  </Button>
                </Link>
              </>
            )}
          </div>
        ) : (
          <div className="rounded-[12px] border border-[#e5e5e5] bg-white overflow-hidden">
            {(() => {
              const filteredDocs = filter === "saved" ? docs.filter(d => d.saved) : docs

              if (filteredDocs.length === 0) {
                return (
                  <div className="p-12 text-center">
                    <FileText className="h-10 w-10 text-[#a3a3a3] mx-auto mb-4" />
                    <h2 className="text-sm font-medium text-[#0a0a0a]">No saved papers yet.</h2>
                  </div>
                )
              }

              return filteredDocs.map((doc, i) => (
                <div
                  key={doc.id}
                  className={`relative px-5 py-4 hover:bg-[#f5f5f5] transition-colors ${
                    i < filteredDocs.length - 1 ? "border-b border-[#e5e5e5]" : ""
                  }`}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2.5 mb-0.5">
                        <button
                          onClick={(e) => handleToggleSaved(e, doc)}
                          className="shrink-0 mt-0.5"
                        >
                          <Star
                            className={`h-4 w-4 ${
                              doc.saved
                                ? "fill-[#ea580c] text-[#ea580c]"
                                : "text-[#a3a3a3] hover:text-[#737373]"
                            }`}
                          />
                        </button>
                        {editingId === doc.id ? (
                          <form onSubmit={(e) => handleRename(e, doc.id)} className="flex items-center gap-1 flex-1">
                            <input
                              type="text"
                              value={editTitle}
                              onChange={(e) => setEditTitle(e.target.value)}
                              onBlur={() => setEditingId(null)}
                              autoFocus
                              className="text-sm font-medium text-[#171717] border border-[#2563eb] rounded px-1.5 py-0.5 flex-1 outline-none"
                            />
                            <button type="submit" className="text-[#16a34a] hover:text-[#15803d]">
                              <Check className="h-4 w-4" />
                            </button>
                            <button type="button" onClick={() => setEditingId(null)} className="text-[#737373] hover:text-[#171717]">
                              <X className="h-4 w-4" />
                            </button>
                          </form>
                        ) : (
                          <button
                            onClick={(e) => {
                              e.preventDefault()
                              e.stopPropagation()
                              setEditingId(doc.id)
                              setEditTitle(doc.title || "")
                            }}
                            className="text-sm font-medium text-[#171717] truncate hover:text-[#2563eb] text-left flex-1"
                          >
                            {doc.title || "Untitled paper"}
                          </button>
                        )}
                      </div>
                      {doc.summary_preview ? (
                        <p className="text-xs text-[#737373] line-clamp-2 ml-6.5 mt-0.5">
                          {doc.summary_preview}
                        </p>
                      ) : null}
                      <div className="flex items-center gap-2 ml-6.5 mt-1.5">
                        <span className="flex items-center gap-1 text-[11px] text-[#a3a3a3]">
                          <Calendar className="h-3 w-3" />
                          {new Date(doc.created_at * 1000).toLocaleDateString(undefined, {
                            year: "numeric",
                            month: "short",
                            day: "numeric",
                          })}
                        </span>
                        {doc.chart_count > 0 ? (
                          <span className="text-[11px] text-[#a3a3a3]">
                            {doc.chart_count} figure{doc.chart_count !== 1 ? "s" : ""}{" "}
                            {doc.explanation_count > 0
                              ? `\u00b7 ${doc.explanation_count} explanation${doc.explanation_count !== 1 ? "s" : ""}`
                              : null}
                          </span>
                        ) : null}
                      </div>
                    </div>
                    <div className="flex items-center gap-2 shrink-0 mt-0.5">
                      <span
                        className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium ${
                          STATUS_STYLES[doc.status] || "bg-gray-100 text-gray-600"
                        }`}
                      >
                        {STATUS_LABELS[doc.status] || doc.status}
                      </span>
                      {deleteConfirmId === doc.id ? (
                        <div className="flex items-center gap-1">
                          <button
                            onClick={(e) => handleDelete(e, doc.id)}
                            className="text-[11px] text-white bg-red-600 hover:bg-red-700 px-2 py-0.5 rounded"
                          >
                            Confirm
                          </button>
                          <button
                            onClick={(e) => { e.preventDefault(); e.stopPropagation(); setDeleteConfirmId(null) }}
                            className="text-[11px] text-[#737373] hover:text-[#171717] px-1"
                          >
                            Cancel
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={(e) => {
                            e.preventDefault()
                            e.stopPropagation()
                            setDeleteConfirmId(doc.id)
                          }}
                          className="text-[#a3a3a3] hover:text-red-600"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              ))
            })()}
          </div>
        )}
      </main>
    </div>
  )
}
