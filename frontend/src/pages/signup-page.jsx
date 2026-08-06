import { useState } from "react"
import { useNavigate, Link } from "react-router-dom"
import { Button } from "@/components/ui/button"

const ERROR_MESSAGES = {
  email_taken: "An account with this email already exists.",
  invalid_request: "Please fill in all fields.",
  invalid_email: "Please enter a valid email address.",
  password_too_short: "Password must be at least 8 characters.",
  internal_error: "Something went wrong. Please try again.",
}

export function SignupPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    setError(null)

    if (!email.trim() || !password) {
      setError(ERROR_MESSAGES.invalid_request)
      return
    }

    setIsSubmitting(true)
    try {
      const res = await fetch("/api/auth/signup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email.trim(), password }),
      })
      const data = await res.json()

      if (!res.ok) {
        setError(ERROR_MESSAGES[data.error] || ERROR_MESSAGES.internal_error)
        setIsSubmitting(false)
        return
      }

      navigate("/dashboard")
    } catch {
      setError(ERROR_MESSAGES.internal_error)
      setIsSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center px-6">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold text-[#0a0a0a]">PaperViz</span>
          </Link>
          <h1 className="font-satoshi text-2xl font-medium text-[#0a0a0a]">Create your account</h1>
          <p className="mt-1 text-sm text-[#737373]">Start simplifying papers today</p>
        </div>

        <form onSubmit={handleSubmit} className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
          <div className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-xs font-medium text-[#737373] mb-1.5">Email</label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => { setEmail(e.target.value); setError(null) }}
                className="w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 focus:border-[#2563eb]"
                placeholder="you@example.com"
                autoComplete="email"
                disabled={isSubmitting}
              />
            </div>
            <div>
              <label htmlFor="password" className="block text-xs font-medium text-[#737373] mb-1.5">Password</label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => { setPassword(e.target.value); setError(null) }}
                className="w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 focus:border-[#2563eb]"
                placeholder="At least 8 characters"
                autoComplete="new-password"
                disabled={isSubmitting}
              />
            </div>
          </div>

          {error && (
            <div className="mt-4 rounded-[6px] bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
              {error}
            </div>
          )}

          <Button type="submit" disabled={isSubmitting} className="mt-4 w-full">
            {isSubmitting ? "Creating account..." : "Create account"}
          </Button>

          <p className="mt-4 text-center text-xs text-[#737373]">
            Already have an account?{" "}
            <Link to="/login" className="text-[#2563eb] hover:underline">Sign in</Link>
          </p>
        </form>
      </div>
    </div>
  )
}
