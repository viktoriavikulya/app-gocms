package extensions

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/fastygo/platform/pkg/contracts"
)

type State string

const (
	StateInstalled   State = "installed"
	StateActive      State = "active"
	StateInactive    State = "inactive"
	StateFailed      State = "failed"
	StateUninstalled State = "uninstalled"
)

type Manifest struct {
	ID           string
	Name         string
	Version      string
	Contract     string
	Capabilities []contracts.CapabilityDefinition
	Settings     []SettingDefinition
	Hooks        []HookRegistration
}

type SettingDefinition struct {
	Key        string
	Type       string
	Default    any
	Public     bool
	Capability contracts.CapabilityID
}

type Route struct {
	Method     string
	Pattern    string
	Capability contracts.CapabilityID
	Handler    http.HandlerFunc
}

type HookRegistration struct {
	HookID   string
	Handler  string
	Priority int
}

type Context struct {
	Routes       []Route
	Capabilities []contracts.CapabilityDefinition
	Settings     []SettingDefinition
	Hooks        []HookRegistration
}

func (c *Context) AddRoute(route Route) {
	c.Routes = append(c.Routes, route)
}

func (c *Context) AddCapabilities(capabilities ...contracts.CapabilityDefinition) {
	c.Capabilities = append(c.Capabilities, capabilities...)
}

func (c *Context) AddSettings(settings ...SettingDefinition) {
	c.Settings = append(c.Settings, settings...)
}

func (c *Context) AddHooks(hooks ...HookRegistration) {
	c.Hooks = append(c.Hooks, hooks...)
}

type Plugin interface {
	Manifest() Manifest
	Register(context.Context, *Context) error
}

type Runtime struct {
	plugins map[string]Plugin
	states  map[string]State
	active  Context
}

func NewRuntime(plugins ...Plugin) (*Runtime, error) {
	runtime := &Runtime{plugins: map[string]Plugin{}, states: map[string]State{}}
	for _, plugin := range plugins {
		manifest := plugin.Manifest()
		if err := ValidateManifest(manifest); err != nil {
			return nil, err
		}
		if _, exists := runtime.plugins[manifest.ID]; exists {
			return nil, fmt.Errorf("duplicate plugin %q", manifest.ID)
		}
		runtime.plugins[manifest.ID] = plugin
		runtime.states[manifest.ID] = StateInstalled
	}
	return runtime, nil
}

func (r *Runtime) Activate(ctx context.Context, ids ...string) error {
	next := Context{}
	nextStates := cloneStates(r.states)
	for _, id := range ids {
		plugin, ok := r.plugins[id]
		if !ok {
			return fmt.Errorf("unknown plugin %q", id)
		}
		before := next
		if err := plugin.Register(ctx, &next); err != nil {
			next = before
			nextStates[id] = StateFailed
			r.states = nextStates
			return fmt.Errorf("activate plugin %q: %w", id, err)
		}
		nextStates[id] = StateActive
	}
	for id := range r.plugins {
		if !contains(ids, id) && nextStates[id] == StateActive {
			nextStates[id] = StateInactive
		}
	}
	r.active = next
	r.states = nextStates
	return nil
}

func (r *Runtime) ActiveContext() Context {
	return r.active
}

func (r *Runtime) State(id string) State {
	return r.states[id]
}

func (r *Runtime) RegisterRoutes(mux *http.ServeMux, authorize func(*http.Request, contracts.CapabilityID) (bool, bool)) {
	for _, route := range r.active.Routes {
		route := route
		mux.HandleFunc(strings.TrimSpace(route.Method+" "+route.Pattern), func(w http.ResponseWriter, req *http.Request) {
			if route.Capability != "" {
				authenticated, allowed := authorize(req, route.Capability)
				if !authenticated {
					http.Error(w, "authorization required", http.StatusUnauthorized)
					return
				}
				if !allowed {
					http.Error(w, "missing capability", http.StatusForbidden)
					return
				}
			}
			route.Handler(w, req)
		})
	}
}

func ValidateManifest(manifest Manifest) error {
	if !pluginIDPattern.MatchString(manifest.ID) {
		return fmt.Errorf("plugin id must be lowercase URL-safe")
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("plugin name is required")
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("plugin version is required")
	}
	if strings.TrimSpace(manifest.Contract) == "" {
		return fmt.Errorf("plugin contract is required")
	}
	return nil
}

var pluginIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func cloneStates(states map[string]State) map[string]State {
	clone := map[string]State{}
	for key, value := range states {
		clone[key] = value
	}
	return clone
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

type HookBus struct {
	actions map[string][]ActionHandler
	filters map[string][]FilterHandler
}

type ActionHandler struct {
	ID       string
	Priority int
	Handle   func(context.Context, any) error
}

type FilterHandler struct {
	ID       string
	Priority int
	Handle   func(context.Context, any) (any, error)
}

func NewHookBus() *HookBus {
	return &HookBus{actions: map[string][]ActionHandler{}, filters: map[string][]FilterHandler{}}
}

func (b *HookBus) AddAction(hook string, handler ActionHandler) {
	b.actions[hook] = append(b.actions[hook], handler)
	sort.SliceStable(b.actions[hook], func(i, j int) bool { return b.actions[hook][i].Priority < b.actions[hook][j].Priority })
}

func (b *HookBus) AddFilter(hook string, handler FilterHandler) {
	b.filters[hook] = append(b.filters[hook], handler)
	sort.SliceStable(b.filters[hook], func(i, j int) bool { return b.filters[hook][i].Priority < b.filters[hook][j].Priority })
}

func (b *HookBus) Dispatch(ctx context.Context, hook string, payload any) error {
	for _, handler := range b.actions[hook] {
		if err := handler.Handle(ctx, payload); err != nil {
			return err
		}
	}
	return nil
}

func ApplyFilter[T any](ctx context.Context, bus *HookBus, hook string, value T) (T, error) {
	if bus == nil {
		return value, nil
	}
	current := any(value)
	for _, handler := range bus.filters[hook] {
		next, err := handler.Handle(ctx, current)
		if err != nil {
			return value, err
		}
		typed, ok := next.(T)
		if !ok {
			return value, fmt.Errorf("filter %q returned incompatible type", hook)
		}
		current = typed
	}
	return current.(T), nil
}
