package preview

import (
	"context"
	"fmt"
	"time"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	domainpreview "github.com/fastygo/app-gocms/internal/domain/preview"
)

type Repository interface {
	SavePreview(context.Context, domainpreview.Access) error
	GetPreview(context.Context, string) (domainpreview.Access, bool, error)
}

type EntryRepository interface {
	Get(context.Context, domaincontent.ID) (domaincontent.Entry, bool, error)
}

type Service struct {
	repo    Repository
	entries EntryRepository
	now     func() time.Time
}

func NewService(repo Repository, entries EntryRepository) Service {
	return Service{repo: repo, entries: entries, now: time.Now}
}

func (s Service) WithClock(now func() time.Time) Service {
	s.now = now
	return s
}

func (s Service) Create(ctx context.Context, entryID domaincontent.ID, ttl time.Duration) (domainpreview.Access, error) {
	if ttl <= 0 {
		return domainpreview.Access{}, fmt.Errorf("preview ttl must be positive")
	}
	if _, ok, err := s.entries.Get(ctx, entryID); err != nil || !ok {
		if err != nil {
			return domainpreview.Access{}, err
		}
		return domainpreview.Access{}, fmt.Errorf("content %q not found", entryID)
	}
	access := domainpreview.NewAccess(entryID, ttl, s.now().UTC())
	return access, s.repo.SavePreview(ctx, access)
}

func (s Service) Validate(ctx context.Context, token string) (domainpreview.Access, bool, error) {
	access, ok, err := s.repo.GetPreview(ctx, token)
	if err != nil || !ok {
		return domainpreview.Access{}, ok, err
	}
	if !access.ValidAt(s.now().UTC()) {
		return domainpreview.Access{}, false, nil
	}
	return access, true, nil
}
