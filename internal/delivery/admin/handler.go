package admin

import (
	"context"
	"net/http"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/application/wiring"
	"github.com/fastygo/app-gocms/internal/extensions"
	"github.com/fastygo/app-gocms/internal/operations"
	"github.com/fastygo/app-gocms/internal/storage"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/app-gocms/pkg/module/capcheck"
	"github.com/fastygo/platform/pkg/contracts"
)

type ActionTokenValidator interface {
	ValidActionToken(raw, scope string) bool
}

type PrincipalResolver interface {
	Principal(r *http.Request) (appauthn.Principal, bool)
}

type Handler struct {
	Provider  storage.StoreProvider
	Workspace contracts.WorkspaceID
	Auth      PrincipalResolver
	Tokens    ActionTokenValidator
	Hooks     *extensions.HookBus
	Audit     operations.AuditRecorder
}

func NewHandler(provider storage.StoreProvider, workspace contracts.WorkspaceID, auth PrincipalResolver, tokens ActionTokenValidator) Handler {
	return NewHandlerWithDeps(provider, workspace, auth, tokens, nil, nil)
}

func NewHandlerWithDeps(provider storage.StoreProvider, workspace contracts.WorkspaceID, auth PrincipalResolver, tokens ActionTokenValidator, hooks *extensions.HookBus, audit operations.AuditRecorder) Handler {
	if workspace == "" {
		workspace = "root"
	}
	return Handler{Provider: provider, Workspace: workspace, Auth: auth, Tokens: tokens, Hooks: hooks, Audit: audit}
}

func (h Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /go-admin/posts", h.contentCreate(domaincontent.KindPost))
	mux.HandleFunc("POST /go-admin/posts/{id}", h.contentUpdate(domaincontent.KindPost))
	mux.HandleFunc("POST /go-admin/posts/{id}/trash", h.contentTrash(domaincontent.KindPost))
	mux.HandleFunc("POST /go-admin/pages", h.contentCreate(domaincontent.KindPage))
	mux.HandleFunc("POST /go-admin/pages/{id}", h.contentUpdate(domaincontent.KindPage))
	mux.HandleFunc("POST /go-admin/pages/{id}/trash", h.contentTrash(domaincontent.KindPage))
	mux.HandleFunc("POST /go-admin/taxonomies", h.taxonomyCreate)
	mux.HandleFunc("POST /go-admin/taxonomies/{type}/terms", h.termCreate)
	mux.HandleFunc("POST /go-admin/taxonomies/{type}/terms/{id}", h.termUpdate)
	mux.HandleFunc("POST /go-admin/terms", h.termCreateFlat)
	mux.HandleFunc("POST /go-admin/terms/{id}", h.termUpdateFlat)
	mux.HandleFunc("POST /go-admin/media", h.mediaSave)
	mux.HandleFunc("POST /go-admin/media/{id}", h.mediaUpdate)
	mux.HandleFunc("POST /go-admin/settings", h.settingsSave)
	mux.HandleFunc("POST /go-admin/menus", h.menuSave)
	mux.HandleFunc("POST /go-admin/menus/{id}", h.menuUpdate)
	mux.HandleFunc("POST /go-admin/content-types", h.contentTypeCreate)
	h.registerTransitions(mux, domaincontent.KindPost, "posts")
	h.registerTransitions(mux, domaincontent.KindPage, "pages")
}

func (h Handler) withWiring(r *http.Request, fn func(ctx context.Context, s wiring.Services) error) error {
	return h.Provider.ForWorkspace(h.Workspace).WithinTx(r.Context(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		return fn(ctx, wiring.Build(repos, wiring.Deps{Hooks: h.Hooks, Audit: h.Audit}))
	})
}

func (h Handler) requirePrincipal(w http.ResponseWriter, r *http.Request) (appauthn.Principal, bool) {
	principal, ok := h.Auth.Principal(r)
	if !ok {
		http.Redirect(w, r, "/go-login", http.StatusSeeOther)
		return appauthn.Principal{}, false
	}
	return principal, true
}

func (h Handler) validateToken(w http.ResponseWriter, r *http.Request, scope string) bool {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return false
	}
	if !h.Tokens.ValidActionToken(r.PostForm.Get("action_token"), scope) {
		http.Error(w, "invalid action token", http.StatusForbidden)
		return false
	}
	return true
}

