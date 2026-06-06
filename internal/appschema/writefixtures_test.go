package appschema

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fastygo/platform/pkg/conformance/bffparity"
)

func TestWriteGoldenFixtures(t *testing.T) {
	if os.Getenv("WRITE_GOLDEN_FIXTURES") == "" {
		t.Skip("set WRITE_GOLDEN_FIXTURES=1 to regenerate Platform fixtures")
	}
	registry, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join("..", "..", "..", "@Platform", "pkg", "conformance", "testdata", "bff", "v1")
	cases := []struct {
		path string
		file string
	}{
		{"/go-admin/posts", "appcms-posts-table.json"},
		{"/go-admin/posts/post-rest/edit", "appcms-post-edit-form.json"},
	}
	for _, tc := range cases {
		screen, err := registry.Screen(tc.path)
		if err != nil {
			t.Fatalf("screen %s: %v", tc.path, err)
		}
		raw, err := bffparity.MarshalStable(screen)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(root, tc.file)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", target, err)
		}
		if err := os.WriteFile(target, append(raw, '\n'), 0o644); err != nil {
			t.Fatalf("write %s: %v", target, err)
		}
	}
}
