package revisions

import (
	"time"

	"github.com/fastygo/app-gocms/internal/domain/content"
)

type Revision struct {
	ID        string        `json:"id"`
	EntryID   content.ID    `json:"entry_id"`
	Entry     content.Entry `json:"entry"`
	CreatedAt time.Time     `json:"created_at"`
}
