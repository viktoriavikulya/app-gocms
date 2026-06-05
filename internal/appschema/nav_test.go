package appschema

import (
	"testing"

	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/panel"
)

func TestNavItemsSortedByOrder(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	items := registry.NavItems()
	if len(items) < 10 {
		t.Fatalf("expected at least 10 nav items (8 resources + menus + settings), got %d", len(items))
	}
	for i := 1; i < len(items); i++ {
		if items[i].Order < items[i-1].Order {
			t.Fatalf("nav items out of order at %d: %#v after %#v", i, items[i], items[i-1])
		}
	}
	seen := map[string]struct{}{}
	for _, item := range items {
		seen[item.Href] = struct{}{}
	}
	for _, href := range []string{
		"/go-admin/posts",
		"/go-admin/pages",
		"/go-admin/content-types",
		"/go-admin/taxonomies",
		"/go-admin/terms",
		"/go-admin/meta",
		"/go-admin/media",
		"/go-admin/authors",
		"/go-admin/menus",
		"/go-admin/settings",
		"/go-admin/import-export",
	} {
		if _, ok := seen[href]; !ok {
			t.Fatalf("missing nav href %q in %#v", href, items)
		}
	}
}

func TestNavItemsCapabilityFromResource(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, resource := range registry.Assembly.Context.Resources {
		want := capabilityFor(resource, panel.OperationList)
		found := false
		for _, item := range registry.NavItems() {
			if item.Href != resource.BasePath {
				continue
			}
			found = true
			if item.Capability != want {
				t.Fatalf("%s capability = %q, want %q", item.Href, item.Capability, want)
			}
		}
		if !found {
			t.Fatalf("nav missing resource %q", resource.BasePath)
		}
	}
	for _, tc := range []struct {
		href string
		want contracts.CapabilityID
	}{
		{href: "/go-admin/menus", want: modulecms.CapabilitySettingsManage},
		{href: "/go-admin/settings", want: modulecms.CapabilitySettingsManage},
		{href: "/go-admin/import-export", want: modulecms.CapabilitySettingsManage},
	} {
		found := false
		for _, item := range registry.NavItems() {
			if item.Href != tc.href {
				continue
			}
			found = true
			if item.Capability != tc.want {
				t.Fatalf("%s capability = %q, want %q", tc.href, item.Capability, tc.want)
			}
		}
		if !found {
			t.Fatalf("missing special nav item %q", tc.href)
		}
	}
}
