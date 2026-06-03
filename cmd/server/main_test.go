package main

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastygo/app-gocms/internal/appschema"
)

func TestCodexRouteSurfaces(t *testing.T) {
	mux := testMux(t)
	for _, path := range []string{"/", "/go-admin/", "/go-login", "/go-json/", "/go-json/go/v2/", "/go-admin/posts", "/go-admin/pages", "/go-admin/taxonomies", "/go-admin/media", "/go-admin/authors"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
		}
	}
}

func TestLogoutRedirectsToLogin(t *testing.T) {
	mux := testMux(t)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/go-logout", nil))
	if response.Code != http.StatusSeeOther {
		t.Fatalf("expected logout redirect, got %d", response.Code)
	}
	if location := response.Header().Get("Location"); location != "/go-login" {
		t.Fatalf("expected logout location /go-login, got %q", location)
	}
}

func TestRESTDiscoveryAndEnvelopes(t *testing.T) {
	mux := testMux(t)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/go-json/go/v2/", nil))
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
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/go-json/go/v2/posts", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected posts list OK, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"pagination"`) {
		t.Fatalf("expected list pagination envelope, got %s", response.Body.String())
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

func testMux(t *testing.T) *http.ServeMux {
	t.Helper()
	registry, err := appschema.NewRegistry()
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	return newMux(filepath.Join("..", "..", "web", "static"), registry)
}
