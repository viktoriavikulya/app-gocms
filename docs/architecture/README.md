# Architecture Overview

AppCMS follows a **product-owned domain, shared host** model. Shared infrastructure lives in Framework and Platform; CMS-specific behavior stays in AppCMS `internal/` and `pkg/module`.

## Design principles

1. **GoCMS codex wins for public behavior** — URLs, REST envelopes, admin entry points, and capability names must stay compatible unless a deliberate codex change is agreed.
2. **Platform stays product-neutral** — Platform must not import `app-gocms`, `app-crm`, or `app-suite`. Products implement `contracts.Module`.
3. **No UI8Kit in product apps** — Admin and public UI use `github.com/fastygo/templ` and product `pkg/templ` shells.
4. **Storage behind a port** — Application services talk to `storage.StoreProvider`, which adapts `contracts.StoragePort`. SQLite is the first driver; Postgres/MySQL are future adapters on the same port.
5. **Bundles for composition** — `pkg/app.Bundle()` exports a reusable profile + module mount for **AppSuite** without duplicating CMS templates in the launcher.

## Layer diagram

```text
┌─────────────────────────────────────────────────────────────┐
│  cmd/server          Thin entry → pkg/app.Run()             │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│  pkg/app             Framework App, routes, auth boundary,    │
│                      runtime mode, extension activation       │
└────────────┬───────────────────────────────┬────────────────┘
             │                               │
   ┌─────────▼─────────┐           ┌─────────▼─────────┐
   │  pkg/module       │           │  pkg/templ        │
   │  descriptors      │           │  product shells   │
   └─────────┬─────────┘           └─────────┬─────────┘
             │                               │
   ┌─────────▼───────────────────────────────▼─────────┐
   │  internal/                                         │
   │  domain → application → storage → delivery         │
   │  extensions, plugins, themes, operations           │
   └─────────┬──────────────────────────────────────────┘
             │
   ┌─────────▼─────────┐     ┌──────────────┐     ┌─────────────┐
   │  Platform         │     │  Framework   │     │  Templ      │
   │  contracts,       │     │  HTTP, auth, │     │  UI atoms   │
   │  toolset, panel,  │     │  security,   │     │  components │
   │  render, profile  │     │  health      │     │             │
   └───────────────────┘     └──────────────┘     └─────────────┘
```

## Request paths (summary)

| Surface | Package | Notes |
|---------|---------|-------|
| Public site | `internal/delivery/publicsite` | Permalinks → assembler → theme render |
| REST API | `internal/delivery/rest` | Codex v2 routes, capability-gated mutations |
| Admin HTML | `internal/appschema` + `pkg/templ` | Schema-driven screens via Platform render models |
| Auth | `pkg/app` + `internal/application/authn` | Cookie sessions (Framework) + RBAC |
| Extensions | `internal/extensions`, `internal/plugins/*` | Compiled-in plugins, active-only routes |

Detailed flow: [Request flow](request-flow.md).

## Further reading

- [Layers and repositories](layers-and-repos.md)
- [Ecosystem packages](../ecosystem/README.md)
- [AppCMS internal layers](../appcms/internal-layers.md)
