import { useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { SAMPLE_DOCUMENT_ID } from "@/lib/sample"

// /sample redirects to the seeded sample document's result view. The seed
// (migration 023) is a real complete document with user_id NULL, readable
// anonymously, so no special rendering is needed here.
export function SamplePage() {
  const navigate = useNavigate()

  useEffect(() => {
    navigate(`/${SAMPLE_DOCUMENT_ID}`, { replace: true })
  }, [navigate])

  return null
}
