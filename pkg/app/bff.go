package app

import (
	"net/http"
	"strings"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/appschema"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/bff"
	"github.com/fastygo/platform/pkg/contracts"
)

func registerBFFRoutes(mux *http.ServeMux, registry *appschema.Registry, authBoundary authBoundary) {
	mux.HandleFunc("GET /bff/screens/{path...}", authBoundary.renderBFFScreen(registry))
	mux.HandleFunc("GET /bff/nav", authBoundary.renderBFFNav(registry))
	mux.HandleFunc("GET /bff/session", authBoundary.renderBFFSession())
}

func (a authBoundary) renderBFFScreen(registry *appschema.Registry) http.HandlerFunc {
	return a.requireBFFJSON(func(w http.ResponseWriter, r *http.Request) {
		path := bffScreenPath(r.PathValue("path"))
		principal, _ := a.principalFromRequest(r)
		page, err := registry.Page(r.Context(), path, principal)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "screen not found"})
			return
		}
		writeJSON(w, http.StatusOK, page.Screen)
	})
}

func (a authBoundary) renderBFFNav(registry *appschema.Registry) http.HandlerFunc {
	return a.requireBFFJSON(func(w http.ResponseWriter, r *http.Request) {
		principal, _ := a.principalFromRequest(r)
		page, err := registry.DashboardPage(r.Context(), principal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, page.Navigation)
	})
}

func (a authBoundary) renderBFFSession() http.HandlerFunc {
	return a.requireBFFJSON(func(w http.ResponseWriter, r *http.Request) {
		principal, _ := a.principalFromRequest(r)
		writeJSON(w, http.StatusOK, sessionModelFromPrincipal("gocms-admin", principal))
	})
}

func (a authBoundary) requireBFFJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := a.principalFromRequest(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if !principal.Has(modulecms.CapabilityAdminAccess) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		}
		runtime := contracts.RuntimeContext{ProfileID: "gocms-admin", WorkspaceID: "root", ModuleID: "cms", PrincipalID: principal.ID()}
		next(w, r.WithContext(contracts.WithRuntimeContext(r.Context(), runtime)))
	}
}

func sessionModelFromPrincipal(profileID string, principal appauthn.Principal) bff.SessionModel {
	capabilities := make([]contracts.CapabilityID, 0, len(principal.Capabilities))
	for capability := range principal.Capabilities {
		capabilities = append(capabilities, capability)
	}
	return bff.NewSessionModel(profileID, principal.ID(), capabilities)
}

func bffScreenPath(raw string) string {
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return "/go-admin"
	}
	return "/" + raw
}
