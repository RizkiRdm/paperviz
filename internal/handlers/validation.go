package handlers

// pdfMagicBytes is the fixed 5-byte signature ("%PDF-") every valid PDF
// file starts with, per the PDF spec. Checking this instead of trusting
// the client's Content-Type header is what AGENTS.md's Security Rules
// require: "MUST NOT trust client-supplied content-type header alone;
// verify actual file content."
var pdfMagicBytes = []byte("%PDF-")

// isPDFContent reports whether b begins with the PDF file signature. This
// is a cheap, dependency-free check — good enough to reject "renamed .txt
// file" style spoofing without needing a full MIME-sniffing library for
// one file type.
func isPDFContent(b []byte) bool {
	if len(b) < len(pdfMagicBytes) {
		return false
	}
	for i, magic := range pdfMagicBytes {
		if b[i] != magic {
			return false
		}
	}
	return true
}
