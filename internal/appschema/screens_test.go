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
	if form.View != render.ViewForm || form.Metadata["form_action"] != "/go-admin/posts" || len(form.Fields) == 0 {
		t.Fatalf("unexpected form screen: %#v", form)
	}
}

func TestTaxonomyTypeTermsPathResolves(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	table, err := registry.Screen("/go-admin/taxonomies/category/terms")
	if err != nil {
		t.Fatal(err)
	}
	if table.View != render.ViewTable || table.Metadata["taxonomy_type"] != "category" {
		t.Fatalf("unexpected taxonomy terms table: %#v", table)
	}
	if table.Metadata["list_path"] != "/go-admin/taxonomies/category/terms" {
		t.Fatalf("unexpected list_path: %#v", table.Metadata)
	}
	form, err := registry.Screen("/go-admin/taxonomies/category/terms/new")
	if err != nil {
		t.Fatal(err)
	}
	if form.View != render.ViewForm || form.Metadata["taxonomy_type"] != "category" {
		t.Fatalf("unexpected taxonomy terms form: %#v", form)
	}
	if form.Metadata["form_action"] != "/go-admin/taxonomies/category/terms" {
		t.Fatalf("unexpected form_action: %#v", form.Metadata)
	}
}
