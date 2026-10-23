# Refactor diagnosis

Inspection found a raw `fmt.Println` of DATABASE_URL in Core startup. It will be removed; any password previously printed should be rotated by its owner.

Migration execution uses embedded SQL and `goose.UpContext(..., ".")`, so it is not inherently dependent on the caller's directory. The migration function discards its underlying error, making the reported generic failure impossible to attribute from the historical output alone. The replacement will preserve redacted diagnostics and inspect actual migration state.

The old wrapper inherits stdio but launches `go run` rather than a compiled executable and assumes the package working directory. The old prompt mixes buffered email reads with direct terminal password reads, has no confirmation, reads redirected input implicitly, and provides only the domain's generic invalid-input error. Schema readiness is never checked before account creation. A missing users relation therefore reaches pgx instead of producing setup instructions.

The Go-only workspace package has no JavaScript dependencies; pnpm's missing-node_modules warning accompanies a failing script and is not proof of a missing Go dependency. Operational commands will move to the root CLI wrapper without pnpm package filtering or Turborepo stdin mediation.

Existing migration 00001 is treated as deployed and kept byte-for-byte intact. Rename and authentication changes will be forward migrations. No existing database will be reset to repair migration history.
