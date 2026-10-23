package migrations

import "embed"

// Files contains versioned migrations for explicit CLI execution only.
//
//go:embed *.sql
var Files embed.FS
