import { useState } from "react"
import { Link, useNavigate, useSearchParams } from "react-router-dom"
import { useSignIn } from "@clerk/clerk-react"
import { clerkErrorMessage } from "@/lib/auth"
import { resolveReturnTo } from "@/lib/draft"
import { Button } from "@/components/ui/button"

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  // Post-login destination: allowlisted returnTo from the auth gate, else
  // the existing /account landing so the normal flow is unchanged.
  const redirectTo = resolveReturnTo(searchParams.get("returnTo"))
  const { signIn, isLoaded } = useSignIn()
  const [error, setError] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleGoogle() {
    if (!isLoaded) return
    setError(null)
    try {
      await signIn.authenticateWithRedirect({
        strategy: "oauth_google",
        redirectUrl: "/sso-callback",
        redirectUrlComplete: redirectTo,
      })
    } catch (err) {
      if (import.meta.env.DEV) console.error("google sign-in failed", err)
      setError(clerkErrorMessage(err?.errors, "Google sign-in failed."))
    }
  }

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
      const attempt = await signIn.create({
        strategy: "password",
        identifier: email,
        password,
      })
      if (attempt.status === "complete") {
        navigate(redirectTo)
        return
      }
      setError(clerkErrorMessage(attempt.error, "Sign-in needs an extra step this page does not support."))
    } catch (err) {
      if (import.meta.env.DEV) console.error("login failed", err)
      setError(clerkErrorMessage(err?.errors, "Sign-in failed. Please try again."))
    } finally {
      setIsSubmitting(false)
    }
  }

  const disabled = isSubmitting || !isLoaded

  return (
    <div className="min-h-screen bg-[#f5f5f5] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center mb-6">
        <Link to="/" className="inline-flex items-center gap-2.5 group">
          <span className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] font-mono text-xs font-semibold text-white transition-transform group-hover:scale-105">
            PV
          </span>
          <span className="font-sans text-lg font-semibold tracking-tight text-[#0a0a0a]">
            PaperViz
          </span>
        </Link>
      </div>

      <div className="sm:mx-auto sm:w-full sm:max-w-[420px] px-4">
        <div className="rounded-[16px] border border-[#e5e5e5] bg-white p-8 shadow-xs">
          <div className="mb-6">
            <h1 className="font-['Satoshi',sans-serif] text-2xl font-medium tracking-tight text-[#0a0a0a]">
              Sign in
            </h1>
            <p className="mt-1 text-xs text-[#737373]">
              Return to your research workspace.
            </p>
          </div>

          <form className="space-y-4" onSubmit={handleSubmit}>
            <Button
              type="button"
              variant="outline"
              size="md"
              onClick={handleGoogle}
              disabled={disabled}
              className="w-full justify-center"
            >
              Continue with Google
            </Button>

            <div className="relative my-4">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-[#e5e5e5]" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-white px-2 font-mono text-[#a3a3a3]">or</span>
              </div>
            </div>

            <div>
              <label htmlFor="email" className="block text-xs font-medium text-[#737373] mb-1.5">
                Email
              </label>
              <input
                id="email"
                name="email"
                type="email"
                placeholder="you@example.com"
                autoComplete="email"
                disabled={disabled}
                className="w-full rounded-[6px] border border-[#e5e5e5] bg-white px-3 py-2 text-sm text-[#171717] placeholder-[#a3a3a3] transition-colors focus:border-[#2563eb] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-xs font-medium text-[#737373] mb-1.5">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                placeholder="Your password"
                autoComplete="current-password"
                disabled={disabled}
                className="w-full rounded-[6px] border border-[#e5e5e5] bg-white px-3 py-2 text-sm text-[#171717] placeholder-[#a3a3a3] transition-colors focus:border-[#2563eb] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20"
              />
            </div>

            {error && (
              <div className="rounded-[6px] border border-[#fecaca] bg-[#fef2f2] p-3 text-xs text-[#b91c1c]">
                {error}
              </div>
            )}

            <Button
              type="submit"
              size="md"
              disabled={disabled}
              className="w-full justify-center"
            >
              {isSubmitting ? "Signing in..." : "Sign in"}
            </Button>
          </form>

          <p className="mt-6 text-center text-xs text-[#737373]">
            Don't have an account?{" "}
            <Link
              to={redirectTo !== "/account" ? `/signup?returnTo=${encodeURIComponent(redirectTo)}` : "/signup"}
              className="text-[#2563eb] font-medium hover:underline"
            >
              Sign up
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
