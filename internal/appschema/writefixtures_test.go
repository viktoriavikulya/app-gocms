package appschema

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/storage"
	storagesqlite "github.com/fastygo/app-gocms/internal/storage/sqlite"
	"github.com/fastygo/platform/pkg/conformance/bffparity"
	"github.com/fastygo/platform/pkg/contracts"
)

func TestWriteGoldenFixtures(t *testing.T) {
	if os.Getenv("WRITE_GOLDEN_FIXTURES") == "" {
		t.Skip("set WRITE_GOLDEN_FIXTURES=1 to regenerate Platform fixtures")
	}
	registry, principal, err := hydratedRegistryForFixtures()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join("..", "..", "..", "@Platform", "pkg", "conformance", "testdata", "bff", "v1")
	cases := []struct {
		path string
		file string
	}{
		{"/go-admin/posts", "appcms-posts-table.json"},
		{"/go-admin/posts/post-rest/edit", "appcms-post-edit-form.json"},
	}
	for _, tc := range cases {
		page, err := registry.Page(context.Background(), tc.path, principal, nil)
		if err != nil {
			t.Fatalf("page %s: %v", tc.path, err)
		}
		raw, err := bffparity.MarshalStable(page.Screen)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(root, tc.file)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", target, err)
		}
		if err := os.WriteFile(target, append(raw, '\n'), 0o644); err != nil {
			t.Fatalf("write %s: %v", target, err)
		}
	}
}

func hydratedRegistryForFixtures() (*Registry, appauthn.Principal, error) {
	ctx := context.Background()
	store, err := storagesqlite.Open("file:appcms-fixtures?mode=memory&cache=shared")
	if err != nil {
		return nil, appauthn.Principal{}, err
	}
	if err := store.Init(ctx); err != nil {
		return nil, appauthn.Principal{}, err
	}
	if err := storagesqlite.SeedMinimalSite(ctx, store, "root"); err != nil {
		return nil, appauthn.Principal{}, err
	}
	registry, err := NewRegistryWithProvider(storage.NewProvider(store))
	if err != nil {
		return nil, appauthn.Principal{}, err
	}
	authStore, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		return nil, appauthn.Principal{}, err
	}
	principal, ok := appauthn.NewService(authStore).Principal(contracts.PrincipalID("admin"))
	if !ok {
		return nil, appauthn.Principal{}, err
	}
	return registry, principal, nil
}
