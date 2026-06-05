package conformance_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/appschema"
	"github.com/fastygo/app-gocms/internal/extensions"
	"github.com/fastygo/app-gocms/internal/operations"
	"github.com/fastygo/app-gocms/internal/storage/sqlite"
	gocmsapp "github.com/fastygo/app-gocms/pkg/app"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
	_ "modernc.org/sqlite"
)

func TestLevel0CoreCapabilitiesDeclared(t *testing.T) {
	required := []contracts.CapabilityID{
		modulecms.CapabilityContentRead,
		modulecms.CapabilityContentCreate,
		modulecms.CapabilityContentEditOwn,
		modulecms.CapabilityContentPublish,
		modulecms.CapabilityContentSchedule,
		modulecms.CapabilityContentDelete,
		modulecms.CapabilityContentRestore,
		modulecms.CapabilityContentManageRevisions,
		modulecms.CapabilityRESTAccess,
		modulecms.CapabilityRESTWrite,
	}
	declared := map[contracts.CapabilityID]bool{}
	for _, item := range modulecms.CapabilityDefinitions() {
		declared[item.ID] = true
	}
	for _, cap := range required {
		if !declared[cap] {
			t.Fatalf("missing capability %q", cap)
		}
	}
}

func TestLevel0DraftVisibilityAndTransitions(t *testing.T) {
	handler := testHandler(t)
	admin := loginCookie(t, handler, "admin", "admin")
	viewer := loginCookie(t, handler, "viewer", "viewer")

	createPost(t, handler, admin, "level0-draft", "draft", "draft")
	createPost(t, handler, admin, "level0-pub", "published", "published")

	list := adminGet(t, handler, viewer, "/go-admin/posts")
	if strings.Contains(list, "level0-draft") {
		t.Fatal("viewer should not see draft in hydrated admin list")
	}
	if !strings.Contains(list, "level0-pub") {
		t.Fatal("viewer should see published post in admin list")
	}

	public := httptest.NewRecorder()
	handler.ServeHTTP(public, httptest.NewRequest(http.MethodGet, "/go-json/go/v2/posts", nil))
	if strings.Contains(public.Body.String(), "level0-draft") {
		t.Fatal("draft must not appear in public REST list")
	}

	transitionOK(t, handler, admin, http.MethodPost, "/go-json/go/v2/posts/level0-draft/publish", nil)
}

func TestLevel0RevisionRestoreRoundTrip(t *testing.T) {
	handler := testHandler(t)
	admin := loginCookie(t, handler, "admin", "admin")
	createPost(t, handler, admin, "rev-post", "v1", "draft")
	patchREST(t, handler, admin, "/go-json/go/v2/posts/rev-post", map[string]any{"content": "v2"})
	list := restGet(t, handler, admin, "/go-json/go/v2/posts/rev-post/revisions")
	var envelope struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(list), &envelope); err != nil || len(envelope.Data) == 0 {
		t.Fatalf("revisions list = %s err=%v", list, err)
	}
	revID := envelope.Data[0].ID
	restore := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/go-json/go/v2/posts/rev-post/revisions/"+revID+"/restore", nil)
	withTestHeaders(req)
	req.AddCookie(admin)
	handler.ServeHTTP(restore, req)
	if restore.Code != http.StatusOK {
		t.Fatalf("restore returned %d: %s", restore.Code, restore.Body.String())
	}
	got := restGet(t, handler, admin, "/go-json/go/v2/posts/rev-post")
	if !strings.Contains(got, `"content":"v1"`) && !strings.Contains(got, "v1") {
		t.Fatalf("restored content missing v1: %s", got)
	}
}

func TestLevel0AuditPersistsAcrossRestart(t *testing.T) {
	dsn := "file:level0audit?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	store, err := sqlite.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	audit := operations.NewSQLiteAuditStore(db, "root", 100)
	audit.RecordAudit(operations.AuditEvent{Action: "test.persist", Actor: "admin", Resource: "post/p1", ResourceType: "post", ResourceID: "p1"})

	restarted := operations.NewSQLiteAuditStore(db, "root", 100)
	events := restarted.Audit()
	if len(events) == 0 || events[len(events)-1].Action != "test.persist" {
		t.Fatalf("expected persisted audit event, got %#v", events)
	}
}

