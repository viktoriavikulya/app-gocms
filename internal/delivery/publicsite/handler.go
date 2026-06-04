package publicsite

import (
	"context"
	"net/http"
	"strings"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	appmenus "github.com/fastygo/app-gocms/internal/application/menus"
	appsettings "github.com/fastygo/app-gocms/internal/application/settings"
	"github.com/fastygo/app-gocms/internal/publicrender"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/app-gocms/internal/themes"
	"github.com/fastygo/platform/pkg/contracts"
)

type Handler struct {
	Provider  storage.StoreProvider
	Workspace contracts.WorkspaceID
	Themes    themes.Registry
	Headless  bool
}

func NewHandler(provider storage.StoreProvider, workspace contracts.WorkspaceID, registry themes.Registry) Handler {
	if workspace == "" {
		workspace = "root"
	}
	return Handler{Provider: provider, Workspace: workspace, Themes: registry}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isSystemPath(r.URL.Path) {
		http.NotFound(w, r)
		return
	}
	if h.Headless {
		http.NotFound(w, r)
		return
	}
	var page publicrender.Page
	err := h.Provider.ForWorkspace(h.Workspace).WithinTx(r.Context(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		contentTypes := appcontenttype.NewService(appRepos)
		if err := contentTypes.InstallBuiltIns(ctx); err != nil {
			return err
		}
		services := services{
			content:      appcontent.NewService(appRepos, appRepos),
			contentTypes: contentTypes,
			settings:     appsettings.NewService(appRepos, defaultSettingsRegistry()),
			menus:        appmenus.NewService(appRepos),
		}
		assembled, err := newAssembler(services).Assemble(ctx, r.URL.Path)
		page = assembled
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	theme, ok := h.Themes.Resolve(page.ThemeID)
	if !ok {
		http.Error(w, "theme not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(page.StatusCode)
	if err := theme.Render(r.Context(), page).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func isSystemPath(path string) bool {
	for _, prefix := range []string{"/go-admin", "/go-json", "/go-login", "/go-logout", "/static", "/healthz", "/readyz"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
