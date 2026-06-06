package main

import (
	"bytes"
	"context"
	"encoding/json"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/appschema"
	gocmsapp "github.com/fastygo/app-gocms/pkg/app"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/bff"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/render"
)

func TestCodexRouteSurfaces(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/"))
	if response.Code != http.StatusNotFound {
		t.Fatalf("ux-pilot default admin mode should not serve public home, got %d", response.Code)
	}
	for _, path := range []string{"/go-login", "/go-json/", "/go-json/go/v2/"} {
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
	posts := httptest.NewRecorder()
	postsRequest := testRequest(http.MethodGet, "/go-admin/posts")
	postsRequest.AddCookie(cookie)
	handler.ServeHTTP(posts, postsRequest)
	body := posts.Body.String()
	for _, want := range []string{"Posts", "Schema-driven Platform preview", "<table"} {
		if !strings.Contains(body, want) {
			t.Fatalf("/go-admin/posts missing %q in pilot screen: %s", want, body)
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
	handler := testHandlerOptions(t, gocmsapp.Options{RuntimeMode: gocmsapp.RuntimeModeFull})
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
	handler := testHandlerOptions(t, gocmsapp.Options{RuntimeMode: gocmsapp.RuntimeModeFull})
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
	handler := testHandlerOptions(t, gocmsapp.Options{RuntimeMode: gocmsapp.RuntimeModeFull})
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
	adminHandler := testHandler(t)
	response := httptest.NewRecorder()
	adminHandler.ServeHTTP(response, testRequest(http.MethodGet, "/"))
	if response.Code != http.StatusNotFound {
		t.Fatalf("default ux-pilot admin mode should disable public home, got %d", response.Code)
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

func TestBFFScreenJSONMatchesInProcessModel(t *testing.T) {
	handler, store := testHandlerWithAuthStore(t)
	registry, err := appschema.NewRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	principal, ok := appauthn.NewService(store).Principal(contracts.PrincipalID("admin"))
	if !ok {
		t.Fatal("expected admin principal")
	}
	expected, err := registry.Page(context.Background(), "/go-admin/posts", principal)
	if err != nil {
		t.Fatalf("resolve in-process page: %v", err)
	}

	cookie := loginCookie(t, handler, "admin", "admin")
	response := httptest.NewRecorder()
	request := testRequest(http.MethodGet, "/bff/screens/go-admin/posts")
	request.AddCookie(cookie)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("/bff/screens/go-admin/posts returned %d: %s", response.Code, response.Body.String())
	}
	var got render.ScreenModel
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode bff screen: %v", err)
	}
	if !reflect.DeepEqual(got, expected.Screen) {
		t.Fatalf("json screen mismatch:\n got: %#v\nwant: %#v", got, expected.Screen)
	}
}

func TestBFFNavAndSessionJSON(t *testing.T) {
	handler, store := testHandlerWithAuthStore(t)
	registry, err := appschema.NewRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	principal, ok := appauthn.NewService(store).Principal(contracts.PrincipalID("admin"))
	if !ok {
		t.Fatal("expected admin principal")
	}
	expectedPage, err := registry.DashboardPage(context.Background(), principal)
	if err != nil {
		t.Fatalf("resolve dashboard page: %v", err)
	}

	cookie := loginCookie(t, handler, "admin", "admin")
	navResponse := httptest.NewRecorder()
	navRequest := testRequest(http.MethodGet, "/bff/nav")
	navRequest.AddCookie(cookie)
	handler.ServeHTTP(navResponse, navRequest)
	if navResponse.Code != http.StatusOK {
		t.Fatalf("/bff/nav returned %d: %s", navResponse.Code, navResponse.Body.String())
	}
	var gotNav bff.NavigationModel
	if err := json.Unmarshal(navResponse.Body.Bytes(), &gotNav); err != nil {
		t.Fatalf("decode bff nav: %v", err)
	}
	if !reflect.DeepEqual(gotNav, expectedPage.Navigation) {
		t.Fatalf("json nav mismatch:\n got: %#v\nwant: %#v", gotNav, expectedPage.Navigation)
	}

	sessionResponse := httptest.NewRecorder()
	sessionRequest := testRequest(http.MethodGet, "/bff/session")
	sessionRequest.AddCookie(cookie)
	handler.ServeHTTP(sessionResponse, sessionRequest)
	if sessionResponse.Code != http.StatusOK {
		t.Fatalf("/bff/session returned %d: %s", sessionResponse.Code, sessionResponse.Body.String())
	}
	var gotSession bff.SessionModel
	if err := json.Unmarshal(sessionResponse.Body.Bytes(), &gotSession); err != nil {
		t.Fatalf("decode bff session: %v", err)
	}
	if !gotSession.Authenticated || gotSession.PrincipalID != "admin" || gotSession.ProfileID != "gocms-admin" {
		t.Fatalf("unexpected session payload: %#v", gotSession)
	}
	if !principal.Has(modulecms.CapabilityAdminAccess) {
		t.Fatal("expected admin capability on principal")
	}

	unauth := httptest.NewRecorder()
	handler.ServeHTTP(unauth, testRequest(http.MethodGet, "/bff/screens/go-admin/posts"))
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated bff 401, got %d", unauth.Code)
	}
}

func TestBFFFormScreenIncludesActionToken(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	response := httptest.NewRecorder()
	request := testRequest(http.MethodGet, "/bff/screens/go-admin/posts/post-rest/edit")
	request.AddCookie(cookie)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("form bff returned %d: %s", response.Code, response.Body.String())
	}
	var screen render.ScreenModel
	if err := json.Unmarshal(response.Body.Bytes(), &screen); err != nil {
		t.Fatalf("decode form screen: %v", err)
	}
	if screen.View != render.ViewForm {
		t.Fatalf("expected form view, got %q", screen.View)
	}
	if screen.Metadata["action_token"] == "" {
		t.Fatalf("expected action_token in BFF form metadata for React/Templ parity: %#v", screen.Metadata)
	}
	if screen.Metadata["form_action"] == "" {
		t.Fatal("expected form_action metadata")
	}
	if !strings.HasPrefix(screen.Metadata["form_action"], "/bff/actions/") {
		t.Fatalf("expected BFF form action, got %q", screen.Metadata["form_action"])
	}
}

func TestBFFPostActionsCreateUpdateTrash(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	token := bffFormActionToken(t, handler, cookie, "/bff/screens/go-admin/posts/new")

	create := url.Values{}
	create.Set("action_token", token)
	create.Set("id", "post-bff")
	create.Set("title", "BFF Post")
	create.Set("slug", "bff-post")
	create.Set("content", "Hello BFF")
	create.Set("status", "draft")
	create.Set("visibility", "public")
	create.Set("author_id", "admin")
	createResponse := httptest.NewRecorder()
	createRequest := testRequest(http.MethodPost, "/bff/actions/post.create")
	createRequest.AddCookie(cookie)
	createRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createRequest.Body = io.NopCloser(strings.NewReader(create.Encode()))
	handler.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusOK {
		t.Fatalf("create action returned %d: %s", createResponse.Code, createResponse.Body.String())
	}
	var createResult bff.ActionResult
	if err := json.Unmarshal(createResponse.Body.Bytes(), &createResult); err != nil || !createResult.OK {
		t.Fatalf("create action result: %#v err=%v body=%s", createResult, err, createResponse.Body.String())
	}

	updateToken := bffFormActionToken(t, handler, cookie, "/bff/screens/go-admin/posts/post-bff/edit")
	update := url.Values{}
	update.Set("action_token", updateToken)
	update.Set("excerpt", "updated via BFF")
	updateResponse := httptest.NewRecorder()
	updateRequest := testRequest(http.MethodPost, "/bff/actions/post.update?id=post-bff")
	updateRequest.AddCookie(cookie)
	updateRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateRequest.Body = io.NopCloser(strings.NewReader(update.Encode()))
	handler.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("update action returned %d: %s", updateResponse.Code, updateResponse.Body.String())
	}

	trashToken := updateToken
	trash := url.Values{}
	trash.Set("action_token", trashToken)
	trashResponse := httptest.NewRecorder()
	trashRequest := testRequest(http.MethodPost, "/bff/actions/post.trash?id=post-bff")
	trashRequest.AddCookie(cookie)
	trashRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	trashRequest.Body = io.NopCloser(strings.NewReader(trash.Encode()))
	handler.ServeHTTP(trashResponse, trashRequest)
	if trashResponse.Code != http.StatusOK {
		t.Fatalf("trash action returned %d: %s", trashResponse.Code, trashResponse.Body.String())
	}
}

