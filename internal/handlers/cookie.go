package handlers

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

// cookieSecure reports whether the session cookie gets the Secure attribute.
//
// It is an explicit flag rather than a check of r.TLS because the deployment
// terminates TLS at Cloudflare and forwards plain HTTP to the origin, so r.TLS
// is nil in production too. Defaulting to true fails closed: a misconfigured
// plain-HTTP host drops the cookie and login silently stops working, which is
// an obvious break. Defaulting to false would instead ship a session cookie
// over plaintext to anyone who forgot the variable, which is not.
func cookieSecure() bool {
	raw := os.Getenv("COOKIE_SECURE")
	if raw == "" {
		return true
	}
	secure, err := strconv.ParseBool(raw)
	if err != nil {
		slog.Warn("ignoring invalid COOKIE_SECURE, treating as secure", "value", raw)
		return true
	}
	return secure
}

// sessionCookie builds the session cookie. Login and logout share it so the
// clearing cookie cannot drift from the one it has to match.
func sessionCookie(token string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	}
}
