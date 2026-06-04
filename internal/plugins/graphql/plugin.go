package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	"github.com/fastygo/app-gocms/internal/extensions"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts"
)

type Plugin struct {
	Provider storage.StoreProvider
}

func New(provider storage.StoreProvider) Plugin {
	return Plugin{Provider: provider}
}

func (p Plugin) Manifest() extensions.Manifest {
	return extensions.Manifest{ID: "graphql", Name: "GraphQL", Version: "0.1.0", Contract: "go-codex.plugin.v0.1"}
}

func (p Plugin) Register(_ context.Context, registry *extensions.Context) error {
	registry.AddRoute(extensions.Route{Method: http.MethodGet, Pattern: "/go-graphql", Handler: p.handle})
	registry.AddRoute(extensions.Route{Method: http.MethodPost, Pattern: "/go-graphql", Handler: p.handle})
	return nil
}

func (p Plugin) handle(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	request.Query = r.URL.Query().Get("query")
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&request)
	}
	query := strings.ToLower(request.Query)
	if strings.Contains(query, "mutation") {
		write(w, http.StatusForbidden, map[string]any{"errors": []map[string]any{{"message": "GraphQL mutations are not enabled in this profile"}}})
		return
	}
	status := http.StatusOK
	payload := map[string]any{"data": map[string]any{}}
	err := p.Provider.ForWorkspace("root").WithinTx(r.Context(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		contentTypes := appcontenttype.NewService(appRepos)
		if err := contentTypes.InstallBuiltIns(ctx); err != nil {
			return err
		}
		data := payload["data"].(map[string]any)
		switch {
		case strings.Contains(query, "settings"):
			settings, err := repos.Settings.List(ctx)
			if err != nil {
				return err
			}
			data["settings"] = settings.Records
		case strings.Contains(query, "menus"):
			menus, err := repos.Menus.List(ctx)
			if err != nil {
				return err
			}
			data["menus"] = menus.Records
		case strings.Contains(query, "pages"):
			page, err := repos.Pages.List(ctx)
			if err != nil {
				return err
			}
			data["pages"] = publicRecords(page.Records)
		default:
			page, err := repos.Posts.List(ctx)
			if err != nil {
				return err
			}
			data["posts"] = publicRecords(page.Records)
		}
		return nil
	})
	if err != nil {
		status = http.StatusInternalServerError
		payload = map[string]any{"errors": []map[string]any{{"message": err.Error()}}}
	}
	write(w, status, payload)
}

func publicRecords(items []contracts.Record) []contracts.Record {
	filtered := []contracts.Record{}
	for _, item := range items {
		if item["status"] == "draft" || item["visibility"] == "private" {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func write(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
