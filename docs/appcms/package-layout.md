# Package Layout

Directory reference for the AppCMS repository.

## Top level

```text
app-gocms/
├── cmd/server/          # main + integration tests
├── pkg/
│   ├── app/             # Framework host, routes, auth, bundle
│   ├── module/          # Platform Module + records/relations/panels
│   └── templ/           # Admin page shell (templ)
├── internal/            # Product implementation (not imported by other apps)
├── web/static/          # CSS, theme assets
├── docs/                # This documentation
└── go.mod
```

## `cmd/server`

| File | Purpose |
|------|---------|
| `main.go` | Entry: `app.Run()` |
| `main_test.go` | HTTP smoke: routes, auth, REST CRUD, public site, extensions, profiles |

Integration tests are the primary guard for codex URL compatibility.

## `pkg/app`

| File | Purpose |
|------|---------|
| `app.go` | `NewApp`, `NewMux`, `Options`, route registration, auth boundary, extension runtime |
| `bundle.go` | `Bundle()`, `DefaultProfile()` for AppSuite |
| `bundle_test.go` | Bundle shape and profile mounts |

Key types:

- `Options` — storage DSN, seed, auth store, `RuntimeMode`, injectable `contracts.StoragePort`
- `RuntimeMode` — `full`, `headless`, `admin`, `conformance`
- `feature` — Framework feature wiring routes

## `pkg/module`

| Path | Purpose |
|------|---------|
| `cms.go` | `Module` implementing `contracts.Module` |
| `capabilities.go` | Capability ID constants and definitions |
| `records/` | Toolset record type descriptors |
| `relations/` | Toolset relation descriptors |
| `panels/` | Admin resources, views, workflows |
| `codex/` | Discovery/envelope helpers for REST tests and stubs |
| `migration/` | Migration anchors (module-level metadata) |

## `pkg/templ`

Product admin HTML wrapper around Platform-rendered screens. Generated Go from `.templ` sources.

## `internal/domain`

Pure domain types per bounded context:

| Package | Entities |
|---------|----------|
| `content` | Entries, status, visibility, scheduling |
| `contenttype` | Content type registry |
| `settings` | Key/value settings |
| `taxonomy` | Definitions, terms |
| `media` | Assets |
| `users` | Author projection |
| `menus` | Menu locations and items |
| `revisions` | Content revisions |
| `preview` | Preview access tokens |

No imports from HTTP, SQL, or templ.

## `internal/application`

Services orchestrate domain rules and call repository interfaces:

| Package | Responsibility |
|---------|----------------|
| `content` | Draft, publish, schedule, trash, list |
| `contenttype` | Built-in types installation |
| `settings` | Registry of known keys, get/set |
| `taxonomy` | Definitions and terms |
| `media` | Asset CRUD |
| `users` | Author CRUD |
| `menus` | Menu CRUD and location lookup |
| `revisions` | Revision storage |
| `preview` | Preview token storage |
| `authn` | Password auth, app tokens, lockout, roles |

Services depend on interfaces implemented by `internal/storage` adapters.

## `internal/storage`

| Path | Purpose |
|------|---------|
| `storage.go` | `StoreProvider`, `Repositories`, generic `RecordRepository` |
| `adapters.go` | Maps records ↔ domain types for all services |
| `sqlite/` | SQLite `StoragePort`, migrations, seed |

## `internal/delivery`

| Path | Purpose |
|------|---------|
| `rest/` | `/go-json` HTTP API |
| `publicsite/` | Public HTML handler and assembler |

## `internal/appschema`

Builds `appschema.Registry` from module panel descriptors: dashboard, resource tables, forms, settings/menus special screens.

## Other `internal` packages

| Path | Purpose |
|------|---------|
| `permalinks/` | Public URL resolution |
| `publicrender/` | Public page DTO |
| `themes/` | Theme manifests and templ views |
| `extensions/` | Plugin runtime, hook bus |
| `plugins/graphql/` | `/go-graphql` |
| `plugins/ops/` | Health, audit, snapshot, import-export UI |
| `operations/` | Audit/diagnostics store, snapshot export/import, health registry |

## `web/static`

| Path | Purpose |
|------|---------|
| `css/input.css` | Tailwind source |
| `css/app.css` | Built stylesheet |
| `themes/blank/`, `themes/gocms-default/` | Theme CSS |

## Import rules for contributors

1. `internal/*` must not be imported by AppSuite or AppCRM.
2. Other apps consume only `pkg/app` bundle and public module types.
3. Do not import `ui8kit` anywhere in this repo.
