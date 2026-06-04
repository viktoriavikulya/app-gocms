package main

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastygo/app-gocms/internal/appschema"
	gocmsapp "github.com/fastygo/app-gocms/pkg/app"
)

func TestCodexRouteSurfaces(t *testing.T) {
	handler := testHandler(t)
	for _, path := range []string{"/", "/go-admin/", "/go-login", "/go-json/", "/go-json/go/v2/", "/go-admin/posts", "/go-admin/pages", "/go-admin/taxonomies", "/go-admin/media", "/go-admin/authors"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, testRequest(http.MethodGet, path))
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
}

func TestRESTContentCRUDRoundTrip(t *testing.T) {
	handler := testHandler(t)
	create := `{"id":"post-rest","title":{"en":"REST Post"},"slug":"rest-post","content":"Hello REST","visibility":"public","status":"published","author_id":"admin"}`
	response := httptest.NewRecorder()
	request := testRequest(http.MethodPost, "/go-json/go/v2/posts")
	request.Header.Set("Authorization", "Bearer admin-token")
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
	request.Header.Set("Authorization", "Bearer admin-token")
	request.Body = io.NopCloser(bytes.NewBufferString(`{"excerpt":"updated excerpt"}`))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("patch returned %d: %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	request = testRequest(http.MethodDelete, "/go-json/go/v2/posts/post-rest")
	request.Header.Set("Authorization", "Bearer admin-token")
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
	for _, path := range []string{"/go-admin/posts/new", "/go-admin/posts/post-rest/edit", "/go-admin/media/new", "/go-admin/settings/new"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, testRequest(http.MethodGet, path))
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
		}
		body := response.Body.String()
		if !strings.Contains(body, "<form") || !strings.Contains(body, "aria-label=") {
			t.Fatalf("%s should render accessible form screen: %s", path, body)
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-admin/missing"))
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing admin route returned %d", response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("admin 404 should be HTML, got %q", contentType)
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
	registry, err := appschema.NewRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	application, err := gocmsapp.NewApp(gocmsapp.Options{
		Addr:      "127.0.0.1:0",
		StaticDir: filepath.Join("..", "..", "web", "static"),
		Registry:  registry,
		Seed:      true,
	})
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	return application
}

func testRequest(method, path string) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("User-Agent", "app-gocms-test")
	return request
}
