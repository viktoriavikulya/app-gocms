package storage

import (
	"context"
	"testing"

	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestProviderScopesContentByWorkspace(t *testing.T) {
	provider := NewProvider(contractstest.NewMemoryStorage())
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos Repositories) error {
		return repos.Posts.Put(ctx, "post-1", contracts.Record{"title": "Root post"})
	})
	if err != nil {
		t.Fatal(err)
	}
	err = provider.ForWorkspace("sales").WithinTx(context.Background(), func(ctx context.Context, repos Repositories) error {
		result, err := repos.Posts.List(ctx)
		if err != nil {
			return err
		}
		if result.TotalItems != 0 {
			t.Fatalf("sales workspace sees %d root content records, want 0", result.TotalItems)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestProviderSeparatesCMSRecordTypes(t *testing.T) {
	provider := NewProvider(contractstest.NewMemoryStorage())
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos Repositories) error {
		if err := repos.Posts.Put(ctx, "post-1", contracts.Record{"title": "Post"}); err != nil {
			return err
		}
		return repos.Pages.Put(ctx, "page-1", contracts.Record{"title": "Page"})
	})
	if err != nil {
		t.Fatal(err)
	}
	err = provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos Repositories) error {
		posts, err := repos.Posts.List(ctx)
		if err != nil {
			return err
		}
		pages, err := repos.Pages.List(ctx)
		if err != nil {
			return err
		}
		if posts.TotalItems != 1 || pages.TotalItems != 1 {
			t.Fatalf("posts=%d pages=%d, want 1/1", posts.TotalItems, pages.TotalItems)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
