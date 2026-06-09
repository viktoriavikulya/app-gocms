# GoCMS To AppCMS Migration Mapping

This document preserves the legacy GoCMS replacement mapping. For serialized CMS
data shapes, use [`schema/codex/v1/README.md`](../../../schema/codex/v1/README.md)
as the canonical contract — not this prose table.

AppCMS is the runnable replacement assembly. GoCMS remains the **behavior oracle**
for observable parity. The CMS module in `pkg/module` owns domain schemas and
panel descriptors.

## Architecture mapping

| Legacy GoCMS area | AppCMS owner |
| --- | --- |
| `internal/domain/content` | `internal/domain/content` + `internal/application/content` |
| `internal/domain/contenttype` | `internal/domain/contenttype` + `pkg/module/records` content types |
| `internal/domain/meta` and metadata registry | `pkg/module/records` content meta definitions |
| `internal/domain/taxonomy` | `internal/domain/taxonomy` + `pkg/module/records` taxonomies and terms |
| `internal/domain/media` | `internal/domain/media` + `pkg/module/records` media assets |
| `internal/domain/users` public author projection | `internal/domain/users` + `pkg/module/records` authors |
| `internal/platform/cmspanel` | `pkg/module/panels` descriptors |
| `internal/delivery/rest` | `internal/delivery/rest` — `/go-json` codex routes |
| `internal/delivery/admin` | `internal/appschema` + Platform render + `pkg/templ` |

## Preserved URL contracts

The replacement keeps:

- `/go-admin`
- `/go-login`
- `/go-logout`
- `/go-json`
- `/go-json/go/v2/`

Additional workspace overlays may be added under `/go-admin/spaces/*` and
`/go-json/spaces/*`, but they do not replace the root GoCMS contracts.

See [URL contracts](../../appcms/url-contracts.md) for runtime profile gating.

## UI migration

The replacement admin does not import UI8Kit.

- AppCMS uses Platform's official Templ renderer.
- ModuleCMS provides panel schemas and descriptors.
- AppCMS owns Tailwind input, tokens, static files, and product shell.

## Current compatibility level

AppCMS has progressed through product stack slices including Framework host,
SQLite storage, application services, REST CRUD, templ admin UI, auth/RBAC,
public themes, and compiled-in extensions (GraphQL, operations, snapshots).

Remaining gaps vs full GoCMS (see Platform `current-progress.md`):

- Postgres/MySQL storage drivers (SQLite is production path today)
- Plugin marketplace and dynamic loading
- Full GraphQL schema and mutation parity
- Binary media pipeline and marketplace packages

When AppCMS behavior intentionally differs from GoCMS, document the gap in
[product progress notes](../../ecosystem/gocms-oracle.md) or Platform
`.project/current-progress.md` — not by silently changing URLs.

## Related

- [Codex schema alignment](../../appcms/codex-schema-alignment.md)
- [AppCMS package layout](../../appcms/package-layout.md)
- [Extensions and operations](../../appcms/extensions-and-operations.md)
- [GoCMS compatibility references](README.md)
