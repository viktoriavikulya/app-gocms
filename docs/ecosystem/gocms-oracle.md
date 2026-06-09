# GoCMS Oracle (`@GoCMS`, `go-codex/en`)

**GoCMS** is the legacy monolith and the **compatibility oracle** for AppCMS. AppCMS targets observable parity with GoCMS public behavior—not a line-by-line port of internal packages.

## What to use GoCMS for

| Need | Where |
|------|-------|
| Normative URL and API contracts | `@GoCMS/go-codex/en` |
| Expected admin capabilities and plugin rules | `05-plugin-contract.md`, `06-hooks-contract.md`, theme contract |
| Reference implementation while porting | `@GoCMS/internal/...` |
| Conformance tests ideas | GoCMS + Platform `pkg/conformance` |

## What not to copy

- UI8Kit admin components and styling model
- Monolithic `internal/platform` wiring — replaced by Framework + Platform modulehost
- Direct imports of GoCMS from AppCMS production code

## Codex documents

**Machine-readable contract (canonical):** `@AppCMS/schema/codex/v1/` — JSON schemas
for content entries, content types, menus, REST envelopes, errors, and seed
bundles. See [Codex v1 README](../../schema/codex/v1/README.md).

**GoCMS oracle prose (reference):** `@AppCMS/docs/compat/gocms/` — summarized
REST, content, domain, storage, and adapter guidance preserved after `@GoCMS`
leaves the workspace.

| Document | Topic |
|----------|-------|
| REST / API shape | Discovery, list envelopes, error codes |
| Content contract | Kinds, statuses, fields, lifecycle |
| Domain / storage | Package boundaries, repository ports |
| Admin contract | Root `/go-admin` behavior (GoCMS codex) |
| Theme contract | Manifest fields, template roles (GoCMS codex) |
| Plugin contract | Lifecycle, active-only exposure (GoCMS codex) |
| Hooks contract | Action/filter priority and error policy (GoCMS codex) |

When AppCMS behavior intentionally differs, document the gap in [migration mapping](../compat/gocms/migration-mapping.md), [Codex schema alignment](../appcms/codex-schema-alignment.md), or product progress notes—not by silently changing URLs.

## Migration mapping

See [migration-mapping.md](../compat/gocms/migration-mapping.md) for the architecture mapping table:

| GoCMS area | AppCMS owner |
|------------|--------------|
| Domain content | `internal/domain/content` + `internal/application/content` |
| REST | `internal/delivery/rest` |
| Admin delivery | `internal/appschema` + Platform render + `pkg/templ` |
| Panel descriptors | `pkg/module/panels` |

## Parity slices (stack progress)

The product stack migrates in slices (Framework host → storage → services → REST → admin UI → auth → public → extensions). AppCMS `docs` describe the **current** architecture; slice status lives in Platform `.project/current-progress.md`.

## Testing against the oracle

AppCMS integration tests assert:

- Codex route surfaces respond (discovery, admin paths)
- REST envelopes include `pagination` where required
- Public themed markers (`data-gocms-theme`, `data-gocms-public-screen`)
- No UI8Kit imports in product paths

When in doubt, add a test that encodes codex behavior rather than relying on manual comparison with GoCMS.
