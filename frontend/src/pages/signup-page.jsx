import { useState } from "react"
import { Link, useNavigate, useSearchParams } from "react-router-dom"
import { useSignUp, useClerk } from "@clerk/clerk-react"
import { clerkErrorMessage } from "@/lib/auth"
import { capture, identify } from "@/lib/analytics"
import { resolveReturnTo } from "@/lib/draft"
import { Button } from "@/components/ui/button"

const STEP_DETAILS = "details"
const STEP_CODE = "code"

export function SignupPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  // Post-signup destination: allowlisted returnTo from the auth gate,
  // else the existing /account landing.
  const redirectTo = resolveReturnTo(searchParams.get("returnTo"))
  const { signUp, isLoaded } = useSignUp()
  const { setActive } = useClerk()

  const [step, setStep] = useState(STEP_DETAILS)
  const [email, setEmail] = useState("")
  const [error, setError] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleGoogle() {
    if (!isLoaded) return
    setError(null)
    try {
      await signUp.authenticateWithRedirect({
        strategy: "oauth_google",
        redirectUrl: "/sso-callback",
        redirectUrlComplete: redirectTo,
      })
    } catch (err) {
      if (import.meta.env.DEV) console.error("google sign-up failed", err)
      setError(clerkErrorMessage(err?.errors, "Google sign-up failed."))
    }
  }

  async function activateSession() {
    await setActive({ session: signUp.session?.id })
    identify(email)
    capture("signup_completed")
    navigate(redirectTo)
  }

  async function submitDetails(e) {
    e.preventDefault()
    const value = e.currentTarget.email.value.trim()
    const password = e.currentTarget.password.value
    if (!value) {
      setError("Please enter your email address.")
      return
    }
    if (!password) {
      setError("Please choose a password.")
      return
    }
    setError(null)
    setIsSubmitting(true)
    try {
      const attempt = await signUp.create({ emailAddress: value, password })
      setEmail(value)
      if (attempt.status === "needs_email_verification") {
        setStep(STEP_CODE)
        return
      }
      await activateSession()
    } catch (err) {
      if (import.meta.env.DEV) console.error("signup failed", err)
      setError(clerkErrorMessage(err?.errors, "We could not create your account."))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function submitCode(e) {
    e.preventDefault()
    const code = e.currentTarget.code.value.trim()
    if (!code) {
      setError("Enter the code from your email.")
      return
    }
    setError(null)
    setIsSubmitting(true)
    try {
      const attempt = await signUp.attemptEmailAddressVerification({ code })
      if (attempt.status === "complete") {
        await activateSession()
        return
      }
      setError(clerkErrorMessage(attempt.error, "That code was not accepted."))
    } catch (err) {
      if (import.meta.env.DEV) console.error("signup verification step failed", err)
      setError(clerkErrorMessage(err?.errors, "That code was not accepted."))
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
              {step === STEP_CODE ? "Check your email" : "Create your account"}
            </h1>
            <p className="mt-1 text-xs text-[#737373]">
              {step === STEP_CODE ? `We sent a code to ${email}` : "A simple session for your research workspace."}
            </p>
          </div>

          {step === STEP_DETAILS && (
            <form className="space-y-4" onSubmit={submitDetails}>
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
                  placeholder="At least 8 characters"
                  autoComplete="new-password"
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
                {isSubmitting ? "Creating account..." : "Create account"}
              </Button>
            </form>
          )}

          {step === STEP_CODE && (
            <form className="space-y-4" onSubmit={submitCode}>
              <div>
                <label htmlFor="code" className="block text-xs font-medium text-[#737373] mb-1.5">
                  Verification code
                </label>
                <input
                  id="code"
                  name="code"
                  type="text"
                  inputMode="numeric"
                  placeholder="123456"
                  autoComplete="one-time-code"
                  disabled={disabled}
                  className="w-full rounded-[6px] border border-[#e5e5e5] bg-white px-3 py-2 font-mono text-center text-lg text-[#171717] tracking-widest placeholder-[#a3a3a3] transition-colors focus:border-[#2563eb] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20"
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
                {isSubmitting ? "Verifying..." : "Verify code"}
              </Button>
            </form>
          )}

          {step !== STEP_CODE && (
            <p className="mt-6 text-center text-xs text-[#737373]">
              Already have an account?{" "}
              <Link
                to={redirectTo !== "/account" ? `/login?returnTo=${encodeURIComponent(redirectTo)}` : "/login"}
                className="text-[#2563eb] font-medium hover:underline"
              >
                Sign in
              </Link>
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
