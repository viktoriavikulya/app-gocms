package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/appschema"
	"github.com/fastygo/app-gocms/internal/delivery/publicsite"
	"github.com/fastygo/app-gocms/internal/delivery/rest"
	"github.com/fastygo/app-gocms/internal/extensions"
	"github.com/fastygo/app-gocms/internal/operations"
	graphqlplugin "github.com/fastygo/app-gocms/internal/plugins/graphql"
	opsplugin "github.com/fastygo/app-gocms/internal/plugins/ops"
	"github.com/fastygo/app-gocms/internal/storage"
	storagesqlite "github.com/fastygo/app-gocms/internal/storage/sqlite"
	"github.com/fastygo/app-gocms/internal/themes"
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	views "github.com/fastygo/app-gocms/pkg/templ"
	frameworkapp "github.com/fastygo/framework/pkg/app"
	"github.com/fastygo/platform/pkg/render"
	frameworkauth "github.com/fastygo/framework/pkg/auth"
	"github.com/fastygo/framework/pkg/web/security"
	"github.com/fastygo/platform/pkg/contracts"
)

type Options struct {
	Addr        string
	StaticDir   string
	Registry    *appschema.Registry
	Storage     contracts.StoragePort
	StorageDSN  string
	Seed        bool
	AuthStore   *appauthn.MemoryStore
	SessionKey  string
	Headless    bool
	RuntimeMode RuntimeMode
}

type RuntimeMode string

const (
	RuntimeModeFull        RuntimeMode = "full"
	RuntimeModeHeadless    RuntimeMode = "headless"
	RuntimeModeAdmin       RuntimeMode = "admin"
	RuntimeModeConformance RuntimeMode = "conformance"
)

func Run() error {
	application, err := NewApp(Options{})
	if err != nil {
		return err
	}
	log.Printf("app-gocms http://%s/", application.Config().AppBind)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := application.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func NewApp(options Options) (*frameworkapp.App, error) {
	registry, err := registryFromOptions(options)
	if err != nil {
		return nil, err
	}
	cfg, err := frameworkConfig(options)
	if err != nil {
		return nil, err
	}
	provider, err := providerFromOptions(context.Background(), options)
	if err != nil {
		return nil, err
	}
	authBoundary, err := authFromOptions(options, cfg)
	if err != nil {
		return nil, err
	}
	return frameworkapp.New(cfg).
		WithSecurity(security.LoadConfig()).
		WithHealthEndpoints(cfg.HealthLivePath, cfg.HealthReadyPath).
		WithFeature(feature{registry: registry, provider: provider, auth: authBoundary, mode: runtimeMode(options)}).
		Build(), nil
}

func NewMux(options Options) *http.ServeMux {
	registry, err := registryFromOptions(options)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(options.StaticDir))))
	provider, err := providerFromOptions(context.Background(), options)
	if err != nil {
		panic(err)
	}
	cfg, err := frameworkConfig(options)
	if err != nil {
		panic(err)
	}
	authBoundary, err := authFromOptions(options, cfg)
	if err != nil {
		panic(err)
	}
	registerRoutesWithOptions(mux, registry, provider, authBoundary, runtimeMode(options))
	return mux
}

type feature struct {
	registry *appschema.Registry
	provider storage.StoreProvider
	auth     authBoundary
	mode     RuntimeMode
}

func (f feature) ID() string {
	return "app-gocms"
}

func (f feature) NavItems() []frameworkapp.NavItem {
	return nil
}

func (f feature) Routes(mux *http.ServeMux) {
	registerRoutesWithOptions(mux, f.registry, f.provider, f.auth, f.mode)
}

func registerRoutes(mux *http.ServeMux, registry *appschema.Registry, provider storage.StoreProvider, authBoundary authBoundary) {
	registerRoutesWithOptions(mux, registry, provider, authBoundary, RuntimeModeFull)
}

func registerRoutesWithOptions(mux *http.ServeMux, registry *appschema.Registry, provider storage.StoreProvider, authBoundary authBoundary, mode RuntimeMode) {
	public := publicsite.NewHandler(provider, "root", themes.DefaultRegistry())
	public.Headless = mode == RuntimeModeHeadless || mode == RuntimeModeAdmin
	if mode != RuntimeModeHeadless && mode != RuntimeModeAdmin {
		mux.Handle("GET /{$}", public)
	}
	mux.HandleFunc("GET /go-login", authBoundary.renderLogin)
	mux.HandleFunc("POST /go-login", authBoundary.completeLogin)
	mux.HandleFunc("GET /go-logout", authBoundary.completeLogout)
	mux.HandleFunc("POST /go-logout", authBoundary.completeLogout)
	registerBFFRoutes(mux, registry, provider, authBoundary)
	if mode != RuntimeModeHeadless {
		mux.HandleFunc("GET /go-admin/{$}", authBoundary.renderAdminDashboard(registry))
		mux.HandleFunc("GET /go-admin/{path...}", authBoundary.renderAdminScreen(registry))
	}
	rest.NewHandler(provider, "root", authBoundary.Authorize).Register(mux)
	if mode != RuntimeModeAdmin {
		runtime, err := extensionRuntime(provider, mode)
		if err != nil {
			panic(err)
		}
		runtime.RegisterRoutes(mux, authBoundary.Authorize)
	}
	if mode != RuntimeModeHeadless && mode != RuntimeModeAdmin {
		mux.Handle("GET /posts/{slug}", public)
		mux.Handle("GET /{slug}", public)
	}
}

func renderHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.Home().Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a authBoundary) renderAdminDashboard(registry *appschema.Registry) http.HandlerFunc {
	return a.requireAdmin(func(w http.ResponseWriter, r *http.Request) {
		principal, _ := a.principalFromRequest(r)
		page, err := registry.DashboardPage(r.Context(), principal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := views.Page(page).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (a authBoundary) renderAdminScreen(registry *appschema.Registry) http.HandlerFunc {
	return a.requireAdmin(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimRight(r.URL.Path, "/")
		principal, _ := a.principalFromRequest(r)
		page, err := registry.Page(r.Context(), path, principal)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`<!doctype html><html><body><main><h1>Admin screen not found</h1><p>The requested admin route is not available.</p></main></body></html>`))
			return
		}
		if err := a.injectFormActionToken(&page.Screen); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := views.Page(page).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

type authBoundary struct {
	service appauthn.Service
	session frameworkauth.CookieSession[contracts.SessionClaims]
	secret  string
}

type actionToken struct {
	Action string `json:"action"`
	Exp    int64  `json:"exp"`
}

func authFromOptions(options Options, cfg frameworkapp.Config) (authBoundary, error) {
	store := options.AuthStore
	if store == nil {
		seeded, err := appauthn.NewSeededMemoryStore()
		if err != nil {
			return authBoundary{}, err
		}
		store = seeded
	}
	secret := firstNonEmpty(options.SessionKey, cfg.SessionKey, "appcms-development-session-secret-32-bytes")
	return authBoundary{
		service: appauthn.NewService(store),
		secret:  secret,
		session: frameworkauth.CookieSession[contracts.SessionClaims]{
			Name:     "appcms_session",
			Path:     "/",
			Secret:   secret,
			TTL:      8 * time.Hour,
			SameSite: http.SameSiteLaxMode,
			HTTPOnly: true,
		},
	}, nil
}

func (a authBoundary) renderLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	token, err := a.actionToken("login", 10*time.Minute)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = fmt.Fprintf(w, `<!doctype html><html><body><main><h1>GoCMS Login</h1><form method="post" action="/go-login"><input type="hidden" name="action_token" value="%s"><label>Username <input name="identifier" autocomplete="username"></label><label>Password <input name="password" type="password" autocomplete="current-password"></label><button type="submit">Sign in</button></form></main></body></html>`, token)
}

func (a authBoundary) completeLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	if !a.validActionToken(r.FormValue("action_token"), "login") {
		http.Error(w, "invalid action token", http.StatusForbidden)
		return
	}
	principal, err := a.service.AuthenticatePassword(r.Context(), r.FormValue("identifier"), r.FormValue("password"), r.RemoteAddr)
	if errors.Is(err, appauthn.ErrLoginLocked) {
		http.Error(w, "login temporarily locked", http.StatusTooManyRequests)
		return
	}
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err := a.session.Issue(w, contracts.SessionClaims{PrincipalID: string(principal.ID()), ProfileID: "gocms-admin"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/go-admin/", http.StatusSeeOther)
}

func (a authBoundary) completeLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		if token := r.FormValue("action_token"); token != "" && !a.validActionToken(token, "logout") {
			http.Error(w, "invalid action token", http.StatusForbidden)
			return
		}
	}
	a.session.Clear(w)
	http.Redirect(w, r, "/go-login", http.StatusSeeOther)
}

func (a authBoundary) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := a.principalFromRequest(r)
		if !ok {
			http.Redirect(w, r, "/go-login", http.StatusSeeOther)
			return
		}
		if !principal.Has(modulecms.CapabilityAdminAccess) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		runtime := contracts.RuntimeContext{ProfileID: "gocms-admin", WorkspaceID: "root", ModuleID: "cms", PrincipalID: principal.ID()}
		next(w, r.WithContext(contracts.WithRuntimeContext(r.Context(), runtime)))
	}
}

func (a authBoundary) Authorize(r *http.Request, capability contracts.CapabilityID) (bool, bool) {
	principal, ok := a.principalFromRequest(r)
	if !ok {
		return false, false
	}
	return true, principal.Has(capability)
}

