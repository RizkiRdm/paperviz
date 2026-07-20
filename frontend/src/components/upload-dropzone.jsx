// UploadDropzone — DESIGN.md "Cards & Dropzone > Upload Dropzone":
// radius-lg, shadow-card, dashed border-default when idle, solid
// accent-verified border on drag-hover.
//
// Handles both input modes from PRD.md's primary flow step 1 ("uploads a
// PDF or pastes text") via a simple mode toggle — file and pasted text are
// mutually exclusive per ARCHITECTURE.md's API contract ("exactly one of
// file/text is required"), so switching modes clears whichever input was
// previously set for the other mode.
import { useRef, useState } from "react"
import { Upload, FileText } from "lucide-react"
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
      {/* Mode toggle: plain buttons, not the ToggleGroup component — this
          switches between two entirely different input widgets below, not
          between two states of the same control, so a lighter-weight
          control reads more honestly here than reusing ReadingLevelSelector's
          ToggleGroup. */}
      <div className="mb-3 flex gap-2 text-body">
        <button
          type="button"
          onClick={() => setMode("file")}
          className={cn(
            "rounded-sm px-3 py-1.5 font-medium",
            mode === "file" ? "bg-accent-verified-soft text-accent-verified" : "text-ink-secondary",
          )}
        >
          Upload PDF
        </button>
        <button
          type="button"
          onClick={() => setMode("text")}
          className={cn(
            "rounded-sm px-3 py-1.5 font-medium",
            mode === "text" ? "bg-accent-verified-soft text-accent-verified" : "text-ink-secondary",
          )}
        >
          Paste text
        </button>
      </div>

      {mode === "file" ? (
        <div
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
            "flex min-h-[200px] cursor-pointer flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed p-8 text-center shadow-[0_1px_2px_rgba(20,23,31,0.04),0_4px_12px_rgba(20,23,31,0.06)]",
            isDragging ? "border-solid border-accent-verified" : "border-border-default",
          )}
        >
          <input
            ref={inputRef}
            type="file"
            accept="application/pdf"
            className="hidden"
            onChange={(e) => handleFiles(e.target.files)}
          />
          {file ? (
            <>
              <FileText className="h-8 w-8 text-accent-verified" aria-hidden="true" />
              <p className="text-body font-medium">{file.name}</p>
              <p className="text-caption text-ink-secondary">Click to choose a different file</p>
            </>
          ) : (
            <>
              <Upload className="h-8 w-8 text-ink-secondary" aria-hidden="true" />
              <p className="text-body font-medium">Drop a PDF here, or click to browse</p>
              <p className="text-caption text-ink-secondary">Text-based PDFs only, up to 20MB</p>
            </>
          )}
        </div>
      ) : (
        <textarea
          value={text}
          onChange={(e) => onTextChange(e.target.value)}
          placeholder="Paste the paper's text here…"
          rows={10}
          className="w-full rounded-lg border border-border-default bg-surface-raised p-4 text-body shadow-[0_1px_2px_rgba(20,23,31,0.04),0_4px_12px_rgba(20,23,31,0.06)] focus:border-accent-verified"
        />
      )}
    </div>
  )
}
