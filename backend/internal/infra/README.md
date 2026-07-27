# infra

Infrastructure adapters: config, security, persistence, observability.

- **Schema**: `persistence/migrator` runs goose SQL (`goose/postgres` + `goose/sqlite`). Do not AutoMigrate in production.
- **GORM**: `persistence/relational` maps models and repositories only.
- Business providers/egress/media were removed in the Foam scaffold.
