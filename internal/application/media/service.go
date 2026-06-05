package media

import (
	"context"
	"fmt"
	"time"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	domainmedia "github.com/fastygo/app-gocms/internal/domain/media"
	"github.com/fastygo/app-gocms/internal/extensions"
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
	hooks   *extensions.HookBus
	now     func() time.Time
}

func NewService(repo Repository, entries EntryRepository) Service {
	return Service{repo: repo, entries: entries, now: time.Now}
}

func (s Service) WithHooks(bus *extensions.HookBus) Service {
	s.hooks = bus
	return s
}

func (s Service) SaveMetadata(ctx context.Context, asset domainmedia.Asset) error {
	if asset.ID == "" {
		return fmt.Errorf("media asset id is required")
	}
	if err := s.repo.SaveAsset(ctx, asset); err != nil {
		return err
	}
	return s.dispatch(ctx, extensions.HookMediaUploadAfter, asset)
}

func (s Service) Delete(ctx context.Context, id string) error {
	asset, ok, err := s.repo.GetAsset(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("media asset %q not found", id)
	}
	if err := s.dispatch(ctx, extensions.HookMediaDeleteBefore, asset); err != nil {
		return err
	}
	// Metadata-only delete: overwrite with empty marker via save is not ideal; repo has no Delete on media adapter.
	// For now mark as trashed in metadata.
	asset.Metadata = map[string]any{"deleted": true}
	if err := s.repo.SaveAsset(ctx, asset); err != nil {
		return err
	}
	return s.dispatch(ctx, extensions.HookMediaDeleteAfter, asset)
}

func (s Service) Get(ctx context.Context, id string) (domainmedia.Asset, bool, error) {
	return s.repo.GetAsset(ctx, id)
}

func (s Service) List(ctx context.Context) ([]domainmedia.Asset, error) {
	return s.repo.ListAssets(ctx)
}

func (s Service) Update(ctx context.Context, asset domainmedia.Asset) error {
	if asset.ID == "" {
		return fmt.Errorf("media asset id is required")
	}
	existing, ok, err := s.repo.GetAsset(ctx, asset.ID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("media asset %q not found", asset.ID)
	}
	merged := mergeMedia(existing, asset)
	return s.repo.SaveAsset(ctx, merged)
}

func mergeMedia(existing, patch domainmedia.Asset) domainmedia.Asset {
	if patch.Title != "" {
		existing.Title = patch.Title
	}
	if patch.MIMEType != "" {
		existing.MIMEType = patch.MIMEType
	}
	if patch.AltText != "" {
		existing.AltText = patch.AltText
	}
	if patch.PublicURL != "" {
		existing.PublicURL = patch.PublicURL
	}
	if patch.ProviderRef != "" {
		existing.ProviderRef = patch.ProviderRef
	}
	if patch.Metadata != nil {
		existing.Metadata = patch.Metadata
	}
	if patch.Variants != nil {
		existing.Variants = patch.Variants
	}
	return existing
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

func (s Service) dispatch(ctx context.Context, hook string, entity any) error {
	if s.hooks == nil {
		return nil
	}
	return s.hooks.Dispatch(ctx, hook, extensions.HookPayload{Hook: hook, Entity: entity, OccurredAt: s.now().UTC()})
}
