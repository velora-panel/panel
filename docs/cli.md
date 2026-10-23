# Unified CLI

`apps/cli` builds the sole `velora` executable. Core retains module identity `github.com/velora/panel`; its `application` integration surface exposes shared services to the CLI. `go.work` joins both modules. No business rules live in the JavaScript launcher.

From the repository root:

```sh
pnpm velora --help
pnpm velora migrate up
pnpm velora migrate status
pnpm velora migrate version
pnpm velora admin create
pnpm velora admin list
pnpm velora admin reset-password --email admin@example.com
pnpm velora user create --role user
pnpm velora user list
pnpm velora user disable --email user@example.com
pnpm velora user enable --email user@example.com
pnpm velora doctor
pnpm velora serve
```

`pnpm velora` loads root `.env` without overriding exported variables, builds the native binary and replaces the launcher process with it using Node 24 `execve`. Arguments, terminal file descriptors, signals and exit status reach the native process. Run on Linux. Production can invoke `apps/cli/bin/velora` directly with exported configuration; it needs neither Node nor pnpm and does not load `.env` implicitly.

Interactive create/reset commands hide passwords and ask for confirmation. Passwords must contain 12–128 bytes. Do not put secrets in command arguments. Noninteractive operation requires explicit `--password-stdin` and one JSON object containing `email`, `password`, and `password_confirmation`. Feed it from a protected secret source; avoid shell tracing and plaintext command history. Unknown fields, trailing objects, incomplete reads and mismatched passwords fail. No default account or password is shipped.

Administrative CLI operations are privileged local operations: protect executable, host/container access and database credentials. They share the web account services, including last-active-administrator protection. Reset tokens are generated internally, stored hashed, consumed immediately and never printed.

`migrate down` rolls back one migration only with `--allow-destructive`. It is not an automated recovery command; see [migration procedures](migrations.md).
