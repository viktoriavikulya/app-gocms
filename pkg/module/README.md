# ModuleCMS

ModuleCMS is the Platform-native CMS module for the GoCMS replacement path.

It is a separate Go module with module path `github.com/fastygo/app-gocms/pkg/module`.
It implements `contracts.Module` and registers CMS records, relations,
capabilities, and Panel resources through the Platform ABI.

## Scope

ModuleCMS owns CMS semantics:

- posts and pages;
- content lifecycle status;
- metadata and custom field definitions;
- taxonomies, terms, and content assignments;
- media metadata;
- public author profiles;
- GoCMS codex compatibility helpers.

ModuleCMS does not own:

- the executable server;
- renderer selection;
- product theme tokens;
- static assets;
- binary media storage;
- private user credentials.

## Packages

- `records`: Toolset record definitions.
- `relations`: Toolset relation definitions.
- `panels`: Panel resources, views, actions, workflows, and relation views.
- `codex`: GoCMS REST/admin compatibility shapes.
- `migration`: mapping from current GoCMS internals to ModuleCMS concepts.

## Validation

```bash
go test ./...
```

The tests prove module registration, schema coverage, descriptor coverage,
codex discovery shapes, and migration mapping anchors.
