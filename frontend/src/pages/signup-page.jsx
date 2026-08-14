import { useNavigate } from "react-router-dom"
import { AuthForm } from "@/components/auth-form"
import { useAuthSubmit } from "@/hooks/use-auth-submit"

const ERROR_MESSAGES = {
  email_taken: "An account with this email already exists.",
  invalid_request: "Please fill in all fields.",
  invalid_email: "Please enter a valid email address.",
  password_too_short: "Password must be at least 8 characters.",
  internal_error: "Something went wrong. Please try again.",
}

export function SignupPage() {
  const navigate = useNavigate()
  const { error, isSubmitting, submit, setError } = useAuthSubmit("/api/auth/signup", ERROR_MESSAGES)

  async function handleSubmit(e) {
    e.preventDefault()
    setError(null)
    const form = e.currentTarget
    const email = form.email.value.trim()
    const password = form.password.value
    if (!email || !password) {
      setError(ERROR_MESSAGES.invalid_request)
      return
    }
    const ok = await submit({ email, password })
    if (ok) navigate("/dashboard")
  }

  return (
    <AuthForm
      title="Create your account"
      subtitle="Start simplifying papers today"
      submitLabel="Create account"
      submittingLabel="Creating account..."
      footerQuestion="Already have an account?"
      footerLinkText="Sign in"
      footerLinkTo="/login"
      autoCompletePassword="new-password"
      error={error}
      isSubmitting={isSubmitting}
      onSubmit={handleSubmit}
    />
  )
}