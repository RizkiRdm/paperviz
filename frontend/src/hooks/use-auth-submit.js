import { useState, useCallback } from "react"

export function useAuthSubmit(endpoint, errorMessages) {
  const [error, setError] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const submit = useCallback(
    async (payload) => {
      setError(null)
      setIsSubmitting(true)
      try {
        const res = await fetch(endpoint, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        })
        const data = await res.json()
        if (!res.ok) {
          setError(errorMessages[data.error] || errorMessages.internal_error)
          setIsSubmitting(false)
          return false
        }
        return true
      } catch {
        setError(errorMessages.internal_error)
        setIsSubmitting(false)
        return false
      }
    },
    [endpoint, errorMessages]
  )

  return { error, isSubmitting, submit, setError }
}