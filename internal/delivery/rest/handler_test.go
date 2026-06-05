package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	appmedia "github.com/fastygo/app-gocms/internal/application/media"
	appmenus "github.com/fastygo/app-gocms/internal/application/menus"
	appsettings "github.com/fastygo/app-gocms/internal/application/settings"
	apptaxonomy "github.com/fastygo/app-gocms/internal/application/taxonomy"
	appusers "github.com/fastygo/app-gocms/internal/application/users"
	"github.com/fastygo/app-gocms/internal/delivery/rest"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/domain/taxonomy"
	"github.com/fastygo/app-gocms/internal/storage"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/app-gocms/pkg/module/codex"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestRESTDiscoveryShape(t *testing.T) {
	handler, _ := newTestHandler(t, nil)
	rec := request(handler, http.MethodGet, "/go-json/go/v2/", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("discovery status = %d body = %s", rec.Code, rec.Body.String())
	}
	var discovery codex.Discovery
	if err := json.Unmarshal(rec.Body.Bytes(), &discovery); err != nil {
		t.Fatal(err)
	}
	if discovery.Version != "2" {
		t.Fatalf("version = %q", discovery.Version)
	}
	if discovery.Routes["posts"] != "/go-json/go/v2/posts" || discovery.Routes["contentTypes"] == "" {
		t.Fatalf("routes = %#v", discovery.Routes)
	}
}

func TestContentDTOProjectionAndDraftHiding(t *testing.T) {
	handler, _ := newTestHandler(t, func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		content := appcontent.NewService(appRepos, appRepos)
		if _, err := content.CreateDraft(ctx, domaincontent.Entry{
			ID: "draft-post", Kind: domaincontent.KindPost, Title: map[string]string{"en": "Draft"}, Slug: "draft-post", AuthorID: "admin",
		}); err != nil {
			return err
		}
		published, err := content.CreateDraft(ctx, domaincontent.Entry{
			ID: "pub-post", Kind: domaincontent.KindPost, Title: map[string]string{"en": "Published"}, Slug: "pub-post", AuthorID: "admin",
			Metadata: map[string]any{"seo": "ok", "secret_private": "hidden"},
		})
		if err != nil {
			return err
		}
		if _, err := content.Publish(ctx, published.ID); err != nil {
			return err
		}
		return nil
	})

	rec := request(handler, http.MethodGet, "/go-json/go/v2/posts", "", "")
	var list struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Data) != 1 {
		t.Fatalf("public list should hide draft, got %#v", list.Data)
	}
	if list.Data[0]["taxonomy_ids"] == nil {
		t.Fatalf("expected taxonomy_ids field: %#v", list.Data[0])
	}
	metadata, _ := list.Data[0]["metadata"].(map[string]any)
	if _, ok := metadata["secret_private"]; ok {
		t.Fatalf("private metadata leaked: %#v", metadata)
	}

	rec = request(handler, http.MethodGet, "/go-json/go/v2/posts/draft-post", "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("draft get status = %d", rec.Code)
	}
}

