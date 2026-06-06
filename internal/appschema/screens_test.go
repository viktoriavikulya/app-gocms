package appschema

import (
	"context"
	"testing"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/render"
)

func TestScreensUsePlatformRendererModels(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	table, err := registry.Screen("/go-admin/posts")
	if err != nil {
		t.Fatal(err)
	}
	if table.View != render.ViewTable || table.Metadata["new_path"] != "/go-admin/posts/new" || table.Metadata["list_api"] != "/go-json/go/v2/posts" {
		t.Fatalf("unexpected table screen: %#v", table)
	}
	form, err := registry.Screen("/go-admin/posts/new")
	if err != nil {
		t.Fatal(err)
	}
	if form.View != render.ViewForm || form.Metadata["form_action"] != "/go-json/go/v2/posts" || len(form.Fields) == 0 {
		t.Fatalf("unexpected form screen: %#v", form)
	}
}

func TestPageUsesPlatformBFFRuntime(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	store, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	principal, ok := appauthn.NewService(store).Principal(contracts.PrincipalID("admin"))
	if !ok {
		t.Fatal("expected admin principal")
	}
	page, err := registry.Page(context.Background(), "/go-admin/posts", principal)
	if err != nil {
		t.Fatal(err)
	}
	if page.Shell.Product != "GoCMS Admin" || len(page.Navigation.Items) == 0 {
		t.Fatalf("unexpected page shell/nav: %#v", page)
	}
	if page.Screen.View != render.ViewTable || page.Screen.Title != "Posts" {
		t.Fatalf("unexpected page screen: %#v", page.Screen)
	}
	dashboard, err := registry.DashboardPage(context.Background(), principal)
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.Screen.View != render.ViewType("dashboard") {
		t.Fatalf("unexpected dashboard screen: %#v", dashboard.Screen)
	}
}
