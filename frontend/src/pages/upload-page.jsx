// ponytail: upload page redesign with Satoshi display headline, pill feature tags & hairline card surface
import { useState } from "react"
import { useNavigate, Link } from "react-router-dom"
import { UploadDropzone } from "@/components/upload-dropzone"
import { ReadingLevelSelector } from "@/components/ui/reading-level-selector"
import { Button } from "@/components/ui/button"
import { ErrorBanner } from "@/components/ui/status-banners"
import { createDocument } from "@/lib/api"
import { Sparkles, FileText, BarChart3, ArrowRight } from "lucide-react"

const ERROR_MESSAGES = {
  no_text_layer:
    "This PDF appears to be a scanned image without selectable text. PaperViz requires a text-based PDF.",
  invalid_reading_level: "Please choose a reading level before submitting.",
  file_too_large: "That file is too large. PaperViz accepts PDFs up to 20MB.",
  missing_input: "Please upload a PDF or paste some text first.",
  invalid_file_type: "That file doesn't look like a valid PDF.",
  network_timeout:
    "The request timed out. Please check your connection and try again.",
  unknown_error: "Something went wrong. Please try again.",
}

const MAX_FILE_SIZE_BYTES = 20 * 1024 * 1024

export function UploadPage() {
  const navigate = useNavigate()
  const [file, setFile] = useState(null)
  const [text, setText] = useState("")
  const [readingLevel, setReadingLevel] = useState("simplified")
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState(null)

  async function handleSubmit() {
    setError(null)

    if (!file && !text.trim()) {
      setError(ERROR_MESSAGES.missing_input)
      return
    }

    setIsSubmitting(true)
    try {
      const result = await createDocument({ file, text, readingLevel })
      navigate(`/${result.document_id}`)
    } catch (err) {
      setError(ERROR_MESSAGES[err.code] || ERROR_MESSAGES.unknown_error)
      setIsSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen bg-[#ffffff] bg-dotted-grid text-[#171717]">
      {/* Top Bar per DESIGN.md */}
      <header className="border-b border-[#e5e5e5] bg-white/80 backdrop-blur-xs">
        <div className="mx-auto flex max-w-[1200px] items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold tracking-tight text-[#0a0a0a]">PaperViz</span>
          </div>
          <div className="flex items-center gap-3">
            <Link to="/dashboard" className="text-xs text-[#737373] hover:text-[#0a0a0a] transition-colors">
              Dashboard
            </Link>
            <span className="h-3.5 w-px bg-[#e5e5e5]" />
            <span className="inline-flex items-center gap-1 rounded-full border border-[#e5e5e5] bg-[#f5f5f5] px-3 py-1 text-xs text-[#737373]">
              Ephemeral 7-Day Storage
            </span>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="mx-auto max-w-[1200px] px-6 py-16">
        {/* Hero Header Stack with Pill Feature Tags */}
        <div className="mx-auto max-w-3xl text-center">
          <div className="mb-6 inline-flex flex-wrap items-center justify-center gap-2">
            <span className="inline-flex items-center gap-1.5 rounded-full border border-[#e5e5e5] bg-white px-3.5 py-1 text-xs font-medium text-[#171717] shadow-2xs">
              <span className="text-[#ea580c]">⚡</span> Plain Language
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-full border border-[#e5e5e5] bg-white px-3.5 py-1 text-xs font-medium text-[#171717] shadow-2xs">
              <span className="text-[#7c3aed]">📊</span> Re-visualized Charts
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-full border border-[#e5e5e5] bg-white px-3.5 py-1 text-xs font-medium text-[#171717] shadow-2xs">
              <span className="text-[#16a34a]">🛡️</span> Claim Verification
            </span>
          </div>

          <h1 className="font-satoshi text-4xl sm:text-5xl font-medium tracking-tight text-[#0a0a0a] leading-tight">
            Papers, in plain language.
          </h1>
          <p className="mt-4 text-base text-[#737373] max-w-xl mx-auto leading-relaxed">
            Transform dense academic PDFs into clear, verified summaries with interactive charts. Ready to share in seconds.
          </p>
        </div>

        {/* Dashboard Card Container per DESIGN.md */}
        <div className="mt-10 mx-auto max-w-2xl rounded-[16px] border border-[#e5e5e5] bg-white p-6 sm:p-8 shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]">
          <UploadDropzone
            file={file}
            text={text}
            onFileChange={(f) => {
              if (f.size > MAX_FILE_SIZE_BYTES) {
                setFile(null)
                setError(ERROR_MESSAGES.file_too_large)
                return
              }
              if (f.type !== "application/pdf") {
                setFile(null)
                setError(ERROR_MESSAGES.invalid_file_type)
                return
              }
              setFile(f)
              setError(null)
            }}
            onTextChange={(t) => {
              setText(t)
              setError(null)
            }}
          />

          <div className="mt-6 flex flex-col sm:flex-row items-center justify-between gap-4 pt-4 border-t border-[#f5f5f5]">
            <div className="flex items-center gap-2">
              <span className="text-xs text-[#737373] font-medium">Target Level:</span>
              <ReadingLevelSelector value={readingLevel} onChange={setReadingLevel} />
            </div>

            <Button
              onClick={handleSubmit}
              disabled={isSubmitting}
              className="w-full sm:w-auto min-w-[140px]"
            >
              {isSubmitting ? (
                "Processing…"
              ) : (
                <span className="flex items-center justify-center gap-1.5">
                  Simplify Paper <ArrowRight className="h-4 w-4" />
                </span>
              )}
            </Button>
          </div>

          {error && (
            <div className="mt-4">
              <ErrorBanner message={error} />
            </div>
          )}
        </div>
      </main>
    </div>
  )
}

