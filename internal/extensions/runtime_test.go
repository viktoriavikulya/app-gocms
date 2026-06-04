package extensions

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/fastygo/platform/pkg/contracts"
)

type testPlugin struct {
	manifest Manifest
	register func(context.Context, *Context) error
}

func (p testPlugin) Manifest() Manifest {
	return p.manifest
}

func (p testPlugin) Register(ctx context.Context, registry *Context) error {
	if p.register != nil {
		return p.register(ctx, registry)
	}
	return nil
}

func TestRuntimeManifestValidationAndLifecycle(t *testing.T) {
	if _, err := NewRuntime(testPlugin{manifest: Manifest{ID: "Bad ID", Name: "Bad", Version: "0.1", Contract: "go-codex.plugin.v0.1"}}); err == nil {
		t.Fatalf("expected invalid manifest")
	}
	runtime, err := NewRuntime(testPlugin{
		manifest: Manifest{ID: "demo", Name: "Demo", Version: "0.1", Contract: "go-codex.plugin.v0.1"},
		register: func(_ context.Context, registry *Context) error {
			registry.AddRoute(Route{Method: http.MethodGet, Pattern: "/demo", Handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }})
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if runtime.State("demo") != StateInstalled {
		t.Fatalf("initial state = %s", runtime.State("demo"))
	}
	if err := runtime.Activate(context.Background(), "demo"); err != nil {
		t.Fatal(err)
	}
	if runtime.State("demo") != StateActive || len(runtime.ActiveContext().Routes) != 1 {
		t.Fatalf("plugin should be active with route")
	}
}

func TestRuntimeRollbackOnFailedActivation(t *testing.T) {
	runtime, err := NewRuntime(
		testPlugin{manifest: Manifest{ID: "good", Name: "Good", Version: "0.1", Contract: "go-codex.plugin.v0.1"}, register: func(_ context.Context, registry *Context) error {
			registry.AddRoute(Route{Method: http.MethodGet, Pattern: "/good", Handler: func(http.ResponseWriter, *http.Request) {}})
			return nil
		}},
		testPlugin{manifest: Manifest{ID: "bad", Name: "Bad", Version: "0.1", Contract: "go-codex.plugin.v0.1"}, register: func(context.Context, *Context) error {
			return fmt.Errorf("boom")
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Activate(context.Background(), "good", "bad"); err == nil {
		t.Fatalf("expected activation failure")
	}
	if runtime.State("bad") != StateFailed || len(runtime.ActiveContext().Routes) != 0 {
		t.Fatalf("failed activation must rollback active exposure")
	}
}

func TestRuntimeActiveOnlyRoutes(t *testing.T) {
	runtime, err := NewRuntime(testPlugin{manifest: Manifest{ID: "demo", Name: "Demo", Version: "0.1", Contract: "go-codex.plugin.v0.1"}, register: func(_ context.Context, registry *Context) error {
		registry.AddRoute(Route{Method: http.MethodGet, Pattern: "/demo", Handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }})
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	runtime.RegisterRoutes(mux, nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/demo", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("inactive route should not be exposed")
	}
	if err := runtime.Activate(context.Background(), "demo"); err != nil {
		t.Fatal(err)
	}
	mux = http.NewServeMux()
	runtime.RegisterRoutes(mux, func(*http.Request, contracts.CapabilityID) (bool, bool) { return true, true })
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/demo", nil))
	if response.Code != http.StatusOK || response.Body.String() != "ok" {
		t.Fatalf("active route got %d %q", response.Code, response.Body.String())
	}
}

func TestHookBusOrderingAndFilterType(t *testing.T) {
	bus := NewHookBus()
	order := []string{}
	bus.AddAction("demo.action", ActionHandler{ID: "late", Priority: 20, Handle: func(context.Context, any) error {
		order = append(order, "late")
		return nil
	}})
	bus.AddAction("demo.action", ActionHandler{ID: "early", Priority: 10, Handle: func(context.Context, any) error {
		order = append(order, "early")
		return nil
	}})
	if err := bus.Dispatch(context.Background(), "demo.action", nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(order, []string{"early", "late"}) {
		t.Fatalf("action order = %#v", order)
	}
	bus.AddFilter("demo.filter", FilterHandler{ID: "suffix", Priority: 10, Handle: func(_ context.Context, value any) (any, error) {
		return value.(string) + "-filtered", nil
	}})
	got, err := ApplyFilter(context.Background(), bus, "demo.filter", "value")
	if err != nil {
		t.Fatal(err)
	}
	if got != "value-filtered" {
		t.Fatalf("filter result = %q", got)
	}
}
