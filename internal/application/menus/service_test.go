package menus_test

import (
	"context"
	"testing"

	appmenus "github.com/fastygo/app-gocms/internal/application/menus"
	"github.com/fastygo/app-gocms/internal/domain/menus"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestMenuByLocation(t *testing.T) {
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		service := appmenus.NewService(storage.NewApplicationRepositories(repos))
		if err := service.Save(ctx, menus.Menu{ID: "primary", Location: "primary", Items: []menus.Item{{ID: "home", Label: "Home", URL: "/"}}}); err != nil {
			return err
		}
		menu, ok, err := service.ByLocation(ctx, "primary")
		if err != nil {
			return err
		}
		if !ok || len(menu.Items) != 1 {
			t.Fatalf("menu = %#v, ok=%v", menu, ok)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
