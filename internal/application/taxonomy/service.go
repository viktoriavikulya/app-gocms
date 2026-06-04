package taxonomy

import (
	"context"
	"fmt"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	domaintaxonomy "github.com/fastygo/app-gocms/internal/domain/taxonomy"
)

type Repository interface {
	SaveDefinition(context.Context, domaintaxonomy.Definition) error
	GetDefinition(context.Context, string) (domaintaxonomy.Definition, bool, error)
	SaveTerm(context.Context, domaintaxonomy.Term) error
	ListTerms(context.Context, string) ([]domaintaxonomy.Term, error)
}

type EntryRepository interface {
	Get(context.Context, domaincontent.ID) (domaincontent.Entry, bool, error)
	Save(context.Context, domaincontent.Entry) error
}

type Service struct {
	repo    Repository
	entries EntryRepository
}

func NewService(repo Repository, entries EntryRepository) Service {
	return Service{repo: repo, entries: entries}
}

func (s Service) Register(ctx context.Context, definition domaintaxonomy.Definition) error {
	if definition.Type == "" {
		return fmt.Errorf("taxonomy type is required")
	}
	if definition.Mode == "" {
		definition.Mode = domaintaxonomy.ModeFlat
	}
	return s.repo.SaveDefinition(ctx, definition)
}

func (s Service) CreateTerm(ctx context.Context, term domaintaxonomy.Term) error {
	if term.ID == "" || term.TaxonomyType == "" {
		return fmt.Errorf("term id and taxonomy type are required")
	}
	if _, ok, err := s.repo.GetDefinition(ctx, term.TaxonomyType); err != nil || !ok {
		if err != nil {
			return err
		}
		return fmt.Errorf("taxonomy %q is not registered", term.TaxonomyType)
	}
	return s.repo.SaveTerm(ctx, term)
}

func (s Service) ListTerms(ctx context.Context, taxonomyType string) ([]domaintaxonomy.Term, error) {
	return s.repo.ListTerms(ctx, taxonomyType)
}

func (s Service) AssignTerms(ctx context.Context, entryID domaincontent.ID, termIDs []string) (domaincontent.Entry, error) {
	entry, ok, err := s.entries.Get(ctx, entryID)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	if !ok {
		return domaincontent.Entry{}, fmt.Errorf("content %q not found", entryID)
	}
	entry.TermIDs = append([]string{}, termIDs...)
	return entry, s.entries.Save(ctx, entry)
}
