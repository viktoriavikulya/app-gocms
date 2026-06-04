package contenttype_test

import (
	"context"
	"testing"

	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestInstallBuiltIns(t *testing.T) {
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		service := appcontenttype.NewService(storage.NewApplicationRepositories(repos))
		if err := service.InstallBuiltIns(ctx); err != nil {
			return err
		}
		post, ok, err := service.Get(ctx, contenttype.Post)
		if err != nil {
			return err
		}
		if !ok || post.PermalinkPattern != "/posts/{slug}" {
			t.Fatalf("post type = %#v, ok=%v", post, ok)
		}
		types, err := service.List(ctx)
		if err != nil {
			return err
		}
		if len(types) != 2 {
			t.Fatalf("types = %d, want 2", len(types))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
