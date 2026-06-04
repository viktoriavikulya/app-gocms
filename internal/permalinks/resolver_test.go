package permalinks

import (
	"testing"

	"github.com/fastygo/app-gocms/internal/domain/content"
)

func TestResolvePublicCandidates(t *testing.T) {
	settings := DefaultSettings()
	tests := []struct {
		path string
		want []Candidate
	}{
		{path: "/", want: []Candidate{{Kind: CandidateHome}}},
		{path: "/about", want: []Candidate{{Kind: CandidatePage, Slug: "about"}}},
		{path: "/posts/hello-world", want: []Candidate{{Kind: CandidatePost, Slug: "hello-world"}}},
		{path: "/loose", want: []Candidate{{Kind: CandidatePage, Slug: "loose"}}},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := Resolve(tt.path, settings)
			if len(got) != len(tt.want) {
				t.Fatalf("Resolve(%q) len=%d, want %d: %#v", tt.path, len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("Resolve(%q)[%d]=%#v, want %#v", tt.path, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestEntryPath(t *testing.T) {
	settings := DefaultSettings()
	post := content.Entry{Kind: content.KindPost, Slug: "hello-world"}
	page := content.Entry{Kind: content.KindPage, Slug: "about"}
	if got := EntryPath(post, settings); got != "/posts/hello-world" {
		t.Fatalf("post path = %q", got)
	}
	if got := EntryPath(page, settings); got != "/about" {
		t.Fatalf("page path = %q", got)
	}
}
