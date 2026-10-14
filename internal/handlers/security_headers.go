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
//
// Clerk needs the same treatment, and its absence fails much louder and much
// more confusingly than PostHog's: clerk.browser.js is refused, so
// useAuth().getToken never resolves, so useCurrentUser's fetch never settles,
// so `checked` stays false and every "Sign in" link silently disappears from
// the page. Nothing throws; the app just looks signed-out forever. The same
// block also stops the SPA fallback for unknown routes, which made a stray
// path render as "Reading document..." instead of the 404 page.
//
// The frontend also links woff2/woff/ttf files on cdn.fontshare.com, while
// only api.fontshare.com was allowed — those fonts were blocked too.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				// Clerk's bot protection loads a bootstrap from
				// cdn.protect.clerk.com and then probes an instance-specific
				// subdomain (<hash>-l.client.protect.clerk.com, and -p for the
				// fallback probe), so both need wildcards. Blocked, no sign-in
				// attempt reaches Clerk at all and the form shows a generic
				// error — invisible except as "login is broken for everyone".
				"script-src 'self' https://us-assets.i.posthog.com https://*.clerk.accounts.dev https://cdn.protect.clerk.com https://*.client.protect.clerk.com; "+
				"font-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com https://api.fontshare.com https://cdn.fontshare.com; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://api.fontshare.com; "+
				"img-src 'self' data: https://us-assets.i.posthog.com; "+
				"worker-src 'self' blob:; "+
				// Clerk resolves the FAPI origin from the publishable key, so a
				// literal *.clerk.accounts.dev wildcard is required here. It is
				// one vendor's domain, still narrow, and not 'unsafe-inline'.
				"connect-src 'self' https://us.i.posthog.com https://us-assets.i.posthog.com https://*.clerk.accounts.dev https://cdn.protect.clerk.com https://*.client.protect.clerk.com")
		next.ServeHTTP(w, r)
	})
}
