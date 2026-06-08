package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/appschema"
	"github.com/fastygo/app-gocms/internal/storage"
	storagesqlite "github.com/fastygo/app-gocms/internal/storage/sqlite"
	"github.com/fastygo/platform/pkg/conformance/bffparity"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/render"
)

func TestBFFProofScreensMatchGoldenFixtures(t *testing.T) {
	handler := testHandler(t)
	registry, principal := testHydratedRegistry(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	cases := []struct {
		path    string
		bffPath string
		fixture string
	}{
		{
			path:    "/go-admin/posts",
			bffPath: "/bff/screens/go-admin/posts",
			fixture: "appcms-posts-table.json",
		},
		{
			path:    "/go-admin/posts/post-rest/edit",
			bffPath: "/bff/screens/go-admin/posts/post-rest/edit",
			fixture: "appcms-post-edit-form.json",
		},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			page, err := registry.Page(context.Background(), tc.path, principal, nil)
			if err != nil {
				t.Fatalf("resolve page: %v", err)
			}
			bffparity.AssertScreenMatchesFixture(t, tc.fixture, page.Screen)

			response := httptest.NewRecorder()
			request := testRequest(http.MethodGet, tc.bffPath)
			request.AddCookie(cookie)
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("%s returned %d: %s", tc.bffPath, response.Code, response.Body.String())
			}
			var httpScreen render.ScreenModel
			if err := json.Unmarshal(response.Body.Bytes(), &httpScreen); err != nil {
				t.Fatalf("decode bff screen: %v", err)
			}
			bffparity.AssertScreenMatchesFixture(t, tc.fixture, httpScreen)
			if !reflect.DeepEqual(bffparity.NormalizeScreen(page.Screen), bffparity.NormalizeScreen(httpScreen)) {
				t.Fatalf("in-process and HTTP screens differ after normalization:\n in-process: %#v\n http: %#v", page.Screen, httpScreen)
			}
			if tc.fixture == "appcms-post-edit-form.json" && httpScreen.Metadata["action_token"] == "" {
				t.Fatal("expected volatile action_token on HTTP form screen")
			}
		})
	}
}

func TestBFFErrorScreensReturnStructuredModels(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")

	unauth := httptest.NewRecorder()
	handler.ServeHTTP(unauth, testRequest(http.MethodGet, "/bff/screens/go-admin/posts"))
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", unauth.Code, unauth.Body.String())
	}
	var unauthScreen render.ScreenModel
	if err := json.Unmarshal(unauth.Body.Bytes(), &unauthScreen); err != nil {
		t.Fatalf("decode unauthorized screen: %v", err)
	}
	if unauthScreen.View != render.ViewError || unauthScreen.Error == nil || unauthScreen.Error.Code != render.ErrorCodeUnauthorized {
		t.Fatalf("expected structured unauthorized screen: %#v", unauthScreen)
	}

	notFound := httptest.NewRecorder()
	notFoundRequest := testRequest(http.MethodGet, "/bff/screens/go-admin/posts/missing-id/edit")
	notFoundRequest.AddCookie(cookie)
	handler.ServeHTTP(notFound, notFoundRequest)
	if notFound.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", notFound.Code, notFound.Body.String())
	}
	var notFoundScreen render.ScreenModel
	if err := json.Unmarshal(notFound.Body.Bytes(), &notFoundScreen); err != nil {
		t.Fatalf("decode not-found screen: %v", err)
	}
	if notFoundScreen.View != render.ViewError || notFoundScreen.Error == nil || notFoundScreen.Error.Code != render.ErrorCodeNotFound {
		t.Fatalf("expected structured not-found screen: %#v", notFoundScreen)
	}
}

func TestBFFEmptyPostsTableMatchesGoldenFixture(t *testing.T) {
	ctx := context.Background()
	store, err := storagesqlite.Open("file:appcms-empty-bff?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Init(ctx); err != nil {
		t.Fatalf("init store: %v", err)
	}
	provider := storage.NewProvider(store)
	registry, err := appschema.NewRegistryWithProvider(provider)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	authStore, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		t.Fatalf("seed auth store: %v", err)
	}
	principal, ok := appauthn.NewService(authStore).Principal(contracts.PrincipalID("admin"))
	if !ok {
		t.Fatal("expected admin principal")
	}
	page, err := registry.Page(ctx, "/go-admin/posts", principal, nil)
	if err != nil {
		t.Fatalf("resolve page: %v", err)
	}
	bffparity.AssertScreenMatchesFixture(t, "appcms-posts-table-empty.json", page.Screen)
}

func testHydratedRegistry(t *testing.T) (*appschema.Registry, appauthn.Principal) {
	t.Helper()
	ctx := context.Background()
	store, err := storagesqlite.Open("file:appcms-bff-parity?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Init(ctx); err != nil {
		t.Fatalf("init store: %v", err)
	}
	if err := storagesqlite.SeedMinimalSite(ctx, store, "root"); err != nil {
		t.Fatalf("seed store: %v", err)
	}
	provider := storage.NewProvider(store)
	registry, err := appschema.NewRegistryWithProvider(provider)
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	authStore, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		t.Fatalf("seed auth store: %v", err)
	}
	principal, ok := appauthn.NewService(authStore).Principal(contracts.PrincipalID("admin"))
	if !ok {
		t.Fatal("expected admin principal")
	}
	return registry, principal
}
