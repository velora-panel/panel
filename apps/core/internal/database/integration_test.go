//go:build integration

package database

import (
	"context"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/velora/panel/internal/database/generated"
	"github.com/velora/panel/internal/domain"
	"github.com/velora/panel/internal/modules/auth"
	"github.com/velora/panel/migrations"
)

func TestPostgres(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	require.NotEmpty(t, url, "TEST_DATABASE_URL must point at a dedicated database ending in _test")
	cfg, e := pgxpool.ParseConfig(url)
	require.NoError(t, e)
	require.True(t, strings.HasSuffix(cfg.ConnConfig.Database, "_test"), "refusing a non-test database")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, e := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, e)
	defer admin.Close()
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, e = admin.Exec(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, e)
	defer func() {
		_, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		require.NoError(t, err)
	}()
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	db := stdlib.OpenDB(*cfg.ConnConfig)
	defer db.Close()
	provider, e := goose.NewProvider(goose.DialectPostgres, db, migrations.Files)
	require.NoError(t, e)
	_, e = provider.UpTo(ctx, 1)
	require.NoError(t, e)
	var legacyID string
	e = db.QueryRowContext(ctx, "INSERT INTO stations(name) VALUES ('preserved') RETURNING id").Scan(&legacyID)
	require.NoError(t, e)
	_, e = db.ExecContext(ctx, "INSERT INTO instances(station_id,name) VALUES ($1,'preserved-instance')", legacyID)
	require.NoError(t, e)
	_, e = provider.Up(ctx)
	require.NoError(t, e)
	applied, e := provider.Up(ctx)
	require.NoError(t, e)
	require.Empty(t, applied)
	version, e := provider.GetDBVersion(ctx)
	require.NoError(t, e)
	require.EqualValues(t, 3, version)
	statuses, e := provider.Status(ctx)
	require.NoError(t, e)
	require.Len(t, statuses, 3)
	var linked string
	e = db.QueryRowContext(ctx, "SELECT node_id FROM instances WHERE name='preserved-instance'").Scan(&linked)
	require.NoError(t, e)
	require.Equal(t, legacyID, linked)
	_, e = db.ExecContext(ctx, "DELETE FROM instances; DELETE FROM nodes")
	require.NoError(t, e)
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, e)
	defer pool.Close()
	repo := New(generated.New(pool), Transactions(pool))
	require.NoError(t, repo.CreateUser(ctx, "admin@example.com", "hash"))
	require.ErrorIs(t, repo.CreateUser(ctx, "admin@example.com", "hash"), domain.ErrConflict)
	user, hash, e := repo.Credentials(ctx, "admin@example.com")
	require.NoError(t, e)
	require.Equal(t, "hash", hash)
	require.NoError(t, repo.SaveSession(ctx, "token", user.ID, "hash", time.Now().Add(time.Hour)))
	found, e := repo.SessionUser(ctx, "token")
	require.NoError(t, e)
	require.Equal(t, user, found)
	require.NoError(t, repo.DeleteSession(ctx, "token"))
	_, e = repo.SessionUser(ctx, "token")
	require.ErrorIs(t, e, domain.ErrNotFound)
	require.NoError(t, repo.SaveSession(ctx, "expired", user.ID, "hash", time.Now().Add(-time.Hour)))
	_, e = repo.SessionUser(ctx, "expired")
	require.ErrorIs(t, e, domain.ErrNotFound)
	require.Error(t, repo.SaveSession(ctx, "orphan", uuid.NewString(), "hash", time.Now()))
	node, e := repo.CreateNode(ctx, domain.NodeInput{Name: "West", Description: "Test"})
	require.NoError(t, e)
	require.Equal(t, "unconfigured", node.Status)
	_, e = repo.CreateNode(ctx, domain.NodeInput{Name: "West"})
	require.ErrorIs(t, e, domain.ErrConflict)
	_, e = repo.CreateNode(ctx, domain.NodeInput{Name: ""})
	require.Error(t, e)
	rows, e := repo.ListNodes(ctx)
	require.NoError(t, e)
	require.Len(t, rows, 1)
	got, e := repo.Node(ctx, node.ID)
	require.NoError(t, e)
	require.Equal(t, node, got)
	tx, e := pool.Begin(ctx)
	require.NoError(t, e)
	_, e = tx.Exec(ctx, "INSERT INTO nodes(name) VALUES ('temporary')")
	require.NoError(t, e)
	_, e = tx.Exec(ctx, "INSERT INTO nodes(name) VALUES ('West')")
	require.Error(t, e)
	require.Error(t, tx.Commit(ctx))
	rows, e = repo.ListNodes(ctx)
	require.NoError(t, e)
	require.Len(t, rows, 1)

	accounts := auth.NewAccounts(repo)
	actor := domain.User{Role: "admin"}
	second, e := accounts.Create(ctx, actor, domain.AccountInput{Email: "second@example.com", Password: "long password here", Role: "admin"})
	require.NoError(t, e)
	login := auth.New(repo)
	_, token, e := login.Login(ctx, second.Email, "long password here")
	require.NoError(t, e)
	_, other, e := login.Login(ctx, second.Email, "long password here")
	require.NoError(t, e)
	_, oldHash, e := repo.Credentials(ctx, second.Email)
	require.NoError(t, e)
	require.NoError(t, accounts.Password(ctx, second, token, domain.PasswordInput{CurrentPassword: "long password here", NewPassword: "changed password"}))
	require.ErrorIs(t, repo.SaveSession(ctx, "stale-login", second.ID, oldHash, time.Now().Add(time.Hour)), domain.ErrNotFound)
	_, e = login.Current(ctx, token)
	require.NoError(t, e)
	_, e = login.Current(ctx, other)
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	reset, e := accounts.IssueReset(ctx, actor, second.ID)
	require.NoError(t, e)
	require.NoError(t, accounts.Reset(ctx, reset, "reset password here"))
	require.ErrorIs(t, accounts.Reset(ctx, reset, "reset password here"), domain.ErrUnauthorized)
	_, e = login.Current(ctx, token)
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	require.NoError(t, repo.SaveReset(ctx, second.ID, auth.TokenHash("expired-reset"), time.Now().Add(-time.Minute)))
	require.ErrorIs(t, repo.ConsumeReset(ctx, auth.TokenHash("expired-reset"), "never-used"), domain.ErrUnauthorized)
	require.NoError(t, accounts.OperatorUpdate(ctx, second.ID, true))
	_, _, e = login.Login(ctx, second.Email, "reset password here")
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	require.ErrorIs(t, accounts.OperatorUpdate(ctx, user.ID, true), domain.ErrLastAdmin)
	require.NoError(t, accounts.OperatorUpdate(ctx, second.ID, false))
	// Competing administrator removals serialize: exactly one active administrator survives.
	failures := make(chan error, 2)
	for _, id := range []string{user.ID, second.ID} {
		go func() { failures <- accounts.OperatorUpdate(ctx, id, true) }()
	}
	first, secondErr := <-failures, <-failures
	if first == nil {
		require.ErrorIs(t, secondErr, domain.ErrLastAdmin)
	} else {
		require.ErrorIs(t, first, domain.ErrLastAdmin)
		require.NoError(t, secondErr)
	}
	_, e = provider.Down(ctx)
	require.NoError(t, e)

	_, e = provider.Down(ctx)
	require.NoError(t, e)
	var preservedName string
	e = db.QueryRowContext(ctx, "SELECT name FROM stations WHERE id=$1", node.ID).Scan(&preservedName)
	require.NoError(t, e)
	require.Equal(t, "West", preservedName)
	_, e = provider.Down(ctx)
	require.NoError(t, e)
	var exists bool
	e = db.QueryRowContext(ctx, "SELECT to_regclass('users') IS NOT NULL").Scan(&exists)
	require.NoError(t, e)
	require.False(t, exists)
	_, e = provider.Up(ctx)
	require.NoError(t, e)
	// A failed Goose migration rolls back both schema writes and version advancement.
	files := fstest.MapFS{}
	entries, e := fs.ReadDir(migrations.Files, ".")
	require.NoError(t, e)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := fs.ReadFile(migrations.Files, entry.Name())
		require.NoError(t, err)
		files[entry.Name()] = &fstest.MapFile{Data: content}
	}
	files["00004_failure.sql"] = &fstest.MapFile{Data: []byte("-- +goose Up\nCREATE TABLE atomic_probe(id int);\nSELECT * FROM deliberately_missing;\n-- +goose Down\nDROP TABLE atomic_probe;\n")}
	failing, e := goose.NewProvider(goose.DialectPostgres, db, files)
	require.NoError(t, e)
	_, e = failing.Up(ctx)
	require.Error(t, e)
	e = db.QueryRowContext(ctx, "SELECT to_regclass('atomic_probe') IS NOT NULL").Scan(&exists)
	require.NoError(t, e)
	require.False(t, exists)
	version, e = provider.GetDBVersion(ctx)
	require.NoError(t, e)
	require.EqualValues(t, 3, version)

}
