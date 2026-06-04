# Internal Layers

AppCMS uses a classic **domain → application → infrastructure → delivery** split. This keeps HTTP and SQL out of business rules and keeps Platform descriptors separate from persistence.

## Layer responsibilities

```text
delivery (HTTP)
    ↓ calls
application (services)
    ↓ uses interfaces implemented by
storage (adapters + port)
    ↓ reads/writes
domain (types + validation)
```

## Domain layer

**Location:** `internal/domain/*`

Contains structs, enums, and validation helpers. Examples:

- Content `Status`: draft, published, scheduled, trashed, archived
- Content `Visibility`: public, private
- Settings value with key, group, public flag

Domain packages must remain **framework-agnostic**.

## Application layer

**Location:** `internal/application/*`

Each service:

1. Accepts a `context.Context`
2. Enforces invariants (e.g. content type must exist before save)
3. Calls repository interfaces (`Save`, `Get`, `List`, …)

Example flow — publish post:

```text
REST handler → storage.WithinTx → ApplicationRepositories
    → content.Service.Publish(ctx, id)
        → repository.Get / Save
```

### Authn service

`internal/application/authn` is special: it backs the HTTP auth boundary but is not a classic domain aggregate. It implements `contracts.Principal` and maps roles to `pkg/module` capability IDs.

## Storage layer

**Location:** `internal/storage`

### StoreProvider

```go
provider.ForWorkspace("root").WithinTx(ctx, func(ctx, repos) error { ... })
```

`Repositories` exposes typed record repos (`Posts`, `Pages`, `Settings`, …).

### ApplicationRepositories

`storage.NewApplicationRepositories(repos)` implements all application repository interfaces in one struct, converting:

- Domain struct → `contracts.Record` (JSON-friendly map) on write
- `contracts.Record` → domain struct on read

This avoids duplicating conversion logic in each service.

### SQLite driver

`internal/storage/sqlite`:

- One table per record family (`posts`, `pages`, …)
- `payload_json` column for document bodies
- `workspace_id` for multi-tenant readiness
- Versioned migrations in `migrations.go`

## Delivery layer

### REST (`internal/delivery/rest`)

- Registers `/go-json/` and `/go-json/go/v2/{path...}`
- Dispatches to resource handlers in `routes.go`
- Uses `Authorizer` callback for mutations
- Runs `contenttype.InstallBuiltIns` inside transactions where needed

### Public site (`internal/delivery/publicsite`)

- `Handler.ServeHTTP` — mode gating, theme render
- `assembler` — builds `publicrender.Page` from services

## Appschema (admin composition)

Not a classic delivery handler for all logic—`internal/appschema` maps URLs to `render.ScreenModel`:

- Resource tables and forms from module descriptors
- Special cases: settings/new, menus

Rendering is delegated to Platform templ renderer inside `pkg/templ.Page`.

## Extensions layer

Plugins register routes into `extensions.Runtime` at startup. Domain services remain unaware of GraphQL; the GraphQL plugin reads storage/repos similarly to REST list handlers.

Prefer hooks for cross-cutting filters (public content projection, REST shaping) rather than duplicating service logic in plugins.

## Testing strategy by layer

| Layer | Test style |
|-------|------------|
| Domain | Pure unit tests (validation) |
| Application | Service tests with fake repos |
| Storage | SQLite memory DSN, repository round-trips |
| Delivery | `cmd/server` HTTP integration tests |

## Adding a new feature (checklist)

1. Define or extend types in `internal/domain`
2. Add service methods in `internal/application`
3. Extend storage adapter conversions if new fields
4. Add migration + record type if new persistence family
5. Wire REST route and/or appschema screen
6. Register panel/toolset descriptors in `pkg/module` if admin-visible
7. Add capability if mutation requires new permission
8. Integration test in `cmd/server/main_test.go`
