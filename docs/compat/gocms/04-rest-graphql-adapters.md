# REST And GraphQL Adapters (GoCMS Oracle Reference)

> Summarized from `@GoCMS/go-stack/en/04-rest-graphql-adapters.md`.

## Core rule

REST and GraphQL are adapters over the **same application services**. They must
not own separate business rules, separate authorization, or direct storage
shortcuts.

```text
Admin  -> Services
REST   -> Services
GraphQL -> Services
CLI    -> Services
Plugins -> Services
```

AppCMS:

- REST: `internal/delivery/rest`
- GraphQL: `internal/plugins/graphql`
- Admin: `internal/appschema` + Platform render

## REST adapter responsibilities

REST adapters should:

- Parse HTTP input
- Authenticate principal
- Build commands or queries
- Call services
- Map results to stable DTOs (Codex envelopes + resource schemas)
- Map service errors to compatibility errors
- Apply cache headers after visibility decisions

REST adapters must not:

- Query storage directly
- Reimplement publish visibility
- Bypass capability checks
- Return storage models as responses

Base contract: [02-rest-api-contract.md](02-rest-api-contract.md).

## GraphQL adapter responsibilities

GraphQL is an optional extension. Resolvers must:

- Use the same services as REST
- Respect the same capabilities and visibility
- Reuse the same ID and status semantics
- Expose private fields only with authorization
- Support stable pagination

GraphQL schema extensions register through plugin or extension descriptors.

Output shapes must align with Codex resource schemas — not independent GraphQL-only
semantics.

## Headless mode

Public HTML may be disabled while REST remains the control plane. GraphQL may be
the primary content delivery surface. Headless mode must not weaken authorization.

## DTOs and view models

Adapters define DTOs near delivery code. Service result types map into REST DTOs,
GraphQL types, admin view models, and public render models.

Do not use one delivery DTO as the internal domain model.

## Error mapping

Consistent classification:

| Service error | REST | GraphQL |
|---------------|------|---------|
| Not found | `not_found` | Not found error |
| Validation | Field errors | Validation error |
| Unauthorized | Auth error | Auth error |
| Forbidden | Capability error | Forbidden error |
| Conflict | `conflict` | Conflict error |
| Internal | Generic internal + request ID | Generic internal |

Internal causes logged, not exposed.

Schema: `error.schema.json`.

## Uploads

Media uploads may use REST even when GraphQL is enabled. Recommended: REST handles
transport; media service validates and stores; GraphQL references media by ID.

## Schema discovery

REST discovery and GraphQL introspection reflect active plugins and capabilities
where appropriate. Private routes/fields should not be advertised to principals
that cannot use them.

## Testing

Adapter tests verify:

- Same service rules through REST and GraphQL
- Same authorization outcomes
- Same content visibility outcomes
- Same error classification

AppCMS conformance: `pkg/conformance/codex/`.
