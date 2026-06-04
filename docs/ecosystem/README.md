# FastyGo Ecosystem

AppCMS is one product in a multi-repo stack. This section describes each ecosystem module and how AppCMS consumes it.

## Stack at a glance

```text
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   AppSuite   │     │   AppCMS     │     │   AppCRM     │
│  (launcher)  │     │  (this app)  │     │  (CRM app)   │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       └────────────────────┼────────────────────┘
                            │
              ┌─────────────▼─────────────┐
              │        Platform           │
              │  contracts · toolset ·    │
              │  panel · profile · render │
              └─────────────┬─────────────┘
                            │
              ┌─────────────▼─────────────┐
              │        Framework          │
              │  HTTP · auth · security   │
              └─────────────┬─────────────┘
                            │
              ┌─────────────▼─────────────┐
              │         Templ             │
              │  shared UI components     │
              └───────────────────────────┘

        GoCMS (oracle) ── go-codex/en contracts
```

## Documents in this section

| Document | Contents |
|----------|----------|
| [Framework](framework.md) | `AppBuilder`, middleware, sessions, health |
| [Platform](platform.md) | Modules, toolset, panel, storage port, profiles |
| [Templ](templ.md) | UI primitives; how admin/public views compose them |
| [GoCMS oracle](gocms-oracle.md) | `go-codex/en`, what to port vs what to reference |
| [Related apps](related-apps.md) | AppCRM, AppSuite, ModuleMonitoring |

## Versioning and local development

AppCMS `go.mod` uses **replace directives** to sibling folders:

```go
replace github.com/fastygo/platform => ../@Platform
replace github.com/fastygo/framework => ../@Framework
replace github.com/fastygo/templ => ../@Templ
```

When changing Platform contracts, run tests in both Platform and AppCMS before merging.

## Guardrails (product stack)

- Do not add `github.com/fastygo/ui8kit` to AppCMS, AppCRM, AppSuite, or new templ views.
- Do not move the GoCMS monolith into Platform.
- Do not change root codex URLs; add spaces overlays **beside** them.
- Platform must not import product apps.
- Shared packages appear only after real duplication across CMS/CRM/Suite.

These rules keep AppCMS replaceable and embeddable without entangling Platform with CMS domain logic.
