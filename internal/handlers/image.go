package handlers

// detectImageMIME sniffs a binary blob's magic bytes and returns the
// corresponding image Content-Type. Empty string means the blob is not a
// recognized image format.
func detectImageMIME(b []byte) string {
	switch {
	case len(b) >= 8 && b[0] == 0x89 && string(b[1:4]) == "PNG":
		return "image/png"
	case len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff:
		return "image/jpeg"
	case len(b) >= 6 && (string(b[:6]) == "GIF87a" || string(b[:6]) == "GIF89a"):
		return "image/gif"
	case len(b) >= 12 && string(b[:4]) == "RIFF" && string(b[8:12]) == "WEBP":
		return "image/webp"
	default:
		return ""
	}
}
