package sanitize

import (
	"regexp"
	"strings"
)

var (
	scriptTagPattern = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	onEventPattern   = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
)

// HTML applies a minimal allowlist sanitizer for rich content fields.
func HTML(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	clean := scriptTagPattern.ReplaceAllString(raw, "")
	clean = onEventPattern.ReplaceAllString(clean, "")
	return clean
}
