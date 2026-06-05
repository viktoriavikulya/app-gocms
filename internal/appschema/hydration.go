package appschema

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/application/wiring"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	domainmedia "github.com/fastygo/app-gocms/internal/domain/media"
	"github.com/fastygo/app-gocms/internal/domain/taxonomy"
	"github.com/fastygo/app-gocms/internal/delivery/rest"
	"github.com/fastygo/app-gocms/internal/extensions"
	"github.com/fastygo/app-gocms/internal/operations"
	"github.com/fastygo/app-gocms/internal/storage"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/app-gocms/pkg/module/capcheck"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/panel"
	"github.com/fastygo/platform/pkg/render"
	"github.com/fastygo/platform/pkg/toolset"
)

type Hydrator struct {
	Provider  storage.StoreProvider
	Workspace contracts.WorkspaceID
	Hooks     *extensions.HookBus
	Audit     operations.AuditRecorder
}

func (h Hydrator) Hydrate(ctx context.Context, screen render.ScreenModel, path string, query url.Values, principal appauthn.Principal) (render.ScreenModel, error) {
	if h.Provider == nil {
		return screen, nil
	}
	trimmed := strings.TrimRight(path, "/")
	if trimmed == "/go-admin/runtime" {
		return h.hydrateRuntime(screen)
	}
	if strings.HasSuffix(trimmed, "/edit") {
		return h.hydrateEdit(ctx, screen, trimmed, principal)
	}
	if binding, ok := resourceBindingForList(trimmed); ok && screen.View == render.ViewTable {
		return h.hydrateList(ctx, screen, binding, query, principal)
	}
	return screen, nil
}

func (h Hydrator) hydrateList(ctx context.Context, screen render.ScreenModel, binding ResourceBinding, query url.Values, principal appauthn.Principal) (render.ScreenModel, error) {
	page, perPage := parsePagination(query)
	publicOnly := !principal.Has(modulecms.CapabilityContentPrivate)
	kind := contentKindForRecord(binding.Record.ID)
	if kind == "" {
		return screen, nil
	}
	var rows []render.Row
	var total int
	err := h.Provider.ForWorkspace(h.workspace()).WithinTx(ctx, func(ctx context.Context, repos storage.Repositories) error {
		services := wiring.Build(repos, wiring.Deps{Hooks: h.Hooks, Audit: h.Audit})
		items, err := services.Content.ListFiltered(ctx, domaincontent.Query{Kind: kind}, publicOnly)
		if err != nil {
			return err
		}
		total = len(items)
		start := (page - 1) * perPage
		if start > total {
			start = total
		}
		end := start + perPage
		if end > total {
			end = total
		}
		pageItems := items[start:end]
		rows = make([]render.Row, 0, len(pageItems))
		for _, entry := range pageItems {
			if !canViewEntry(principal, entry, publicOnly) {
				continue
			}
			rows = append(rows, render.Row{
				ID: string(entry.ID),
				Cells: map[string]string{
					"id":     string(entry.ID),
					"title":  localized(entry.Title),
					"slug":   entry.Slug,
					"status": string(entry.Status),
				},
				Actions: rowActionsForEntry(screen, entry, principal),
			})
		}
		return nil
	})
	if err != nil {
		return screen, err
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	screen.Rows = rows
	screen.Pagination = &render.Pagination{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}
	return screen, nil
}

func (h Hydrator) hydrateEdit(ctx context.Context, screen render.ScreenModel, path string, principal appauthn.Principal) (render.ScreenModel, error) {
	base := strings.TrimSuffix(path, "/edit")
	id := ""
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		id = base[idx+1:]
		base = base[:idx]
	}
	if id == "" {
		return screen, nil
	}
	kind := contentKindForPath(base)
	if kind == "" {
		return screen, nil
	}
	err := h.Provider.ForWorkspace(h.workspace()).WithinTx(ctx, func(ctx context.Context, repos storage.Repositories) error {
		services := wiring.Build(repos, wiring.Deps{Hooks: h.Hooks, Audit: h.Audit})
		entry, ok, err := services.Content.Get(ctx, domaincontent.ID(id))
		if err != nil {
			return err
		}
		if !ok || entry.Kind != kind {
			return fmt.Errorf("content not found")
		}
		if !capcheck.CanEdit(principal, entry.AuthorID) {
			return fmt.Errorf("forbidden")
		}
		screen.FormRecord = &render.FormRecord{Values: entryFormValues(entry)}
		if terms, err := services.Taxonomy.ListTerms(ctx, "category"); err == nil {
			screen = appendTermOptions(screen, terms)
		}
		if assets, err := services.Media.List(ctx); err == nil {
			screen = appendMediaOptions(screen, assets)
		}
		screen.RowActions = formWorkflowActions(base, id, entry, principal)
		return nil
	})
	if err != nil {
		return screen, err
	}
	return screen, nil
}