func TestTaxonomyListAndPaths(t *testing.T) {
	handler, auth := newTestHandler(t, seedTaxonomy)
	admin := loginToken(t, auth, []contracts.CapabilityID{
		modulecms.CapabilityContentWrite,
		modulecms.CapabilityTaxonomyManage,
		modulecms.CapabilityTaxonomyAssign,
	})

	rec := request(handler, http.MethodGet, "/go-json/go/v2/taxonomies", "", "")
	var envelope struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data) == 0 {
		t.Fatalf("expected taxonomy definitions, got %s", rec.Body.String())
	}

	rec = request(handler, http.MethodGet, "/go-json/go/v2/taxonomies/category", "", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data) != 1 || envelope.Data[0]["slug"].(map[string]any)["en"] != "news" {
		t.Fatalf("terms list = %#v", envelope.Data)
	}

	rec = request(handler, http.MethodGet, "/go-json/go/v2/taxonomies/category/news", "", "")
	var term struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &term); err != nil {
		t.Fatal(err)
	}
	if term.Data["id"] != "news" {
		t.Fatalf("term detail = %#v", term.Data)
	}

	body := `{"id":"post-terms","title":{"en":"Terms"},"slug":"terms-post","author_id":"admin"}`
	rec = request(handler, http.MethodPost, "/go-json/go/v2/posts", body, admin)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create post status = %d %s", rec.Code, rec.Body.String())
	}
	assign := `{"term_ids":["news"]}`
	rec = request(handler, http.MethodPost, "/go-json/go/v2/content/post-terms/terms", assign, admin)
	if rec.Code != http.StatusOK {
		t.Fatalf("assign terms status = %d %s", rec.Code, rec.Body.String())
	}
	var content struct {
		Data struct {
			TaxonomyIDs []string `json:"taxonomy_ids"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &content); err != nil {
		t.Fatal(err)
	}
	if len(content.Data.TaxonomyIDs) != 1 || content.Data.TaxonomyIDs[0] != "news" {
		t.Fatalf("taxonomy_ids = %#v", content.Data.TaxonomyIDs)
	}
}

func TestMediaPatchAndFeatured(t *testing.T) {
	handler, auth := newTestHandler(t, seedTaxonomy)
	token := loginToken(t, auth, []contracts.CapabilityID{
		modulecms.CapabilityContentWrite,
		modulecms.CapabilityMediaUpload,
		modulecms.CapabilityMediaEdit,
	})

	createMedia := `{"id":"media-hero","title":"Hero","mime_type":"image/png"}`
	rec := request(handler, http.MethodPost, "/go-json/go/v2/media", createMedia, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create media status = %d %s", rec.Code, rec.Body.String())
	}

	patch := `{"alt_text":"Hero alt"}`
	rec = request(handler, http.MethodPatch, "/go-json/go/v2/media/media-hero", patch, token)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Hero alt") {
		t.Fatalf("patch media status = %d %s", rec.Code, rec.Body.String())
	}

	postBody := `{"id":"featured-post","title":{"en":"Featured"},"slug":"featured-post","author_id":"admin"}`
	rec = request(handler, http.MethodPost, "/go-json/go/v2/posts", postBody, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create post status = %d %s", rec.Code, rec.Body.String())
	}
	rec = request(handler, http.MethodPost, "/go-json/go/v2/media/media-hero/featured/featured-post", "", token)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "media-hero") {
		t.Fatalf("featured media status = %d %s", rec.Code, rec.Body.String())
	}
}

func TestSettingsDTOStripsPrivate(t *testing.T) {
	handler, _ := newTestHandler(t, func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		service := appsettings.NewService(appRepos, appsettings.NewRegistry())
		if err := service.Save(ctx, settings.Value{Key: "site.title", Value: "AppCMS", Public: true}); err != nil {
			return err
		}
		return service.Save(ctx, settings.Value{Key: "site.secret_private", Value: "hidden", Public: false, Private: true})
	})
	rec := request(handler, http.MethodGet, "/go-json/go/v2/settings", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("settings status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret_private") {
		t.Fatalf("private setting leaked: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "site.title") {
		t.Fatalf("public setting missing: %s", rec.Body.String())
	}
}

func newTestHandler(t *testing.T, seed func(context.Context, storage.Repositories) error) (http.Handler, appauthn.Service) {
	t.Helper()
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	store, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	auth := appauthn.NewService(store)
	authorizer := func(r *http.Request, capability contracts.CapabilityID) (bool, bool) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return false, false
		}
		principal, err := auth.AuthenticateAppToken(r.Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return false, false
		}
		return true, principal.Has(capability)
	}
	handler := rest.NewHandler(provider, "root", authorizer)
	mux := http.NewServeMux()
	handler.Register(mux)
	if seed != nil {
		if err := provider.ForWorkspace("root").WithinTx(context.Background(), seed); err != nil {
			t.Fatal(err)
		}
	}
	return mux, auth
}

func seedTaxonomy(ctx context.Context, repos storage.Repositories) error {
	appRepos := storage.NewApplicationRepositories(repos)
	if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
		return err
	}
	taxonomyService := apptaxonomy.NewService(appRepos, appRepos)
	if err := taxonomyService.Register(ctx, taxonomy.Definition{Type: "category", Label: "Category", Mode: taxonomy.ModeFlat, Public: true}); err != nil {
		return err
	}
	return taxonomyService.CreateTerm(ctx, taxonomy.Term{ID: "news", TaxonomyType: "category", Name: "News", Slug: "news"})
}

func loginToken(t *testing.T, auth appauthn.Service, capabilities []contracts.CapabilityID) string {
	t.Helper()
	raw, _, err := auth.CreateAppToken(context.Background(), "admin", capabilities, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func request(handler http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// Ensure service packages are referenced for seed helpers.
var (
	_ = appmenus.NewService
	_ = appmedia.NewService
	_ = appusers.NewService
)