func requireCapability(w http.ResponseWriter, principal appauthn.Principal, capability contracts.CapabilityID) bool {
	if !principal.Has(capability) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (h Handler) contentCreate(kind domaincontent.Kind) http.HandlerFunc {
	listPath := "/go-admin/posts"
	if kind == domaincontent.KindPage {
		listPath = "/go-admin/pages"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := h.requirePrincipal(w, r)
		if !ok {
			return
		}
		if !capcheck.CanCreate(principal) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !h.validateToken(w, r, ActionContentWrite) {
			return
		}
		entry, err := bindContent(r.PostForm, kind, "")
		if err != nil {
			setFlashError(w, err.Error())
			http.Redirect(w, r, listPath+"/new", http.StatusSeeOther)
			return
		}
		entry.Content = SanitizeHTML(entry.Content)
		entry.Excerpt = SanitizeHTML(entry.Excerpt)
		var createdID domaincontent.ID
		err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
			created, err := s.Content.CreateDraft(ctx, entry)
			if err != nil {
				return err
			}
			createdID = created.ID
			if entry.Status == domaincontent.StatusPublished {
				_, err = s.Content.Publish(ctx, created.ID)
			}
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.recordAudit(principal, operations.ActionAdminContentCreate, string(kind), string(createdID), nil)
		http.Redirect(w, r, listPath, http.StatusSeeOther)
	}
}

func (h Handler) contentUpdate(kind domaincontent.Kind) http.HandlerFunc {
	listPath := "/go-admin/posts"
	if kind == domaincontent.KindPage {
		listPath = "/go-admin/pages"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := h.requirePrincipal(w, r)
		if !ok {
			return
		}
		if !h.validateToken(w, r, ActionContentWrite) {
			return
		}
		entry, err := bindContent(r.PostForm, kind, domaincontent.ID(r.PathValue("id")))
		if err != nil {
			setFlashError(w, err.Error())
			http.Redirect(w, r, listPath+"/"+r.PathValue("id")+"/edit", http.StatusSeeOther)
			return
		}
		entry.Content = SanitizeHTML(entry.Content)
		entry.Excerpt = SanitizeHTML(entry.Excerpt)
		err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
			existing, found, err := s.Content.Get(ctx, entry.ID)
			if err != nil {
				return err
			}
			if !found || existing.Kind != kind {
				return errNotFound("content not found")
			}
			if !capcheck.CanEdit(principal, existing.AuthorID) {
				return errNotFound("forbidden")
			}
			updated := mergeContentEntry(existing, entry)
			if r.PostForm.Get("status") == "" {
				updated.Status = existing.Status
			}
			if r.PostForm.Get("visibility") == "" {
				updated.Visibility = existing.Visibility
			}
			if err := s.Content.Update(ctx, updated); err != nil {
				return err
			}
			if updated.Status == domaincontent.StatusPublished && existing.Status != domaincontent.StatusPublished {
				_, err = s.Content.Publish(ctx, entry.ID)
			}
			return err
		})
		if err != nil {
			status := http.StatusBadRequest
			if isNotFound(err) {
				status = http.StatusNotFound
			}
			http.Error(w, err.Error(), status)
			return
		}
		h.recordAudit(principal, operations.ActionAdminContentUpdate, string(kind), string(entry.ID), nil)
		http.Redirect(w, r, listPath, http.StatusSeeOther)
	}
}

func (h Handler) contentTrash(kind domaincontent.Kind) http.HandlerFunc {
	listPath := "/go-admin/posts"
	if kind == domaincontent.KindPage {
		listPath = "/go-admin/pages"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := h.requirePrincipal(w, r)
		if !ok {
			return
		}
		if !capcheck.CanDelete(principal) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !h.validateToken(w, r, ActionContentWrite) {
			return
		}
		id := domaincontent.ID(r.PathValue("id"))
		err := h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
			entry, found, err := s.Content.Get(ctx, id)
			if err != nil {
				return err
			}
			if !found || entry.Kind != kind {
				return errNotFound("content not found")
			}
			_, err = s.Content.Trash(ctx, id)
			return err
		})
		if err != nil {
			status := http.StatusBadRequest
			if isNotFound(err) {
				status = http.StatusNotFound
			}
			http.Error(w, err.Error(), status)
			return
		}
		h.recordAudit(principal, operations.ActionAdminContentTrash, string(kind), string(id), nil)
		http.Redirect(w, r, listPath, http.StatusSeeOther)
	}
}

func (h Handler) taxonomyCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	if !requireCapability(w, principal, modulecms.CapabilityTaxonomyManage) {
		return
	}
	if !h.validateToken(w, r, ActionContentWrite) {
		return
	}
	definition, err := bindTaxonomy(r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
		return s.Taxonomy.Register(ctx, definition)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/go-admin/taxonomies", http.StatusSeeOther)
}

func (h Handler) termCreate(w http.ResponseWriter, r *http.Request) {
	h.termCreateWithType(w, r, r.PathValue("type"))
}

