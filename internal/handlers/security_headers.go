package handlers

import "net/http"

// SecurityHeaders sets basic security headers on every response.
// CSP allows Google Fonts (fonts.googleapis.com, fonts.gstatic.com) and
// Fontshare (api.fontshare.com) per frontend/index.html <link> tags.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"font-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com https://api.fontshare.com; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://api.fontshare.com; "+
				"img-src 'self' data:; "+
				"connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}
