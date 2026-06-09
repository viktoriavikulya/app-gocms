# Codex Schema Alignment

AppCMS adapters must serialize CMS data through `@AppCMS/schema/codex/v1/` — not
parallel ad hoc JSON shapes.

## Alignment matrix

| Surface | Package / path | Codex schema | Notes |
|---------|----------------|--------------|-------|
| REST posts/pages | `internal/delivery/rest` | `content-entry.schema.json` + `envelope.schema.json` | Domain `Entry` is the DTO |
| REST menus | `internal/delivery/rest` | `menu.schema.json` + `envelope.schema.json` | `menus.Menu` |
| REST content-types | `internal/delivery/rest` | `content-type.schema.json` | Built-ins from `contenttype` domain |
| REST errors | `internal/delivery/rest/errors.go` | `error.schema.json` | `codex.ErrorEnvelope` |
| REST discovery | `pkg/module/codex/routes.go` | `envelope.schema.json` | `Discovery`, routes, links |
| GraphQL | `internal/plugins/graphql` | Same resource schemas as REST | **Gap:** returns storage records today |
| SQLite seed | `internal/storage/sqlite/seed.go` | `seed-site.schema.json` | Logical bundle: `fixtures/seed-minimal-site.json` |
| BuildY CMS fixtures | `@BuildY/internal/fixtures/cms/` (planned) | `seed-site.schema.json` | UI locale JSON stays separate from Codex |
| Platform BFF | `@Platform/schema/bff/v1` | **Out of scope** | Renderer screens only |

## Validation

Golden fixtures live in `schema/codex/v1/fixtures/`. Conformance tests in
`pkg/conformance/codex/` validate:

1. Fixture files against schemas
2. Marshaled domain types and codex envelopes against schemas
3. Seed bundle inventory against the SQLite minimal seed set
4. Representative REST `/go-json/go/v2/*` responses against Codex envelopes and resources

```bash
go test ./pkg/conformance/codex/...
```

## GraphQL boundary

`internal/plugins/graphql/plugin.go` currently returns `contracts.Record` slices
from storage repositories. Those payloads omit Codex-required fields such as `id`
and `kind`, so they do **not** yet validate against `content-entry.schema.json`.

New GraphQL work must map through the same domain types and DTOs as REST before
Codex can be considered enforced on that surface.

## Known gaps (codex.v1 migration)

These GoCMS contract fields are in schemas but not yet on all domain structs or
storage paths:

| Field | Schema | Domain / storage today |
|-------|--------|------------------------|
| `deleted_at` | `content-entry.schema.json` | Not on `content.Entry` yet |
| `template` | `content-entry.schema.json` | Optional in fixtures only |
| `taxonomy_ids` | alias of `term_ids` | AppCMS uses `term_ids` |

New adapter work should close gaps toward schema validation rather than add
alternate JSON keys.

## Schema strictness path

`content-entry.schema.json` currently requires only `id`, `kind`, `status`, and
`visibility` so existing REST and domain output stays valid during migration.

Do **not** promote these to required until all adapters populate them consistently:

- `slug`, `title`, `content`, `excerpt`
- `created_at`, `updated_at`, `published_at`
- `deleted_at`, `template`

When promotion happens, update domain structs, REST handlers, GraphQL resolvers,
seed bundles, and conformance tests in the same change set.

## Compatibility status

AppCMS implements the core GoCMS URL and REST envelope contract. Remaining product
gaps vs full GoCMS parity are tracked in
[migration mapping](../compat/gocms/migration-mapping.md) and Platform
`.project/current-progress.md`.

## BuildY boundary

BuildY prototype fixtures (`internal/fixtures/locale/`) are UI copy — not Codex
CMS resources. When BuildY adds CMS seed JSON, place files under
`internal/fixtures/cms/` and validate with `seed-site.schema.json`.

Do not embed BuildY shell or screen props in AppCMS Codex schemas.

## Related

- [Codex v1 README](../../schema/codex/v1/README.md)
- [GoCMS compatibility references](../compat/gocms/README.md)
- [Migration mapping](../compat/gocms/migration-mapping.md)
