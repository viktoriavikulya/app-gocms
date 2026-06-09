# AppCMS Codex v1 JSON Schema

Contract version: `codex.v1`

These schemas are the **canonical machine-readable contract** for AppCMS CMS
resource data. Serialized Codex payloads — REST responses, GraphQL output,
storage seed bundles, import fixtures, and BuildY CMS seeds — must conform to
these schemas.

Go domain structs, REST DTOs, and storage adapters are **implementation
artifacts**. They must mirror Codex schemas; schemas must not be inferred from
accidental code drift.

## Ownership

| Layer | Owner | Role |
|-------|-------|------|
| Codex schemas | `@AppCMS/schema/codex/v1` | Single source of truth for CMS data shapes |
| Go domain | `internal/domain/*` | In-process entities and validation |
| REST adapters | `internal/delivery/rest` | HTTP mapping to Codex envelopes |
| GraphQL adapter | `internal/plugins/graphql` | Optional delivery; same semantics as REST |
| Storage seeds | `internal/storage/sqlite/seed.go` | Persistence bootstrap |
| BuildY CMS fixtures | `@BuildY/internal/fixtures/cms/` (future) | Prototype content seeds |
| Platform BFF | `@Platform/schema/bff/v1` | Renderer-facing only; **not** CMS resources |

## Files

| Schema | Go source | Primary consumers |
|--------|-----------|-------------------|
| `content-entry.schema.json` | `internal/domain/content/content.go` | REST posts/pages, seeds, GraphQL nodes |
| `content-type.schema.json` | `internal/domain/contenttype/contenttype.go` | REST content-types, built-in install |
| `menu.schema.json` | `internal/domain/menus/menus.go` | REST menus, public nav assembly |
| `envelope.schema.json` | `pkg/module/codex/routes.go` | Discovery, list/resource envelopes, pagination |
| `error.schema.json` | `pkg/module/codex/routes.go` (`ErrorEnvelope`) | REST error responses |
| `seed-site.schema.json` | — | JSON seed bundles, fixture imports, BuildY CMS seeds |

Historical GoCMS oracle docs live under [docs/compat/gocms/](../../docs/compat/gocms/).
They explain compatibility intent; they are **not** the machine-readable contract.

## Boundaries

- **Do not** move CMS resource shapes into `@Platform/schema/bff/v1`.
- **Do not** let GraphQL define independent data semantics.
- **Do not** treat removed `@GoCMS` docs as the long-term schema source.
- **Do not** embed BuildY prototype UI shapes in Codex schemas.

BFF describes renderer screens (tables, forms, nav chrome). Codex describes CMS
resources (content entries, menus, content types).

## JSON casing

Codex resource schemas follow Go `json` tags on domain structs:

- Resource fields use **snake_case** (`author_id`, `featured_media_id`, `term_ids`).
- Envelope wrappers use snake_case (`per_page`, `total_pages`, `request_id`).

GoCMS REST docs sometimes refer to `taxonomy_ids`. AppCMS currently serializes
`term_ids`. Both names are accepted in `content-entry.schema.json` during the
compatibility window; new fixtures should prefer `term_ids`.

## GoCMS compatibility fields

Schemas include fields expected by the GoCMS content contract even when current
Go structs omit them, so adapters can grow without a schema version bump:

- `deleted_at` — soft delete / trash restoration
- `template` — page template selection
- Localized `title` as `object` (locale map) or plain `string`

Optional fields remain optional until domain and adapters populate them.

## Versioning

### `codex.v1` policy

- **Additive-only** within v1: new optional fields and fixture growth when Go
  structs, JSON Schema, golden fixtures, and validators stay aligned.
- **Breaking changes** require `codex.v2`: field removal, rename, semantic change,
  or making optional fields required.
- Unknown JSON fields MUST be ignored by clients; servers SHOULD preserve
  unknown metadata keys where storage allows.

## Adapter alignment

| Adapter | Alignment target | Status |
|---------|------------------|--------|
| REST list/resource responses | `envelope.schema.json` + resource schema | Validated by conformance tests |
| REST errors | `error.schema.json` | Validated by conformance tests |
| Domain structs | Resource schemas | Round-trip JSON validation in tests |
| SQLite seeds | `seed-site.schema.json` | Example bundle fixture; flat store records map logically |
| GraphQL | Same resource schemas | Resolver output must match REST semantics |
| BuildY CMS fixtures | `seed-site.schema.json` | Future `internal/fixtures/cms/*.json` |

Implementation code may lag schemas briefly, but new work must close gaps toward
validation rather than introduce parallel shapes.

## Golden fixtures

Example payloads under `fixtures/`:

| Fixture | Schema |
|---------|--------|
| `content-entry-post.json` | `content-entry.schema.json` |
| `content-entry-page.json` | `content-entry.schema.json` |
| `content-type-post.json` | `content-type.schema.json` |
| `menu-primary.json` | `menu.schema.json` |
| `list-posts.json` | `envelope.schema.json` |
| `resource-post.json` | `envelope.schema.json` |
| `error-not-found.json` | `error.schema.json` |
| `seed-minimal-site.json` | `seed-site.schema.json` |

Run validation:

```bash
go test ./pkg/conformance/codex/...
```

## Related docs

- [GoCMS compatibility references](../../docs/compat/gocms/README.md)
- [Migration from GoCMS](../../docs/migration-from-gocms.md)
- [GoCMS oracle (ecosystem)](../../docs/ecosystem/gocms-oracle.md)
- [Platform BFF v1](../../../@Platform/schema/bff/v1/README.md) — separate renderer contract
