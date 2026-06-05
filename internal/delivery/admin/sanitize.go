package admin

import "github.com/fastygo/app-gocms/internal/sanitize"

// SanitizeHTML applies a minimal allowlist sanitizer for rich content fields.
func SanitizeHTML(raw string) string {
	return sanitize.HTML(raw)
}
