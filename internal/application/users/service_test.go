package users_test

import (
	"context"
	"testing"

	appusers "github.com/fastygo/app-gocms/internal/application/users"
	"github.com/fastygo/app-gocms/internal/domain/users"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestPublicAuthorHidesPrivateEmail(t *testing.T) {
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		service := appusers.NewService(storage.NewApplicationRepositories(repos))
		if err := service.Save(ctx, users.User{ID: "admin", Email: "admin@example.test", DisplayName: "Admin", Slug: "admin", Active: true}); err != nil {
			return err
		}
		author, ok, err := service.PublicAuthor(ctx, "admin")
		if err != nil {
			return err
		}
		if !ok || author.ID != "admin" || author.DisplayName != "Admin" {
			t.Fatalf("author = %#v, ok=%v", author, ok)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
