package contenttype

type ID string

const (
	Post ID = "post"
	Page ID = "page"
)

type Supports struct {
	Title         bool `json:"title"`
	Editor        bool `json:"editor"`
	Excerpt       bool `json:"excerpt"`
	FeaturedMedia bool `json:"featured_media"`
}

type Type struct {
	ID               ID       `json:"id"`
	Label            string   `json:"label"`
	PermalinkPattern string   `json:"permalink_pattern"`
	Public           bool     `json:"public"`
	Supports         Supports `json:"supports"`
}

func BuiltInPost() Type {
	return Type{ID: Post, Label: "Post", PermalinkPattern: "/posts/{slug}", Public: true, Supports: Supports{Title: true, Editor: true, Excerpt: true, FeaturedMedia: true}}
}

func BuiltInPage() Type {
	return Type{ID: Page, Label: "Page", PermalinkPattern: "/{slug}", Public: true, Supports: Supports{Title: true, Editor: true, Excerpt: false, FeaturedMedia: true}}
}
