# Domain Model (GoCMS Oracle Reference)

> Summarized from `@GoCMS/go-stack/en/01-domain-model.md`.

## Package boundaries

Recommended domain packages:

```text
content, taxonomy, media, users, authz, settings, menus, themes,
plugins, hooks, revisions, search
```

Domain packages define entities, value objects, typed constants, validation
invariants, and domain errors. They do not own storage, HTTP handlers, rendering,
or session logic.

AppCMS mirrors this layout under `internal/domain/` and `internal/application/`.

## Content

Represents posts, pages, custom kinds, statuses, slugs, excerpts, metadata, and
publish windows.

Stable kinds: `post`, `page`.

Stable statuses: `draft`, `scheduled`, `published`, `archived`, `trashed`.

Codex schema: `content-entry.schema.json`, `content-type.schema.json`.

## Taxonomy

Taxonomy types, terms, hierarchy, localized names, slugs, assignments.

Default types: `category`, `tag`.

## Media

Assets, variants, upload metadata, ownership, captions, alt text, resolver keys.
URLs resolved through a service — not string concatenation in domain objects.

## Users

Identities, profile data, account state, login metadata. Password hashes and
sessions separated from public profile data.

## Authorization

Capabilities, roles, grants, ownership checks, policy decisions. Capabilities are
the real permission boundary; roles are named capability sets.

## Settings

Typed settings with visibility, validation, groups, defaults. Public and private
settings distinguishable at domain level.

## Themes

Manifests, template roles, slots, assets, settings, activation. Theme entities do
not encode content visibility rules.

## Plugins

Manifests, lifecycle state, dependencies, migrations, capabilities, routes, hooks,
jobs, assets. State transitions explicit and auditable.

## Hooks

Hook IDs, handler descriptors, priority, category, arguments, failure policy.

## Domain invariants

Examples enforced in domain or services:

- Invalid statuses rejected
- Content kind identifiers normalized
- Slugs non-empty when publishable
- Public settings cannot contain private values
- Plugin IDs URL-safe
- Capability identifiers stable

Invariants requiring storage lookups belong in application services.
