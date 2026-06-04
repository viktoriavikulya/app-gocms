# Extensions and Operations

Slice 8 adds a **compiled-in plugin runtime**, operational diagnostics, GraphQL extension route, and snapshot import/export—without a marketplace or dynamic loading.

## Plugin runtime

**Location:** `internal/extensions`

| Concept | Description |
|---------|-------------|
| `Plugin` interface | `Manifest()` + `Register(ctx, *Context)` |
| `Runtime` | Installs plugins, tracks state, exposes active routes |
| States | `installed`, `active`, `inactive`, `failed`, `uninstalled` |
| Activation | Failed registration rolls back routes; failed plugin marked `failed` |

Active plugins only expose routes, settings, and hooks collected in `Runtime.ActiveContext()`.

### Hook bus

| Type | Semantics |
|------|-----------|
| Actions | Ordered by priority; first error stops dispatch |
| Filters | Transform payload; type must match |

Use hooks for cross-cutting concerns (content projection, settings side effects). Keep business rules in application services.

Codex alignment: `go-codex/en/06-hooks-contract.md`.

## Built-in plugins

### GraphQL (`internal/plugins/graphql`)

| Route | Methods |
|-------|---------|
| `/go-graphql` | GET (query param), POST (JSON body) |

Baseline behavior:

- Simple query string matching (`posts`, `pages`, `settings`, `menus`)
- Public list filtering removes draft/private records
- Mutations return **403** (not enabled in this profile)

Future work: full schema, resolver auth parity with REST mutations, service-layer hooks.

### Operations (`internal/plugins/ops`)

| Route | Capability | Purpose |
|-------|------------|---------|
| `GET /go-json/go/v2/ops/health` | `admin.access` | Storage + plugin + audit checks |
| `GET /go-json/go/v2/ops/audit` | `admin.access` | Recent audit events |
| `GET /go-json/go/v2/ops/errors` | `admin.access` | Recent error records |
| `GET /go-json/go/v2/ops/snapshot` | `settings.manage` | Export `gocms.snapshot.v1` |
| `POST /go-json/go/v2/ops/snapshot` | `settings.manage` | Import snapshot |
| `GET /go-admin/import-export` | `settings.manage` | Minimal HTML UX |

## Operations store

**Location:** `internal/operations`

| Type | Purpose |
|------|---------|
| `AuditEvent` | Bounded in-memory audit trail with redaction |
| `ErrorRecord` | Recent application errors |
| `Health` | Aggregates checks (storage list, plugin state, stores) |
| `Snapshot` | Versioned export/import across record families |

Audit redacts map keys containing `token`, `secret`, `password`.

Framework `/healthz` remains process-level; CMS health is the ops plugin registry.

## Activation wiring

`pkg/app.extensionRuntime` builds:

```go
extensions.NewRuntime(
    graphqlplugin.New(provider),
    opsplugin.New(provider, opsStore, runtimeState),
)
runtime.Activate(ctx, "graphql", "operations")
```

Routes register on the same mux as REST with shared `authBoundary.Authorize`.

## Plugin contract reference

Normative rules: `go-codex/en/05-plugin-contract.md`

Tests in `internal/extensions/runtime_test.go` cover:

- Invalid manifest rejection
- Activate/deactivate
- Active-only routes
- Failed activation rollback
- Hook ordering

## What is explicitly out of scope

- Plugin marketplace
- Remote/dynamic plugin loading
- GraphQL as authorization bypass
- Moving snapshot format or GraphQL schema to Platform

## Extending with a new compiled-in plugin

1. Implement `extensions.Plugin` in `internal/plugins/{name}`
2. Validate manifest ID (lowercase URL-safe)
3. Register routes/settings/hooks in `Register`
4. Add to `extensionRuntime` plugin list
5. Activate in `Activate(...)` call
6. Add capability checks on sensitive routes
7. Integration test route + auth behavior
