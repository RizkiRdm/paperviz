// ponytail: upload dropzone redesign matching DESIGN.md hairline cards & pill feature tags
import { useRef, useState } from "react"
import { Upload, FileText, Sparkles, ShieldCheck, BarChart3 } from "lucide-react"
import { cn } from "@/lib/utils"

export function UploadDropzone({ file, text, onFileChange, onTextChange }) {
  const [mode, setMode] = useState("file") // "file" | "text"
  const [isDragging, setIsDragging] = useState(false)
  const inputRef = useRef(null)

  function handleFiles(fileList) {
    const selected = fileList?.[0]
    if (selected) onFileChange(selected)
  }

  return (
    <div>
      {/* Mode toggle pill selector per DESIGN.md */}
      <div className="mb-4 flex items-center justify-between">
        <div className="inline-flex gap-1 rounded-full border border-[#e5e5e5] bg-[#ffffff] p-1 shadow-xs">
          <button
            type="button"
            onClick={() => setMode("file")}
            className={cn(
              "rounded-full px-3.5 py-1 text-xs font-medium transition-colors",
              mode === "file" ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717]",
            )}
          >
            Upload PDF
          </button>
          <button
            type="button"
            onClick={() => setMode("text")}
            className={cn(
              "rounded-full px-3.5 py-1 text-xs font-medium transition-colors",
              mode === "text" ? "bg-[#0a0a0a] text-white" : "text-[#737373] hover:text-[#171717]",
            )}
          >
            Paste Text
          </button>
        </div>

        <span className="text-xs text-[#737373] flex items-center gap-1">
          <ShieldCheck className="h-3.5 w-3.5 text-[#16a34a]" />
          Instant verification
        </span>
      </div>

      {mode === "file" ? (
        <>
          <input
            ref={inputRef}
            type="file"
            accept="application/pdf"
            className="hidden"
            onChange={(e) => handleFiles(e.target.files)}
          />
          <button
            type="button"
            aria-label="Upload a PDF file"
            onDragOver={(e) => {
              e.preventDefault()
              setIsDragging(true)
            }}
            onDragLeave={() => setIsDragging(false)}
            onDrop={(e) => {
              e.preventDefault()
              setIsDragging(false)
              handleFiles(e.dataTransfer.files)
            }}
            onClick={() => inputRef.current?.click()}
            className={cn(
              "flex w-full min-h-[220px] cursor-pointer flex-col items-center justify-center gap-3 rounded-[12px] border border-dashed p-8 text-center bg-white transition-all hover:border-[#a3a3a3]",
              isDragging ? "border-solid border-[#2563eb] bg-[#f5f5f5]" : "border-[#e5e5e5]",
            )}
          >
            {file ? (
              <div className="flex flex-col items-center gap-2">
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[#dbeaff] text-[#2563eb]">
                  <FileText className="h-6 w-6" aria-hidden="true" />
                </div>
                <p className="text-sm font-medium text-[#171717]">{file.name}</p>
                <p className="text-xs text-[#737373]">Click or drag to replace PDF</p>
              </div>
            ) : (
              <div className="flex flex-col items-center gap-2">
                <div className="flex h-12 w-12 items-center justify-center rounded-full bg-[#f5f5f5] text-[#737373]">
                  <Upload className="h-5 w-5" aria-hidden="true" />
                </div>
                <p className="text-sm font-medium text-[#171717]">
                  Drop research paper PDF here, or <span className="text-[#2563eb] underline">browse</span>
                </p>
                <p className="text-xs text-[#737373]">Text-based PDFs up to 20MB</p>
              </div>
            )}
          </button>
        </>
      ) : (
        <>
          <label htmlFor="paper-text" className="mb-2 block text-xs font-medium text-[#737373] uppercase tracking-wider">
            Paper Content
          </label>
          <textarea
            id="paper-text"
            value={text}
            onChange={(e) => onTextChange(e.target.value)}
            placeholder="Paste raw academic paper text here…"
            rows={8}
            className="w-full rounded-[6px] border border-[#000000] bg-white p-4 text-sm text-[#171717] placeholder-[#737373] focus:outline-none focus:ring-2 focus:ring-[#2563eb]"
          />
        </>
      )}
    </div>
  )
}

