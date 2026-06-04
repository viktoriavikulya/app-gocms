# Related Product Apps

AppCMS shares architecture with other FastyGo product repositories. This page explains how they relate without duplicating their domain documentation.

## AppCRM (`github.com/fastygo/app-crm`)

| Aspect | AppCRM | AppCMS |
|--------|--------|--------|
| Domain | Leads, CRM pipelines | Posts, pages, taxonomies, media |
| Module ID | `crm` | `cms` |
| Capability prefix | `crm.*` | `content.*`, `admin.access`, … |
| Default routes | `/go-admin`, `/go-json` (profile-ready for `/y-admin`) | Same root codex URLs |
| Stack | Framework + Platform + Templ | Same |

AppCRM validates that Platform abstractions work for **non-CMS** products. Shared code should be promoted to Platform or a future shared package only after **real duplication** appears— not preemptively.

## AppSuite (`github.com/fastygo/app-suite`)

AppSuite is the **launcher and composer**:

- Imports AppCMS and AppCRM bundles via `pkg/compose`
- Owns workspace directory, profile selection, spaces overlay routes
- Does **not** own CMS/CRM domain logic or duplicate product templates

Spaces overlay (beside root contracts):

- `/go-admin/spaces/{space}/...`
- `/go-json/spaces/{space}/...`

AppCMS `pkg/app.Bundle()` must keep working in:

1. Standalone mode (`cmd/server`)
2. Composed AppSuite profile (different theme context, workspace in path)

## ModuleMonitoring (`@ModuleMonitoring`)

Optional monitoring module mounted in AppSuite for health/telemetry demos. AppCMS operations plugin (`/go-json/go/v2/ops/health`) is **product-local** CMS diagnostics, not a replacement for suite-level monitoring modules.

## GoCMS monolith (`@GoCMS`)

Still the behavioral reference and codex home. New features land in AppCMS first for the product stack; GoCMS receives selective backports only when maintaining dual stacks is required.

## Choosing where to implement a feature

| Feature type | Repository |
|--------------|------------|
| CMS content model, REST, themes | AppCMS |
| CRM leads, stages | AppCRM |
| Launcher chrome, workspace switcher | AppSuite |
| Generic module contract, render model | Platform |
| HTTP session, middleware | Framework |
| Shared button/table component | Templ |

## Cross-product storage and auth

Platform `StoragePort` and session claims support:

- Single-product DB (AppCMS alone with SQLite)
- Composed profile (AppSuite with workspace-scoped access)

Each product owns its schema migrations and authn service; Platform provides ports and policy evaluation shapes only.