func (h Handler) termCreateFlat(w http.ResponseWriter, r *http.Request) {
	h.termCreateWithType(w, r, "")
}

func (h Handler) termCreateWithType(w http.ResponseWriter, r *http.Request, taxonomyType string) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	if !requireCapability(w, principal, modulecms.CapabilityTaxonomyAssign) {
		return
	}
	if !h.validateToken(w, r, ActionContentWrite) {
		return
	}
	if taxonomyType == "" {
		taxonomyType = r.PostForm.Get("taxonomy_type")
		if taxonomyType == "" {
			taxonomyType = "category"
		}
	}
	term, err := bindTerm(r.PostForm, taxonomyType, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
		return s.Taxonomy.CreateTerm(ctx, term)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	redirect := "/go-admin/terms"
	if taxonomyType != "" {
		redirect = "/go-admin/taxonomies/" + taxonomyType + "/terms"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (h Handler) termUpdate(w http.ResponseWriter, r *http.Request) {
	h.termUpdateWithType(w, r, r.PathValue("type"), r.PathValue("id"))
}

func (h Handler) termUpdateFlat(w http.ResponseWriter, r *http.Request) {
	h.termUpdateWithType(w, r, "", r.PathValue("id"))
}

func (h Handler) termUpdateWithType(w http.ResponseWriter, r *http.Request, taxonomyType, id string) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	if !requireCapability(w, principal, modulecms.CapabilityTaxonomyAssign) {
		return
	}
	if !h.validateToken(w, r, ActionContentWrite) {
		return
	}
	if taxonomyType == "" {
		taxonomyType = r.PostForm.Get("taxonomy_type")
		if taxonomyType == "" {
			taxonomyType = "category"
		}
	}
	term, err := bindTerm(r.PostForm, taxonomyType, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
		return s.Taxonomy.CreateTerm(ctx, term)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	redirect := "/go-admin/terms"
	if taxonomyType != "" {
		redirect = "/go-admin/taxonomies/" + taxonomyType + "/terms"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (h Handler) mediaSave(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	if !requireCapability(w, principal, modulecms.CapabilityMediaUpload) {
		return
	}
	if !h.validateToken(w, r, ActionMediaUpload) {
		return
	}
	asset, err := bindMediaAsset(r.PostForm, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
		return s.Media.SaveMetadata(ctx, asset)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/go-admin/media", http.StatusSeeOther)
}

func (h Handler) mediaUpdate(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	if !requireCapability(w, principal, modulecms.CapabilityMediaUpload) {
		return
	}
	if !h.validateToken(w, r, ActionMediaUpload) {
		return
	}
	asset, err := bindMediaAsset(r.PostForm, r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
		return s.Media.Update(ctx, asset)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/go-admin/media", http.StatusSeeOther)
}

func (h Handler) settingsSave(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	if !requireCapability(w, principal, modulecms.CapabilitySettingsManage) {
		return
	}
	if !h.validateToken(w, r, ActionSettingsWrite) {
		return
	}
	values := bindSettingValues(r.PostForm)
	if len(values) == 0 {
		http.Error(w, "no settings submitted", http.StatusBadRequest)
		return
	}
	err := h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
		for _, value := range values {
			if err := s.Settings.Save(ctx, value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/go-admin/settings", http.StatusSeeOther)
}

func (h Handler) menuSave(w http.ResponseWriter, r *http.Request) {
	h.menuSaveWithID(w, r, "")
}

func (h Handler) menuUpdate(w http.ResponseWriter, r *http.Request) {
	h.menuSaveWithID(w, r, r.PathValue("id"))
}

func (h Handler) menuSaveWithID(w http.ResponseWriter, r *http.Request, id string) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	if !requireCapability(w, principal, modulecms.CapabilitySettingsManage) {
		return
	}
	if !h.validateToken(w, r, ActionMenusWrite) {
		return
	}
	menu, err := bindMenu(r.PostForm, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
		return s.Menus.Save(ctx, menu)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/go-admin/menus", http.StatusSeeOther)
}

func (h Handler) contentTypeCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
		if !capcheck.CanCreate(principal) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !h.validateToken(w, r, ActionContentWrite) {
			return
		}
		item, err := bindContentType(r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
		return s.ContentTypes.Register(ctx, item)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/go-admin/content-types", http.StatusSeeOther)
}

func mergeContentEntry(existing, patch domaincontent.Entry) domaincontent.Entry {
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
	if patch.TermIDs != nil {
		existing.TermIDs = patch.TermIDs
	}
	return existing
}

type notFoundError string

func (e notFoundError) Error() string { return string(e) }

func errNotFound(message string) error { return notFoundError(message) }

func isNotFound(err error) bool {
	_, ok := err.(notFoundError)
	return ok
}
