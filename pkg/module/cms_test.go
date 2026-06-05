package modulecms_test

import (
	"testing"

	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/app-gocms/pkg/module/codex"
	"github.com/fastygo/app-gocms/pkg/module/migration"
	"github.com/fastygo/app-gocms/pkg/module/records"
	"github.com/fastygo/app-gocms/pkg/module/relations"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
	"github.com/fastygo/platform/pkg/toolset"
)

func TestModuleRegistersCMSDescriptors(t *testing.T) {
	ctx := contractstest.New(contracts.WorkspaceInfo{ID: "root", Title: "Root", Modules: []contracts.ModuleID{"cms"}})
	if err := (modulecms.Module{}).Register(ctx); err != nil {
		t.Fatalf("register module: %v", err)
	}
	if len(ctx.Records) != 8 {
		t.Fatalf("expected 8 CMS record types, got %d", len(ctx.Records))
	}
	if len(ctx.Relations) < 8 {
		t.Fatalf("expected CMS relations, got %d", len(ctx.Relations))
	}
	if len(ctx.Resources) != 8 {
		t.Fatalf("expected 8 CMS resources, got %d", len(ctx.Resources))
	}
	if len(ctx.RegisteredCapabilities) == 0 {
		t.Fatalf("expected capability definitions")
	}
}

func TestContentRecordsPreserveGoCMSLifecycleFields(t *testing.T) {
	for _, record := range []toolset.RecordTypeDefinition{records.Post(), records.Page()} {
		fields := fieldSet(record)
		for _, field := range []toolset.FieldID{"id", "title", "slug", "content", "status", "visibility", "author_id", "featured_media_id", "metadata", "published_at", "scheduled_for"} {
			if !fields[field] {
				t.Fatalf("%s missing field %s", record.ID, field)
			}
		}
		if !hasOption(fieldByID(record, "status"), "archived") {
			t.Fatalf("%s status must include archived for GoCMS content parity", record.ID)
		}
	}
}

func TestMetadataTaxonomyMediaAndAuthorsAreModeled(t *testing.T) {
	required := map[toolset.RecordTypeID]bool{
		records.RecordContentMeta: false,
		records.RecordTaxonomy:    false,
		records.RecordTerm:        false,
		records.RecordMediaAsset:  false,
		records.RecordAuthor:      false,
	}
	for _, record := range records.All() {
		if _, ok := required[record.ID]; ok {
			required[record.ID] = true
		}
	}
	for id, ok := range required {
		if !ok {
			t.Fatalf("missing record %s", id)
		}
	}
	if defs := records.DefaultMetaDefinitions(); len(defs) != 4 {
		t.Fatalf("expected default SEO metadata definitions, got %v", defs)
	}
}

func TestRelationsAndPanelDescriptors(t *testing.T) {
	if len(relations.All()) < 8 {
		t.Fatalf("expected taxonomy/media/author relations")
	}
	if len(modulecms.Views()) != 8 {
		t.Fatalf("expected table view for every CMS resource")
	}
	if len(modulecms.Workflows()) != 2 {
		t.Fatalf("expected post and page workflows")
	}
	if len(modulecms.RelationViews()) == 0 {
		t.Fatalf("expected relation views")
	}
}

func TestCodexDiscoveryShapes(t *testing.T) {
	root := codex.RootDiscovery()
	if root.Links.Self != "/go-json" || root.Links.Namespace != "/go-json/go/v2/" {
		t.Fatalf("root discovery links do not preserve GoCMS codex paths: %#v", root.Links)
	}
	v2 := codex.V2Discovery()
	if v2.Version != "2" || v2.Routes["posts"] == "" || v2.Routes["contentTypes"] == "" {
		t.Fatalf("v2 discovery requires map routes: %#v", v2.Routes)
	}
	if root.Version != "2" || root.Routes["go/v2"] != "/go-json/go/v2/" {
		t.Fatalf("root discovery version/routes mismatch: %#v", root)
	}
	if codex.NotFound("missing").Error.Code != "not_found" {
		t.Fatalf("expected stable not_found error code")
	}
}

func TestMigrationMappingDocumentsCurrentGoCMSAnchors(t *testing.T) {
	if len(migration.CurrentGoCMSMappings()) < 6 {
		t.Fatalf("expected migration mapping for core GoCMS domains")
	}
}

func fieldSet(record toolset.RecordTypeDefinition) map[toolset.FieldID]bool {
	result := map[toolset.FieldID]bool{}
	for _, field := range record.Fields {
		result[field.ID] = true
	}
	return result
}

func fieldByID(record toolset.RecordTypeDefinition, id toolset.FieldID) toolset.FieldDefinition {
	for _, field := range record.Fields {
		if field.ID == id {
			return field
		}
	}
	return toolset.FieldDefinition{}
}

func hasOption(field toolset.FieldDefinition, value string) bool {
	for _, option := range field.Options {
		if option.Value == value {
			return true
		}
	}
	return false
}
