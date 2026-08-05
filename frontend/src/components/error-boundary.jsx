// Error boundary — class component required (hooks can't catch render errors).
// Catches uncaught render errors, shows ErrorBanner + reload button.
import { Component } from "react"
import { ErrorBanner } from "./ui/status-banners"
import { Button } from "./ui/button"

export class ErrorBoundary extends Component {
  constructor(props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError() {
    return { hasError: true }
  }

  componentDidCatch(error, errorInfo) {
    console.error("ErrorBoundary caught:", error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-white bg-dotted-grid flex flex-col items-center justify-center p-6">
          <div className="max-w-md w-full">
            <ErrorBanner message="Something went wrong. Please refresh the page." />
            <Button
              onClick={() => window.location.reload()}
              variant="secondary"
              className="mt-4 w-full"
            >
              Reload
            </Button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
