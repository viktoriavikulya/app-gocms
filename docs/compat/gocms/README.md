# GoCMS Compatibility References

These documents preserve the **GoCMS oracle** material that informed AppCMS
Codex v1. They are explanatory compatibility references — not the machine-readable
source of truth.

## Canonical contract

Use `@AppCMS/schema/codex/v1/` JSON schemas for serialized CMS data shapes.
When prose and schema disagree during migration, schemas win for new work; file
gaps in [migration-from-gocms.md](../../migration-from-gocms.md).

## Preserved references

| Document | Original source | Topic |
|----------|-----------------|-------|
| [02-rest-api-contract.md](02-rest-api-contract.md) | `@GoCMS/go-codex/en/02-rest-api-contract.md` | REST discovery, envelopes, pagination, errors |
| [03-content-contract.md](03-content-contract.md) | `@GoCMS/go-codex/en/03-content-contract.md` | Content kinds, statuses, fields, lifecycle |
| [01-domain-model.md](01-domain-model.md) | `@GoCMS/go-stack/en/01-domain-model.md` | Recommended domain package boundaries |
| [03-storage-repositories.md](03-storage-repositories.md) | `@GoCMS/go-stack/en/03-storage-repositories.md` | Repository ports, IDs, metadata, migrations |
| [04-rest-graphql-adapters.md](04-rest-graphql-adapters.md) | `@GoCMS/go-stack/en/04-rest-graphql-adapters.md` | REST/GraphQL adapter rules over services |

## Related

- [Codex v1 schemas](../../../schema/codex/v1/README.md)
- [GoCMS oracle (ecosystem)](../../ecosystem/gocms-oracle.md)
- [Migration from GoCMS](../../migration-from-gocms.md)

## Note on `@GoCMS` removal

The legacy `@GoCMS` workspace may be removed. These references remain inside AppCMS
so parity work does not depend on an external monolith checkout.
