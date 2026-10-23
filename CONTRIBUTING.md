# Contributing

Read README.md, docs/architecture.md and AGENTS.md. Keep changes small and domain-focused. Use Vue Composition API with script setup and strict TypeScript. Keep handlers thin, dependencies explicit and SQL parameterized. Do not add Beacon code here or expose temporary design references.

Run `pnpm check` and relevant integration/browser tests before proposing changes. Preserve hard coverage gates and document any legitimate wiring exemption. Regenerate API/sqlc artifacts and include lockfile changes. Explain behavior, validation and operational impact in pull requests. Never commit credentials, test traces containing private deployments or production database snapshots.

This project uses the MIT license. Contributions must be original or appropriately licensed and attributed. User-facing changes should include release notes once versioned releases begin. Security reports should follow SECURITY.md.

Use generated shadcn-vue primitives for generic controls and semantic Velora theme tokens. Application compositions retain full coverage. Keep the unified CLI in apps/cli and reuse Core services through its application integration surface. Apply forward migrations; never edit previously deployed history. See docs/cli.md and docs/migrations.md for bootstrap and recovery.
