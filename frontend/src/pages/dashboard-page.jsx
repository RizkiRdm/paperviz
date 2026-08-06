import { useEffect, useState } from "react"
import { useNavigate, Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { ArrowRight, FileText, Plus } from "lucide-react"

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

        {loading ? (
          <div className="flex justify-center py-12">
            <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#2563eb] border-t-transparent" />
          </div>
        ) : docs.length === 0 ? (
          <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-12 text-center">
            <FileText className="h-10 w-10 text-[#a3a3a3] mx-auto mb-4" />
            <h2 className="text-sm font-medium text-[#0a0a0a]">No documents yet</h2>
            <p className="mt-1 text-xs text-[#737373]">Upload your first PDF to get started.</p>
            <Link to="/">
              <Button className="mt-6 gap-1.5">
                Upload your first PDF <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>
        ) : (
          <div className="rounded-[12px] border border-[#e5e5e5] bg-white overflow-hidden">
            {docs.map((doc, i) => (
              <Link
                key={doc.id}
                to={`/${doc.id}`}
                className={`flex items-center justify-between px-5 py-3.5 hover:bg-[#f5f5f5] transition-colors ${
                  i < docs.length - 1 ? "border-b border-[#e5e5e5]" : ""
                }`}
              >
                <div className="flex items-center gap-3 min-w-0">
                  <FileText className="h-4 w-4 text-[#737373] shrink-0" />
                  <span className="text-sm text-[#171717] truncate font-mono">{doc.id}</span>
                </div>
                <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-medium shrink-0 ml-4 ${STATUS_STYLES[doc.status] || "bg-gray-100 text-gray-600"}`}>
                  {STATUS_LABELS[doc.status] || doc.status}
                </span>
              </Link>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
