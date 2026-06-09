# AppCMS Product Guide

This section documents **this repository** (`github.com/fastygo/app-gocms`): packages, internal layers, routes, auth, storage, extensions, and public site.

## Documents

| Guide | Topics |
|-------|--------|
| [Package layout](package-layout.md) | `cmd/`, `pkg/`, `internal/`, `web/` |
| [Internal layers](internal-layers.md) | Domain, application, storage, delivery |
| [URL contracts](url-contracts.md) | Admin, REST, public, extension routes |
| [Auth and capabilities](auth-and-capabilities.md) | Sessions, roles, app tokens, RBAC |
| [Storage](storage.md) | StoragePort, SQLite, repositories, future drivers |
| [Public site and themes](public-site-and-themes.md) | Permalinks, assembler, theme registry |
| [Extensions and operations](extensions-and-operations.md) | Plugins, GraphQL, snapshots, health |
| [Runtime profiles](runtime-profiles.md) | full, headless, admin, conformance |
| [Codex schema alignment](codex-schema-alignment.md) | JSON schemas, adapter mapping, validation |

## Module summary (`pkg/module`)

The CMS **Platform module** (`Module` struct) registers:

| Record families | Examples |
|-----------------|----------|
| Content | posts, pages, content types, meta definitions |
| Taxonomy | taxonomies, terms |
| Media | media assets |
| Users (public projection) | authors |
| App-owned (storage) | settings, menus, revisions, preview (via SQLite tables) |

Capabilities (namespace-safe IDs):

| ID | Label (short) |
|----|----------------|
| `admin.access` | Enter admin |
| `content.read` | Read content |
| `content.write` | Create/update content |
| `content.read_private` | See private/draft via API |
| `media.upload` / `media.edit` | Media mutations |
| `taxonomies.manage` / `taxonomies.assign` | Taxonomy admin |
| `users.manage` | Authors |
| `settings.manage` | Settings and snapshot ops |

## Executable entry

`cmd/server/main.go` calls `pkg/app.Run()` with default options. Tests build apps via `gocmsapp.NewApp(Options{...})` for seeded SQLite and auth stores.

## Bundle export

`pkg/app.Bundle()` returns `appbundle.StaticBundle` for embedding in AppSuite. `DefaultProfile()` defines workspace `root` and panel mount at `/go-admin`.

## Where business logic lives

| Layer | Location |
|-------|----------|
| Rules and invariants | `internal/domain/*` |
| Use cases | `internal/application/*` |
| Persistence | `internal/storage` + driver |
| HTTP | `internal/delivery/*` |
| Admin routing | `internal/appschema` |
| Plugins | `internal/extensions`, `internal/plugins/*` |

Platform `pkg/toolset` holds **schemas**; AppCMS services hold **behavior**.
