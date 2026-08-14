import { useEffect, useRef, useState, useCallback } from "react"
import { getDocument } from "@/lib/api"

const POLL_INTERVAL_MS = 2000
const POLL_SOFT_WARN_MS = 120000
const POLL_TIMEOUT_MS = 600000

export function useDocumentPoll(documentId) {
  const [doc, setDoc] = useState(null)
  const [error, setError] = useState(null)
  const [notFound, setNotFound] = useState(false)
  const [timedOut, setTimedOut] = useState(false)
  const [takingLong, setTakingLong] = useState(false)
  const [retryNonce, setRetryNonce] = useState(0)

  const pollTimer = useRef(null)
  const pollStartRef = useRef(Date.now())

  const retry = useCallback(() => setRetryNonce((n) => n + 1), [])

  useEffect(() => {
    let cancelled = false
    setTimedOut(false)
    setTakingLong(false)
    pollStartRef.current = Date.now()

    async function poll() {
      try {
        const fetched = await getDocument(documentId)
        if (cancelled) return
        setDoc(fetched)
        if (fetched.status === "processing") {
          const elapsed = Date.now() - pollStartRef.current
          if (elapsed > POLL_SOFT_WARN_MS) {
            setTakingLong(true)
          }
          if (elapsed > POLL_TIMEOUT_MS) {
            setTimedOut(true)
            return
          }
          pollTimer.current = setTimeout(poll, POLL_INTERVAL_MS)
        }
      } catch (err) {
        if (!cancelled) {
          if (err.code === "not_found") {
            setNotFound(true)
          } else {
            setError(
              err.code === "network_timeout"
                ? "The request timed out. Please check your connection."
                : "We couldn't load this document. Please try again."
            )
          }
        }
      }
    }
    poll()
    return () => { cancelled = true; clearTimeout(pollTimer.current) }
  }, [documentId, retryNonce])

  return { doc, error, notFound, timedOut, takingLong, retry }
}