func (h Hydrator) hydrateRuntime(screen render.ScreenModel) (render.ScreenModel, error) {
	if h.Audit == nil {
		return screen, nil
	}
	events := h.Audit.Audit()
	limit := 50
	if len(events) > limit {
		events = events[:limit]
	}
	rows := make([]render.Row, 0, len(events))
	for _, event := range events {
		resource := event.Resource
		if resource == "" && event.ResourceType != "" {
			resource = event.ResourceType + "/" + event.ResourceID
		}
		rows = append(rows, render.Row{
			ID: event.Action + event.CreatedAt.Format(time.RFC3339Nano),
			Cells: map[string]string{
				"action":     event.Action,
				"actor":      event.Actor,
				"resource":   resource,
				"created_at": event.CreatedAt.Format(time.RFC3339),
			},
		})
	}
	screen.Rows = rows
	if screen.Pagination == nil {
		screen.Pagination = &render.Pagination{Page: 1, PerPage: 50, Total: len(rows), TotalPages: 1}
	}
	return screen, nil
}

func formWorkflowActions(base, id string, entry domaincontent.Entry, principal appauthn.Principal) []render.RowAction {
	collection := adminCollectionForRecord(toolset.RecordTypeID(entry.Kind))
	if collection == "" {
		collection = strings.TrimPrefix(base, "/go-admin/")
	}
	prefix := fmt.Sprintf("/go-admin/%s/%s", collection, id)
	actions := []render.RowAction{}
	if capcheck.CanPublish(principal) && entry.Status != domaincontent.StatusPublished {
		actions = append(actions, render.RowAction{ID: "publish", Label: "Publish", Method: "POST", Path: prefix + "/publish", Scope: "admin.content.publish"})
	}
	if capcheck.CanSchedule(principal) && entry.Status != domaincontent.StatusPublished {
		actions = append(actions, render.RowAction{ID: "schedule", Label: "Schedule", Method: "POST", Path: prefix + "/schedule", Scope: "admin.content.schedule"})
	}
	if capcheck.CanDelete(principal) && entry.Status != domaincontent.StatusTrashed {
		actions = append(actions, render.RowAction{ID: "trash", Label: "Trash", Method: "POST", Path: prefix + "/trash", Scope: "admin.content.write", Destructive: true})
	}
	if capcheck.CanRestore(principal) && entry.Status == domaincontent.StatusTrashed {
		actions = append(actions, render.RowAction{ID: "restore", Label: "Restore", Method: "POST", Path: prefix + "/restore", Scope: "admin.content.restore"})
	}
	if capcheck.CanManageRevisions(principal) {
		actions = append(actions, render.RowAction{ID: "revisions", Label: "View revisions API", Method: "GET", Path: fmt.Sprintf("/go-json/go/v2/%s/%s/revisions", collection, id)})
	}
	return actions
}

func (h Hydrator) workspace() contracts.WorkspaceID {
	if h.Workspace == "" {
		return "root"
	}
	return h.Workspace
}

func resourceBindingForList(path string) (ResourceBinding, bool) {
	switch path {
	case "/go-admin/posts":
		return ResourceBinding{Record: toolset.RecordTypeDefinition{ID: "post"}}, true
	case "/go-admin/pages":
		return ResourceBinding{Record: toolset.RecordTypeDefinition{ID: "page"}}, true
	default:
		return ResourceBinding{}, false
	}
}

func contentKindForRecord(id toolset.RecordTypeID) domaincontent.Kind {
	switch id {
	case "post":
		return domaincontent.KindPost
	case "page":
		return domaincontent.KindPage
	default:
		return ""
	}
}

