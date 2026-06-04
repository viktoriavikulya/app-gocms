# Platform (`github.com/fastygo/platform`)

Platform defines **how product modules plug into a host**: capabilities, record schemas, admin panel resources, API/public surface registration, storage transactions, and render models. It intentionally excludes CMS/CRM business rules.

## Core concepts

### Module (`pkg/contracts`)

Every product exposes a type implementing:

```go
type Module interface {
    Manifest() ModuleManifest
    Register(ModuleContext) error
}
```

AppCMS: `pkg/module/cms.go` — module ID `cms`, compiled-in kind.

`Register` adds:

- Capability definitions (`CapabilityAdminAccess`, `content.read`, …)
- Toolset record types and relations
- Panel admin resources (posts, pages, taxonomies, …)

### Toolset (`pkg/toolset`)

Describes **record types and relations** for admin/metadata and future custom fields—not SQL tables.

AppCMS registers types via `pkg/module/records` and `pkg/module/relations`.

Toolset supports review/diff descriptors for change tracking; it is **not** an ORM. Persistence is the product's `StoragePort` adapter.

### Panel (`pkg/panel`, `pkg/panelschema`)

Admin UI metadata:

- **Resources** — list/form routes, capability requirements
- **Views / workflows / actions** — schema descriptors consumed by renderers
- **Policies** — capability checks at panel level

AppCMS bindings: `pkg/module/panels`.

### Render (`pkg/render`, `pkg/renderers/templ`)

Generic screen model:

```go
type ScreenModel struct {
    ID, Title, View, Resource string
    Metadata map[string]string
    // fields, actions, etc.
}
```

Platform's templ renderer turns `ScreenModel` into HTML. AppCMS `pkg/templ` wraps pages with product chrome (title, layout).

### Profile (`pkg/profile`)

Deployment profile: admin base, API base, public base, workspace list, panel mounts.

AppCMS `DefaultProfile()` in `pkg/app/bundle.go`:

- Profile ID: `gocms-admin`
- Workspace: `root` with module `cms`
- Bases: `/go-admin`, `/go-json`, `/`

AppSuite may compose multiple product bundles under different workspace paths.

### App bundle (`pkg/appbundle`)

Exports a product as a reusable unit:

```go
appbundle.StaticBundle{
    AppManifest: ...,
    AppModule:   modulecms.Module{},
    Profile:     DefaultProfile(),
    Mount:       AdminMount{BasePath: "/go-admin", ...},
}
```

Used by AppSuite `pkg/compose` without importing AppCMS `internal/`.

## Storage port (`pkg/contracts/storage.go`)

```go
type StoragePort interface {
    WithinWorkspaceTx(ctx, workspace, fn) error
}

type StorageTx interface {
    List, Get, Put, Delete(ctx, recordType, ...) 
}
```

Records are `map[string]any` JSON payloads keyed by workspace + record type + id.

AppCMS implements the port in `internal/storage/sqlite` and adapts to typed domain repos in `internal/storage/adapters.go`.

## Policy and runtime context

| Type | Package | Use |
|------|---------|-----|
| `CapabilityID` | `contracts` | RBAC strings (`admin.access`, `content.write`, …) |
| `Principal` | product authn | Implements capability checks |
| `RuntimeContext` | `contracts` | Profile, workspace, module, principal on `context` |
| `PolicyEvaluator` | `contracts` | Workspace-aware grants (AppSuite) |

AppCMS sets `RuntimeContext` on admin requests after session validation.

## Hooks and audit (contracts)

Platform defines hook registration and audit event shapes. AppCMS implements a **product-owned** hook bus in `internal/extensions` and audit store in `internal/operations` for the operations slice baseline.

Firing domain hooks from services is a product responsibility; Platform stays transport/metadata only.

## Module host (`pkg/modulehost`)

Hosts module registration for tests and composed profiles. Production AppCMS wiring is primarily through `pkg/app` + Framework, but modulehost tests guard descriptor integrity.

## Conformance (`pkg/conformance`, `pkg/codex`)

Helpers to validate profiles and codex-shaped behavior. Use when adding routes or envelopes that must match GoCMS discovery contracts.

## Import boundary rule

**Platform must never import** `github.com/fastygo/app-gocms` (or other product apps). Products import Platform; not the reverse.

## Further reading

- [AppCMS package layout](../appcms/package-layout.md)
- [Storage](../appcms/storage.md)
