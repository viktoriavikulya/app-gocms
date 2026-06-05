package revisions

import (
	"context"
	"fmt"
	"time"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	domainrevisions "github.com/fastygo/app-gocms/internal/domain/revisions"
)

type Repository interface {
	SaveRevision(context.Context, domainrevisions.Revision) error
	GetRevision(context.Context, string) (domainrevisions.Revision, bool, error)
	ListRevisionsByEntry(context.Context, domaincontent.ID) ([]domainrevisions.Revision, error)
}

type EntryRepository interface {
	Get(context.Context, domaincontent.ID) (domaincontent.Entry, bool, error)
	Save(context.Context, domaincontent.Entry) error
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

func (s Service) Create(ctx context.Context, id string, entryID domaincontent.ID) (domainrevisions.Revision, error) {
	entry, ok, err := s.entries.Get(ctx, entryID)
	if err != nil {
		return domainrevisions.Revision{}, err
	}
	if !ok {
		return domainrevisions.Revision{}, fmt.Errorf("content %q not found", entryID)
	}
	revision := domainrevisions.Revision{ID: id, EntryID: entryID, Entry: entry, CreatedAt: s.now().UTC()}
	return revision, s.repo.SaveRevision(ctx, revision)
}

func (s Service) Get(ctx context.Context, id string) (domainrevisions.Revision, bool, error) {
	return s.repo.GetRevision(ctx, id)
}

func (s Service) ListByEntry(ctx context.Context, entryID domaincontent.ID, limit int) ([]domainrevisions.Revision, error) {
	items, err := s.repo.ListRevisionsByEntry(ctx, entryID)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(items) > limit {
		return items[:limit], nil
	}
	return items, nil
}

func (s Service) Restore(ctx context.Context, id string) (domaincontent.Entry, error) {
	revision, ok, err := s.repo.GetRevision(ctx, id)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	if !ok {
		return domaincontent.Entry{}, fmt.Errorf("revision %q not found", id)
	}
	return revision.Entry, s.entries.Save(ctx, revision.Entry)
}
