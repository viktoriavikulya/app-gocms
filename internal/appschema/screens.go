package appschema

import (
	"fmt"

	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/app-gocms/pkg/module/codex"
	"github.com/fastygo/platform/pkg/panel"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
	"github.com/fastygo/platform/pkg/modulehost"
	"github.com/fastygo/platform/pkg/profile"
	"github.com/fastygo/platform/pkg/render"
	"github.com/fastygo/platform/pkg/toolset"
)

type Registry struct {
	Assembly modulehost.WorkspaceAssembly
	ByPath   map[string]ResourceBinding
}

type ResourceBinding struct {
	Resource panel.Resource[contracts.CapabilityID]
	Record   string
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
	registry := &Registry{Assembly: assemblies[0], ByPath: map[string]ResourceBinding{}}
	for _, resource := range registry.Assembly.Context.Resources {
		registry.ByPath[resource.BasePath] = ResourceBinding{Resource: resource, Record: string(resource.ID)}
	}
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
	if binding, ok := r.ByPath[path]; ok {
		return render.ScreenModel{
			ID:         string(binding.Resource.ID) + "-table",
			Title:      binding.Resource.Label,
			View:       render.ViewTable,
			Record:     toolset.RecordTypeID(binding.Record),
			Resource:   binding.Resource.ID,
			Columns:    binding.Resource.Table.Columns,
			Fields:     binding.Resource.Form.Fields,
			Capability: capabilityFor(binding.Resource, panel.OperationList),
		}, nil
	}
	return render.ScreenModel{}, fmt.Errorf("unknown screen path %q", path)
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
