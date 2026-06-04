package content

import (
	"fmt"
	"time"
)

type ID string
type Kind string
type Status string
type Visibility string

const (
	KindPost Kind = "post"
	KindPage Kind = "page"

	StatusDraft     Status = "draft"
	StatusScheduled Status = "scheduled"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
	StatusTrashed   Status = "trashed"

	VisibilityPublic   Visibility = "public"
	VisibilityPrivate  Visibility = "private"
	VisibilityPassword Visibility = "password"
)

type Entry struct {
	ID              ID                `json:"id"`
	Kind            Kind              `json:"kind"`
	Title           map[string]string `json:"title"`
	Slug            string            `json:"slug"`
	Content         string            `json:"content"`
	Excerpt         string            `json:"excerpt"`
	Status          Status            `json:"status"`
	Visibility      Visibility        `json:"visibility"`
	AuthorID        string            `json:"author_id"`
	FeaturedMediaID string            `json:"featured_media_id"`
	TermIDs         []string          `json:"term_ids"`
	Metadata        map[string]any    `json:"metadata"`
	PublishedAt     time.Time         `json:"published_at"`
	ScheduledFor    time.Time         `json:"scheduled_for"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type Query struct {
	Kind   Kind
	Status Status
}

func (e Entry) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("content id is required")
	}
	if e.Kind != KindPost && e.Kind != KindPage {
		return fmt.Errorf("unsupported content kind %q", e.Kind)
	}
	if e.Status == "" {
		return fmt.Errorf("content status is required")
	}
	if e.Visibility == "" {
		return fmt.Errorf("content visibility is required")
	}
	return nil
}
