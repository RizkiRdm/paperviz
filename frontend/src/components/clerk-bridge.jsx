// ClerkBridge registers Clerk's getToken with the api layer.
//
// useAuth() only works inside ClerkProvider, and lib/auth.js cannot call a
// hook, so this component mounts inside the provider and hands the getter over.
import { useAuth } from "@clerk/clerk-react"
import { registerTokenProviderOnce } from "@/lib/auth"

export function ClerkBridge() {
  const { getToken } = useAuth()
  registerTokenProviderOnce(getToken)
  return null
}