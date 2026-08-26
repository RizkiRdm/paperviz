// App — top-level router for PaperViz. Replaces manual popstate/pushState
// routing with react-router-dom now that we have >2 screens (Chunk 5).
import { BrowserRouter, Routes, Route } from "react-router-dom"
import { UploadPage } from "@/pages/upload-page"
import { ResultPage } from "@/pages/result-page"
import { LoginPage } from "@/pages/login-page"
import { SignupPage } from "@/pages/signup-page"
import { DashboardPage } from "@/pages/dashboard-page"
import { ComparePage } from "@/pages/compare-page"
import { ShareFigurePage } from "@/pages/share-figure-page"
import { SharePaperPage } from "@/pages/share-paper-page"
import { ExplainPage } from "@/pages/explain-page"
import { NotFoundPage } from "@/pages/not-found-page"
import { ErrorBoundary } from "@/components/error-boundary"
import { TooltipProvider } from "@/components/ui/tooltip"

export default function App() {
  return (
    <ErrorBoundary>
      <TooltipProvider delayDuration={200}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<UploadPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/compare" element={<ComparePage />} />
          <Route path="/research-paper-summarizer" element={<SharePaperPage />} />
          <Route path="/figure-explanation" element={<ShareFigurePage />} />
          <Route path="/compare-research-papers" element={<ComparePage />} />
          <Route path="/share/fig/:shareToken" element={<ShareFigurePage />} />
          <Route path="/share/doc/:shareToken" element={<SharePaperPage />} />
          <Route path="/explain/:slug" element={<ExplainPage />} />
          <Route path="/:documentId" element={<ResultPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </BrowserRouter>
      </TooltipProvider>
    </ErrorBoundary>
  )
}
