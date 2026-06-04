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
		return repos.Content.Put(ctx, "post-1", contracts.Record{"title": "Root post"})
	})
	if err != nil {
		t.Fatal(err)
	}
	err = provider.ForWorkspace("sales").WithinTx(context.Background(), func(ctx context.Context, repos Repositories) error {
		result, err := repos.Content.List(ctx, "content")
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
