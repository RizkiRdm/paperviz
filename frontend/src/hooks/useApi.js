import { useState, useCallback, useRef } from "react"

// useApi wraps async API calls with loading/error/retry state — eliminates
// the per-page boilerplate of manual setLoading/setError/catch patterns.
export function useApi(fn) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const fnRef = useRef(fn)
  fnRef.current = fn

  // execute runs the wrapped function, managing loading and error state.
  const execute = useCallback(async (...args) => {
    setLoading(true)
    setError(null)
    try {
      const result = await fnRef.current(...args)
      setLoading(false)
      return result
    } catch (err) {
      setLoading(false)
      setError(err)
      throw err
    }
  }, [])

  // retry re-runs the last execute call with the same arguments.
  const retry = useCallback(() => execute(), [execute])

  return { execute, loading, error, setError, retry }
}
