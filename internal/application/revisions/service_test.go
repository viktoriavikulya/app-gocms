package revisions_test

import (
	"context"
	"testing"
	"time"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	apprevisions "github.com/fastygo/app-gocms/internal/application/revisions"
	"github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestCreateAndRestoreRevision(t *testing.T) {
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		contentService := appcontent.NewService(appRepos, appRepos)
		entry, err := contentService.CreateDraft(ctx, content.Entry{ID: "post-1", Kind: content.KindPost, Title: map[string]string{"en": "Original"}, Slug: "original"})
		if err != nil {
			return err
		}
		service := apprevisions.NewService(appRepos, appRepos).WithClock(func() time.Time { return time.Unix(1, 0).UTC() })
		if _, err := service.Create(ctx, "rev-1", entry.ID); err != nil {
			return err
		}
		entry.Title = map[string]string{"en": "Changed"}
		if err := contentService.Update(ctx, entry); err != nil {
			return err
		}
		restored, err := service.Restore(ctx, "rev-1")
		if err != nil {
			return err
		}
		if restored.Title["en"] != "Original" {
			t.Fatalf("restored title = %q", restored.Title["en"])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
