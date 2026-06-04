# URL Contracts

AppCMS preserves **GoCMS root URL contracts**. This page lists the route surfaces and how runtime modes affect them.

## Root contracts (must not break)

| Path | Method | Surface |
|------|--------|---------|
| `/go-admin` | GET | Admin entry (dashboard) |
| `/go-admin/{path...}` | GET | Resource screens (posts, pages, …) |
| `/go-login` | GET, POST | Login form and submission |
| `/go-logout` | GET, POST | Session clear → redirect login |
| `/go-json` | GET | API root discovery |
| `/go-json/go/v2` | GET | v2 discovery |
| `/go-json/go/v2/{path...}` | * | Resource API (see below) |

Spaces overlays (AppSuite) add parallel paths under `/go-admin/spaces/*` and `/go-json/spaces/*` without replacing these roots.

## REST v2 resources (current)

Typical paths (non-exhaustive; see discovery JSON for canonical list):

| Path | Notes |
|------|-------|
| `/go-json/go/v2/posts` | List/create posts |
| `/go-json/go/v2/posts/{id}` | Get/update/delete |
| `/go-json/go/v2/posts/by-slug/{slug}` | Public-safe by-slug read |
| `/go-json/go/v2/pages` | Pages collection |
| `/go-json/go/v2/pages/by-slug/{slug}` | Page by slug |
| `/go-json/go/v2/media` | Media metadata |
| `/go-json/go/v2/taxonomies` | Taxonomies |
| `/go-json/go/v2/settings` | Settings |
| `/go-json/go/v2/search` | Search |
| `/go-json/go/v2/ops/health` | CMS runtime health (plugin) |
| `/go-json/go/v2/ops/audit` | Audit log (plugin) |
| `/go-json/go/v2/ops/errors` | Error log (plugin) |
| `/go-json/go/v2/ops/snapshot` | Export/import snapshot (plugin) |

List responses use codex-style envelopes with `data` and `pagination`. Errors use structured JSON with `code` / `message` fields (see REST tests).

## Admin screens (current)

| Path | Screen type |
|------|-------------|
| `/go-admin/` | Dashboard |
| `/go-admin/posts` | Table |
| `/go-admin/posts/new` | Form |
| `/go-admin/posts/{id}/edit` | Form |
| `/go-admin/pages` | Table |
| `/go-admin/content-types` | Table |
| `/go-admin/taxonomies` | Table |
| `/go-admin/terms` | Table |
| `/go-admin/meta` | Table |
| `/go-admin/media` | Table |
| `/go-admin/authors` | Table |
| `/go-admin/settings/new` | Settings form |
| `/go-admin/import-export` | Operations plugin UI |

## Public site

| Path | Content |
|------|---------|
| `/` | Home (latest/public content in theme) |
| `/posts/{slug}` | Single post |
| `/{slug}` | Page by slug (non-reserved paths) |

Reserved: paths starting with `/go-admin`, `/go-json`, `/go-login`, `/go-logout`, `/static`, `/healthz`, `/readyz` are not handled as public slugs.

## Extensions

| Path | Plugin |
|------|--------|
| `/go-graphql` | GraphQL (GET query param or POST JSON body) |

## Framework probes

| Path | Purpose |
|------|---------|
| `/healthz` | Liveness |
| `/readyz` | Readiness |
| `/static/*` | CSS, theme assets |

## Runtime mode gating

| Mode | Public routes | Admin routes | REST + plugins |
|------|---------------|--------------|----------------|
| `full` | Registered | Registered | Registered |
| `headless` | Not registered | Not registered | Registered |
| `admin` | Not registered | Registered | Registered |
| `conformance` | Registered (seeded fixtures) | Registered | Registered |

`Options.Headless: true` is equivalent to `RuntimeModeHeadless`.

## Discovery

Hit `GET /go-json/go/v2/` for the live route list returned by codex helpers. Tests assert non-empty `routes` in the discovery payload.
