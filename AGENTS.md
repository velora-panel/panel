# Velora contributor instructions

Keep Core a modular Go monolith and Console a Vue SPA. Never implement Beacon here. Use PostgreSQL only. Root assets and assets_temp are private design references, never public roots. Do not hand-edit generated sqlc or OpenAPI types. Run pnpm check and service-dependent tests after meaningful changes. Never weaken coverage thresholds or report unexecuted checks as passing. Keep credentials and environment files out of Git.

Use shadcn-vue primitives for generic Console controls. Keep the sole executable in apps/cli and reuse Core services. Preserve deployed migration history; generate sqlc/OpenAPI output using codegen. Test authentication, last-admin protection and CLI terminal handling after related changes.
