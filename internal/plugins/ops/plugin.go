package ops

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fastygo/app-gocms/internal/extensions"
	"github.com/fastygo/app-gocms/internal/operations"
	"github.com/fastygo/app-gocms/internal/storage"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
)

type Plugin struct {
	Provider storage.StoreProvider
	Store    *operations.Store
	Runtime  func() string
}

func New(provider storage.StoreProvider, store *operations.Store, runtime func() string) Plugin {
	return Plugin{Provider: provider, Store: store, Runtime: runtime}
}

func (p Plugin) Manifest() extensions.Manifest {
	return extensions.Manifest{ID: "operations", Name: "Operations", Version: "0.1.0", Contract: "go-codex.plugin.v0.1"}
}

func (p Plugin) Register(_ context.Context, registry *extensions.Context) error {
	registry.AddRoute(extensions.Route{Method: http.MethodGet, Pattern: "/go-json/go/v2/ops/health", Capability: modulecms.CapabilityAdminAccess, Handler: p.health})
	registry.AddRoute(extensions.Route{Method: http.MethodGet, Pattern: "/go-json/go/v2/ops/audit", Capability: modulecms.CapabilityAdminAccess, Handler: p.audit})
	registry.AddRoute(extensions.Route{Method: http.MethodGet, Pattern: "/go-json/go/v2/ops/errors", Capability: modulecms.CapabilityAdminAccess, Handler: p.errors})
	registry.AddRoute(extensions.Route{Method: http.MethodGet, Pattern: "/go-json/go/v2/ops/snapshot", Capability: modulecms.CapabilitySettingsManage, Handler: p.exportSnapshot})
	registry.AddRoute(extensions.Route{Method: http.MethodPost, Pattern: "/go-json/go/v2/ops/snapshot", Capability: modulecms.CapabilitySettingsManage, Handler: p.importSnapshot})
	registry.AddRoute(extensions.Route{Method: http.MethodGet, Pattern: "/go-admin/import-export", Capability: modulecms.CapabilitySettingsManage, Handler: p.importExportScreen})
	return nil
}

func (p Plugin) health(w http.ResponseWriter, r *http.Request) {
	write(w, http.StatusOK, map[string]any{"checks": operations.Health(r.Context(), p.Provider, p.Store, p.runtimeState())})
}

func (p Plugin) audit(w http.ResponseWriter, _ *http.Request) {
	write(w, http.StatusOK, map[string]any{"events": p.Store.Audit()})
}

func (p Plugin) errors(w http.ResponseWriter, _ *http.Request) {
	write(w, http.StatusOK, map[string]any{"errors": p.Store.Errors()})
}

func (p Plugin) exportSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := operations.ExportSnapshot(r.Context(), p.Provider)
	if err != nil {
		p.Store.RecordError(operations.ErrorRecord{Source: "snapshot.export", Message: err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p.Store.RecordAudit(operations.AuditEvent{Action: "snapshot.export", Actor: "admin", Resource: "snapshot"})
	write(w, http.StatusOK, snapshot)
}

func (p Plugin) importSnapshot(w http.ResponseWriter, r *http.Request) {
	var snapshot operations.Snapshot
	if err := json.NewDecoder(r.Body).Decode(&snapshot); err != nil {
		http.Error(w, "invalid snapshot", http.StatusBadRequest)
		return
	}
	if snapshot.Version != operations.SnapshotVersion {
		http.Error(w, "unsupported snapshot version", http.StatusBadRequest)
		return
	}
	if err := operations.ImportSnapshot(r.Context(), p.Provider, snapshot); err != nil {
		p.Store.RecordError(operations.ErrorRecord{Source: "snapshot.import", Message: err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p.Store.RecordAudit(operations.AuditEvent{Action: "snapshot.import", Actor: "admin", Resource: "snapshot"})
	write(w, http.StatusOK, map[string]any{"status": "imported"})
}

func (p Plugin) importExportScreen(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><body><main><h1>Import / Export</h1><p>Export and restore gocms.snapshot.v1 bundles.</p><a href="/go-json/go/v2/ops/snapshot">Export snapshot</a></main></body></html>`))
}

func (p Plugin) runtimeState() string {
	if p.Runtime == nil {
		return "unknown"
	}
	return p.Runtime()
}

func write(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