func contentKindForPath(base string) domaincontent.Kind {
	switch base {
	case "/go-admin/posts":
		return domaincontent.KindPost
	case "/go-admin/pages":
		return domaincontent.KindPage
	default:
		return ""
	}
}

func parsePagination(query url.Values) (int, int) {
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}
	perPage, _ := strconv.Atoi(query.Get("per_page"))
	if perPage <= 0 {
		perPage = 20
	}
	return page, perPage
}

func canViewEntry(principal appauthn.Principal, entry domaincontent.Entry, publicOnly bool) bool {
	if !publicOnly {
		return true
	}
	return rest.IsPublicEntry(entry, time.Now().UTC())
}

func rowActionsForEntry(screen render.ScreenModel, entry domaincontent.Entry, principal appauthn.Principal) []render.RowAction {
	base := screen.Metadata["list_path"]
	if base == "" {
		base = "/go-admin"
	}
	collection := adminCollectionForRecord(toolset.RecordTypeID(entry.Kind))
	actions := []render.RowAction{
		{ID: "edit", Label: "Edit", Method: "GET", Path: fmt.Sprintf("%s/%s/edit", strings.TrimRight(base, "/"), entry.ID)},
	}
	if capcheck.CanPublish(principal) && entry.Status != domaincontent.StatusPublished {
		actions = append(actions, render.RowAction{ID: "publish", Label: "Publish", Method: "POST", Path: fmt.Sprintf("/go-admin/%s/%s/publish", collection, entry.ID), Scope: "admin.content.publish"})
	}
	if capcheck.CanDelete(principal) && entry.Status != domaincontent.StatusTrashed {
		actions = append(actions, render.RowAction{ID: "trash", Label: "Trash", Method: "POST", Path: fmt.Sprintf("/go-admin/%s/%s/trash", collection, entry.ID), Scope: "admin.content.write", Destructive: true})
	}
	if capcheck.CanRestore(principal) && entry.Status == domaincontent.StatusTrashed {
		actions = append(actions, render.RowAction{ID: "restore", Label: "Restore", Method: "POST", Path: fmt.Sprintf("/go-admin/%s/%s/restore", collection, entry.ID), Scope: "admin.content.restore"})
	}
	return actions
}

func entryFormValues(entry domaincontent.Entry) map[string]string {
	values := map[string]string{
		"id":                string(entry.ID),
		"title":             localized(entry.Title),
		"slug":              entry.Slug,
		"content":           entry.Content,
		"excerpt":           entry.Excerpt,
		"status":            string(entry.Status),
		"visibility":        string(entry.Visibility),
		"author_id":         entry.AuthorID,
		"featured_media_id": entry.FeaturedMediaID,
	}
	if !entry.ScheduledFor.IsZero() {
		values["scheduled_at"] = entry.ScheduledFor.Format("2006-01-02T15:04")
	}
	if len(entry.TermIDs) > 0 {
		values["taxonomy_term_ids"] = strings.Join(entry.TermIDs, ",")
	}
	return values
}

func localized(value any) string {
	switch typed := value.(type) {
	case map[string]string:
		if v, ok := typed["en"]; ok {
			return v
		}
		for _, v := range typed {
			return v
		}
	case string:
		return typed
	}
	return fmt.Sprint(value)
}

func appendTermOptions(screen render.ScreenModel, terms []taxonomy.Term) render.ScreenModel {
	options := make([]panel.Option, 0, len(terms))
	for _, term := range terms {
		options = append(options, panel.Option{Value: term.ID, Label: localized(term.Slug)})
	}
	for i, field := range screen.Fields {
		if field.ID == "taxonomy_term_ids" {
			screen.Fields[i].Options = options
		}
	}
	return screen
}

func appendMediaOptions(screen render.ScreenModel, assets []domainmedia.Asset) render.ScreenModel {
	options := make([]panel.Option, 0, len(assets))
	for _, asset := range assets {
		options = append(options, panel.Option{Value: asset.ID, Label: asset.Title})
	}
	for i, field := range screen.Fields {
		if field.ID == "featured_media_id" {
			screen.Fields[i].Options = options
		}
	}
	return screen
}
