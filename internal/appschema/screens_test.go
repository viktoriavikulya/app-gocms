package appschema

import (
	"testing"

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
