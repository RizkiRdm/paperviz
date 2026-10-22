import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { useSignUp, useClerk } from "@clerk/clerk-react"
import { clerkErrorMessage } from "@/lib/auth"
import { capture, identify } from "@/lib/analytics"

const STEP_DETAILS = "details"
const STEP_CODE = "code"

export function SignupPage() {
  const navigate = useNavigate()
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
        redirectUrlComplete: "/account",
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
    navigate("/account")
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
    <div className="auth-screen dot-grid">
      <div className="auth-card-wrap">
        <Link to="/" className="brand auth-brand"><span className="brand__mark">PV</span><span>PaperViz</span></Link>
        <div className="auth-card">
          <p className="eyebrow">ACCOUNT</p>
          <h1>{step === STEP_CODE ? "Check your email" : "Create your account"}</h1>
          <p className="muted">{step === STEP_CODE ? `We sent a code to ${email}` : "A simple session for your research workspace."}</p>
          {step === STEP_DETAILS && (
            <form className="stack" onSubmit={submitDetails}>
              <button type="button" onClick={handleGoogle} disabled={disabled} className="btn btn--outline btn--full">Continue with Google</button>
              <div className="divider"><span>or</span></div>
              <label className="field"><span>Email</span><input id="email" name="email" type="email" placeholder="you@example.com" autoComplete="email" disabled={disabled} /></label>
              <label className="field"><span>Password</span><input id="password" name="password" type="password" placeholder="At least 8 characters" autoComplete="new-password" disabled={disabled} /></label>
              {error && <div className="alert alert--error">{error}</div>}
              <button type="submit" className="btn btn--dark btn--full" disabled={disabled}>{isSubmitting ? "Creating account..." : "Create account"}</button>
            </form>
          )}
          {step === STEP_CODE && (
            <form className="stack" onSubmit={submitCode}>
              <label className="field"><span>Verification code</span><input id="code" name="code" type="text" inputMode="numeric" placeholder="123456" autoComplete="one-time-code" disabled={disabled} /></label>
              {error && <div className="alert alert--error">{error}</div>}
              <button type="submit" className="btn btn--dark btn--full" disabled={disabled}>{isSubmitting ? "Verifying..." : "Verify"}</button>
            </form>
          )}
          {step !== STEP_CODE && <p className="auth-meta">Already have an account? <Link to="/login">Sign in</Link></p>}
        </div>
      </div>
    </div>
  )
}
