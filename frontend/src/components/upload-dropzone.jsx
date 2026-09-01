// ponytail: upload dropzone redesign matching DESIGN.md hairline cards & pill feature tags
import { useRef, useState } from "react"
import { Upload, FileText } from "lucide-react"
import { cn } from "@/lib/utils"

export function UploadDropzone({ file, text, mode, onFileChange, onTextChange }) {
  const [isDragging, setIsDragging] = useState(false)
  const inputRef = useRef(null)

  function handleFiles(fileList) {
    const selected = fileList?.[0]
    if (selected) onFileChange(selected)
  }

  return (
    <div>
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

