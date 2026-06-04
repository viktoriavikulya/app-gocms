package main

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/appschema"
	gocmsapp "github.com/fastygo/app-gocms/pkg/app"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/contracts"
)

func TestCodexRouteSurfaces(t *testing.T) {
	handler := testHandler(t)
	for _, path := range []string{"/", "/go-login", "/go-json/", "/go-json/go/v2/"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, testRequest(http.MethodGet, path))
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
		}
	}
	cookie := loginCookie(t, handler, "admin", "admin")
	for _, path := range []string{"/go-admin/", "/go-admin/posts", "/go-admin/pages", "/go-admin/taxonomies", "/go-admin/media", "/go-admin/authors"} {
		response := httptest.NewRecorder()
		request := testRequest(http.MethodGet, path)
		request.AddCookie(cookie)
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
		}
	}
}

func TestLogoutRedirectsToLogin(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodPost, "/go-logout"))
	if response.Code != http.StatusSeeOther {
		t.Fatalf("expected logout redirect, got %d", response.Code)
	}
	if location := response.Header().Get("Location"); location != "/go-login" {
		t.Fatalf("expected logout location /go-login, got %q", location)
	}
}

func TestRESTDiscoveryAndEnvelopes(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/"))
	if response.Code != http.StatusOK {
		t.Fatalf("expected v2 discovery OK, got %d", response.Code)
	}
	var discovery struct {
		Routes []struct {
			Path string `json:"path"`
		} `json:"routes"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &discovery); err != nil {
		t.Fatalf("decode discovery: %v", err)
	}
	if len(discovery.Routes) == 0 {
		t.Fatalf("expected v2 routes")
	}

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts"))
	if response.Code != http.StatusOK {
		t.Fatalf("expected posts list OK, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"pagination"`) {
		t.Fatalf("expected list pagination envelope, got %s", response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts/by-slug/hello-world"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Hello world") {
		t.Fatalf("expected by-slug post envelope, got %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/pages/by-slug/about"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "About") {
		t.Fatalf("expected by-slug page envelope, got %d %s", response.Code, response.Body.String())
	}
}

func TestRESTContentCRUDRoundTrip(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	create := `{"id":"post-rest","title":{"en":"REST Post"},"slug":"rest-post","content":"Hello REST","visibility":"public","status":"published","author_id":"admin"}`
	response := httptest.NewRecorder()
	request := testRequest(http.MethodPost, "/go-json/go/v2/posts")
	request.AddCookie(cookie)
	request.Body = io.NopCloser(bytes.NewBufferString(create))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create returned %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts?per_page=1"))
	if response.Code != http.StatusOK {
		t.Fatalf("list returned %d: %s", response.Code, response.Body.String())
	}
	var list struct {
		Data       []map[string]any `json:"data"`
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if list.Pagination.Total == 0 {
		t.Fatalf("expected persisted post in list: %s", response.Body.String())
	}

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts/post-rest"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "REST Post") {
		t.Fatalf("get returned %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	request = testRequest(http.MethodPatch, "/go-json/go/v2/posts/post-rest")
	request.AddCookie(cookie)
	request.Body = io.NopCloser(bytes.NewBufferString(`{"excerpt":"updated excerpt"}`))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("patch returned %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	request = testRequest(http.MethodDelete, "/go-json/go/v2/posts/post-rest")
	request.AddCookie(cookie)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "trashed") {
		t.Fatalf("delete returned %d: %s", response.Code, response.Body.String())
	}
}

func TestRESTSearchAndErrorEnvelope(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/search?q=Hello"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"pagination"`) {
		t.Fatalf("search returned %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts?page=bad"))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid page returned %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"validation_error"`) {
		t.Fatalf("expected validation error envelope: %s", response.Body.String())
	}
}

func TestAdminFormScreensAndHTMLNotFound(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	for _, path := range []string{"/go-admin/posts/new", "/go-admin/posts/post-rest/edit", "/go-admin/media/new", "/go-admin/settings/new"} {
		response := httptest.NewRecorder()
		request := testRequest(http.MethodGet, path)
		request.AddCookie(cookie)
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
		}
		body := response.Body.String()
		if !strings.Contains(body, "<form") || !strings.Contains(body, "aria-label=") || !strings.Contains(body, `name="action_token"`) {
			t.Fatalf("%s should render accessible form screen: %s", path, body)
		}
	}
	response := httptest.NewRecorder()
	request := testRequest(http.MethodGet, "/go-admin/missing")
	request.AddCookie(cookie)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing admin route returned %d", response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("admin 404 should be HTML, got %q", contentType)
	}
}

func TestAuthSessionAndRESTAuthorization(t *testing.T) {
	handler, store := testHandlerWithAuthStore(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-admin/"))
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/go-login" {
		t.Fatalf("expected unauthenticated admin redirect, got %d %q", response.Code, response.Header().Get("Location"))
	}
	create := `{"id":"post-denied","title":{"en":"Denied"},"slug":"denied","content":"Denied"}`
	response = httptest.NewRecorder()
	request := testRequest(http.MethodPost, "/go-json/go/v2/posts")
	request.Body = io.NopCloser(bytes.NewBufferString(create))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated REST mutation 401, got %d: %s", response.Code, response.Body.String())
	}
	viewer := loginCookie(t, handler, "viewer", "viewer")
	response = httptest.NewRecorder()
	request = testRequest(http.MethodPost, "/go-json/go/v2/posts")
	request.AddCookie(viewer)
	request.Body = io.NopCloser(bytes.NewBufferString(create))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected viewer REST mutation 403, got %d: %s", response.Code, response.Body.String())
	}
	service := appauthn.NewService(store)
	raw, _, err := service.CreateAppToken(t.Context(), "editor", []contracts.CapabilityID{modulecms.CapabilityContentWrite}, time.Hour)
	if err != nil {
		t.Fatalf("create app token: %v", err)
	}
	response = httptest.NewRecorder()
	request = testRequest(http.MethodPost, "/go-json/go/v2/posts")
	request.Header.Set("Authorization", "Bearer "+raw)
	request.Body = io.NopCloser(bytes.NewBufferString(create))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected scoped app token create, got %d: %s", response.Code, response.Body.String())
	}
}

func TestLoginLockout(t *testing.T) {
	handler := testHandler(t)
	for i := 0; i < 3; i++ {
		response := postLogin(t, handler, "admin", "wrong")
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("bad login %d returned %d", i+1, response.Code)
		}
	}
	response := postLogin(t, handler, "admin", "admin")
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("expected lockout after repeated failures, got %d", response.Code)
	}
}

func TestPublicSiteRendersThemeAndSlugs(t *testing.T) {
	handler := testHandler(t)
	for _, tc := range []struct {
		path string
		want []string
		code int
	}{
		{path: "/", code: http.StatusOK, want: []string{`data-gocms-theme="gocms-default"`, `data-gocms-public-screen="home"`, "Hello world"}},
		{path: "/posts/hello-world", code: http.StatusOK, want: []string{`data-gocms-public-screen="post"`, "Hello world"}},
		{path: "/about", code: http.StatusOK, want: []string{`data-gocms-public-screen="page"`, "About"}},
		{path: "/missing", code: http.StatusNotFound, want: []string{`data-gocms-public-screen="not_found"`, "Not found"}},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, testRequest(http.MethodGet, tc.path))
			if response.Code != tc.code {
				t.Fatalf("%s returned %d: %s", tc.path, response.Code, response.Body.String())
			}
			for _, want := range tc.want {
				if !strings.Contains(response.Body.String(), want) {
					t.Fatalf("%s missing %q in %s", tc.path, want, response.Body.String())
				}
			}
		})
	}
}

func TestHeadlessDisablesPublicButKeepsREST(t *testing.T) {
	handler := testHandlerOptions(t, gocmsapp.Options{Headless: true})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/"))
	if response.Code != http.StatusNotFound {
		t.Fatalf("headless public home returned %d", response.Code)
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts/by-slug/hello-world"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Hello world") {
		t.Fatalf("headless REST should remain available, got %d %s", response.Code, response.Body.String())
	}
}

func TestFrameworkHealthEndpoints(t *testing.T) {
	handler := testHandler(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, testRequest(http.MethodGet, path))
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
		}
	}
}

func TestGraphQLExtensionRouteAndVisibility(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-graphql?query="+url.QueryEscape(`{posts{id slug title}}`)))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"posts"`) || !strings.Contains(response.Body.String(), "Hello world") {
		t.Fatalf("expected public GraphQL posts, got %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	request := testRequest(http.MethodPost, "/go-graphql")
	request.Body = io.NopCloser(strings.NewReader(`{"query":"mutation { createPost { id } }"}`))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("GraphQL mutations should be disabled in baseline, got %d %s", response.Code, response.Body.String())
	}
}

func TestOperationsHealthAuditAndSnapshot(t *testing.T) {
	handler := testHandler(t)
	admin := loginCookie(t, handler, "admin", "admin")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/ops/health"))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected ops health auth, got %d", response.Code)
	}
	response = httptest.NewRecorder()
	request := testRequest(http.MethodGet, "/go-json/go/v2/ops/health")
	request.AddCookie(admin)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"storage"`) || !strings.Contains(response.Body.String(), `"plugin_runtime"`) {
		t.Fatalf("expected health checks, got %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	request = testRequest(http.MethodGet, "/go-admin/import-export")
	request.AddCookie(admin)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "gocms.snapshot.v1") {
		t.Fatalf("expected import/export admin screen, got %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	request = testRequest(http.MethodGet, "/go-json/go/v2/ops/snapshot")
	request.AddCookie(admin)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"version":"gocms.snapshot.v1"`) || !strings.Contains(response.Body.String(), "Hello world") {
		t.Fatalf("expected snapshot export, got %d %s", response.Code, response.Body.String())
	}
	snapshot := response.Body.String()
	response = httptest.NewRecorder()
	request = testRequest(http.MethodPost, "/go-json/go/v2/ops/snapshot")
	request.AddCookie(admin)
	request.Body = io.NopCloser(strings.NewReader(snapshot))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"imported"`) {
		t.Fatalf("expected snapshot import, got %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	request = testRequest(http.MethodGet, "/go-json/go/v2/ops/audit")
	request.AddCookie(admin)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "snapshot.import") {
		t.Fatalf("expected audit events, got %d %s", response.Code, response.Body.String())
	}
}

func TestRuntimeProfileModesGateSurfaces(t *testing.T) {
	adminHandler := testHandlerOptions(t, gocmsapp.Options{RuntimeMode: gocmsapp.RuntimeModeAdmin})
	response := httptest.NewRecorder()
	adminHandler.ServeHTTP(response, testRequest(http.MethodGet, "/"))
	if response.Code != http.StatusNotFound {
		t.Fatalf("admin mode should disable public home, got %d", response.Code)
	}
	response = httptest.NewRecorder()
	adminHandler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts/by-slug/hello-world"))
	if response.Code != http.StatusOK {
		t.Fatalf("admin mode should keep REST, got %d", response.Code)
	}
	conformanceHandler := testHandlerOptions(t, gocmsapp.Options{RuntimeMode: gocmsapp.RuntimeModeConformance})
	response = httptest.NewRecorder()
	conformanceHandler.ServeHTTP(response, testRequest(http.MethodGet, "/"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Hello world") {
		t.Fatalf("conformance mode should keep deterministic seeded public fixture, got %d %s", response.Code, response.Body.String())
	}
}

func TestAppCMSUsesModuleCMSDescriptors(t *testing.T) {
	registry, err := appschema.NewRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	ctx := appschema.Context(registry)
	if ctx == nil || len(ctx.Records) != 8 || len(ctx.Resources) != 8 {
		t.Fatalf("expected ModuleCMS records/resources, got records=%d resources=%d", len(ctx.Records), len(ctx.Resources))
	}
	if _, err := registry.Screen("/go-admin/posts"); err != nil {
		t.Fatalf("expected posts screen from ModuleCMS descriptors: %v", err)
	}
}

func TestAppCMSDoesNotImportUI8Kit(t *testing.T) {
	for _, path := range []string{
		filepath.Join("..", "..", "internal", "appschema", "screens.go"),
		filepath.Join("..", "..", "internal", "delivery", "publicsite", "handler.go"),
		filepath.Join("..", "..", "internal", "themes", "registry.go"),
		filepath.Join("..", "..", "pkg", "app", "app.go"),
		"main.go",
	} {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse imports for %s: %v", path, err)
		}
		for _, imported := range file.Imports {
			if strings.Contains(imported.Path.Value, "ui8kit") {
				t.Fatalf("%s must not import UI8Kit", path)
			}
		}
	}
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	handler, _ := testHandlerWithAuthStore(t)
	return handler
}

func testHandlerWithAuthStore(t *testing.T) (http.Handler, *appauthn.MemoryStore) {
	t.Helper()
	return testHandlerWithAuthStoreOptions(t, gocmsapp.Options{})
}

func testHandlerOptions(t *testing.T, options gocmsapp.Options) http.Handler {
	t.Helper()
	handler, _ := testHandlerWithAuthStoreOptions(t, options)
	return handler
}

func testHandlerWithAuthStoreOptions(t *testing.T, options gocmsapp.Options) (http.Handler, *appauthn.MemoryStore) {
	t.Helper()
	registry, err := appschema.NewRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	authStore, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		t.Fatalf("seed auth store: %v", err)
	}
	application, err := gocmsapp.NewApp(gocmsapp.Options{
		Addr:        "127.0.0.1:0",
		StaticDir:   filepath.Join("..", "..", "web", "static"),
		Registry:    registry,
		Seed:        true,
		AuthStore:   authStore,
		Headless:    options.Headless,
		RuntimeMode: options.RuntimeMode,
	})
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	return application, authStore
}

func testRequest(method, path string) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("User-Agent", "app-gocms-test")
	return request
}

func loginCookie(t *testing.T, handler http.Handler, identifier string, password string) *http.Cookie {
	t.Helper()
	response := postLogin(t, handler, identifier, password)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("login returned %d: %s", response.Code, response.Body.String())
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == "appcms_session" {
			return cookie
		}
	}
	t.Fatalf("login did not issue appcms_session cookie")
	return nil
}

func postLogin(t *testing.T, handler http.Handler, identifier string, password string) *httptest.ResponseRecorder {
	t.Helper()
	tokenResponse := httptest.NewRecorder()
	handler.ServeHTTP(tokenResponse, testRequest(http.MethodGet, "/go-login"))
	token := hiddenValue(tokenResponse.Body.String(), "action_token")
	form := url.Values{}
	form.Set("action_token", token)
	form.Set("identifier", identifier)
	form.Set("password", password)
	request := testRequest(http.MethodPost, "/go-login")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Body = io.NopCloser(strings.NewReader(form.Encode()))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func hiddenValue(body string, name string) string {
	needle := `name="` + name + `" value="`
	start := strings.Index(body, needle)
	if start < 0 {
		return ""
	}
	start += len(needle)
	end := strings.Index(body[start:], `"`)
	if end < 0 {
		return ""
	}
	return body[start : start+end]
}
