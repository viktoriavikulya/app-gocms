package modulecms

import (
	"github.com/fastygo/app-gocms/pkg/module/panels"
	"github.com/fastygo/app-gocms/pkg/module/records"
	"github.com/fastygo/app-gocms/pkg/module/relations"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/panel"
	"github.com/fastygo/platform/pkg/panelschema"
	"github.com/fastygo/platform/pkg/render"
)

type Module struct{}

func (Module) Manifest() contracts.ModuleManifest {
	return contracts.ModuleManifest{
		ID:          "cms",
		Name:        "GoCMS CMS Module",
		Version:     "0.1.0",
		Description: "Platform-native CMS module for posts, pages, metadata, taxonomies, media, and authors.",
		Capabilities: []contracts.CapabilityID{
			CapabilityAdminAccess,
			CapabilityContentRead,
			CapabilityContentWrite,
			CapabilityContentPrivate,
			CapabilityMediaUpload,
			CapabilityMediaEdit,
			CapabilityTaxonomyManage,
			CapabilityTaxonomyAssign,
			CapabilityUsersManage,
			CapabilitySettingsManage,
		},
		Kind: contracts.ModuleCompiledIn,
	}
}

func (Module) Register(ctx contracts.ModuleContext) error {
	if err := ctx.Capabilities().AddCapabilities(CapabilityDefinitions()...); err != nil {
		return err
	}
	if err := ctx.Toolset().AddRecordTypes(records.All()...); err != nil {
		return err
	}
	if err := ctx.Toolset().AddRelations(relations.All()...); err != nil {
		return err
	}
	resources := []panel.Resource[contracts.CapabilityID]{}
	for _, binding := range panels.Resources() {
		resources = append(resources, binding.Resource)
	}
	return ctx.Panel().AddResources(resources...)
}

func ResourceBindings() []panels.ResourceBinding {
	return panels.Resources()
}

func Views() []panelschema.ViewDescriptor {
	return panels.Views()
}

func Workflows() []panelschema.WorkflowDescriptor {
	return panels.Workflows()
}

func Actions() []panelschema.ActionDescriptor {
	return panels.Actions()
}

func RelationViews() []panelschema.RelationViewDescriptor {
	return panels.RelationViews()
}

func DashboardScreen() render.ScreenModel {
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
