package rest

import (
	"context"
	"net/http"
	"strings"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	appmedia "github.com/fastygo/app-gocms/internal/application/media"
	appmenus "github.com/fastygo/app-gocms/internal/application/menus"
	appsettings "github.com/fastygo/app-gocms/internal/application/settings"
	apptaxonomy "github.com/fastygo/app-gocms/internal/application/taxonomy"
	appusers "github.com/fastygo/app-gocms/internal/application/users"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/app-gocms/pkg/module/codex"
	"github.com/fastygo/platform/pkg/contracts"
)

type Handler struct {
	Provider  storage.StoreProvider
	Workspace contracts.WorkspaceID
	Authorize Authorizer
}

type Authorizer func(*http.Request, contracts.CapabilityID) (authenticated bool, allowed bool)

func NewHandler(provider storage.StoreProvider, workspace contracts.WorkspaceID, authorizers ...Authorizer) Handler {
	if workspace == "" {
		workspace = "root"
	}
	var authorizer Authorizer
	if len(authorizers) > 0 {
		authorizer = authorizers[0]
	}
	return Handler{Provider: provider, Workspace: workspace, Authorize: authorizer}
}

func (h Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /go-json/{$}", h.root)
	mux.HandleFunc("GET /go-json/go/v2/{$}", h.v2)
	mux.HandleFunc("/go-json/go/v2/{path...}", h.resource)
}

func (h Handler) root(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, codex.RootDiscovery())
}

func (h Handler) v2(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, codex.V2Discovery())
}

func (h Handler) resource(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/go-json/go/v2/"), "/")
	status, payload := h.handle(r, path)
	writeJSON(w, status, payload)
}

func (h Handler) handle(r *http.Request, path string) (int, any) {
	var status int
	var payload any
	err := h.Provider.ForWorkspace(h.Workspace).WithinTx(r.Context(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		services := services{
			content:      appcontent.NewService(appRepos, appRepos),
			contentTypes: appcontenttype.NewService(appRepos),
			settings:     appsettings.NewService(appRepos, appsettings.NewRegistry(settings.Definition{Key: "site.title", Group: "site", DefaultValue: "AppCMS", Public: true})),
			users:        appusers.NewService(appRepos),
			taxonomy:     apptaxonomy.NewService(appRepos, appRepos),
			media:        appmedia.NewService(appRepos, appRepos),
			menus:        appmenus.NewService(appRepos),
		}
		status, payload = h.dispatch(ctx, r, path, services)
		return nil
	})
	if err != nil {
		return http.StatusInternalServerError, apiError(http.StatusInternalServerError, "internal_error", err.Error())
	}
	return status, payload
}

type services struct {
	content      appcontent.Service
	contentTypes appcontenttype.Service
	settings     appsettings.Service
	users        appusers.Service
	taxonomy     apptaxonomy.Service
	media        appmedia.Service
	menus        appmenus.Service
}
