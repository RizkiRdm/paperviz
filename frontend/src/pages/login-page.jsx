import { useNavigate } from "react-router-dom"
import { AuthForm } from "@/components/auth-form"
import { useAuthSubmit } from "@/hooks/use-auth-submit"

const ERROR_MESSAGES = {
  invalid_credentials: "Invalid email or password.",
  invalid_request: "Please fill in all fields.",
  invalid_email: "Please enter a valid email address.",
  password_too_short: "Password must be at least 8 characters.",
  internal_error: "Something went wrong. Please try again.",
}

export function LoginPage() {
  const navigate = useNavigate()
  const { error, isSubmitting, submit, setError } = useAuthSubmit("/api/auth/login", ERROR_MESSAGES)

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
      title="Welcome back"
      subtitle="Sign in to your account"
      submitLabel="Sign in"
      submittingLabel="Signing in..."
      footerQuestion="Don't have an account?"
      footerLinkText="Sign up"
      footerLinkTo="/signup"
      autoCompletePassword="current-password"
      error={error}
      isSubmitting={isSubmitting}
      onSubmit={handleSubmit}
    />
  )
}