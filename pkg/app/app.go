package app

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fastygo/app-gocms/internal/appschema"
	views "github.com/fastygo/app-gocms/pkg/templ"
	frameworkapp "github.com/fastygo/framework/pkg/app"
	"github.com/fastygo/framework/pkg/web/security"
)

type Options struct {
	Addr      string
	StaticDir string
	Registry  *appschema.Registry
}

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
	return frameworkapp.New(cfg).
		WithSecurity(security.LoadConfig()).
		WithHealthEndpoints(cfg.HealthLivePath, cfg.HealthReadyPath).
		WithFeature(feature{registry: registry}).
		Build(), nil
}

func NewMux(options Options) *http.ServeMux {
	registry, err := registryFromOptions(options)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(options.StaticDir))))
	registerRoutes(mux, registry)
	return mux
}

type feature struct {
	registry *appschema.Registry
}

func (f feature) ID() string {
	return "app-gocms"
}

func (f feature) NavItems() []frameworkapp.NavItem {
	return nil
}

func (f feature) Routes(mux *http.ServeMux) {
	registerRoutes(mux, f.registry)
}

func registerRoutes(mux *http.ServeMux, registry *appschema.Registry) {
	mux.HandleFunc("GET /{$}", renderHome)
	mux.HandleFunc("GET /go-login", renderLogin)
	mux.HandleFunc("POST /go-login", completeLogin)
	mux.HandleFunc("GET /go-logout", completeLogout)
	mux.HandleFunc("POST /go-logout", completeLogout)
	mux.HandleFunc("GET /go-admin/{$}", renderAdminDashboard(registry))
	mux.HandleFunc("GET /go-admin/{path...}", renderAdminScreen(registry))
	mux.HandleFunc("GET /go-json/{$}", renderAPIRoot)
	mux.HandleFunc("GET /go-json/go/v2/{$}", renderAPIV2)
	mux.HandleFunc("GET /go-json/go/v2/{path...}", renderAPIResource)
}

func renderHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.Home().Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderAdminDashboard(registry *appschema.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		screen := registry.DashboardScreen()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := views.Page("GoCMS Admin", screen).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func renderAdminScreen(registry *appschema.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimRight(r.URL.Path, "/")
		screen, err := registry.Screen(path)
		if err != nil {
			writeJSON(w, http.StatusNotFound, appschema.NotFound("admin screen not found"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := views.Page(screen.Title, screen).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func renderLogin(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html><html><body><main><h1>GoCMS Login</h1><form method="post" action="/go-login"><button type="submit">Continue</button></form></main></body></html>`))
}

func completeLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/go-admin/", http.StatusSeeOther)
}

func completeLogout(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/go-login", http.StatusSeeOther)
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
