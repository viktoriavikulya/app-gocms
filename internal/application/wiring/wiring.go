package wiring

import (
	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	appmedia "github.com/fastygo/app-gocms/internal/application/media"
	appmenus "github.com/fastygo/app-gocms/internal/application/menus"
	apprevisions "github.com/fastygo/app-gocms/internal/application/revisions"
	appsettings "github.com/fastygo/app-gocms/internal/application/settings"
	apptaxonomy "github.com/fastygo/app-gocms/internal/application/taxonomy"
	appusers "github.com/fastygo/app-gocms/internal/application/users"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/extensions"
	"github.com/fastygo/app-gocms/internal/operations"
	"github.com/fastygo/app-gocms/internal/storage"
)

type Deps struct {
	Hooks *extensions.HookBus
	Audit operations.AuditRecorder
}

type Services struct {
	Content      appcontent.Service
	ContentTypes appcontenttype.Service
	Settings     appsettings.Service
	Users        appusers.Service
	Taxonomy     apptaxonomy.Service
	Media        appmedia.Service
	Menus        appmenus.Service
	Revisions    apprevisions.Service
}

func Build(repos storage.Repositories, deps Deps) Services {
	appRepos := storage.NewApplicationRepositories(repos)
	content := appcontent.NewService(appRepos, appRepos)
	if deps.Hooks != nil {
		content = content.WithHooks(deps.Hooks)
	}
	revisions := apprevisions.NewService(appRepos, appRepos)
	content = content.WithRevisions(revisions)
	media := appmedia.NewService(appRepos, appRepos)
	if deps.Hooks != nil {
		media = media.WithHooks(deps.Hooks)
	}
	settingsSvc := appsettings.NewService(appRepos, appsettings.NewRegistry(
		settings.Definition{Key: "site.title", Group: "site", DefaultValue: "AppCMS", Public: true},
		settings.Definition{Key: "site.description", Group: "site", DefaultValue: "", Public: true},
	))
	if deps.Hooks != nil {
		settingsSvc = settingsSvc.WithHooks(deps.Hooks)
	}
	return Services{
		Content:      content,
		ContentTypes: appcontenttype.NewService(appRepos),
		Settings:     settingsSvc,
		Users:        appusers.NewService(appRepos),
		Taxonomy:     apptaxonomy.NewService(appRepos, appRepos),
		Media:        media,
		Menus:        appmenus.NewService(appRepos),
		Revisions:    revisions,
	}
}
