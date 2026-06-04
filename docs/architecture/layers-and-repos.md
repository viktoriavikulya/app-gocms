# Layers and Repositories

This document maps **who owns what** across the FastyGo product stack and related repositories.

## Repository roles

| Repository | Go module | Role in AppCMS |
|------------|-----------|----------------|
| **AppCMS** | `github.com/fastygo/app-gocms` | Runnable CMS app: services, storage, REST, admin, public site, extensions |
| **Platform** | `github.com/fastygo/platform` | Contracts, module host, toolset, panel, profiles, render registry |
| **Framework** | `github.com/fastygo/framework` | Production HTTP app builder, middleware, sessions, security, health |
| **Templ** | `github.com/fastygo/templ` | Shared UI components (buttons, forms, tables, layout primitives) |
| **GoCMS** | `github.com/fastygo/cms` (legacy) | Compatibility **oracle** — behavior and `go-codex/en` specs |
| **AppCRM** | `github.com/fastygo/app-crm` | Sibling product proving the same stack for non-CMS domain |
| **AppSuite** | `github.com/fastygo/app-suite` | Launcher/composer: root admin + workspace spaces overlay |
| **Panel** | (Platform-related) | Admin composition helpers consumed via Platform `pkg/panel` |

AppCMS does **not** depend on UI8Kit. Legacy GoCMS may still reference UI8Kit; new work must not.

## Dependency direction

Allowed imports (simplified):

```text
cmd/server
    → pkg/app
        → pkg/module, pkg/templ
        → internal/*
        → github.com/fastygo/framework
        → github.com/fastygo/platform
        → github.com/fastygo/templ

internal/*
    → pkg/module (types, records)
    → platform/pkg/contracts (ports only where needed)
    → must NOT import AppCRM, AppSuite, GoCMS monolith

platform
    → must NOT import app-gocms, app-crm, app-suite
```

## AppCMS directory ownership

| Path | Layer | Responsibility |
|------|-------|----------------|
| `cmd/server` | Entry | `main.go`, HTTP integration tests |
| `pkg/app` | Host | Framework `App`, route table, auth boundary, storage wiring, runtime mode |
| `pkg/app/bundle.go` | Composition | `appbundle.Bundle` for AppSuite / standalone profile |
| `pkg/module` | Module | `contracts.Module` implementation: toolset record types, panel resources |
| `pkg/module/records` | Schema | Record type IDs and field shapes for CMS entities |
| `pkg/module/relations` | Schema | Relation definitions (toolset) |
| `pkg/module/panels` | Admin metadata | Resource bindings, views, workflows |
| `pkg/module/codex` | API helpers | Discovery and envelope helpers aligned with codex |
| `pkg/templ` | UI shell | Product admin page wrapper around Platform screen renderer |
| `internal/domain/*` | Domain | Pure types and validation per bounded context |
| `internal/application/*` | Services | Use cases: content, settings, authn, taxonomy, etc. |
| `internal/storage` | Persistence port | `StoreProvider`, repositories, adapters to domain |
| `internal/storage/sqlite` | Driver | `contracts.StoragePort` + migrations (today) |
| `internal/delivery/rest` | HTTP API | `/go-json` handlers |
| `internal/delivery/publicsite` | HTTP public | Permalink resolution and page assembly |
| `internal/appschema` | Admin routing | Screen registry from module descriptors |
| `internal/permalinks` | Public routing | URL → content candidate resolution |
| `internal/publicrender` | Public DTOs | Theme-facing page model (breaks import cycles) |
| `internal/themes` | Public UI | Theme registry, manifests, templ views |
| `internal/extensions` | Plugins | Compiled-in plugin runtime and hook bus |
| `internal/plugins/*` | Plugins | GraphQL, operations/import-export |
| `internal/operations` | Ops | Audit store, diagnostics, health checks, snapshots |
| `web/static` | Assets | CSS, theme static files |

## Platform packages used by AppCMS

| Platform package | Purpose |
|------------------|---------|
| `pkg/contracts` | `Module`, `StoragePort`, `CapabilityID`, `RuntimeContext`, hooks, audit types |
| `pkg/toolset` | Record types, relations, review/diff descriptors (not an ORM) |
| `pkg/panel` | Admin resource registry, policies, datasources |
| `pkg/panelschema` | View/workflow/action descriptors |
| `pkg/render` | `ScreenModel`, view types for generic admin rendering |
| `pkg/renderers/templ` | Platform templ screen renderer |
| `pkg/profile` | Deployment profile: admin/API/public bases, workspaces |
| `pkg/appbundle` | Static bundle export for composed apps |
| `pkg/modulehost` | Module registration host (used in tests and composition) |
| `pkg/codex` | Profile/conformance helpers |

## Framework packages used by AppCMS

| Framework package | Purpose |
|-------------------|---------|
| `pkg/app` | `AppBuilder`, config, features, graceful shutdown |
| `pkg/auth` | `CookieSession`, signed action tokens |
| `pkg/web/security` | Security headers, method guards, config loading |
| `pkg/web/middleware` | Recover, request ID, logging (via builder) |
| `pkg/web/health` | `/healthz`, `/readyz` integration |

## What stays in GoCMS (oracle only)

Use GoCMS when you need the **normative contract** or legacy reference implementation:

- `go-codex/en/*` — plugin, theme, hooks, REST, admin contracts
- `internal/domain`, `internal/delivery` — reference behavior until AppCMS parity is explicit

Do not copy UI8Kit admin patterns into AppCMS.

## Composition with AppSuite

AppSuite imports `pkg/app.Bundle()` from AppCMS and mounts it under:

- `/go-admin/spaces/{space}/...`
- `/go-json/spaces/{space}/...`

AppCMS internals should not assume they only run standalone; screens should accept render context metadata for launcher vs standalone themes (see bundle tests).
