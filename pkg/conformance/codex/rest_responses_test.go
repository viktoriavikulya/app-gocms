package codex_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	gocmsapp "github.com/fastygo/app-gocms/pkg/app"
)

func TestRESTResponsesValidateAgainstCodexSchemas(t *testing.T) {
	handler := testRESTHandler(t)

	envelopeSchema := mustLoadSchema(t, "envelope.schema.json")
	entrySchema := mustLoadSchema(t, "content-entry.schema.json")
	menuSchema := mustLoadSchema(t, "menu.schema.json")
	errorSchema := mustLoadSchema(t, "error.schema.json")

	t.Run("discovery", func(t *testing.T) {
		body := getJSON(t, handler, "/go-json/go/v2/")
		mustValidateDecoded(t, envelopeSchema, body)
	})

	t.Run("posts list", func(t *testing.T) {
		body := getJSON(t, handler, "/go-json/go/v2/posts")
		mustValidateDecoded(t, envelopeSchema, body)
		data, ok := body["data"].([]any)
		if !ok {
			t.Fatalf("posts list data must be array, got %T", body["data"])
		}
		for i, item := range data {
			mustValidateDecoded(t, entrySchema, item)
			if i == 0 {
				break
			}
		}
	})

	t.Run("post by slug", func(t *testing.T) {
		body := getJSON(t, handler, "/go-json/go/v2/posts/by-slug/hello-world")
		mustValidateDecoded(t, envelopeSchema, body)
		mustValidateDecoded(t, entrySchema, body["data"])
	})

	t.Run("page by slug", func(t *testing.T) {
		body := getJSON(t, handler, "/go-json/go/v2/pages/by-slug/about")
		mustValidateDecoded(t, envelopeSchema, body)
		mustValidateDecoded(t, entrySchema, body["data"])
	})

	t.Run("menu by location", func(t *testing.T) {
		body := getJSON(t, handler, "/go-json/go/v2/menus/primary")
		mustValidateDecoded(t, envelopeSchema, body)
		mustValidateDecoded(t, menuSchema, body["data"])
	})

	t.Run("not found error", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/go-json/go/v2/unknown-route", nil)
		request.Header.Set("User-Agent", "app-gocms-test")
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", response.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode error body: %v", err)
		}
		mustValidateDecoded(t, errorSchema, body)
	})
}

func TestGraphQLPostsOutputDocumentsCodexGap(t *testing.T) {
	handler := testRESTHandler(t)
	entrySchema := mustLoadSchema(t, "content-entry.schema.json")

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/go-graphql?query=posts", nil)
	request.Header.Set("User-Agent", "app-gocms-test")
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected GraphQL OK, got %d: %s", response.Code, response.Body.String())
	}

	var payload struct {
		Data struct {
			Posts []map[string]any `json:"posts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode GraphQL payload: %v", err)
	}
	if len(payload.Data.Posts) == 0 {
		t.Fatal("expected seeded GraphQL posts")
	}

	first := payload.Data.Posts[0]
	if err := entrySchema.Validate(first); err == nil {
		t.Fatalf("expected GraphQL storage records to fail content-entry schema until mapped to domain Entry")
	}
	if first["kind"] != nil {
		t.Fatalf("expected storage record without kind field, got %#v", first["kind"])
	}
}

func testRESTHandler(t *testing.T) http.Handler {
	t.Helper()
	authStore, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		t.Fatalf("seed auth store: %v", err)
	}
	application, err := gocmsapp.NewApp(gocmsapp.Options{
		Addr:        "127.0.0.1:0",
		StaticDir:   filepath.Join("..", "..", "..", "..", "web", "static"),
		StorageDSN:  fmt.Sprintf("file:appcms-codex-test-%d?mode=memory&cache=private", time.Now().UnixNano()),
		Seed:        true,
		AuthStore:   authStore,
		RuntimeMode: gocmsapp.RuntimeModeFull,
	})
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	return application
}

func getJSON(t *testing.T, handler http.Handler, path string) map[string]any {
	t.Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("User-Agent", "app-gocms-test")
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return body
}

func mustValidateDecoded(t *testing.T, schema *jsonschema.Schema, value any) {
	t.Helper()
	if err := schema.Validate(value); err != nil {
		t.Fatalf("schema validation failed: %v", err)
	}
}
