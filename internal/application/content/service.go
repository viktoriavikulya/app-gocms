package content

import (
	"context"
	"fmt"
	"time"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
)

type Repository interface {
	Save(context.Context, domaincontent.Entry) error
	Get(context.Context, domaincontent.ID) (domaincontent.Entry, bool, error)
	List(context.Context, domaincontent.Query) ([]domaincontent.Entry, error)
}

type TypeRegistry interface {
	GetContentType(context.Context, contenttype.ID) (contenttype.Type, bool, error)
}

type Service struct {
	repo  Repository
	types TypeRegistry
	now   func() time.Time
}

func NewService(repo Repository, types TypeRegistry) Service {
	return Service{repo: repo, types: types, now: time.Now}
}

func (s Service) WithClock(now func() time.Time) Service {
	s.now = now
	return s
}

func (s Service) CreateDraft(ctx context.Context, entry domaincontent.Entry) (domaincontent.Entry, error) {
	if entry.ID == "" {
		return domaincontent.Entry{}, fmt.Errorf("content id is required")
	}
	if err := s.ensureType(ctx, entry.Kind); err != nil {
		return domaincontent.Entry{}, err
	}
	now := s.now().UTC()
	entry.Status = domaincontent.StatusDraft
	if entry.Visibility == "" {
		entry.Visibility = domaincontent.VisibilityPublic
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now
	if err := entry.Validate(); err != nil {
		return domaincontent.Entry{}, err
	}
	return entry, s.repo.Save(ctx, entry)
}

func (s Service) Update(ctx context.Context, entry domaincontent.Entry) error {
	if err := s.ensureType(ctx, entry.Kind); err != nil {
		return err
	}
	entry.UpdatedAt = s.now().UTC()
	if err := entry.Validate(); err != nil {
		return err
	}
	return s.repo.Save(ctx, entry)
}

func (s Service) Publish(ctx context.Context, id domaincontent.ID) (domaincontent.Entry, error) {
	return s.transition(ctx, id, func(entry domaincontent.Entry) domaincontent.Entry {
		entry.Status = domaincontent.StatusPublished
		entry.PublishedAt = s.now().UTC()
		entry.ScheduledFor = time.Time{}
		return entry
	})
}

func (s Service) Schedule(ctx context.Context, id domaincontent.ID, publishAt time.Time) (domaincontent.Entry, error) {
	if publishAt.IsZero() {
		return domaincontent.Entry{}, fmt.Errorf("schedule time is required")
	}
	return s.transition(ctx, id, func(entry domaincontent.Entry) domaincontent.Entry {
		entry.Status = domaincontent.StatusScheduled
		entry.ScheduledFor = publishAt
		return entry
	})
}

func (s Service) Trash(ctx context.Context, id domaincontent.ID) (domaincontent.Entry, error) {
	return s.transitionStatus(ctx, id, domaincontent.StatusTrashed)
}

func (s Service) Restore(ctx context.Context, id domaincontent.ID) (domaincontent.Entry, error) {
	return s.transitionStatus(ctx, id, domaincontent.StatusDraft)
}

func (s Service) Archive(ctx context.Context, id domaincontent.ID) (domaincontent.Entry, error) {
	return s.transitionStatus(ctx, id, domaincontent.StatusArchived)
}

func (s Service) Get(ctx context.Context, id domaincontent.ID) (domaincontent.Entry, bool, error) {
	return s.repo.Get(ctx, id)
}

func (s Service) List(ctx context.Context, query domaincontent.Query) ([]domaincontent.Entry, error) {
	return s.repo.List(ctx, query)
}

func (s Service) ListFiltered(ctx context.Context, query domaincontent.Query, publicOnly bool) ([]domaincontent.Entry, error) {
	items, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, err
	}
	if !publicOnly {
		return items, nil
	}
	return filterPublicEntries(items, s.now().UTC()), nil
}

func (s Service) GetBySlug(ctx context.Context, kind domaincontent.Kind, slug string, publicOnly bool) (domaincontent.Entry, bool, error) {
	items, err := s.ListFiltered(ctx, domaincontent.Query{Kind: kind}, publicOnly)
	if err != nil {
		return domaincontent.Entry{}, false, err
	}
	for _, entry := range items {
		if entry.Slug == slug {
			return entry, true, nil
		}
	}
	return domaincontent.Entry{}, false, nil
}

func filterPublicEntries(items []domaincontent.Entry, now time.Time) []domaincontent.Entry {
	filtered := make([]domaincontent.Entry, 0, len(items))
	for _, entry := range items {
		if entry.IsPublicAt(now) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (s Service) transitionStatus(ctx context.Context, id domaincontent.ID, status domaincontent.Status) (domaincontent.Entry, error) {
	return s.transition(ctx, id, func(entry domaincontent.Entry) domaincontent.Entry {
		entry.Status = status
		return entry
	})
}

func (s Service) transition(ctx context.Context, id domaincontent.ID, mutate func(domaincontent.Entry) domaincontent.Entry) (domaincontent.Entry, error) {
	entry, ok, err := s.repo.Get(ctx, id)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	if !ok {
		return domaincontent.Entry{}, fmt.Errorf("content %q not found", id)
	}
	entry = mutate(entry)
	entry.UpdatedAt = s.now().UTC()
	if err := entry.Validate(); err != nil {
		return domaincontent.Entry{}, err
	}
	return entry, s.repo.Save(ctx, entry)
}

func (s Service) ensureType(ctx context.Context, kind domaincontent.Kind) error {
	_, ok, err := s.types.GetContentType(ctx, contenttype.ID(kind))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("content type %q is not registered", kind)
	}
	return nil
}
