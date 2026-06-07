package appschema

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/storage"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/bff"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/render"
)

const cmsWorkspace = "root"

// CMSHydrator fills CMS proof screens from application services.
type CMSHydrator struct {
	Provider storage.StoreProvider
}

func (h CMSHydrator) Hydrate(ctx context.Context, screen render.ScreenModel, req bff.HydrateRequest) (render.ScreenModel, error) {
	if h.Provider == nil {
		return screen, nil
	}
	switch screen.Record {
	case "post":
		switch screen.View {
		case render.ViewTable:
			return h.hydratePostTable(ctx, screen, req)
		case render.ViewForm:
			return h.hydratePostForm(ctx, screen, req)
		}
	}
	return screen, nil
}

func (h CMSHydrator) hydratePostTable(ctx context.Context, screen render.ScreenModel, req bff.HydrateRequest) (render.ScreenModel, error) {
	var entries []domaincontent.Entry
	err := h.Provider.ForWorkspace(cmsWorkspace).WithinTx(ctx, func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		items, err := appcontent.NewService(appRepos, appRepos).List(ctx, domaincontent.Query{Kind: domaincontent.KindPost})
		if err != nil {
			return err
		}
		entries = items
		return nil
	})
	if err != nil {
		return render.ScreenModel{}, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return string(entries[i].ID) < string(entries[j].ID)
	})
	paged := bff.PaginateSlice(entries, req.Query)
	screen.Pagination = bff.PaginationFromQuery(req.Query, len(entries))
	screen.Rows = make([]render.Row, 0, len(paged))
	for _, entry := range paged {
		if entry.Kind == "" {
			entry.Kind = domaincontent.KindPost
		}
		screen.Rows = append(screen.Rows, postRow(entry, req))
	}
	if len(entries) == 0 {
		empty := render.TableEmptyState(screen)
		screen.EmptyState = &empty
	}
	return screen, nil
}

func (h CMSHydrator) hydratePostForm(ctx context.Context, screen render.ScreenModel, req bff.HydrateRequest) (render.ScreenModel, error) {
	id := recordIDFromEditPath(req.Path)
	if id == "" {
		return screen, nil
	}
	var entry domaincontent.Entry
	var ok bool
	err := h.Provider.ForWorkspace(cmsWorkspace).WithinTx(ctx, func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		var getErr error
		entry, ok, getErr = appcontent.NewService(appRepos, appRepos).Get(ctx, domaincontent.ID(id))
		return getErr
	})
	if err != nil {
		return render.ScreenModel{}, err
	}
	if !ok {
		return render.ScreenModel{}, fmt.Errorf("content %q not found", id)
	}
	if entry.Kind == "" {
		entry.Kind = domaincontent.KindPost
	}
	if entry.Kind != domaincontent.KindPost {
		return render.ScreenModel{}, fmt.Errorf("content %q not found", id)
	}
	screen.FormRecord = &render.FormRecord{Values: postFormValues(entry)}
	return screen, nil
}

func postRow(entry domaincontent.Entry, req bff.HydrateRequest) render.Row {
	row := render.Row{
		ID:    string(entry.ID),
		Cells: postCells(entry),
	}
	if req.Principal != nil && req.Principal.Has(modulecms.CapabilityContentWrite) {
		row.Capabilities = []contracts.CapabilityID{modulecms.CapabilityContentWrite}
		row.Actions = []render.RowAction{
			{
				ID:         "edit",
				Label:      "Edit",
				Method:     "GET",
				Path:       "/go-admin/posts/" + string(entry.ID) + "/edit",
				Capability: modulecms.CapabilityContentWrite,
			},
			{
				ID:          "trash",
				Label:       "Trash",
				Method:      "POST",
				Path:        "/bff/actions/post.trash?id=" + string(entry.ID),
				Scope:       cmsContentActionScope,
				Capability:  modulecms.CapabilityContentWrite,
				Destructive: true,
			},
		}
	}
	return row
}

func postCells(entry domaincontent.Entry) map[string]string {
	return map[string]string{
		"id":                string(entry.ID),
		"title":             localizedJSON(entry.Title),
		"slug":              entry.Slug,
		"content":           entry.Content,
		"excerpt":           entry.Excerpt,
		"status":            string(entry.Status),
		"visibility":        string(entry.Visibility),
		"author_id":         entry.AuthorID,
		"featured_media_id": entry.FeaturedMediaID,
		"metadata":          metadataJSON(entry.Metadata),
		"permalink":         permalinkFor(entry),
		"published_at":      formatTime(entry.PublishedAt),
		"scheduled_for":     formatTime(entry.ScheduledFor),
		"created_at":        formatTime(entry.CreatedAt),
		"updated_at":        formatTime(entry.UpdatedAt),
	}
}

func postFormValues(entry domaincontent.Entry) map[string]string {
	return map[string]string{
		"id":                string(entry.ID),
		"title":             localizedJSON(entry.Title),
		"slug":              localizedSlugJSON(entry),
		"content":           entry.Content,
		"excerpt":           entry.Excerpt,
		"status":            string(entry.Status),
		"visibility":        string(entry.Visibility),
		"author_id":         entry.AuthorID,
		"featured_media_id": entry.FeaturedMediaID,
		"metadata":          metadataJSON(entry.Metadata),
		"permalink":         permalinkFor(entry),
		"published_at":      formatTime(entry.PublishedAt),
		"scheduled_for":     formatTime(entry.ScheduledFor),
		"created_at":        formatTime(entry.CreatedAt),
		"updated_at":        formatTime(entry.UpdatedAt),
	}
}

func localizedJSON(values map[string]string) string {
	if len(values) == 0 {
		return "{}"
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func localizedSlugJSON(entry domaincontent.Entry) string {
	if entry.Slug == "" {
		return "{}"
	}
	return localizedJSON(map[string]string{"en": entry.Slug})
}

func metadataJSON(values map[string]any) string {
	if len(values) == 0 {
		return "{}"
	}
	payload, err := json.Marshal(values)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func permalinkFor(entry domaincontent.Entry) string {
	if entry.Slug == "" {
		return ""
	}
	switch entry.Kind {
	case domaincontent.KindPost:
		return "/posts/" + entry.Slug
	case domaincontent.KindPage:
		return "/" + entry.Slug
	default:
		return ""
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func QueryFromValues(values map[string][]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	query := make(map[string]string, len(values))
	for key, items := range values {
		if len(items) > 0 {
			query[key] = items[0]
		}
	}
	return query
}

func pageRequest(path string, principal bff.Principal, query map[string]string) bff.PageRequest {
	return bff.PageRequest{
		Path:      path,
		Principal: principal,
		Hydrate: bff.HydrateRequest{
			Path:      path,
			Query:     query,
			Principal: principal,
		},
	}
}

// ContentNotFound reports hydration misses for missing records.
func ContentNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