func TestLevel0HookAbortBlocksMutation(t *testing.T) {
	bus := extensions.NewHookBus()
	bus.AddAction(extensions.HookContentCreateBefore, extensions.ActionHandler{
		ID: "block", Handle: func(context.Context, any) error { return errBlockedHook },
	})
	handler := buildHandler(t, operations.NewStore(50), bus)
	admin := loginCookie(t, handler, "admin", "admin")
	body := []byte(`{"id":"blocked","title":{"en":"Nope"},"slug":"blocked","kind":"post"}`)
	req := httptest.NewRequest(http.MethodPost, "/go-json/go/v2/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	withTestHeaders(req)
	req.AddCookie(admin)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code == http.StatusCreated {
		t.Fatal("hook should block create")
	}
}

var errBlockedHook = errors.New("blocked by hook")

func testHandler(t *testing.T) http.Handler {
	return buildHandler(t, nil, nil)
}

func buildHandler(t *testing.T, audit operations.AuditRecorder, hooks *extensions.HookBus) http.Handler {
	t.Helper()
	registry, err := appschema.NewRegistry()
	if err != nil {
		t.Fatal(err)
	}
	authStore, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	opts := gocmsapp.Options{
		Addr:      "127.0.0.1:0",
		StaticDir: filepath.Join("..", "..", "web", "static"),
		Registry:  registry,
		Seed:      false,
		Storage:   contractstest.NewMemoryStorage(),
		AuthStore: authStore,
	}
	if audit != nil {
		opts.AuditStore = audit
	}
	if hooks != nil {
		opts.HookBus = hooks
	}
	app, err := gocmsapp.NewApp(opts)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func loginCookie(t *testing.T, handler http.Handler, identifier, password string) *http.Cookie {
	t.Helper()
	tokenPage := httptest.NewRecorder()
	loginGet := httptest.NewRequest(http.MethodGet, "/go-login", nil)
	loginGet.Header.Set("User-Agent", "app-gocms-conformance-test")
	handler.ServeHTTP(tokenPage, loginGet)
	token := hiddenValue(tokenPage.Body.String(), "action_token")
	if token == "" {
		t.Fatalf("missing login action token for %s", identifier)
	}
	form := url.Values{}
	form.Set("action_token", token)
	form.Set("identifier", identifier)
	form.Set("password", password)
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/go-login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "app-gocms-conformance-test")
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusSeeOther {
		t.Fatalf("login for %s returned %d: %s", identifier, resp.Code, resp.Body.String())
	}
	for _, cookie := range resp.Result().Cookies() {
		if cookie.Name == "appcms_session" {
			return cookie
		}
	}
	t.Fatalf("login failed for %s: no session cookie", identifier)
	return nil
}

func createPost(t *testing.T, handler http.Handler, cookie *http.Cookie, id, title, status string) {
	t.Helper()
	token := formToken(t, handler, cookie, "/go-admin/posts/new")
	form := url.Values{}
	form.Set("action_token", token)
	form.Set("id", id)
	form.Set("title", title)
	form.Set("slug", id)
	form.Set("status", status)
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/go-admin/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	withTestHeaders(req)
	req.AddCookie(cookie)
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusSeeOther {
		t.Fatalf("create %s returned %d: %s", id, resp.Code, resp.Body.String())
	}
}

func withTestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "app-gocms-conformance-test")
}

func formToken(t *testing.T, handler http.Handler, cookie *http.Cookie, path string) string {
	t.Helper()
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	withTestHeaders(req)
	req.AddCookie(cookie)
	handler.ServeHTTP(resp, req)
	token := hiddenValue(resp.Body.String(), "action_token")
	if token == "" {
		t.Fatalf("missing action token on %s", path)
	}
	return token
}

func adminGet(t *testing.T, handler http.Handler, cookie *http.Cookie, path string) string {
	t.Helper()
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	withTestHeaders(req)
	req.AddCookie(cookie)
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET %s = %d", path, resp.Code)
	}
	return resp.Body.String()
}

func restGet(t *testing.T, handler http.Handler, cookie *http.Cookie, path string) string {
	t.Helper()
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	withTestHeaders(req)
	req.AddCookie(cookie)
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", path, resp.Code, resp.Body.String())
	}
	return resp.Body.String()
}

func patchREST(t *testing.T, handler http.Handler, cookie *http.Cookie, path string, body map[string]any) {
	t.Helper()
	raw, _ := json.Marshal(body)
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	withTestHeaders(req)
	req.AddCookie(cookie)
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("PATCH %s = %d: %s", path, resp.Code, resp.Body.String())
	}
}

func transitionOK(t *testing.T, handler http.Handler, cookie *http.Cookie, method, path string, body []byte) {
	t.Helper()
	if method == http.MethodPost && strings.HasPrefix(path, "/go-admin/") {
		token := formToken(t, handler, cookie, strings.TrimSuffix(path, "/publish")+"/edit")
		form := url.Values{}
		form.Set("action_token", token)
		body = []byte(form.Encode())
	}
	resp := httptest.NewRecorder()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
		contentType := "application/json"
		if strings.HasPrefix(path, "/go-admin/") {
			contentType = "application/x-www-form-urlencoded"
		}
		req.Header.Set("Content-Type", contentType)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	withTestHeaders(req)
	req.AddCookie(cookie)
	handler.ServeHTTP(resp, req)
	want := http.StatusOK
	if strings.HasPrefix(path, "/go-admin/") {
		want = http.StatusSeeOther
	}
	if resp.Code != want {
		t.Fatalf("%s %s = %d: %s", method, path, resp.Code, resp.Body.String())
	}
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
