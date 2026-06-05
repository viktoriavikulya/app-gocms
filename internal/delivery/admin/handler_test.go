package admin_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/appschema"
	gocmsapp "github.com/fastygo/app-gocms/pkg/app"
)

func TestAdminPostCreatePost(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	response := adminFormPost(t, handler, cookie, "/go-admin/posts/new", "/go-admin/posts", url.Values{
		"id":       {"post-admin-create"},
		"title":    {"Admin Post"},
		"slug":     {"admin-post"},
		"content":  {"Created via admin POST"},
		"status":   {"published"},
		"author_id": {"admin"},
	})
	if response.Code != http.StatusSeeOther {
		t.Fatalf("create returned %d: %s", response.Code, response.Body.String())
	}
	if location := response.Header().Get("Location"); location != "/go-admin/posts" {
		t.Fatalf("expected redirect to list, got %q", location)
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts/post-admin-create"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Admin Post") {
		t.Fatalf("REST get returned %d: %s", response.Code, response.Body.String())
	}
}

func TestAdminPostCreatePage(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	response := adminFormPost(t, handler, cookie, "/go-admin/pages/new", "/go-admin/pages", url.Values{
		"id":      {"page-admin-create"},
		"title":   {"Admin Page"},
		"slug":    {"admin-page"},
		"content": {"Page body"},
		"status":  {"draft"},
	})
	if response.Code != http.StatusSeeOther {
		t.Fatalf("create returned %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/pages/page-admin-create"))
	if response.Code != http.StatusNotFound {
		t.Fatalf("draft page should be hidden from public REST, got %d", response.Code)
	}
}

func TestAdminPostUpdatePost(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	adminFormPost(t, handler, cookie, "/go-admin/posts/new", "/go-admin/posts", url.Values{
		"id":     {"post-admin-update"},
		"title":  {"Before"},
		"slug":   {"before-update"},
		"status": {"published"},
	})
	response := adminFormPost(t, handler, cookie, "/go-admin/posts/post-admin-update/edit", "/go-admin/posts/post-admin-update", url.Values{
		"title": {"After Update"},
	})
	if response.Code != http.StatusSeeOther {
		t.Fatalf("update returned %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts/post-admin-update"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "After Update") {
		t.Fatalf("updated post not visible: %d %s", response.Code, response.Body.String())
	}
}

func TestAdminPostTrashPost(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	adminFormPost(t, handler, cookie, "/go-admin/posts/new", "/go-admin/posts", url.Values{
		"id":     {"post-admin-trash"},
		"title":  {"Trash Me"},
		"slug":   {"trash-me"},
		"status": {"published"},
	})
	token := formToken(t, handler, cookie, "/go-admin/posts/post-admin-trash/edit")
	form := url.Values{}
	form.Set("action_token", token)
	response := httptest.NewRecorder()
	request := testRequest(http.MethodPost, "/go-admin/posts/post-admin-trash/trash")
	request.AddCookie(cookie)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Body = io.NopCloser(strings.NewReader(form.Encode()))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("trash returned %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/posts/post-admin-trash"))
	if response.Code != http.StatusNotFound {
		t.Fatalf("trashed post should be hidden, got %d", response.Code)
	}
}

