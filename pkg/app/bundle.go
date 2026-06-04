package app

import (
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/appbundle"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/profile"
)

func Bundle() appbundle.Bundle {
	module := modulecms.Module{}
	manifest := module.Manifest()
	return appbundle.StaticBundle{
		AppManifest: appbundle.Manifest{
			ID:          "app-gocms",
			Name:        "AppCMS",
			Version:     manifest.Version,
			Description: "Standalone GoCMS app and reusable root workspace bundle.",
			ModuleID:    manifest.ID,
			Kind:        contracts.ModuleCompiledIn,
		},
		AppModule: module,
		Profile:   DefaultProfile(),
		Mount: appbundle.AdminMount{
			PanelID:  "cms",
			BasePath: "/go-admin",
			Default:  true,
		},
	}
}

func DefaultProfile() profile.Profile {
	return profile.Profile{
		ID:         "gocms-admin",
		Title:      "GoCMS Admin",
		AdminBase:  "/go-admin",
		APIBase:    "/go-json",
		PublicBase: "/",
		Workspaces: []profile.Workspace{
			{
				ID:          "root",
				Title:       "Content Admin",
				Description: "GoCMS root admin workspace.",
				Icon:        "layout-dashboard",
				Category:    "content",
				Order:       0,
				DefaultPath: "/go-admin",
				Capability:  "workspace.root.access",
				Modules:     []contracts.ModuleID{"cms"},
				Panels:      []profile.PanelMount{{ID: "cms", BasePath: "/go-admin", Default: true}},
			},
		},
	}
}
