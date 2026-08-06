// App — top-level router for PaperViz. Replaces manual popstate/pushState
// routing with react-router-dom now that we have >2 screens (Chunk 5).
import { BrowserRouter, Routes, Route } from "react-router-dom"
import { UploadPage } from "@/pages/upload-page"
import { ResultPage } from "@/pages/result-page"
import { LoginPage } from "@/pages/login-page"
import { SignupPage } from "@/pages/signup-page"
import { DashboardPage } from "@/pages/dashboard-page"
import { NotFoundPage } from "@/pages/not-found-page"
import { ErrorBoundary } from "@/components/error-boundary"

export default function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<UploadPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/:documentId" element={<ResultPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  )
}
