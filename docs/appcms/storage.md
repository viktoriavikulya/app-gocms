# Storage

AppCMS persists data through Platform **`contracts.StoragePort`**, adapted to typed repositories for application services.

## Architecture

```text
application.Service
    → storage.ApplicationRepositories (adapters)
        → storage.RecordRepository
            → contracts.StorageTx
                → sqlite.Store (today)
```

Products depend on the **port**, not on SQLite types, in `pkg/app` and tests.

## StoragePort API

```go
type StoragePort interface {
    WithinWorkspaceTx(ctx, workspace, fn) error
}

type StorageTx interface {
    List(ctx, Query) (PageResult, error)
    Get(ctx, recordType, id) (Record, error)
    Put(ctx, recordType, id, Record) error
    Delete(ctx, recordType, id) error
}
```

Records are `map[string]any` documents serialized to JSON in SQLite.

## Workspace model

Every row is scoped by `workspace_id` (string). Default standalone workspace: `"root"`.

This supports future AppSuite multi-workspace databases without changing service signatures.

## Record families

| Record type | Table | Domain adapter |
|-------------|-------|----------------|
| post | `posts` | `domain/content` |
| page | `pages` | `domain/content` |
| content_type | `content_types` | `domain/contenttype` |
| taxonomy / term | `taxonomies`, `terms` | `domain/taxonomy` |
| media_asset | `media_assets` | `domain/media` |
| author | `authors` | `domain/users` |
| setting | `settings` | `domain/settings` |
| menu | `menus` | `domain/menus` |
| revision | `revisions` | `domain/revisions` |
| preview | `preview_access` | `domain/preview` |

Auth tables (`auth_users`, `auth_app_tokens`, `auth_login_attempts`) are separate from `StoragePort` today; authn uses dedicated SQLite access in migrations.

## Migrations

`internal/storage/sqlite/migrations.go`:

- `0001_cms_schema` — CMS record tables + indexes
- `0002_auth_schema` — auth tables

`Store.Init` applies pending migrations. `MigrationStatus` reports drift.

## Seeding

`storagesqlite.SeedMinimalSite` inserts demo content for tests and local dev:

- Sample post (`hello-world`), page (`about`)
- Site settings (`site.title`, `theme.active`, …)
- Primary menu

Enable with `Options.Seed: true` in tests or custom hosts.

## Configuration

| Option | Effect |
|--------|--------|
| `Options.Storage` | Inject custom `contracts.StoragePort` (tests, future drivers) |
| `Options.StorageDSN` | SQLite DSN (default in-memory shared) |
| `Options.Seed` | Run minimal site seed after init |

Default DSN for tests: `file:appcms?mode=memory&cache=shared`

## Injecting storage in tests

```go
gocmsapp.NewApp(gocmsapp.Options{
    Storage:   myPort,
    Seed:      true,
    AuthStore: authStore,
})
```

## Multi-storage profiles (future)

The **interface boundary is ready**:

- Platform `StoragePort` is driver-agnostic
- AppCMS services use `StoreProvider`, not `sql.DB`

Still required for Postgres/MySQL:

| Piece | Status |
|-------|--------|
| `internal/storage/postgres` adapter | Not yet |
| `internal/storage/mysql` adapter | Not yet |
| Dialect-neutral migrations | Not yet (SQL is SQLite-specific) |
| `StorageDriver` config in `pkg/app` | Not yet |
| Cross-driver contract tests | Not yet |

Optional Platform extensions (only if needed): `Migrator`, `HealthChecker` interfaces—keep CRUD on `StoragePort`.

## Snapshots

`internal/operations` exports/import `gocms.snapshot.v1` JSON bundles via the operations plugin, reading/writing through the same repositories—not a separate dump format.

## Performance note

Current list operations load JSON payloads per workspace and paginate in memory for simplicity. Driver-specific optimizations (SQL `WHERE`, JSONB indexes) belong in future adapters without changing application services.
