# Storage And Repositories (GoCMS Oracle Reference)

> Summarized from `@GoCMS/go-stack/en/03-storage-repositories.md`.

## Purpose

Storage is an implementation detail. Domain and application services depend on
repository interfaces (ports), not concrete drivers.

AppCMS: `internal/storage/` with SQLite implementation and Platform workspace
transactions.

## Repository ownership

Interfaces live near consuming application services. Concrete implementations live
in infrastructure packages.

Examples:

- `ContentRepository` near content services
- `MediaRepository` near media services
- `SettingsRepository` near settings services

## Required repository capabilities

- Lookup by ID
- Lookup by slug where applicable
- List with filters and pagination
- Create, update
- Soft delete/trash and restore where applicable
- Permanent delete where applicable
- Transaction participation

## Transactions

Application provides a transaction boundary that:

- Accepts `context.Context`
- Commits on success, rolls back on error
- Propagates transaction-scoped repositories

Do not leak driver transaction types into domain packages.

## IDs and timestamps

- Stable IDs across API responses
- `created_at` immutable after create
- `updated_at` changes on mutation
- `published_at` reflects visibility rules
- `deleted_at` supports restoration when soft deletion is used

API serialization must remain stable regardless of internal ID type.

Codex timestamps: RFC3339 strings in JSON (`format: date-time`).

## Metadata storage

Supports content, user, term, media, and plugin metadata with public/private
visibility. Frequently queried metadata needs a documented indexing strategy.

## Migrations

Versioned, ordered, repeat-safe, logged, recoverable where possible. Plugin
migrations separated from core but integrated into activation workflows.

AppCMS: `internal/storage/sqlite/migrations.go`.

## Search storage

Search indexes are derived data:

- Rebuildable from source content
- Draft/private content must not leak into public indexes
- Updates via service workflows or background jobs

## Media storage

Separate binary storage, media metadata, variant metadata, and URL resolution.
Use a resolver service for public URLs.

## Seeds and fixtures

Conformance and prototype seeds should prefer documented public APIs or Codex seed
bundles (`seed-site.schema.json`) over ad hoc database writes.

AppCMS minimal seed: `internal/storage/sqlite/seed.go` — maps logically to
`schema/codex/v1/fixtures/seed-minimal-site.json`.

## Backup and restore

Document database, media, plugin data, theme settings backup, restore order, and
consistency expectations for production deployments.
