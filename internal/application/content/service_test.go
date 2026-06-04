package content_test

import (
	"context"
	"testing"
	"time"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	"github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestContentWorkflowTransitions(t *testing.T) {
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	now := time.Date(2026, 6, 5, 1, 2, 3, 0, time.UTC)
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		service := appcontent.NewService(appRepos, appRepos).WithClock(func() time.Time { return now })
		entry, err := service.CreateDraft(ctx, content.Entry{ID: "post-1", Kind: content.KindPost, Title: map[string]string{"en": "Hello"}, Slug: "hello", AuthorID: "admin"})
		if err != nil {
			return err
		}
		if entry.Status != content.StatusDraft {
			t.Fatalf("status = %s, want draft", entry.Status)
		}
		entry, err = service.Publish(ctx, "post-1")
		if err != nil {
			return err
		}
		if entry.Status != content.StatusPublished || entry.PublishedAt.IsZero() {
			t.Fatalf("published entry = %#v", entry)
		}
		entry, err = service.Trash(ctx, "post-1")
		if err != nil {
			return err
		}
		if entry.Status != content.StatusTrashed {
			t.Fatalf("status = %s, want trashed", entry.Status)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
