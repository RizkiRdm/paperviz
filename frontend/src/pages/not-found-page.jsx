// NotFoundPage — 404 catch-all for invalid routes and expired document links.
import { Link } from "react-router-dom"

export function NotFoundPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-white">
      <div className="text-center">
        <h1 className="text-4xl font-semibold text-[#0a0a0a]">404</h1>
        <p className="mt-2 text-sm text-[#737373]">Page not found</p>
        <Link
          to="/"
          className="mt-6 inline-block text-sm text-[#2563eb] hover:underline"
        >
          Back to upload
        </Link>
      </div>
    </div>
  )
}
