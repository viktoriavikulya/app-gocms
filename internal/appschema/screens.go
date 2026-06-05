package appschema

import (
	"fmt"
	"strings"

	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/app-gocms/pkg/module/codex"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
	"github.com/fastygo/platform/pkg/modulehost"
	"github.com/fastygo/platform/pkg/panel"
	"github.com/fastygo/platform/pkg/profile"
	"github.com/fastygo/platform/pkg/render"
	"github.com/fastygo/platform/pkg/toolset"
)

type Registry struct {
	Assembly modulehost.WorkspaceAssembly
	ByPath   map[string]ResourceBinding
	Special  map[string]render.ScreenModel
}

type ResourceBinding struct {
	Resource panel.Resource[contracts.CapabilityID]
	Record   toolset.RecordTypeDefinition
}

func NewRegistry() (*Registry, error) {
	host, err := modulehost.New(modulecms.Module{})
	if err != nil {
		return nil, err
	}
	assemblies, err := host.Assemble(profile.Profile{
		ID:         "gocms-admin",
		Title:      "GoCMS Admin",
		AdminBase:  "/go-admin",
		APIBase:    "/go-json",
		PublicBase: "/",
		Workspaces: []profile.Workspace{
			{ID: "root", Title: "Root Admin", Modules: []contracts.ModuleID{"cms"}, DefaultPath: "/go-admin"},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(assemblies) == 0 {
		return nil, fmt.Errorf("profile assembled without workspaces")
	}
	registry := &Registry{Assembly: assemblies[0], ByPath: map[string]ResourceBinding{}, Special: map[string]render.ScreenModel{}}
	records := recordsByID(registry.Assembly.Context.Records)
	for _, resource := range registry.Assembly.Context.Resources {
		record, ok := records[toolset.RecordTypeID(resource.ID)]
		if !ok {
			record = toolset.RecordTypeDefinition{ID: toolset.RecordTypeID(resource.ID), Label: resource.Label}
		}
		registry.ByPath[resource.BasePath] = ResourceBinding{Resource: resource, Record: record}
	}
	registry.Special["/go-admin/settings"] = settingsScreen("/go-admin/settings")
	registry.Special["/go-admin/settings/new"] = settingsScreen("/go-admin/settings")
	registry.Special["/go-admin/menus"] = menusScreen("/go-admin/menus")
	registry.Special["/go-admin/menus/new"] = menusScreen("/go-admin/menus")
	return registry, nil
}

func MustRegistry() *Registry {
	registry, err := NewRegistry()
	if err != nil {
		panic(err)
	}
	return registry
}

func (r *Registry) Screen(path string) (render.ScreenModel, error) {
	trimmed := strings.TrimRight(path, "/")
	if screen, ok := r.Special[trimmed]; ok {
		return screen, nil
	}
	if taxonomyType, isNew, ok := parseTaxonomyTermsPath(trimmed); ok {
		return r.taxonomyTermsScreen(taxonomyType, isNew)
	}
	if strings.HasSuffix(trimmed, "/new") {
		base := strings.TrimSuffix(trimmed, "/new")
		if binding, ok := r.ByPath[base]; ok {
			screen := render.ResourceFormScreen(binding.Resource, binding.Record)
			screen.Metadata = screenMetadata(base, binding.Record.ID)
			return screen, nil
		}
	}
	if strings.HasSuffix(trimmed, "/edit") {
		base := strings.TrimSuffix(trimmed, "/edit")
		if idx := strings.LastIndex(base, "/"); idx >= 0 {
			base = base[:idx]
		}
		if binding, ok := r.ByPath[base]; ok {
			screen := render.ResourceFormScreen(binding.Resource, binding.Record)
			screen.ID = string(binding.Resource.ID) + "-edit"
			screen.Title = "Edit " + binding.Resource.Singular
			screen.Metadata = screenMetadata(base, binding.Record.ID)
			return screen, nil
		}
	}
	if binding, ok := r.ByPath[path]; ok {
		screen := render.ResourceTableScreen(binding.Resource, binding.Record)
		screen.Metadata = screenMetadata(path, binding.Record.ID)
		return screen, nil
	}
	return render.ScreenModel{}, fmt.Errorf("unknown screen path %q", path)
}

func settingsScreen(path string) render.ScreenModel {
	return render.ScreenModel{
		ID:       "settings-form",
		Title:    "Settings",
		View:     render.ViewForm,
		Resource: "settings",
		Record:   "setting",
		Fields: []panel.Field{
			{ID: "site.title", Label: "Site title", Type: panel.FieldText, Required: true},
			{ID: "site.description", Label: "Site description", Type: panel.FieldTextarea},
		},
		Metadata: screenMetadata(path, "setting"),
	}
}

func menusScreen(path string) render.ScreenModel {
	return render.ScreenModel{
		ID:       "menus-form",
		Title:    "Menus",
		View:     render.ViewForm,
		Resource: "menus",
		Record:   "menu",
		Fields: []panel.Field{
			{ID: "id", Label: "Menu ID", Type: panel.FieldText, Required: true},
			{ID: "location", Label: "Location", Type: panel.FieldText, Required: true},
			{ID: "items", Label: "Items JSON", Type: panel.FieldJSON},
		},
		Metadata: screenMetadata(path, "menu"),
	}
}

func (r *Registry) DashboardScreen() render.ScreenModel {
	return render.ScreenModel{
		ID:       "admin-dashboard",
		Title:    "GoCMS Admin",
		View:     render.ViewType("dashboard"),
		Resource: "dashboard",
		Metadata: map[string]string{
			"admin_base": "/go-admin",
			"api_base":   "/go-json",
		},
	}
}

func RootDiscovery() codex.Discovery {
	return codex.RootDiscovery()
}

func V2Discovery() codex.Discovery {
	return codex.V2Discovery()
}

func EmptyList[T any]() codex.ListEnvelope[T] {
	return codex.EmptyList[T]()
}

func NotFound(message string) codex.ErrorEnvelope {
	return codex.NotFound(message)
}

func Context(registry *Registry) *contractstest.Context {
	if registry == nil {
		return nil
	}
	return registry.Assembly.Context
}

func capabilityFor(resource panel.Resource[contracts.CapabilityID], operation panel.ResourceOperation) contracts.CapabilityID {
	for _, capability := range resource.Capabilities {
		if capability.Operation == operation {
			return capability.Capability
		}
	}
	return ""
}

func recordsByID(records []toolset.RecordTypeDefinition) map[toolset.RecordTypeID]toolset.RecordTypeDefinition {
	result := map[toolset.RecordTypeID]toolset.RecordTypeDefinition{}
	for _, record := range records {
		result[record.ID] = record
	}
	return result
}

func screenMetadata(path string, record toolset.RecordTypeID) map[string]string {
	return map[string]string{
		"admin_base":  "/go-admin",
		"api_base":    "/go-json",
		"new_path":    strings.TrimRight(path, "/") + "/new",
		"list_path":   path,
		"list_api":    "/go-json/go/v2/" + collectionForRecord(record),
		"form_action": "/go-admin/" + adminCollectionForRecord(record),
	}
}

func parseTaxonomyTermsPath(path string) (taxonomyType string, isNew bool, ok bool) {
	const prefix = "/go-admin/taxonomies/"
	if !strings.HasPrefix(path, prefix) {
		return "", false, false
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) < 2 || parts[1] != "terms" {
		return "", false, false
	}
	if parts[0] == "" {
		return "", false, false
	}
	switch len(parts) {
	case 2:
		return parts[0], false, true
	case 3:
		if parts[2] == "new" {
			return parts[0], true, true
		}
	}
	return "", false, false
}

func (r *Registry) taxonomyTermsScreen(taxonomyType string, isNew bool) (render.ScreenModel, error) {
	binding, ok := r.ByPath["/go-admin/terms"]
	if !ok {
		return render.ScreenModel{}, fmt.Errorf("terms resource not registered")
	}
	basePath := "/go-admin/taxonomies/" + taxonomyType + "/terms"
	var screen render.ScreenModel
	if isNew {
		screen = render.ResourceFormScreen(binding.Resource, binding.Record)
	} else {
		screen = render.ResourceTableScreen(binding.Resource, binding.Record)
	}
	screen.Metadata = screenMetadata(basePath, binding.Record.ID)
	screen.Metadata["form_action"] = basePath
	screen.Metadata["taxonomy_type"] = taxonomyType
	return screen, nil
}

func adminCollectionForRecord(id toolset.RecordTypeID) string {
	switch id {
	case "post":
		return "posts"
	case "page":
		return "pages"
	case "media_asset":
		return "media"
	case "taxonomy":
		return "taxonomies"
	case "term":
		return "terms"
	case "content_type":
		return "content-types"
	case "author":
		return "authors"
	case "setting":
		return "settings"
	case "menu":
		return "menus"
	default:
		return string(id) + "s"
	}
}

func collectionForRecord(id toolset.RecordTypeID) string {
	switch id {
	case "post":
		return "posts"
	case "page":
		return "pages"
	case "media_asset":
		return "media"
	case "taxonomy":
		return "taxonomies"
	case "content_type":
		return "content-types"
	case "author":
		return "authors"
	case "setting":
		return "settings"
	case "menu":
		return "menus"
	default:
		return string(id) + "s"
	}
}
