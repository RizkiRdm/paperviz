// auth.js — attaches the Clerk session token to backend requests.
//
// The Go backend verifies `Authorization: Bearer <token>` with Clerk's JWKS,
// so every authenticated call needs the header. Doing that per call site would
// mean twenty near-identical getToken() awaits; doing it here keeps api.js
// signatures unchanged.
//
// Clerk's useAuth() only works inside its provider, so the token getter is
// registered once by <ClerkBridge/> rather than threaded through every
// component.
let tokenProvider = null

export function registerTokenProviderOnce(fn) {
  tokenProvider = fn
}

// authFetch behaves like fetch and additionally sends the Clerk bearer token
// when one is available. Public endpoints still work with no token.
export async function authFetch(input, init = {}) {
  const headers = new Headers(init.headers)
  if (tokenProvider) {
    try {
      const token = await tokenProvider()
      if (token) headers.set("Authorization", `Bearer ${token}`)
    } catch (error) {
      // The request continues unauthenticated rather than being aborted: a
      // token refresh failure should surface as a 401 the UI already handles,
      // not as a network error with no recovery path.
      if (import.meta.env.DEV) {
        console.error("authFetch: could not read a Clerk session token", error)
      }
    }
  }
  return fetch(input, { ...init, headers })
}

// clerkErrorMessage maps a Clerk attempt's error array to something a person
// can act on.
//
// Credential failures deliberately collapse to one message: Clerk reports a
// missing account and a wrong password differently, and distinguishing them
// would turn this form into a user-enumeration oracle. Do not "improve" these
// two entries into distinct messages.
const CLERK_ERROR_MESSAGES = {
  form_param_nil: "Please fill in all fields.",
  too_many_requests: "Too many attempts. Please wait a moment and try again.",
  form_session_exists: "You are already signed in.",
  form_identifier_exists: "An account with this email already exists.",
  form_password_incorrect: "That email and password combination did not match.",
  form_identifier_not_found: "That email and password combination did not match.",
  form_code_incorrect: "That verification code is not correct.",
  form_code_expired: "That verification code has expired.",
}

export function clerkErrorMessage(errors, fallback = "Something went wrong. Please try again.") {
  const first = Array.isArray(errors) ? errors[0] : null
  if (!first?.code) return fallback
  return CLERK_ERROR_MESSAGES[first.code] ?? first.longMessage ?? first.message ?? fallback
}