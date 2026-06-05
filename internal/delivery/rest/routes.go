package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
	"github.com/fastygo/app-gocms/internal/domain/media"
	"github.com/fastygo/app-gocms/internal/domain/taxonomy"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/app-gocms/pkg/module/codex"
	"github.com/fastygo/platform/pkg/contracts"
)

type requestContext struct {
	includePrivate bool
}

func (h Handler) readContext(r *http.Request) requestContext {
	rc := requestContext{}
	if h.Authorize != nil {
		authenticated, allowed := h.Authorize(r, modulecms.CapabilityContentPrivate)
		rc.includePrivate = authenticated && allowed
	}
	return rc
}

func (h Handler) dispatch(ctx context.Context, r *http.Request, path string, s services) (int, any) {
	readCtx := h.readContext(r)
	parts := splitPath(path)
	switch parts[0] {
	case "posts", "pages":
		return h.content(ctx, r, readCtx, parts, s, kindForCollection(parts[0]))
	case "content":
		return h.contentTerms(ctx, r, readCtx, parts, s)
	case "content-types":
		return h.contentTypes(ctx, r, readCtx, parts, s)
	case "taxonomies":
		return h.taxonomies(ctx, r, readCtx, parts, s)
	case "media":
		return h.media(ctx, r, readCtx, parts, s)
	case "authors":
		return h.authors(ctx, r, readCtx, parts, s)
	case "settings":
		return h.settings(ctx, r, readCtx, parts, s)
	case "menus":
		return h.menus(ctx, r, readCtx, parts, s)
	case "search":
		if r.Method == http.MethodGet {
			return h.search(ctx, r, readCtx, s)
		}
	}
	return notFound("route not found")
}

