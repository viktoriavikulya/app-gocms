# REST API Contract (GoCMS Oracle Reference)

> Summarized from `@GoCMS/go-codex/en/02-rest-api-contract.md`. Normative JSON
> shapes are defined in `@AppCMS/schema/codex/v1/`.

## Purpose

REST is the base compatibility and control-plane API. It remains required even
when GraphQL is the primary public content API.

## Required base paths

```text
/go-json
/go-json/go/v2/
```

- `GET /go-json` — discovery for available namespaces
- `GET /go-json/go/v2/` — discovery for the GoCMS v2 namespace

## Discovery shape

Discovery responses include:

- `name`
- `version`
- `routes`
- `authentication`
- `links`

Additional fields may appear. Clients must ignore unknown fields.

Schema: `envelope.schema.json` (`discovery` definition).

## Required read endpoints (v2)

Posts, pages, media, taxonomies, menus, settings, and search endpoints as listed
in the original GoCMS contract. AppCMS implements the core subset today; see
`pkg/module/codex/routes.go` for the mounted route inventory.

## Authentication and authorization

- Public reads for public resources unless configured private/headless-private
- Authenticated requests mapped to a principal before capability checks
- Every endpoint enforces capabilities server-side
- Plugin routes use the same authorization model

## Pagination

List endpoints support `page` and `per_page`. Responses include:

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 120,
    "total_pages": 6
  }
}
```

Schema: `envelope.schema.json` (`listEnvelope`, `pagination`).

## Error shape

```json
{
  "error": {
    "code": "validation_error",
    "message": "The request is invalid.",
    "status": 400,
    "details": {},
    "request_id": "req_123"
  }
}
```

Required: `code`, `message`, `status`. Optional: `details`, `request_id`.

Stable codes include: `not_found`, `validation_error`, `unauthorized`,
`forbidden`, `conflict`, `rate_limited`, `unsupported_media_type`,
`payload_too_large`, `internal_error`.

Schema: `error.schema.json`.

## Resource envelope

Single resource:

```json
{ "data": {} }
```

List resource:

```json
{ "data": [], "pagination": {} }
```

Implementations may include `links` and `meta`.

## Content resource fields

Post and page resources include at minimum:

- `id`, `kind`, `status`, `slug`, `title`, `content`, `excerpt`
- `author_id`, `featured_media_id`, taxonomy assignment
- `created_at`, `updated_at`, `published_at`

GoCMS also documents `deleted_at`, `template`, and `metadata`. See
[03-content-contract.md](03-content-contract.md) and `content-entry.schema.json`.

Private fields are returned only when authorized.

## Locale behavior

Locale selection via query parameter, `Accept-Language`, or documented equivalent.
Fallback behavior must be stable and documented.

## Conformance expectations

- Discovery exists
- Required routes exist
- Pagination metadata is stable
- Unauthorized draft access is blocked
- Error shape is stable
- Unknown JSON fields do not break clients
