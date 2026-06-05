package admin

import (
	"context"
	"net/http"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	appmedia "github.com/fastygo/app-gocms/internal/application/media"
	appmenus "github.com/fastygo/app-gocms/internal/application/menus"
	appsettings "github.com/fastygo/app-gocms/internal/application/settings"
	apptaxonomy "github.com/fastygo/app-gocms/internal/application/taxonomy"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/storage"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
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
}

func NewHandler(provider storage.StoreProvider, workspace contracts.WorkspaceID, auth PrincipalResolver, tokens ActionTokenValidator) Handler {
	if workspace == "" {
		workspace = "root"
	}
	return Handler{Provider: provider, Workspace: workspace, Auth: auth, Tokens: tokens}
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
}

type services struct {
	content      appcontent.Service
	contentTypes appcontenttype.Service
	settings     appsettings.Service
	taxonomy     apptaxonomy.Service
	media        appmedia.Service
	menus        appmenus.Service
}

func (h Handler) withServices(r *http.Request, fn func(ctx context.Context, s services) error) error {
	return h.Provider.ForWorkspace(h.Workspace).WithinTx(r.Context(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		s := services{
			content:      appcontent.NewService(appRepos, appRepos),
			contentTypes: appcontenttype.NewService(appRepos),
			settings: appsettings.NewService(appRepos, appsettings.NewRegistry(
				settings.Definition{Key: "site.title", Group: "site", DefaultValue: "AppCMS", Public: true},
				settings.Definition{Key: "site.description", Group: "site", DefaultValue: "", Public: true},
			)),
			taxonomy: apptaxonomy.NewService(appRepos, appRepos),
			media:    appmedia.NewService(appRepos, appRepos),
			menus:    appmenus.NewService(appRepos),
		}
		return fn(ctx, s)
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
		if !requireCapability(w, principal, modulecms.CapabilityContentWrite) {
			return
		}
		if !h.validateToken(w, r, ActionContentWrite) {
			return
		}
		entry, err := bindContent(r.PostForm, kind, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = h.withServices(r, func(ctx context.Context, s services) error {
			created, err := s.content.CreateDraft(ctx, entry)
			if err != nil {
				return err
			}
			if entry.Status == domaincontent.StatusPublished {
				_, err = s.content.Publish(ctx, created.ID)
			}
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
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
		if !requireCapability(w, principal, modulecms.CapabilityContentWrite) {
			return
		}
		if !h.validateToken(w, r, ActionContentWrite) {
			return
		}
		entry, err := bindContent(r.PostForm, kind, domaincontent.ID(r.PathValue("id")))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = h.withServices(r, func(ctx context.Context, s services) error {
			existing, found, err := s.content.Get(ctx, entry.ID)
			if err != nil {
				return err
			}
			if !found || existing.Kind != kind {
				return errNotFound("content not found")
			}
			updated := mergeContentEntry(existing, entry)
			if r.PostForm.Get("status") == "" {
				updated.Status = existing.Status
			}
			if r.PostForm.Get("visibility") == "" {
				updated.Visibility = existing.Visibility
			}
			if err := s.content.Update(ctx, updated); err != nil {
				return err
			}
			if updated.Status == domaincontent.StatusPublished && existing.Status != domaincontent.StatusPublished {
				_, err = s.content.Publish(ctx, entry.ID)
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
		if !requireCapability(w, principal, modulecms.CapabilityContentWrite) {
			return
		}
		if !h.validateToken(w, r, ActionContentWrite) {
			return
		}
		id := domaincontent.ID(r.PathValue("id"))
		err := h.withServices(r, func(ctx context.Context, s services) error {
			entry, found, err := s.content.Get(ctx, id)
			if err != nil {
				return err
			}
			if !found || entry.Kind != kind {
				return errNotFound("content not found")
			}
			_, err = s.content.Trash(ctx, id)
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
	err = h.withServices(r, func(ctx context.Context, s services) error {
		return s.taxonomy.Register(ctx, definition)
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
	err = h.withServices(r, func(ctx context.Context, s services) error {
		return s.taxonomy.CreateTerm(ctx, term)
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
	err = h.withServices(r, func(ctx context.Context, s services) error {
		return s.taxonomy.CreateTerm(ctx, term)
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
	err = h.withServices(r, func(ctx context.Context, s services) error {
		return s.media.SaveMetadata(ctx, asset)
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
	err = h.withServices(r, func(ctx context.Context, s services) error {
		return s.media.Update(ctx, asset)
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
	err := h.withServices(r, func(ctx context.Context, s services) error {
		for _, value := range values {
			if err := s.settings.Save(ctx, value); err != nil {
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
	err = h.withServices(r, func(ctx context.Context, s services) error {
		return s.menus.Save(ctx, menu)
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
	if !requireCapability(w, principal, modulecms.CapabilityContentWrite) {
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
	err = h.withServices(r, func(ctx context.Context, s services) error {
		return s.contentTypes.Register(ctx, item)
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