func TestBFFPostActionsEnforceAuthCapabilityAndToken(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	token := bffFormActionToken(t, handler, cookie, "/bff/screens/go-admin/posts/new")

	unauth := httptest.NewRecorder()
	handler.ServeHTTP(unauth, testRequest(http.MethodPost, "/bff/actions/post.create"))
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauth.Code)
	}

	viewerCookie := loginCookie(t, handler, "viewer", "viewer")
	viewerBody := url.Values{}
	viewerBody.Set("action_token", token)
	viewerBody.Set("id", "post-viewer")
	viewerBody.Set("status", "draft")
	viewerBody.Set("visibility", "public")
	viewerRequest := testRequest(http.MethodPost, "/bff/actions/post.create")
	viewerRequest.AddCookie(viewerCookie)
	viewerRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	viewerRequest.Body = io.NopCloser(strings.NewReader(viewerBody.Encode()))
	viewerResponse := httptest.NewRecorder()
	handler.ServeHTTP(viewerResponse, viewerRequest)
	if viewerResponse.Code != http.StatusForbidden {
		t.Fatalf("expected viewer forbidden 403, got %d: %s", viewerResponse.Code, viewerResponse.Body.String())
	}

	badToken := url.Values{}
	badToken.Set("action_token", "invalid")
	badToken.Set("id", "post-bad")
	badToken.Set("status", "draft")
	badToken.Set("visibility", "public")
	badRequest := testRequest(http.MethodPost, "/bff/actions/post.create")
	badRequest.AddCookie(cookie)
	badRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	badRequest.Body = io.NopCloser(strings.NewReader(badToken.Encode()))
	badResponse := httptest.NewRecorder()
	handler.ServeHTTP(badResponse, badRequest)
	if badResponse.Code != http.StatusForbidden {
		t.Fatalf("expected invalid token 403, got %d", badResponse.Code)
	}

	invalid := url.Values{}
	invalid.Set("action_token", token)
	invalidRequest := testRequest(http.MethodPost, "/bff/actions/post.create")
	invalidRequest.AddCookie(cookie)
	invalidRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	invalidRequest.Body = io.NopCloser(strings.NewReader(invalid.Encode()))
	invalidResponse := httptest.NewRecorder()
	handler.ServeHTTP(invalidResponse, invalidRequest)
	if invalidResponse.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected validation 422, got %d: %s", invalidResponse.Code, invalidResponse.Body.String())
	}
	if !strings.Contains(invalidResponse.Body.String(), "validation") {
		t.Fatalf("expected validation model: %s", invalidResponse.Body.String())
	}
}

func bffFormActionToken(t *testing.T, handler http.Handler, cookie *http.Cookie, screenPath string) string {
	t.Helper()
	response := httptest.NewRecorder()
	request := testRequest(http.MethodGet, screenPath)
	request.AddCookie(cookie)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("%s returned %d: %s", screenPath, response.Code, response.Body.String())
	}
	var screen render.ScreenModel
	if err := json.Unmarshal(response.Body.Bytes(), &screen); err != nil {
		t.Fatalf("decode screen: %v", err)
	}
	if screen.Metadata["action_token"] == "" {
		t.Fatalf("missing action token: %#v", screen.Metadata)
	}
	return screen.Metadata["action_token"]
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
