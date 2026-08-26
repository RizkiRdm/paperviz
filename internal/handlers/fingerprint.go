package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// GenerateFingerprint creates a SHA256 hash from IP + User-Agent + Accept-Language.
func GenerateFingerprint(r *http.Request) string {
	ip := clientIP(r)
	ua := r.UserAgent()
	al := r.Header.Get("Accept-Language")

	h := sha256.Sum256([]byte(ip + ua + al))
	return hex.EncodeToString(h[:])
}

// GetFingerprint returns X-Fingerprint header if present, otherwise generates one.
func GetFingerprint(r *http.Request) string {
	if fp := r.Header.Get("X-Fingerprint"); fp != "" {
		return fp
	}
	return GenerateFingerprint(r)
}
