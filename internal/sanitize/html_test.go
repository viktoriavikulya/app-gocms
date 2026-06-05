package sanitize_test

import (
	"strings"
	"testing"

	"github.com/fastygo/app-gocms/internal/sanitize"
)

func TestHTMLStripsScriptAndEventHandlers(t *testing.T) {
	raw := `<p><strong>Hello</strong></p><script>alert(1)</script><img src=x onerror="alert(2)">`
	clean := sanitize.HTML(raw)
	if strings.Contains(clean, "<script") || strings.Contains(strings.ToLower(clean), "onerror") {
		t.Fatalf("sanitizer left unsafe markup: %q", clean)
	}
	if !strings.Contains(clean, "<p><strong>Hello</strong></p>") {
		t.Fatalf("sanitizer removed safe markup: %q", clean)
	}
}