func (h Handler) content(ctx context.Context, r *http.Request, readCtx requestContext, parts []string, s services, kind domaincontent.Kind) (int, any) {
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		return h.contentList(ctx, r, readCtx, s, domaincontent.Query{Kind: kind})
	case len(parts) == 1 && r.Method == http.MethodPost:
		if status, payload, ok := h.authorize(r, modulecms.CapabilityContentWrite); !ok {
			return status, payload
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
		if err != nil {
			return errorResponse(err)
		}
		return http.StatusCreated, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(created, true)}
	case len(parts) == 3 && parts[1] == "by-slug" && r.Method == http.MethodGet:
		entry, ok, err := s.content.GetBySlug(ctx, kind, parts[2], !readCtx.includePrivate)
		if err != nil {
			return serverError(err)
		}
		if !ok {
			return notFound("content not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(entry, readCtx.includePrivate)}
	case len(parts) == 2 && r.Method == http.MethodGet:
		entry, ok, err := s.content.Get(ctx, domaincontent.ID(parts[1]))
		if err != nil {
			return serverError(err)
		}
		if !ok || entry.Kind != kind {
			return notFound("content not found")
		}
		if !readCtx.includePrivate && !IsPublicEntry(entry, time.Now().UTC()) {
			return notFound("content not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(entry, readCtx.includePrivate)}
	case len(parts) == 2 && r.Method == http.MethodPatch:
		if status, payload, ok := h.authorize(r, modulecms.CapabilityContentWrite); !ok {
			return status, payload
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
		if err := s.content.Update(ctx, updated); err != nil {
			return errorResponse(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(updated, true)}
	case len(parts) == 2 && r.Method == http.MethodDelete:
		if status, payload, ok := h.authorize(r, modulecms.CapabilityContentWrite); !ok {
			return status, payload
		}
		entry, err := s.content.Trash(ctx, domaincontent.ID(parts[1]))
		if err != nil {
			return errorResponse(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(entry, true)}
	}
	return notFound("route not found")
}

func (h Handler) contentList(ctx context.Context, r *http.Request, readCtx requestContext, s services, query domaincontent.Query) (int, any) {
	page, perPage, err := pagination(r)
	if err != nil {
		return validationError(err.Error())
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query.Status = domaincontent.Status(status)
	}
	items, err := s.content.ListFiltered(ctx, query, !readCtx.includePrivate)
	if err != nil {
		return serverError(err)
	}
	return http.StatusOK, contentListEnvelope(items, readCtx.includePrivate, page, perPage)
}

func (h Handler) contentTerms(ctx context.Context, r *http.Request, _ requestContext, parts []string, s services) (int, any) {
	if len(parts) == 3 && parts[2] == "terms" && r.Method == http.MethodPost {
		if status, payload, ok := h.authorize(r, modulecms.CapabilityTaxonomyAssign); !ok {
			return status, payload
		}
		var request struct {
			TermIDs []string `json:"term_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return validationError(err.Error())
		}
		entry, err := s.taxonomy.AssignTerms(ctx, domaincontent.ID(parts[1]), request.TermIDs)
		if err != nil {
			return errorResponse(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(entry, true)}
	}
	return notFound("route not found")
}

func (h Handler) contentTypes(ctx context.Context, r *http.Request, _ requestContext, parts []string, s services) (int, any) {
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		items, err := s.contentTypes.List(ctx)
		if err != nil {
			return serverError(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[[]ContentTypeDTO]{Data: projectContentTypes(items)}
	case len(parts) == 1 && r.Method == http.MethodPost:
		if status, payload, ok := h.authorize(r, modulecms.CapabilityContentWrite); !ok {
			return status, payload
		}
		var item contenttype.Type
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			return validationError(err.Error())
		}
		if err := s.contentTypes.Register(ctx, item); err != nil {
			return errorResponse(err)
		}
		return http.StatusCreated, codex.ResourceEnvelope[ContentTypeDTO]{Data: ContentTypeProjection(item)}
	default:
		return notFound("route not found")
	}
}

func (h Handler) search(ctx context.Context, r *http.Request, readCtx requestContext, s services) (int, any) {
	status, payload := h.contentList(ctx, r, readCtx, s, domaincontent.Query{})
	if status != http.StatusOK {
		return status, payload
	}
	q := strings.ToLower(firstNonEmpty(r.URL.Query().Get("q"), r.URL.Query().Get("search")))
	envelope := payload.(codex.ListEnvelope[ContentDTO])
	if q == "" {
		return status, envelope
	}
	filtered := []ContentDTO{}
	for _, entry := range envelope.Data {
		if strings.Contains(strings.ToLower(delocalize(entry.Slug)), q) ||
			strings.Contains(strings.ToLower(delocalize(entry.Content)), q) ||
			strings.Contains(strings.ToLower(delocalize(entry.Title)), q) {
			filtered = append(filtered, entry)
		}
	}
	return http.StatusOK, listEnvelope(filtered, envelope.Pagination.Page, envelope.Pagination.PerPage)
}

func (h Handler) taxonomies(ctx context.Context, r *http.Request, _ requestContext, parts []string, s services) (int, any) {
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		items, err := s.taxonomy.ListDefinitions(ctx)
		if err != nil {
			return serverError(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[[]TaxonomyDTO]{Data: projectTaxonomies(items)}
	case len(parts) == 1 && r.Method == http.MethodPost:
		if status, payload, ok := h.authorize(r, modulecms.CapabilityTaxonomyManage); !ok {
			return status, payload
		}
		var definition taxonomy.Definition
		if err := json.NewDecoder(r.Body).Decode(&definition); err != nil {
			return validationError(err.Error())
		}
		if err := s.taxonomy.Register(ctx, definition); err != nil {
			return errorResponse(err)
		}
		return http.StatusCreated, codex.ResourceEnvelope[TaxonomyDTO]{Data: TaxonomyProjection(definition)}
	case len(parts) == 2 && r.Method == http.MethodGet:
		items, err := s.taxonomy.ListTerms(ctx, parts[1])
		if err != nil {
			return serverError(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[[]TermDTO]{Data: projectTerms(items)}
	case len(parts) == 3 && parts[2] == "terms" && r.Method == http.MethodPost:
		if status, payload, ok := h.authorize(r, modulecms.CapabilityTaxonomyAssign); !ok {
			return status, payload
		}
		var term taxonomy.Term
		if err := json.NewDecoder(r.Body).Decode(&term); err != nil {
			return validationError(err.Error())
		}
		term.TaxonomyType = parts[1]
		if err := s.taxonomy.CreateTerm(ctx, term); err != nil {
			return errorResponse(err)
		}
		return http.StatusCreated, codex.ResourceEnvelope[TermDTO]{Data: TermProjection(term)}
	case len(parts) == 3 && r.Method == http.MethodGet:
		term, ok, err := s.taxonomy.GetTerm(ctx, parts[2])
		if err != nil {
			return serverError(err)
		}
		if !ok || term.TaxonomyType != parts[1] {
			return notFound("term not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[TermDTO]{Data: TermProjection(term)}
	default:
		return notFound("route not found")
	}
}

func (h Handler) media(ctx context.Context, r *http.Request, _ requestContext, parts []string, s services) (int, any) {
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		items, err := s.media.List(ctx)
		if err != nil {
			return serverError(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[[]MediaDTO]{Data: projectMediaList(items)}
	case len(parts) == 1 && r.Method == http.MethodPost:
		if status, payload, ok := h.authorize(r, modulecms.CapabilityMediaUpload); !ok {
			return status, payload
		}
		var asset media.Asset
		if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
			return validationError(err.Error())
		}
		if err := s.media.SaveMetadata(ctx, asset); err != nil {
			return errorResponse(err)
		}
		return http.StatusCreated, codex.ResourceEnvelope[MediaDTO]{Data: MediaProjection(asset)}
	case len(parts) == 2 && r.Method == http.MethodGet:
		asset, ok, err := s.media.Get(ctx, parts[1])
		if err != nil {
			return serverError(err)
		}
		if !ok {
			return notFound("media asset not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[MediaDTO]{Data: MediaProjection(asset)}
	case len(parts) == 2 && r.Method == http.MethodPatch:
		if status, payload, ok := h.authorize(r, modulecms.CapabilityMediaEdit); !ok {
			return status, payload
		}
		var asset media.Asset
		if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
			return validationError(err.Error())
		}
		asset.ID = parts[1]
		if err := s.media.Update(ctx, asset); err != nil {
			return errorResponse(err)
		}
		updated, ok, err := s.media.Get(ctx, parts[1])
		if err != nil {
			return serverError(err)
		}
		if !ok {
			return notFound("media asset not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[MediaDTO]{Data: MediaProjection(updated)}
	case len(parts) == 4 && parts[2] == "featured" && r.Method == http.MethodPost:
		if status, payload, ok := h.authorize(r, modulecms.CapabilityMediaEdit); !ok {
			return status, payload
		}
		entry, err := s.media.AttachFeatured(ctx, domaincontent.ID(parts[3]), parts[1])
		if err != nil {
			return errorResponse(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(entry, true)}
	default:
		return notFound("route not found")
	}
}

func (h Handler) authors(ctx context.Context, r *http.Request, _ requestContext, parts []string, s services) (int, any) {
	if r.Method == http.MethodGet && len(parts) == 2 {
		author, ok, err := s.users.PublicAuthor(ctx, parts[1])
		if err != nil {
			return serverError(err)
		}
		if !ok {
			return notFound("author not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[AuthorDTO]{Data: AuthorProjection(author, resolveAvatarURL(ctx, s.media, author.AvatarID))}
	}
	return notFound("route not found")
}

func (h Handler) settings(ctx context.Context, r *http.Request, _ requestContext, parts []string, s services) (int, any) {
	if r.Method == http.MethodGet && len(parts) == 1 {
		items, err := s.settings.Public(ctx)
		if err != nil {
			return serverError(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[[]SettingDTO]{Data: projectPublicSettings(items)}
	}
	return notFound("route not found")
}

func (h Handler) menus(ctx context.Context, r *http.Request, _ requestContext, parts []string, s services) (int, any) {
	switch {
	case r.Method == http.MethodGet && len(parts) == 1:
		items, err := s.menus.List(ctx)
		if err != nil {
			return serverError(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[[]MenuDTO]{Data: projectMenus(items)}
	case r.Method == http.MethodGet && len(parts) == 2:
		menu, ok, err := s.menus.ByLocation(ctx, parts[1])
		if err != nil {
			return serverError(err)
		}
		if !ok {
			return notFound("menu not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[MenuDTO]{Data: MenuProjection(menu)}
	default:
		return notFound("route not found")
	}
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
	if patch.TermIDs != nil {
		existing.TermIDs = patch.TermIDs
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

func contentListEnvelope(items []domaincontent.Entry, includePrivate bool, page int, perPage int) codex.ListEnvelope[ContentDTO] {
	return listEnvelope(projectContentList(items, includePrivate), page, perPage)
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

func (h Handler) authorize(r *http.Request, capability contracts.CapabilityID) (int, any, bool) {
	if h.Authorize == nil {
		return unauthorizedStatus()
	}
	authenticated, allowed := h.Authorize(r, capability)
	if !authenticated {
		return unauthorizedStatus()
	}
	if !allowed {
		status, payload := forbidden()
		return status, payload, false
	}
	return 0, nil, true
}

func unauthorizedStatus() (int, any, bool) {
	status, payload := unauthorized()
	return status, payload, false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
