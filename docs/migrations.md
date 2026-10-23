# PostgreSQL setup and migrations

Use a dedicated PostgreSQL database. Set `DATABASE_URL` for the runtime role. When using separate migration and runtime roles, set `MIGRATION_DATABASE_URL` to a migration-capable owner role in local `.env` or your secret manager. Never paste credentials into diagnostics or commit environment files.

The migration role needs CONNECT to the database, USAGE/CREATE on its schema, and ownership of objects it must alter. The runtime role needs USAGE on the schema and SELECT/INSERT/UPDATE/DELETE on application tables, but need not have CREATE. Configure grants and default privileges as the database/schema owner for the role that runs migrations. PostgreSQL grants are installation-specific; granting CREATE to every runtime role is not required.

```sh
pnpm velora migrate status
pnpm velora migrate up
pnpm velora migrate version
pnpm velora doctor
pnpm velora admin create
pnpm velora serve
```

Migrations are embedded in Core and discovered with a Goose provider, independently of the working directory. Commands use bounded contexts and preserve redacted underlying PostgreSQL errors. No startup path or administrator command creates tables implicitly. `doctor` and account commands reject missing required tables with an instruction to run `velora migrate up`.

Migration 00001 is unchanged historical SQL. Migration 00002 renames the legacy compute table to `nodes`, its unique/check/primary-key constraints, the instance foreign-key column and constraint, and its index. Existing UUIDs, data and instance links are preserved. Down reverses the names. Migration 00003 adds disabled accounts and hashed password-reset persistence. Goose tracks applied versions, applies migrations transactionally, and repeated `up` is safe.

Before a production change, take and verify a PostgreSQL backup, including roles/grants, and rehearse restore on a separate database. Stop application writes before rollback. Review the specific Down SQL: rolling back 00003 removes reset tokens and disabled flags; rolling back 00001 deletes the foundation tables. Prefer restoring a verified backup or a corrective forward migration when rollback would lose information. Never drop volumes, delete Goose history, or edit deployed migration files to hide a failure.

For SQLSTATE 42501, inspect database/schema ownership and grants for the exact migration connection. For 42P01, check the database, search_path and migration version before bootstrap. Correct the underlying cause, inspect `migrate status`, then retry `up`; failed transactional migrations do not require a database reset.
