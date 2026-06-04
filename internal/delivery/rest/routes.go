package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
	"github.com/fastygo/app-gocms/internal/domain/media"
	"github.com/fastygo/app-gocms/internal/domain/menus"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/domain/taxonomy"
	"github.com/fastygo/app-gocms/pkg/module/codex"
)

func (h Handler) dispatch(ctx context.Context, r *http.Request, path string, s services) (int, any) {
	parts := splitPath(path)
	switch parts[0] {
	case "posts", "pages":
		return h.content(ctx, r, parts, s, kindForCollection(parts[0]))
	case "content-types":
		if r.Method == http.MethodGet && len(parts) == 1 {
			items, err := s.contentTypes.List(ctx)
			return result(http.StatusOK, codex.ResourceEnvelope[[]contenttype.Type]{Data: items}, err)
		}
	case "taxonomies":
		return h.taxonomies(ctx, r, parts, s)
	case "media":
		return h.media(ctx, r, parts, s)
	case "authors":
		if r.Method == http.MethodGet && len(parts) == 2 {
			author, ok, err := s.users.PublicAuthor(ctx, parts[1])
			if err != nil {
				return serverError(err)
			}
			if !ok {
				return notFound("author not found")
			}
			return http.StatusOK, codex.ResourceEnvelope[any]{Data: author}
		}
	case "settings":
		if r.Method == http.MethodGet && len(parts) == 1 {
			items, err := s.settings.Public(ctx)
			return result(http.StatusOK, codex.ResourceEnvelope[[]settings.Value]{Data: items}, err)
		}
	case "menus":
		if r.Method == http.MethodGet && len(parts) == 1 {
			items, err := s.menus.List(ctx)
			return result(http.StatusOK, codex.ResourceEnvelope[[]menus.Menu]{Data: items}, err)
		}
		if r.Method == http.MethodGet && len(parts) == 2 {
			menu, ok, err := s.menus.ByLocation(ctx, parts[1])
			if err != nil {
				return serverError(err)
			}
			if !ok {
				return notFound("menu not found")
			}
			return http.StatusOK, codex.ResourceEnvelope[menus.Menu]{Data: menu}
		}
	case "search":
		if r.Method == http.MethodGet {
			return h.search(ctx, r, s)
		}
	}
	return notFound("route not found")
}

func (h Handler) content(ctx context.Context, r *http.Request, parts []string, s services, kind domaincontent.Kind) (int, any) {
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		return h.contentList(ctx, r, s, domaincontent.Query{Kind: kind})
	case len(parts) == 1 && r.Method == http.MethodPost:
		if !canMutate(r) {
			return unauthorized()
		}
		var entry domaincontent.Entry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			return validationError(err.Error())
		}
		entry.Kind = kind
		created, err := s.content.CreateDraft(ctx, entry)
		if err == nil && entry.Status == domaincontent.StatusPublished {
			created, err = s.content.Publish(ctx, created.ID)
		}
		return result(http.StatusCreated, codex.ResourceEnvelope[domaincontent.Entry]{Data: created}, err)
	case len(parts) == 2 && r.Method == http.MethodGet:
		entry, ok, err := s.content.Get(ctx, domaincontent.ID(parts[1]))
		if err != nil {
			return serverError(err)
		}
		if !ok || entry.Kind != kind {
			return notFound("content not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[domaincontent.Entry]{Data: entry}
	case len(parts) == 2 && r.Method == http.MethodPatch:
		if !canMutate(r) {
			return unauthorized()
		}
		existing, ok, err := s.content.Get(ctx, domaincontent.ID(parts[1]))
		if err != nil {
			return serverError(err)
		}
		if !ok {
			return notFound("content not found")
		}
		var patch domaincontent.Entry
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			return validationError(err.Error())
		}
		updated := mergeContent(existing, patch)
		err = s.content.Update(ctx, updated)
		return result(http.StatusOK, codex.ResourceEnvelope[domaincontent.Entry]{Data: updated}, err)
	case len(parts) == 2 && r.Method == http.MethodDelete:
		if !canMutate(r) {
			return unauthorized()
		}
		entry, err := s.content.Trash(ctx, domaincontent.ID(parts[1]))
		return result(http.StatusOK, codex.ResourceEnvelope[domaincontent.Entry]{Data: entry}, err)
	}
	return notFound("route not found")
}

func (h Handler) contentList(ctx context.Context, r *http.Request, s services, query domaincontent.Query) (int, any) {
	page, perPage, err := pagination(r)
	if err != nil {
		return validationError(err.Error())
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query.Status = domaincontent.Status(status)
	}
	items, err := s.content.List(ctx, query)
	if err != nil {
		return serverError(err)
	}
	return http.StatusOK, listEnvelope(items, page, perPage)
}

