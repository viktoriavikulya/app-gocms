# Migration From Current GoCMS

This document describes the first Platform-based GoCMS replacement path.

Current GoCMS remains the compatibility oracle. AppCMS is the runnable
replacement assembly. ModuleCMS owns the CMS domain schemas and descriptors.

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

This slice establishes the replacement shape and route compatibility surfaces.
It includes discovery routes, list envelopes, admin route rendering, login/logout
surfaces, and ModuleCMS schema coverage.

It is not yet a full data migration or binary media migration. Storage adapters,
private content filtering at persistence level, CSRF enforcement, and full CRUD
will be implemented in later compatibility slices.

## Commands

```bash
cd e:/_@Go/@ModuleCMS
go test ./...

cd e:/_@Go/@AppCMS
bun verify
bun go
```
