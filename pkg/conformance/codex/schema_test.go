package codex_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
	"github.com/fastygo/app-gocms/internal/domain/menus"
	"github.com/fastygo/app-gocms/pkg/conformance/codex/schematest"
	"github.com/fastygo/app-gocms/pkg/module/codex"
)

var goldenFixtures = []struct {
	schema  string
	fixture string
}{
	{schema: "content-entry.schema.json", fixture: "content-entry-post.json"},
	{schema: "content-entry.schema.json", fixture: "content-entry-page.json"},
	{schema: "content-type.schema.json", fixture: "content-type-post.json"},
	{schema: "menu.schema.json", fixture: "menu-primary.json"},
	{schema: "envelope.schema.json", fixture: "list-posts.json"},
	{schema: "envelope.schema.json", fixture: "resource-post.json"},
	{schema: "error.schema.json", fixture: "error-not-found.json"},
	{schema: "seed-site.schema.json", fixture: "seed-minimal-site.json"},
}

func TestGoldenFixturesValidateAgainstCodexSchemas(t *testing.T) {
	for _, item := range goldenFixtures {
		t.Run(item.fixture, func(t *testing.T) {
			if err := schematest.ValidateFixture(item.schema, item.fixture); err != nil {
				t.Fatalf("fixture failed schema validation: %v", err)
			}
		})
	}
}

func TestDomainTypesValidateAgainstCodexSchemas(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	entrySchema := mustLoadSchema(t, "content-entry.schema.json")
	typeSchema := mustLoadSchema(t, "content-type.schema.json")
	menuSchema := mustLoadSchema(t, "menu.schema.json")
	envelopeSchema := mustLoadSchema(t, "envelope.schema.json")
	errorSchema := mustLoadSchema(t, "error.schema.json")

	mustValidateJSON(t, entrySchema, domaincontent.Entry{
		ID: "rest-post", Kind: domaincontent.KindPost,
		Title: map[string]string{"en": "REST Post"}, Slug: "rest-post", Content: "Hello REST",
		Status: domaincontent.StatusPublished, Visibility: domaincontent.VisibilityPublic,
		AuthorID: "admin", TermIDs: []string{}, Metadata: map[string]any{},
		PublishedAt: now, CreatedAt: now, UpdatedAt: now,
	})

	mustValidateJSON(t, typeSchema, contenttype.BuiltInPost())
	mustValidateJSON(t, typeSchema, contenttype.BuiltInPage())

	mustValidateJSON(t, menuSchema, menus.Menu{
		ID: "primary", Location: "primary",
		Items: []menus.Item{{ID: "about", Label: "About", URL: "/about"}},
	})

	mustValidateJSON(t, envelopeSchema, codex.RootDiscovery())
	mustValidateJSON(t, envelopeSchema, codex.V2Discovery())
	mustValidateJSON(t, envelopeSchema, codex.ListEnvelope[domaincontent.Entry]{
		Data: []domaincontent.Entry{},
		Pagination: codex.Pagination{Page: 1, PerPage: 20, Total: 0, TotalPages: 0},
	})
	mustValidateJSON(t, envelopeSchema, codex.ResourceEnvelope[domaincontent.Entry]{
		Data: domaincontent.Entry{
			ID: "about", Kind: domaincontent.KindPage,
			Title: map[string]string{"en": "About"}, Slug: "about",
			Status: domaincontent.StatusPublished, Visibility: domaincontent.VisibilityPublic,
			AuthorID: "admin", TermIDs: []string{}, Metadata: map[string]any{},
			PublishedAt: now, CreatedAt: now, UpdatedAt: now,
		},
	})
	mustValidateJSON(t, errorSchema, codex.NotFound("route not found"))
}

func TestSeedBundleMatchesSQLiteSeedInventory(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(schematest.FixtureDir(), "seed-minimal-site.json"))
	if err != nil {
		t.Fatalf("read seed fixture: %v", err)
	}
	var bundle struct {
		ContentEntries []struct {
			ID   string `json:"id"`
			Kind string `json:"kind"`
			Slug string `json:"slug"`
		} `json:"content_entries"`
		Menus []struct {
			Location string `json:"location"`
		} `json:"menus"`
		Settings []struct {
			Key string `json:"key"`
		} `json:"settings"`
	}
	if err := json.Unmarshal(raw, &bundle); err != nil {
		t.Fatalf("decode seed fixture: %v", err)
	}

	wantEntries := map[string]string{
		"hello-world": "post",
		"rest-post":   "post",
		"about":       "page",
	}
	if len(bundle.ContentEntries) != len(wantEntries) {
		t.Fatalf("expected %d content entries, got %d", len(wantEntries), len(bundle.ContentEntries))
	}
	for _, entry := range bundle.ContentEntries {
		kind, ok := wantEntries[entry.ID]
		if !ok {
			t.Fatalf("unexpected seed entry %q", entry.ID)
		}
		if entry.Kind != kind {
			t.Fatalf("entry %q kind = %q, want %q", entry.ID, entry.Kind, kind)
		}
	}

	if len(bundle.Menus) != 1 || bundle.Menus[0].Location != "primary" {
		t.Fatalf("expected primary menu seed, got %#v", bundle.Menus)
	}

	wantSettings := []string{"site.title", "site.description", "theme.active"}
	if len(bundle.Settings) != len(wantSettings) {
		t.Fatalf("expected %d settings, got %d", len(wantSettings), len(bundle.Settings))
	}
	for i, key := range wantSettings {
		if bundle.Settings[i].Key != key {
			t.Fatalf("settings[%d] = %q, want %q", i, bundle.Settings[i].Key, key)
		}
	}
}

func mustLoadSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	schema, err := schematest.LoadSchema(name)
	if err != nil {
		t.Fatalf("load schema %s: %v", name, err)
	}
	return schema
}

func mustValidateJSON(t *testing.T, schema *jsonschema.Schema, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode marshaled value: %v", err)
	}
	if err := schema.Validate(decoded); err != nil {
		t.Fatalf("value failed schema validation: %v", err)
	}
}
