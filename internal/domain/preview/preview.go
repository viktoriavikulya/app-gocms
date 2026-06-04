package preview

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/fastygo/app-gocms/internal/domain/content"
)

type Access struct {
	Token     string     `json:"token"`
	EntryID   content.ID `json:"entry_id"`
	ExpiresAt time.Time  `json:"expires_at"`
}

func NewAccess(entryID content.ID, ttl time.Duration, now time.Time) Access {
	return Access{Token: randomToken(), EntryID: entryID, ExpiresAt: now.Add(ttl)}
}

func (a Access) ValidAt(now time.Time) bool {
	return a.Token != "" && now.Before(a.ExpiresAt)
}

func randomToken() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
