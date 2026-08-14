package handlers

import "testing"

func TestDetectImageMIME(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"png", []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}, "image/png"},
		{"jpeg", []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10}, "image/jpeg"},
		{"gif87", []byte{'G', 'I', 'F', '8', '7', 'a', 0x01, 0x02}, "image/gif"},
		{"gif89", []byte{'G', 'I', 'F', '8', '9', 'a', 0x01, 0x02}, "image/gif"},
		{"webp", []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P'}, "image/webp"},
		{"empty", nil, ""},
		{"too_short", []byte{0x89}, ""},
		{"unknown", []byte("plain text blob"), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectImageMIME(tt.data); got != tt.want {
				t.Errorf("detectImageMIME(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
