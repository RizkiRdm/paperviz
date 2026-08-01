// UploadPage — PRD.md "User Flows > Primary Flow" steps 1-2: land, choose
// input, choose reading level, submit. On success, navigates to the result
// page for the returned document_id (App.jsx owns the actual routing/state
// switch — this component only calls onCreated with the new ID).
import { useState } from "react"
import { UploadDropzone } from "@/components/upload-dropzone"
import { ReadingLevelSelector } from "@/components/ui/reading-level-selector"
import { Button } from "@/components/ui/button"
import { ErrorBanner } from "@/components/ui/status-banners"
import { createDocument } from "@/lib/api"

// Human-readable messages for the snake_case error codes ARCHITECTURE.md's
// API contract can return. Kept as a flat map here (not a fancier lookup
// service) since there are only 5 codes and they're all specific to this
// one submit action.
const ERROR_MESSAGES = {
  no_text_layer:
    "This PDF appears to be a scanned image without selectable text. PaperViz requires a text-based PDF.",
  invalid_reading_level: "Please choose a reading level before submitting.",
  file_too_large: "That file is too large. PaperViz accepts PDFs up to 20MB.",
  missing_input: "Please upload a PDF or paste some text first.",
  invalid_file_type: "That file doesn't look like a valid PDF.",
  unknown_error: "Something went wrong. Please try again.",
}

const MAX_FILE_SIZE_BYTES = 20 * 1024 * 1024

export function UploadPage({ onCreated }) {
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
      onCreated(result.document_id)
    } catch (err) {
      setError(ERROR_MESSAGES[err.code] || ERROR_MESSAGES.unknown_error)
      setIsSubmitting(false)
    }
    // Deliberately not resetting isSubmitting in a finally block on the
    // success path — onCreated navigates away, so this component unmounts
    // and further state updates would be wasted/warn in the console.
  }

  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col justify-center px-6 py-16">
      <h1 className="text-hero text-ink-primary">Papers, in plain language</h1>
      <p className="mt-3 text-body text-ink-secondary">
        Upload a research paper or paste its text. PaperViz rewrites it at a
        reading level you choose, re-draws its charts, and checks that
        nothing important changed along the way.
      </p>

      <div className="mt-8">
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
      </div>

      <div className="mt-6 flex items-center justify-between gap-4">
        <ReadingLevelSelector value={readingLevel} onChange={setReadingLevel} />
        <Button onClick={handleSubmit} disabled={isSubmitting}>
          {isSubmitting ? "Processing…" : "Simplify"}
        </Button>
      </div>

      {error && (
        <div className="mt-4">
          <ErrorBanner message={error} />
        </div>
      )}
    </main>
  )
}
