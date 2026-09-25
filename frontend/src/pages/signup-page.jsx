import { Link, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { useAuthSubmit } from "@/hooks/use-auth-submit"

const ERROR_MESSAGES = {
  email_taken: "An account with this email already exists.",
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
      setError("Please fill in all fields.")
      return
    }
    const ok = await submit({ email, password })
    if (ok) navigate("/account")
  }

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
          <h1 className="font-satoshi text-2xl font-medium text-[#0a0a0a]">Create your account</h1>
          <p className="mt-1 text-sm text-[#737373]">Get started with PaperViz</p>
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
                disabled={isSubmitting}
              />
            </div>
            <div>
              <label htmlFor="password" className="block text-xs font-medium text-[#737373] mb-1.5">Password</label>
              <input
                id="password"
                name="password"
                type="password"
                className="w-full rounded-[6px] border border-[#000000] bg-white px-3 py-2 text-sm text-[#171717] placeholder:text-[#a3a3a3] focus:outline-none focus:ring-2 focus:ring-[#2563eb]/20 focus:border-[#2563eb]"
                placeholder="At least 8 characters"
                autoComplete="new-password"
                disabled={isSubmitting}
              />
            </div>

            {error && (
              <div className="rounded-[6px] bg-red-50 border border-red-200 px-3 py-2 text-xs text-red-700">
                {error}
              </div>
            )}

            <Button type="submit" disabled={isSubmitting} className="w-full">
              {isSubmitting ? "Creating account..." : "Sign up"}
            </Button>
          </form>

          <p className="mt-4 text-center text-xs text-[#737373]">
            Already have an account?{" "}
            <Link to="/login" className="text-[#2563eb] hover:underline">Sign in</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
