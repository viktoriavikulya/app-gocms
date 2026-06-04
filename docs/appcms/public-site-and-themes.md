# Public Site and Themes

AppCMS serves a **codex-shaped public site** with theme manifests and templ views—parallel to admin/API, not a separate application.

## Components

| Package | Role |
|---------|------|
| `internal/permalinks` | Map URL path → content candidates (home, post, page) |
| `internal/delivery/publicsite` | HTTP handler + assembler |
| `internal/publicrender` | Page DTO (decouples themes from delivery) |
| `internal/themes` | Manifest registry + templ views |
| `web/static/themes/*` | Theme CSS |

## Permalink patterns

| Pattern | Resolves to |
|---------|-------------|
| `/` | Home |
| `/posts/{slug}` | Post by slug |
| `/{slug}` | Page by slug (if not a system path) |

Resolver returns candidates; assembler loads data from application services inside a storage transaction.

## Page assembly

`assembler` loads:

- Active theme ID from settings (`theme.active`)
- Site title/description
- Primary menu location
- Content entry for resolved candidate

Output: `publicrender.Page` with HTTP status (200 or 404), screen kind, entry payload.

## Themes

Built-in themes (registry `themes.DefaultRegistry()`):

| ID | Description |
|----|-------------|
| `blank` | Minimal markup |
| `gocms-default` | Simplified default theme with navigation and content regions |

Each theme implements:

- `Manifest()` — codex-shaped fields (name, version, template roles)
- `Render(ctx, page)` — templ component

HTML includes markers for tests and integration:

- `data-gocms-theme="{id}"`
- `data-gocms-public-screen="{home|post|page|not_found}"`

## Static assets

Themes may ship CSS under `web/static/themes/{id}/theme.css`. `gocms-default` imports shared `app.css`.

## Headless and admin-only modes

When runtime mode is `headless` or `admin`:

- Public routes are not registered on the mux
- Direct requests to `/` or slugs return **404** from public handler if reached

REST `by-slug` endpoints remain available for headless consumers.

## REST parity

Public visibility rules also apply to unauthenticated API reads:

- Draft and private content excluded from public GraphQL list projection
- Authenticated reads may use `content.read_private` where implemented

## Adding a theme

1. Add manifest + `Theme` implementation in `internal/themes`
2. Add `views.templ` component
3. Register in `DefaultRegistry()`
4. Add CSS under `web/static/themes/{id}/`
5. Document `theme.active` setting value
6. Test via `TestPublicSiteRendersThemeAndSlugs`

## Theme contract reference

Normative fields: GoCMS `go-codex/en/04-theme-contract.md`. Keep manifests aligned for future marketplace packaging even though AppCMS uses compiled-in themes today.