func (h Handler) search(ctx context.Context, r *http.Request, s services) (int, any) {
	status, payload := h.contentList(ctx, r, s, domaincontent.Query{})
	if status != http.StatusOK {
		return status, payload
	}
	q := strings.ToLower(firstNonEmpty(r.URL.Query().Get("q"), r.URL.Query().Get("search")))
	envelope := payload.(codex.ListEnvelope[domaincontent.Entry])
	if q == "" {
		return status, envelope
	}
	filtered := []domaincontent.Entry{}
	for _, entry := range envelope.Data {
		if strings.Contains(strings.ToLower(entry.Slug), q) || strings.Contains(strings.ToLower(entry.Content), q) || strings.Contains(strings.ToLower(entry.Title["en"]), q) {
			filtered = append(filtered, entry)
		}
	}
	return http.StatusOK, listEnvelope(filtered, envelope.Pagination.Page, envelope.Pagination.PerPage)
}

func (h Handler) taxonomies(ctx context.Context, r *http.Request, parts []string, s services) (int, any) {
	if len(parts) == 1 && r.Method == http.MethodGet {
		// Taxonomies are intentionally exposed as resources without pagination, matching the current AppCMS codex shell.
		return http.StatusOK, codex.ResourceEnvelope[[]taxonomy.Definition]{Data: []taxonomy.Definition{}}
	}
	if len(parts) == 1 && r.Method == http.MethodPost {
		if !canMutate(r) {
			return unauthorized()
		}
		var definition taxonomy.Definition
		if err := json.NewDecoder(r.Body).Decode(&definition); err != nil {
			return validationError(err.Error())
		}
		return result(http.StatusCreated, codex.ResourceEnvelope[taxonomy.Definition]{Data: definition}, s.taxonomy.Register(ctx, definition))
	}
	if len(parts) == 3 && parts[2] == "terms" && r.Method == http.MethodGet {
		items, err := s.taxonomy.ListTerms(ctx, parts[1])
		return result(http.StatusOK, codex.ResourceEnvelope[[]taxonomy.Term]{Data: items}, err)
	}
	if len(parts) == 3 && parts[2] == "terms" && r.Method == http.MethodPost {
		if !canMutate(r) {
			return unauthorized()
		}
		var term taxonomy.Term
		if err := json.NewDecoder(r.Body).Decode(&term); err != nil {
			return validationError(err.Error())
		}
		term.TaxonomyType = parts[1]
		return result(http.StatusCreated, codex.ResourceEnvelope[taxonomy.Term]{Data: term}, s.taxonomy.CreateTerm(ctx, term))
	}
	return notFound("route not found")
}

func (h Handler) media(ctx context.Context, r *http.Request, parts []string, s services) (int, any) {
	if len(parts) == 1 && r.Method == http.MethodGet {
		items, err := s.media.List(ctx)
		return result(http.StatusOK, codex.ResourceEnvelope[[]media.Asset]{Data: items}, err)
	}
	if len(parts) == 1 && r.Method == http.MethodPost {
		if !canMutate(r) {
			return unauthorized()
		}
		var asset media.Asset
		if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
			return validationError(err.Error())
		}
		return result(http.StatusCreated, codex.ResourceEnvelope[media.Asset]{Data: asset}, s.media.SaveMetadata(ctx, asset))
	}
	if len(parts) == 2 && r.Method == http.MethodGet {
		asset, ok, err := s.media.Get(ctx, parts[1])
		if err != nil {
			return serverError(err)
		}
		if !ok {
			return notFound("media asset not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[media.Asset]{Data: asset}
	}
	return notFound("route not found")
}

func kindForCollection(collection string) domaincontent.Kind {
	if collection == "pages" {
		return domaincontent.KindPage
	}
	return domaincontent.KindPost
}

func mergeContent(existing domaincontent.Entry, patch domaincontent.Entry) domaincontent.Entry {
	if patch.Title != nil {
		existing.Title = patch.Title
	}
	if patch.Slug != "" {
		existing.Slug = patch.Slug
	}
	if patch.Content != "" {
		existing.Content = patch.Content
	}
	if patch.Excerpt != "" {
		existing.Excerpt = patch.Excerpt
	}
	if patch.Status != "" {
		existing.Status = patch.Status
	}
	if patch.Visibility != "" {
		existing.Visibility = patch.Visibility
	}
	if patch.AuthorID != "" {
		existing.AuthorID = patch.AuthorID
	}
	if patch.FeaturedMediaID != "" {
		existing.FeaturedMediaID = patch.FeaturedMediaID
	}
	if patch.Metadata != nil {
		existing.Metadata = patch.Metadata
	}
	return existing
}

func splitPath(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return []string{""}
	}
	return parts
}

func pagination(r *http.Request) (int, int, error) {
	page, err := parsePositive(r.URL.Query().Get("page"), 1)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page")
	}
	perPage, err := parsePositive(r.URL.Query().Get("per_page"), 20)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid per_page")
	}
	return page, perPage, nil
}

func parsePositive(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid positive integer")
	}
	return parsed, nil
}

func listEnvelope[T any](items []T, page int, perPage int) codex.ListEnvelope[T] {
	total := len(items)
	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return codex.ListEnvelope[T]{Data: items[start:end], Pagination: codex.Pagination{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}}
}

func canMutate(r *http.Request) bool {
	return r.Header.Get("Authorization") == "Bearer admin-token"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
