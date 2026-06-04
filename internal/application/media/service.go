package media

import (
	"context"
	"fmt"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	domainmedia "github.com/fastygo/app-gocms/internal/domain/media"
)

type Repository interface {
	SaveAsset(context.Context, domainmedia.Asset) error
	GetAsset(context.Context, string) (domainmedia.Asset, bool, error)
	ListAssets(context.Context) ([]domainmedia.Asset, error)
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

func (s Service) SaveMetadata(ctx context.Context, asset domainmedia.Asset) error {
	if asset.ID == "" {
		return fmt.Errorf("media asset id is required")
	}
	return s.repo.SaveAsset(ctx, asset)
}

func (s Service) Get(ctx context.Context, id string) (domainmedia.Asset, bool, error) {
	return s.repo.GetAsset(ctx, id)
}

func (s Service) List(ctx context.Context) ([]domainmedia.Asset, error) {
	return s.repo.ListAssets(ctx)
}

func (s Service) AttachFeatured(ctx context.Context, entryID domaincontent.ID, assetID string) (domaincontent.Entry, error) {
	if _, ok, err := s.repo.GetAsset(ctx, assetID); err != nil || !ok {
		if err != nil {
			return domaincontent.Entry{}, err
		}
		return domaincontent.Entry{}, fmt.Errorf("media asset %q not found", assetID)
	}
	entry, ok, err := s.entries.Get(ctx, entryID)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	if !ok {
		return domaincontent.Entry{}, fmt.Errorf("content %q not found", entryID)
	}
	entry.FeaturedMediaID = assetID
	return entry, s.entries.Save(ctx, entry)
}
