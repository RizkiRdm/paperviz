package handlers

import "net/http"

// SecurityHeaders sets basic security headers on every response.
// CSP allows Google Fonts (fonts.googleapis.com, fonts.gstatic.com) and
// Fontshare (api.fontshare.com) per frontend/index.html <link> tags.
//
// PostHog needs its own script and connect origins. Without them the
// analytics library loads but every event POST is refused by the browser, so
// signup_completed, api_key_added, and paper_analysed silently never arrive
// and the CSP looks healthy in the header dump.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' https://us-assets.i.posthog.com; "+
				"font-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com https://api.fontshare.com; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://api.fontshare.com; "+
				"img-src 'self' data: https://us-assets.i.posthog.com; "+
				"connect-src 'self' https://us.i.posthog.com https://us-assets.i.posthog.com")
		next.ServeHTTP(w, r)
	})
}
