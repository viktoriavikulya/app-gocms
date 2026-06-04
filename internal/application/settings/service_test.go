package settings_test

import (
	"context"
	"testing"

	appsettings "github.com/fastygo/app-gocms/internal/application/settings"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestPublicProjection(t *testing.T) {
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		service := appsettings.NewService(storage.NewApplicationRepositories(repos), appsettings.NewRegistry(
			settings.Definition{Key: "site.title", Group: "site", DefaultValue: "AppCMS", Public: true},
			settings.Definition{Key: "smtp.password", Group: "mail", Public: false},
		))
		if err := service.Save(ctx, settings.Value{Key: "site.title", Value: "Demo"}); err != nil {
			return err
		}
		if err := service.Save(ctx, settings.Value{Key: "smtp.password", Value: "secret"}); err != nil {
			return err
		}
		public, err := service.Public(ctx)
		if err != nil {
			return err
		}
		if len(public) != 1 || public[0].Key != "site.title" {
			t.Fatalf("public settings = %#v", public)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
