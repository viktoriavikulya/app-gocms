# Templ (`github.com/fastygo/templ`)

**Templ** is the official UI kit for FastyGo product apps. It provides type-safe HTML components used by Platform renderers and by AppCMS product views.

AppCMS does **not** use UI8Kit. All new admin and public markup should use `github.com/fastygo/templ` components and explicit Tailwind utility classes in templates or CSS.

## Package layout (Templ repo)

| Area | Path | Purpose |
|------|------|---------|
| UI primitives | `ui/button`, `ui/form`, `ui/input`, `ui/table`, … | Atoms and small molecules |
| Facade | `ui/facade.templ` | Convenient imports for common widgets |
| Layout | `ui/stack`, `ui/grid`, `ui/container` | Composition helpers |

Components follow the Templ component spec (`*.spec.md` next to `*.templ`) for API enums and showcase examples.

## How AppCMS uses Templ

### Admin shell — `pkg/templ`

- `page.templ` — wraps Platform-rendered screens in product layout
- Generated `page_templ.go` via `go tool templ generate`

Admin screens are built as `render.ScreenModel` in `internal/appschema`, then rendered through Platform `pkg/renderers/templ` inside the product page wrapper.

### Public themes — `internal/themes`

- `views.templ` — `BlankPage`, `DefaultPage`, navigation and content regions
- Uses `publicrender.Page` DTO and shared templ primitives
- Theme CSS under `web/static/themes/{theme-id}/`

### Platform renderer

`github.com/fastygo/platform/pkg/renderers/templ` renders generic tables, forms, and dashboard views from metadata. AppCMS supplies descriptors via `pkg/module/panels`; it does not fork the renderer for standard CRUD screens.

## Styling pipeline

AppCMS owns product CSS:

| File | Role |
|------|------|
| `web/static/css/input.css` | Tailwind entry |
| `web/static/css/app.css` | Built output (`bun run build:css`) |
| Theme CSS | Per-theme overrides |

Verification: `bun verify` runs templ generate + Tailwind + `go test`.

## Dependency rules

| Layer | May import Templ? |
|-------|-------------------|
| `pkg/templ` | Yes |
| `internal/themes` | Yes |
| `internal/domain` | **No** |
| `internal/application` | **No** |
| Platform `pkg/renderers/templ` | Yes |

Keep domain and application packages free of HTML/templ imports.

## Accessibility

Platform screen renderer adds metadata-driven attributes (`aria-label`, form actions, empty states). Product forms include signed `action_token` hidden fields when `screen.Metadata["action_token"]` is set from `pkg/app` auth boundary.

## Further reading

- Templ repo `.cursor/rules/templ-component-spec.mdc` for component authoring
- [Public site and themes](../appcms/public-site-and-themes.md)
