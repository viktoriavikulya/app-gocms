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
	if screen, ok := r.Special[strings.TrimRight(path, "/")]; ok {
		return screen, nil
	}
	trimmed := strings.TrimRight(path, "/")
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
		"form_action": "/go-json/go/v2/" + collectionForRecord(record),
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
