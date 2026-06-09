# Content Contract (GoCMS Oracle Reference)

> Summarized from `@GoCMS/go-codex/en/03-content-contract.md`. Normative JSON
> shapes are defined in `content-entry.schema.json`.

## Purpose

Content resources are shared by admin screens, REST, themes, plugins, search,
imports, previews, feeds, and optional GraphQL.

## Content kinds

Required kinds:

- `post`
- `page`

Custom kinds may be added with stable identifiers. They must not redefine `post`
or `page`.

## Statuses

Publishable content supports:

| Status | Public visibility rule |
|--------|------------------------|
| `draft` | Must not appear in public unauthenticated output |
| `scheduled` | Must not appear before publish timestamp |
| `published` | May appear when visibility allows |
| `archived` | Must not appear in normal public listings unless authorized |
| `trashed` | Must not appear publicly; restorable until permanent delete |

## Required fields

Content entries must have:

- `id`, `kind`, `status`, `slug`, `title`, `content`, `excerpt`
- `author_id`, `created_at`, `updated_at`, `published_at`
- `featured_media_id`, `metadata`

GoCMS also requires `deleted_at` and `template` in the contract. Codex v1
includes them as optional until all adapters populate them.

AppCMS domain today: `internal/domain/content/content.go`.

## Localized fields

When localization is enabled, implementations should localize:

- `title`, `slug`, `content`, `excerpt`
- SEO fields (`seo_title`, `seo_description`) when supported

Localized fields need deterministic fallback behavior. AppCMS serializes `title`
as a locale map in JSON fixtures and REST output.

## Slugs

- Normalized consistently
- Unique within content kind and routing scope
- Page slugs may be scoped to parent when hierarchical routing is enabled
- Slug changes should create redirects when public rendering is enabled

## Pages and posts

**Pages** — hierarchical evergreen content with optional template, menu inclusion,
preview, revisions, featured media.

**Posts** — chronological content with author, excerpt, featured media, taxonomy,
archive listing, preview, revisions.

## Metadata

Key-value extension surface:

- Stable string keys
- Typed values where possible
- Private metadata must not leak publicly
- Plugin keys should be plugin-prefixed

## Revisions, autosave, preview

- Revisions required for `post` and `page`
- Autosave should not become public without explicit publish
- Preview must not leak unrelated drafts

## Visibility

Additional states such as `private` and `password` must be enforced consistently
across admin, REST, themes, search, and feeds.

Schema enum: `public`, `private`, `password`.

## Taxonomy assignment

Posts must support taxonomy assignment. Pages may support it. Assignments preserve
taxonomy type, term ID, content ID, and ordering.

AppCMS serializes assigned term IDs as `term_ids` (GoCMS docs use
`taxonomy_ids` — both accepted in Codex v1 during compatibility).

## Lifecycle hooks

Create, update, status transition, trash/delete, and restore events should emit
or expose hooks per the hooks contract.

## Conformance expectations

- Drafts are not public
- Scheduled content hidden before publish time
- Published content appears in REST when authorized
- Slug lookup is stable
- Revisions restorable
- Preview does not leak unrelated drafts
- Unauthorized metadata hidden
