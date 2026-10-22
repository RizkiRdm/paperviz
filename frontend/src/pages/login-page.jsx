import { useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { useSignIn } from "@clerk/clerk-react"
import { clerkErrorMessage } from "@/lib/auth"

export function LoginPage() {
  const navigate = useNavigate()
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
        redirectUrlComplete: "/account",
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
        navigate("/account")
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

  return (
    <div className="auth-screen dot-grid">
      <div className="auth-card-wrap">
        <Link to="/" className="brand auth-brand"><span className="brand__mark">PV</span><span>PaperViz</span></Link>
        <div className="auth-card">
          <p className="eyebrow">ACCOUNT</p>
          <h1>Sign in</h1>
          <p className="muted">Return to your research workspace.</p>
          <form className="stack" onSubmit={handleSubmit}>
            <button type="button" onClick={handleGoogle} disabled={isSubmitting || !isLoaded} className="btn btn--outline btn--full">Continue with Google</button>
            <div className="divider"><span>or</span></div>
            <label className="field"><span>Email</span><input id="email" name="email" type="email" placeholder="you@example.com" autoComplete="email" disabled={isSubmitting || !isLoaded} /></label>
            <label className="field"><span>Password</span><input id="password" name="password" type="password" placeholder="Your password" autoComplete="current-password" disabled={isSubmitting || !isLoaded} /></label>
            {error && <div className="alert alert--error">{error}</div>}
            <button type="submit" className="btn btn--dark btn--full" disabled={isSubmitting || !isLoaded}>{isSubmitting ? "Signing in..." : "Sign in"}</button>
          </form>
          <p className="auth-meta">Don't have an account? <Link to="/signup">Sign up</Link></p>
        </div>
      </div>
    </div>
  )
}