func (a authBoundary) principalFromRequest(r *http.Request) (appauthn.Principal, bool) {
	if claims, ok := a.session.Read(r); ok && claims.PrincipalID != "" {
		if principal, found := a.service.Principal(contracts.PrincipalID(claims.PrincipalID)); found {
			return principal, true
		}
	}
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
		principal, err := a.service.AuthenticateAppToken(r.Context(), strings.TrimPrefix(header, "Bearer "))
		return principal, err == nil
	}
	return appauthn.Principal{}, false
}

func (a authBoundary) actionToken(action string, ttl time.Duration) (string, error) {
	return frameworkauth.SignedEncode(actionToken{Action: action, Exp: time.Now().Add(ttl).Unix()}, a.secret)
}

func (a authBoundary) injectFormActionToken(screen *render.ScreenModel) error {
	if string(screen.View) != "form" {
		return nil
	}
	if screen.Metadata == nil {
		screen.Metadata = map[string]string{}
	}
	scope := screen.Metadata["action_scope"]
	if scope == "" {
		scope = "admin-write"
	}
	token, err := a.actionToken(scope, 10*time.Minute)
	if err != nil {
		return err
	}
	screen.Metadata["action_token"] = token
	return nil
}

func (a authBoundary) validActionToken(raw string, action string) bool {
	var token actionToken
	if err := frameworkauth.SignedDecode(raw, a.secret, &token); err != nil {
		return false
	}
	return token.Action == action && token.Exp >= time.Now().Unix()
}

func renderAPIRoot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, appschema.RootDiscovery())
}

func renderAPIV2(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, appschema.V2Discovery())
}

func renderAPIResource(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/go-json/go/v2/"), "/")
	switch {
	case path == "posts", path == "pages", path == "media", path == "taxonomies", path == "search", path == "settings":
		writeJSON(w, http.StatusOK, appschema.EmptyList[map[string]any]())
	case strings.HasPrefix(path, "posts/"), strings.HasPrefix(path, "pages/"), strings.HasPrefix(path, "media/"), strings.HasPrefix(path, "authors/"):
		writeJSON(w, http.StatusNotFound, appschema.NotFound("resource not found"))
	default:
		writeJSON(w, http.StatusNotFound, appschema.NotFound("route not found"))
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func providerFromOptions(ctx context.Context, options Options) (storage.StoreProvider, error) {
	if options.Storage != nil {
		return storage.NewProvider(options.Storage), nil
	}
	dsn := options.StorageDSN
	if dsn == "" {
		dsn = "file:appcms?mode=memory&cache=shared"
	}
	store, err := storagesqlite.Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := store.Init(ctx); err != nil {
		return nil, err
	}
	if options.Seed {
		if err := storagesqlite.SeedMinimalSite(ctx, store, "root"); err != nil {
			return nil, err
		}
	}
	return storage.NewProvider(store), nil
}

func registryFromOptions(options Options) (*appschema.Registry, error) {
	if options.Registry != nil {
		return options.Registry, nil
	}
	return appschema.NewRegistry()
}

func frameworkConfig(options Options) (frameworkapp.Config, error) {
	cfg, err := frameworkapp.LoadConfig()
	if err != nil {
		return frameworkapp.Config{}, err
	}
	if options.Addr != "" {
		cfg.AppBind = options.Addr
	} else if addr := os.Getenv("ADDR"); addr != "" {
		cfg.AppBind = addr
	}
	if options.StaticDir != "" {
		cfg.StaticDir = options.StaticDir
	} else if os.Getenv("APP_STATIC_DIR") == "" {
		root, err := repoRoot()
		if err != nil {
			return frameworkapp.Config{}, err
		}
		cfg.StaticDir = filepath.Join(root, "web", "static")
	}
	if cfg.HealthLivePath == "" {
		cfg.HealthLivePath = "/healthz"
	}
	if cfg.HealthReadyPath == "" {
		cfg.HealthReadyPath = "/readyz"
	}
	return cfg, nil
}

func runtimeMode(options Options) RuntimeMode {
	if options.Headless {
		return RuntimeModeHeadless
	}
	if options.RuntimeMode != "" {
		return options.RuntimeMode
	}
	return RuntimeModeAdmin
}

func extensionRuntime(provider storage.StoreProvider, mode RuntimeMode) (*extensions.Runtime, error) {
	opsStore := operations.NewStore(100)
	runtime, err := extensions.NewRuntime(
		graphqlplugin.New(provider),
		opsplugin.New(provider, opsStore, func() string { return string(mode) }),
	)
	if err != nil {
		return nil, err
	}
	active := []string{"graphql", "operations"}
	if err := runtime.Activate(context.Background(), active...); err != nil {
		return nil, err
	}
	return runtime, nil
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
