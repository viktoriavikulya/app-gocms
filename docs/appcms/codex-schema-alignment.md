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
| GraphQL | `internal/plugins/graphql` | Same resource schemas as REST | Resolvers must not invent fields |
| SQLite seed | `internal/storage/sqlite/seed.go` | `seed-site.schema.json` | Logical bundle: `fixtures/seed-minimal-site.json` |
| BuildY CMS fixtures | `@BuildY/internal/fixtures/cms/` (planned) | `seed-site.schema.json` | UI locale JSON stays separate from Codex |
| Platform BFF | `@Platform/schema/bff/v1` | **Out of scope** | Renderer screens only |

## Validation

Golden fixtures live in `schema/codex/v1/fixtures/`. Conformance tests in
`pkg/conformance/codex/` validate:

1. Fixture files against schemas
2. Marshaled domain types and codex envelopes against schemas
3. Seed bundle inventory against the SQLite minimal seed set

```bash
go test ./pkg/conformance/codex/...
```

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

## BuildY boundary

BuildY prototype fixtures (`internal/fixtures/locale/`) are UI copy — not Codex
CMS resources. When BuildY adds CMS seed JSON, place files under
`internal/fixtures/cms/` and validate with `seed-site.schema.json`.

Do not embed BuildY shell or screen props in AppCMS Codex schemas.

## Related

- [Codex v1 README](../../schema/codex/v1/README.md)
- [GoCMS compatibility references](../compat/gocms/README.md)
