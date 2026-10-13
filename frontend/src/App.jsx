// App — top-level router for PaperViz. Replaces manual popstate/pushState
// routing with react-router-dom now that we have >2 screens (Chunk 5).
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { ClerkProvider } from "@clerk/clerk-react"
import { UploadPage } from "@/pages/upload-page"
import { ResultPage } from "@/pages/result-page"
import { LoginPage } from "@/pages/login-page"
import { SignupPage } from "@/pages/signup-page"
import { AccountPage } from "@/pages/account-page"
import { ShareFigurePage } from "@/pages/share-figure-page"
import { SharePaperPage } from "@/pages/share-paper-page"
import { NotFoundPage } from "@/pages/not-found-page"
import { AgentsPage } from "@/pages/agents-page"
import { ClerkBridge } from "@/components/clerk-bridge"
import { ErrorBoundary } from "@/components/error-boundary"
import { TooltipProvider } from "@/components/ui/tooltip"

// prefetchUI={false} keeps Clerk's prebuilt UI bundle out of the initial
// payload; PaperViz renders its own forms, so it would be dead weight.
const clerkPublishableKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY

export default function App() {
  // ClerkProvider throws on an empty key, which surfaces as a blank screen.
  // Name the missing variable instead.
  if (!clerkPublishableKey) {
    return (
      <div className="flex min-h-screen items-center justify-center px-6">
        <p className="max-w-md text-center text-sm text-[#737373]">
          VITE_CLERK_PUBLISHABLE_KEY is not set. Copy <code>.env.example</code>{" "}
          to <code>.env</code> and add your Clerk publishable key.
        </p>
      </div>
    )
  }

  return (
    <ErrorBoundary>
      <ClerkProvider publishableKey={clerkPublishableKey} prefetchUI={false}>
      <ClerkBridge />
      <TooltipProvider delayDuration={200}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<UploadPage />} />
          <Route path="/upload" element={<Navigate to="/" replace />} />
          <Route path="/dashboard" element={<Navigate to="/account" replace />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/account" element={<AccountPage />} />
          <Route path="/agents" element={<AgentsPage />} />
          <Route path="/share/fig/:shareToken" element={<ShareFigurePage />} />
          <Route path="/share/doc/:shareToken" element={<SharePaperPage />} />
          <Route path="/:documentId" element={<ResultPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </BrowserRouter>
      </TooltipProvider>
      </ClerkProvider>
    </ErrorBoundary>
  )
}
