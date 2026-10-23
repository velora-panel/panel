# Refactor diagnosis and fixes

Verified on October 23, 2026.

## Migration and bootstrap failures

The original migration call used embedded SQL with a valid `.` filesystem root. It discarded the underlying Goose error. A direct check of the configured installation found no application tables or migration history, and the runtime role had schema USAGE but no CREATE. The repaired command exposed the actual failing operation: creating `goose_db_version` failed with PostgreSQL SQLSTATE 42501, permission denied for schema public. This explains why a later bootstrap attempt found no `users` relation; it was not a missing migration-file path.

The operator corrected database access. A subsequent check now reports migrations 1, 2 and 3 applied, `doctor` succeeds, and the existing administrator is listed. No existing database or volume was reset. Optional MIGRATION_DATABASE_URL supports a separate owner role without granting runtime CREATE privileges. Migration diagnostics preserve technical causes after redaction.

## Terminal and validation

The old launcher inherited stdio but ran `go run`; the prompt mixed buffered email reads and direct terminal reads, had no password confirmation, implicitly accepted redirected input, and returned generic invalid-input errors. The historical EIO terminal state was not available to reproduce conclusively, so no unverified process-group cause is asserted.

The replacement builds one native executable and uses Linux execve to preserve terminal descriptors, signals and exit status. Email reads never buffer ahead into password entry. Passwords use term.ReadPassword with confirmation. Noninteractive input is explicit, bounded credential JSON on stdin. Validation gives actionable email/password/confirmation messages. A real Linux PTY test verifies hidden passwords, persistence, error exits and noninteractive rejection; schema checks occur before prompting.

The old Go-only package's node_modules warning appeared alongside failed scripts despite it having no JavaScript dependencies. Operational commands now run through the root wrapper rather than package-filtered scripts. Root frozen installation passes; no fake dependencies or empty directories were added to disguise warnings.

## Security and architecture

Removed the raw DATABASE_URL print. Redaction covers configured credentials, PostgreSQL/Redis URLs, labeled secrets and complete authorization/cookie header values. Passwords previously exposed in terminal output should be rotated by their owner. Environment files remain ignored. Git metadata is unavailable in this workspace, so historical tracked-secret auditing or history rewriting was not possible.

One CLI module under apps/cli reuses Core through its application facade. Core remains github.com/velora/panel. Forward migrations preserve legacy node data and instance links; migration 00001 remains byte-for-byte unchanged. Account services add disabled accounts, reauthentication, atomic password/session/reset updates, single-use reset tokens and serialized last-active-admin checks. Session insertion also verifies the credential version under a row lock to reject stale concurrent logins.

Official shadcn-vue CLI initialization and component generation replaced generic custom controls. Narrow prop-forwarding/default adaptations preserve strict TypeScript; application compositions retain full coverage. The Velora palette maps to semantic theme variables. No React or Beacon dependency was introduced.

See ../validation.md for actual verification results and ../cli.md, ../authentication.md and ../migrations.md for operation and remaining limitations.