func TestAdminPostCreateTaxonomy(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	response := adminFormPost(t, handler, cookie, "/go-admin/taxonomies/new", "/go-admin/taxonomies", url.Values{
		"type":  {"tag"},
		"label": {"Tags"},
		"mode":  {"flat"},
	})
	if response.Code != http.StatusSeeOther {
		t.Fatalf("taxonomy create returned %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/taxonomies"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"tag"`) {
		t.Fatalf("taxonomy not listed: %d %s", response.Code, response.Body.String())
	}
}

func TestAdminPostCreateTerm(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	response := adminFormPost(t, handler, cookie, "/go-admin/taxonomies/category/terms/new", "/go-admin/taxonomies/category/terms", url.Values{
		"id":   {"term-admin"},
		"name": {"Admin Term"},
		"slug": {"admin-term"},
	})
	if response.Code != http.StatusSeeOther {
		t.Fatalf("term create returned %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/taxonomies/category/term-admin"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Admin Term") {
		t.Fatalf("term not found: %d %s", response.Code, response.Body.String())
	}
}

func TestAdminPostSaveMedia(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	response := adminFormPost(t, handler, cookie, "/go-admin/media/new", "/go-admin/media", url.Values{
		"id":        {"media-admin"},
		"filename":  {"cover.jpg"},
		"mime_type": {"image/jpeg"},
		"alt_text":  {"Cover image"},
		"public_url": {"https://example.test/cover.jpg"},
	})
	if response.Code != http.StatusSeeOther {
		t.Fatalf("media save returned %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/media/media-admin"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Cover image") {
		t.Fatalf("media not found: %d %s", response.Code, response.Body.String())
	}
}

func TestAdminPostSaveSettings(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	response := adminFormPost(t, handler, cookie, "/go-admin/settings/new", "/go-admin/settings", url.Values{
		"site.title": {"Admin Settings Title"},
	})
	if response.Code != http.StatusSeeOther {
		t.Fatalf("settings save returned %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/settings"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Admin Settings Title") {
		t.Fatalf("settings not updated: %d %s", response.Code, response.Body.String())
	}
}

func TestAdminPostSaveMenu(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	response := adminFormPost(t, handler, cookie, "/go-admin/menus/new", "/go-admin/menus", url.Values{
		"id":       {"menu-admin"},
		"location": {"footer"},
		"items":    {`[{"id":"home","label":"Home","url":"/"}]`},
	})
	if response.Code != http.StatusSeeOther {
		t.Fatalf("menu save returned %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/menus/footer"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "footer") {
		t.Fatalf("menu not found: %d %s", response.Code, response.Body.String())
	}
}

func TestAdminPostCreateContentType(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	response := adminFormPost(t, handler, cookie, "/go-admin/content-types/new", "/go-admin/content-types", url.Values{
		"id":    {"event"},
		"label": {"Event"},
	})
	if response.Code != http.StatusSeeOther {
		t.Fatalf("content type create returned %d: %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, testRequest(http.MethodGet, "/go-json/go/v2/content-types"))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"event"`) {
		t.Fatalf("content type not listed: %d %s", response.Code, response.Body.String())
	}
}

func TestAdminPostSecurity(t *testing.T) {
	handler := testHandler(t)
	cookie := loginCookie(t, handler, "admin", "admin")
	viewer := loginCookie(t, handler, "viewer", "viewer")

	t.Run("missing token", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := testRequest(http.MethodPost, "/go-admin/posts")
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Body = io.NopCloser(strings.NewReader("id=x&title=y&slug=z"))
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", response.Code)
		}
	})

	t.Run("wrong scope", func(t *testing.T) {
		token := formToken(t, handler, cookie, "/go-login")
		form := url.Values{"action_token": {token}, "id": {"bad"}, "title": {"Bad"}, "slug": {"bad"}}
		response := httptest.NewRecorder()
		request := testRequest(http.MethodPost, "/go-admin/posts")
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Body = io.NopCloser(strings.NewReader(form.Encode()))
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", response.Code)
		}
	})

	t.Run("anonymous", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := testRequest(http.MethodPost, "/go-admin/posts")
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Body = io.NopCloser(strings.NewReader("action_token=x&id=y"))
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/go-login" {
			t.Fatalf("expected redirect to login, got %d location=%q", response.Code, response.Header().Get("Location"))
		}
	})

	t.Run("insufficient capability", func(t *testing.T) {
		response := adminFormPost(t, handler, viewer, "/go-admin/posts/new", "/go-admin/posts", url.Values{
			"id": {"denied"}, "title": {"Denied"}, "slug": {"denied"},
		})
		if response.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", response.Code, response.Body.String())
		}
	})
}

func adminFormPost(t *testing.T, handler http.Handler, cookie *http.Cookie, getPath, postPath string, fields url.Values) *httptest.ResponseRecorder {
	t.Helper()
	token := formToken(t, handler, cookie, getPath)
	fields.Set("action_token", token)
	response := httptest.NewRecorder()
	request := testRequest(http.MethodPost, postPath)
	request.AddCookie(cookie)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Body = io.NopCloser(strings.NewReader(fields.Encode()))
	handler.ServeHTTP(response, request)
	return response
}

func formToken(t *testing.T, handler http.Handler, cookie *http.Cookie, path string) string {
	t.Helper()
	response := httptest.NewRecorder()
	request := testRequest(http.MethodGet, path)
	request.AddCookie(cookie)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("GET %s returned %d: %s", path, response.Code, response.Body.String())
	}
	token := hiddenValue(response.Body.String(), "action_token")
	if token == "" {
		t.Fatalf("missing action_token on %s", path)
	}
	return token
}

func testHandler(t *testing.T) http.Handler {
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
		Addr:      "127.0.0.1:0",
		StaticDir: filepath.Join("..", "..", "web", "static"),
		Registry:  registry,
		Seed:      true,
		AuthStore: authStore,
	})
	if err != nil {
		t.Fatalf("build app: %v", err)
	}
	return application
}

func testRequest(method, path string) *http.Request {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("User-Agent", "app-gocms-admin-test")
	return request
}

func loginCookie(t *testing.T, handler http.Handler, identifier, password string) *http.Cookie {
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

func hiddenValue(body, name string) string {
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
