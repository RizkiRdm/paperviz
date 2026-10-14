import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { useSignUp, useClerk } from "@clerk/clerk-react"
import { Button } from "@/components/ui/button"
import { clerkErrorMessage } from "@/lib/auth"
import { capture, identify } from "@/lib/analytics"

const STEP_DETAILS = "details"
const STEP_CODE = "code"

const INPUT_CLASS =
  "w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 focus:border-[#2563eb]"

const LABEL_CLASS = "block text-xs font-medium text-[#737373] mb-1.5"

export function SignupPage() {
  const navigate = useNavigate()
  const { signUp, isLoaded } = useSignUp()
  const { setActive } = useClerk()

  const [step, setStep] = useState(STEP_DETAILS)
  const [email, setEmail] = useState("")
  const [error, setError] = useState(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  // activateSession turns a finished signup into a live session, then records
  // the conversion. Clerk returns a session id that only exists once every
  // required step is done.
  async function activateSession(signUpResource) {
    await setActive({ session: signUpResource.session?.id })
    identify(email)
    capture("signup_completed")
    navigate("/account")
  }

  // signUp.create is the only creation entry point on the resource: it takes
  // emailAddress + password together and returns a status. The older
  // attemptEmailAddress / attemptPassword pair does not exist on the resource,
  // so splitting signup across those steps fails for every user.
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
      await activateSession(signUp)
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
        await activateSession(signUp)
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
    <div className="min-h-screen bg-white bg-dotted-grid flex items-center justify-center px-6">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="flex h-8 w-8 items-center justify-center rounded-[6px] bg-[#0a0a0a] text-white font-mono text-xs font-bold">
              PV
            </div>
            <span className="font-mono text-sm font-semibold tracking-tight text-[#0a0a0a]">PaperViz</span>
          </Link>
          <h1 className="font-satoshi text-2xl font-medium text-[#0a0a0a]">
            {step === STEP_CODE ? "Check your email" : "Create your account"}
          </h1>
          <p className="mt-1 text-sm text-[#737373]">
            {step === STEP_CODE
              ? `We sent a code to ${email}`
              : "Get started with PaperViz"}
          </p>
        </div>

        <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-6">
          {step === STEP_DETAILS && (
            <form onSubmit={submitDetails} className="space-y-4">
              <div>
                <label htmlFor="email" className={LABEL_CLASS}>Email</label>
                <input id="email" name="email" type="email" className={INPUT_CLASS}
                  placeholder="you@example.com" autoComplete="email" disabled={disabled} />
              </div>
              <div>
                <label htmlFor="password" className={LABEL_CLASS}>Password</label>
                <input id="password" name="password" type="password" className={INPUT_CLASS}
                  placeholder="At least 8 characters" autoComplete="new-password" disabled={disabled} />
              </div>
              {error && <ErrorNote message={error} />}
              <Button type="submit" disabled={disabled} className="w-full">
                {isSubmitting ? "Creating account..." : "Create account"}
              </Button>
            </form>
          )}

          {step === STEP_CODE && (
            <form onSubmit={submitCode} className="space-y-4">
              <div>
                <label htmlFor="code" className={LABEL_CLASS}>Verification code</label>
                <input id="code" name="code" type="text" inputMode="numeric" className={INPUT_CLASS}
                  placeholder="123456" autoComplete="one-time-code" disabled={disabled} />
              </div>
              {error && <ErrorNote message={error} />}
              <Button type="submit" disabled={disabled} className="w-full">
                {isSubmitting ? "Verifying..." : "Verify"}
              </Button>
            </form>
          )}

          {step !== STEP_CODE && (
            <p className="mt-4 text-center text-xs text-[#737373]">
              Already have an account?{" "}
              <Link to="/login" className="text-[#2563eb] hover:underline">Sign in</Link>
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

// ErrorNote keeps the inline error treatment identical across both steps.
function ErrorNote({ message }) {
  return (
    <div className="rounded-[6px] bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
      {message}
    </div>
  )
}