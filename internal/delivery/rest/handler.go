package rest

import (
	"context"
	"net/http"
	"strings"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	"github.com/fastygo/app-gocms/internal/application/wiring"
	"github.com/fastygo/app-gocms/internal/extensions"
	"github.com/fastygo/app-gocms/internal/operations"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/app-gocms/pkg/module/codex"
	"github.com/fastygo/platform/pkg/contracts"
)

type Handler struct {
	Provider  storage.StoreProvider
	Workspace contracts.WorkspaceID
	Authorize Authorizer
	Principal func(*http.Request) (appauthn.Principal, bool)
	Hooks     *extensions.HookBus
	Audit     operations.AuditRecorder
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

func (h Handler) WithDeps(hooks *extensions.HookBus, audit operations.AuditRecorder, principal func(*http.Request) (appauthn.Principal, bool)) Handler {
	h.Hooks = hooks
	h.Audit = audit
	h.Principal = principal
	return h
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
		services := wiring.Build(repos, wiring.Deps{Hooks: h.Hooks, Audit: h.Audit})
		status, payload = h.dispatch(ctx, r, path, services)
		return nil
	})
	if err != nil {
		return http.StatusInternalServerError, apiError(http.StatusInternalServerError, "internal_error", err.Error())
	}
	return status, payload
}

func (h Handler) principal(r *http.Request) (appauthn.Principal, bool) {
	if h.Principal != nil {
		return h.Principal(r)
	}
	return appauthn.Principal{}, false
}
