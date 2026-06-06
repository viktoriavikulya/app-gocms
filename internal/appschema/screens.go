package appschema

import (
	"context"
	"fmt"
	"strings"

	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/app-gocms/pkg/module/codex"
	"github.com/fastygo/platform/pkg/bff"
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
	resolver bff.Resolver
	page     bff.PageRuntime
	Special  map[string]render.ScreenModel
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
	special := map[string]render.ScreenModel{
		"/go-admin/settings":     settingsScreen("/go-admin/settings"),
		"/go-admin/settings/new": settingsScreen("/go-admin/settings"),
		"/go-admin/menus":        menusScreen("/go-admin/menus"),
		"/go-admin/menus/new":    menusScreen("/go-admin/menus"),
	}
	resolver := bff.NewResolver(bff.Options{
		Bases:    bff.Bases{AdminBase: "/go-admin", APIBase: "/go-json"},
		Bindings: bff.BindingsFromAssembly(assemblies[0]),
		Special:  special,
		Metadata: cmsMetadata(),
	})
	registry := &Registry{
		Assembly: assemblies[0],
		Special:  special,
		resolver: resolver,
		page: bff.NewPageRuntime(
			resolver,
			bff.NewNavigator(assemblies[0].Context.Resources),
			bff.ShellModel{
				Title:       "GoCMS Admin",
				Product:     "GoCMS Admin",
				Workspace:   "Content workspace",
				AdminBase:   "/go-admin",
				APIBase:     "/go-json",
				Description: "Schema-driven Platform preview",
			},
		),
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
	return r.resolver.Screen(path)
}

func (r *Registry) Page(ctx context.Context, path string, principal bff.Principal) (bff.PageModel, error) {
	return r.page.Page(ctx, bff.PageRequest{Path: path, Principal: principal})
}

func (r *Registry) DashboardPage(ctx context.Context, principal bff.Principal) (bff.PageModel, error) {
	return r.page.Dashboard(ctx, "GoCMS Admin", principal)
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
		Metadata: cmsMetadata()(bff.ScreenContext{Path: path, Base: path, Record: "setting", Variant: bff.VariantSpecial}),
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
		Metadata: cmsMetadata()(bff.ScreenContext{Path: path, Base: path, Record: "menu", Variant: bff.VariantSpecial}),
	}
}

func (r *Registry) DashboardScreen() render.ScreenModel {
	return r.resolver.Dashboard("GoCMS Admin")
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

func cmsMetadata() bff.MetadataFunc {
	base := bff.DefaultMetadata{
		Bases:               bff.Bases{AdminBase: "/go-admin", APIBase: "/go-json"},
		CollectionForRecord: collectionForRecord,
		ResourceAPIPath: func(collection string) string {
			return "/go-json/go/v2/" + collection
		},
	}
	return func(ctx bff.ScreenContext) map[string]string {
		meta := base.Metadata(ctx)
		if ctx.Record != "post" {
			return meta
		}
		collection := collectionForRecord(ctx.Record)
		if collection != "" {
			meta["rest_form_action"] = "/go-json/go/v2/" + collection
		}
		meta["action_scope"] = bff.ActionScopeContentWrite
		switch ctx.Variant {
		case bff.VariantNewForm:
			meta["form_action"] = "/bff/actions/post.create"
		case bff.VariantEditForm:
			if id := recordIDFromEditPath(ctx.Path); id != "" {
				meta["form_action"] = "/bff/actions/post.update?id=" + id
			}
		}
		return meta
	}
}

func recordIDFromEditPath(path string) string {
	trimmed := strings.TrimSuffix(strings.TrimRight(path, "/"), "/edit")
	if idx := strings.LastIndex(trimmed, "/"); idx >= 0 {
		return trimmed[idx+1:]
	}
	return ""
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
