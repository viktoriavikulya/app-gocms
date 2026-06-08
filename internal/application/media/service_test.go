package media_test

import (
	"context"
	"testing"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	appmedia "github.com/fastygo/app-gocms/internal/application/media"
	"github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/media"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestAttachFeaturedMedia(t *testing.T) {
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		if _, err := appcontent.NewService(appRepos, appRepos).CreateDraft(ctx, content.Entry{ID: "post-1", Kind: content.KindPost, Title: map[string]string{"en": "Hello"}, Slug: "hello"}); err != nil {
			return err
		}
		service := appmedia.NewService(appRepos, appRepos)
		if err := service.SaveMetadata(ctx, media.Asset{ID: "media-1", Title: "Hero", MIMEType: "image/png"}); err != nil {
			return err
		}
		entry, err := service.AttachFeatured(ctx, "post-1", "media-1")
		if err != nil {
			return err
		}
		if entry.FeaturedMediaID != "media-1" {
			t.Fatalf("featured media = %q", entry.FeaturedMediaID)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
