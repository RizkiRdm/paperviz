import { useEffect, useState } from "react"

// useCurrentUser resolves the signed-in user once per mount. A 401 is the
// anonymous case, not an error, so it resolves to null rather than throwing.
export function useCurrentUser() {
  const [user, setUser] = useState(null)
  const [checked, setChecked] = useState(false)

  useEffect(() => {
    let cancelled = false
    fetch("/api/auth/me")
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (!cancelled) setUser(data)
      })
      .catch((err) => {
        // Offline or a 5xx: leave user null so the header keeps offering Sign in
        // rather than breaking the page.
        console.error("Failed to resolve current user", err)
      })
      .finally(() => {
        if (!cancelled) setChecked(true)
      })
    return () => {
      cancelled = true
    }
  }, [])

  return { user, checked }
}