import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { useSignIn } from "@clerk/clerk-react"
import { ArrowLeft } from "lucide-react"
import { Button } from "@/components/ui/button"
import { clerkErrorMessage } from "@/lib/auth"

export function LoginPage() {
  const navigate = useNavigate()
  const { signIn, isLoaded } = useSignIn()
  const [error, setError] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    const form = e.currentTarget
    const email = form.email.value.trim()
    const password = form.password.value
    if (!email || !password) {
      setError("Please fill in all fields.")
      return
    }

    setError(null)
    setIsSubmitting(true)
    try {
      // signIn.create with the password strategy is the one-call form: it takes
      // identifier + password together. attemptFirstFactor is only valid after
      // an existing sign-in attempt exists, so it 422s here, and the resource
      // has no attemptEmailAddress method at all — both failures look like a
      // generic error for every single sign-in.
      const attempt = await signIn.create({
        strategy: "password",
        identifier: email,
        password,
      })
      if (attempt.status === "complete") {
        navigate("/account")
        return
      }
      // Password sign-in completed but the session is not active yet, which
      // means Clerk wants another factor. PaperViz enables no second factor,
      // so this is reported rather than left on a dead button.
      setError(clerkErrorMessage(attempt.error, "Sign-in needs an extra step this page does not support."))
    } catch (err) {
      if (import.meta.env.DEV) console.error("login failed", err)
      setError(clerkErrorMessage(err?.errors, "Sign-in failed. Please try again."))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="relative min-h-screen bg-white bg-dotted-grid flex items-center justify-center px-6">
      <header className="absolute left-6 top-6">
        <Link
          to="/"
          className="inline-flex items-center gap-1.5 rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5 text-sm font-medium text-[#171717] hover:bg-[#f5f5f5] transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </Link>
      </header>
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold tracking-tight text-[#0a0a0a]">PaperViz</span>
          </Link>
          <h1 className="font-satoshi text-2xl font-medium text-[#0a0a0a]">Welcome back</h1>
          <p className="mt-1 text-sm text-[#737373]">Sign in to your account</p>
        </div>

        <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-xs font-medium text-[#737373] mb-1.5">Email</label>
              <input
                id="email"
                name="email"
                type="email"
                className="w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 focus:border-[#2563eb]"
                placeholder="you@example.com"
                autoComplete="email"
                disabled={isSubmitting || !isLoaded}
              />
            </div>
            <div>
              <label htmlFor="password" className="block text-xs font-medium text-[#737373] mb-1.5">Password</label>
              <input
                id="password"
                name="password"
                type="password"
                className="w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 focus:border-[#2563eb]"
                placeholder="Your password"
                autoComplete="current-password"
                disabled={isSubmitting || !isLoaded}
              />
            </div>

            {error && (
              <div className="rounded-[6px] bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
                {error}
              </div>
            )}

            <Button type="submit" disabled={isSubmitting || !isLoaded} className="w-full">
              {isSubmitting ? "Signing in..." : "Sign in"}
            </Button>
          </form>

          <p className="mt-4 text-center text-xs text-[#737373]">
            Don't have an account?{" "}
            <Link to="/signup" className="text-[#2563eb] hover:underline">Sign up</Link>
          </p>
        </div>
      </div>
    </div>
  )
}