# Migration From Current GoCMS

This document describes the first Platform-based GoCMS replacement path.

For the full AppCMS guide, start at [docs/README.md](README.md).

Current GoCMS remains the compatibility oracle for behavior. AppCMS is the runnable
replacement assembly. The CMS module in `pkg/module` owns domain schemas and descriptors.

**Data contract:** `@AppCMS/schema/codex/v1/` JSON schemas are the canonical source
of truth for serialized CMS data. GoCMS prose references live under
[docs/compat/gocms/](compat/gocms/). Adapter alignment:
[codex-schema-alignment.md](appcms/codex-schema-alignment.md).

## Architecture Mapping

| Current GoCMS area | Replacement owner |
| --- | --- |
| `internal/domain/content` | `module-cms/records` posts and pages |
| `internal/domain/contenttype` | `module-cms/records` content types |
| `internal/domain/meta` and metadata registry | `module-cms/records` content meta definitions |
| `internal/domain/taxonomy` | `module-cms/records` taxonomies and terms plus relations |
| `internal/domain/media` | `module-cms/records` media assets |
| `internal/domain/users` public author projection | `module-cms/records` authors |
| `internal/platform/cmspanel` | `module-cms/panels` descriptors |
| `internal/delivery/rest` | AppCMS `/go-json` codex routes backed by ModuleCMS helpers |
| `internal/delivery/admin` | AppCMS `/go-admin` shell plus Platform Templ renderer |

## Preserved URL Contracts

The replacement keeps:

- `/go-admin`
- `/go-login`
- `/go-logout`
- `/go-json`
- `/go-json/go/v2/`

Additional workspace overlays can be added later under `/go-admin/spaces/*` and
`/go-json/spaces/*`, but they do not replace the root GoCMS contracts.

## UI Migration

The replacement admin does not import UI8Kit.

- AppCMS uses Platform's official Templ renderer.
- ModuleCMS provides Panel schemas and descriptors.
- AppCMS owns Tailwind input, tokens, static files, and product shell.

## Current Compatibility Level

AppCMS has progressed through product stack slices including Framework host,
SQLite storage, application services, REST CRUD, templ admin UI, auth/RBAC,
public themes, and compiled-in extensions (GraphQL, operations, snapshots).

Remaining gaps vs full GoCMS (see Platform `current-progress.md`):

- Postgres/MySQL storage drivers (SQLite is production path today)
- Plugin marketplace and dynamic loading
- Full GraphQL schema and mutation parity
- Binary media pipeline and marketplace packages

For architecture details see [AppCMS package layout](appcms/package-layout.md) and
[extensions](appcms/extensions-and-operations.md).

## Commands

```bash
cd e:/_@Go/@ModuleCMS
go test ./...

cd e:/_@Go/@AppCMS
bun verify
bun go
```
