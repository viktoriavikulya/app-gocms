package preview_test

import (
	"context"
	"testing"
	"time"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	apppreview "github.com/fastygo/app-gocms/internal/application/preview"
	"github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestPreviewTokenTTL(t *testing.T) {
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	start := time.Date(2026, 6, 5, 1, 0, 0, 0, time.UTC)
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		if _, err := appcontent.NewService(appRepos, appRepos).CreateDraft(ctx, content.Entry{ID: "post-1", Kind: content.KindPost, Title: map[string]string{"en": "Preview"}, Slug: "preview"}); err != nil {
			return err
		}
		service := apppreview.NewService(appRepos, appRepos).WithClock(func() time.Time { return start })
		access, err := service.Create(ctx, "post-1", time.Minute)
		if err != nil {
			return err
		}
		if _, ok, err := service.Validate(ctx, access.Token); err != nil || !ok {
			t.Fatalf("valid token ok=%v err=%v", ok, err)
		}
		expiredService := service.WithClock(func() time.Time { return start.Add(2 * time.Minute) })
		if _, ok, err := expiredService.Validate(ctx, access.Token); err != nil || ok {
			t.Fatalf("expired token ok=%v err=%v", ok, err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
