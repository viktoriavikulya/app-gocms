package themes

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/fastygo/app-gocms/internal/publicrender"
)

type TemplateRole string

const (
	RoleIndex    TemplateRole = "index"
	RoleFront    TemplateRole = "front"
	RolePage     TemplateRole = "page"
	RolePost     TemplateRole = "post"
	RoleArchive  TemplateRole = "archive"
	RoleSearch   TemplateRole = "search"
	RoleNotFound TemplateRole = "not_found"
)

type Manifest struct {
	ID        string
	Name      string
	Version   string
	Contract  string
	Templates map[TemplateRole]string
	Assets    []string
}

type Theme interface {
	Manifest() Manifest
	Render(context.Context, publicrender.Page) templ.Component
}

type Registry struct {
	themes map[string]Theme
}

func NewRegistry(themes ...Theme) (Registry, error) {
	registry := Registry{themes: map[string]Theme{}}
	for _, theme := range themes {
		manifest := theme.Manifest()
		if err := ValidateManifest(manifest); err != nil {
			return Registry{}, err
		}
		if _, exists := registry.themes[manifest.ID]; exists {
			return Registry{}, fmt.Errorf("duplicate theme %q", manifest.ID)
		}
		registry.themes[manifest.ID] = theme
	}
	return registry, nil
}

func DefaultRegistry() Registry {
	registry, err := NewRegistry(Blank{}, GoCMSDefault{})
	if err != nil {
		panic(err)
	}
	return registry
}

func (r Registry) Resolve(id string) (Theme, bool) {
	if theme, ok := r.themes[strings.TrimSpace(id)]; ok {
		return theme, true
	}
	theme, ok := r.themes["gocms-default"]
	return theme, ok
}

func ValidateManifest(manifest Manifest) error {
	if !themeIDPattern.MatchString(manifest.ID) {
		return fmt.Errorf("theme id must be lowercase URL-safe")
	}
	if strings.TrimSpace(manifest.Name) == "" {
		return fmt.Errorf("theme name is required")
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("theme version is required")
	}
	if strings.TrimSpace(manifest.Contract) == "" {
		return fmt.Errorf("theme contract is required")
	}
	if _, ok := manifest.Templates[RoleIndex]; !ok {
		return fmt.Errorf("theme %q must provide index template fallback", manifest.ID)
	}
	return nil
}

var themeIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type Blank struct{}

func (Blank) Manifest() Manifest {
	return Manifest{
		ID:       "blank",
		Name:     "Blank",
		Version:  "0.1.0",
		Contract: "go-codex.theme.v0.1",
		Templates: map[TemplateRole]string{
			RoleIndex:    "index",
			RoleFront:    "index",
			RolePage:     "index",
			RolePost:     "index",
			RoleNotFound: "index",
		},
		Assets: []string{"/static/themes/blank/theme.css"},
	}
}

func (Blank) Render(_ context.Context, page publicrender.Page) templ.Component {
	return BlankPage(page)
}

type GoCMSDefault struct{}

func (GoCMSDefault) Manifest() Manifest {
	return Manifest{
		ID:       "gocms-default",
		Name:     "GoCMS Default",
		Version:  "0.1.0",
		Contract: "go-codex.theme.v0.1",
		Templates: map[TemplateRole]string{
			RoleIndex:    "index",
			RoleFront:    "front",
			RolePage:     "page",
			RolePost:     "post",
			RoleArchive:  "archive",
			RoleSearch:   "search",
			RoleNotFound: "not_found",
		},
		Assets: []string{"/static/themes/gocms-default/theme.css"},
	}
}

func (GoCMSDefault) Render(_ context.Context, page publicrender.Page) templ.Component {
	return DefaultPage(page)
}
