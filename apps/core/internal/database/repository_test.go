package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"github.com/velora/panel/internal/database/generated"
	"github.com/velora/panel/internal/domain"
)

type fakeQueries struct {
	generated.Querier
	err error
}

func (f fakeQueries) CreateUser(context.Context, generated.CreateUserParams) (generated.User, error) {
	return generated.User{}, f.err
}
func (f fakeQueries) UserByEmail(context.Context, string) (generated.User, error) {
	return generated.User{}, f.err
}
func (f fakeQueries) CreateSession(context.Context, generated.CreateSessionParams) (string, error) {
	return "", f.err
}
func (f fakeQueries) DeleteSession(context.Context, string) error { return f.err }
func (f fakeQueries) SessionUser(context.Context, string) (generated.User, error) {
	return generated.User{}, f.err
}
func (f fakeQueries) ListNodes(context.Context) ([]generated.Node, error) {
	return []generated.Node{{Name: "west"}}, f.err
}
func (f fakeQueries) NodeByID(context.Context, uuid.UUID) (generated.Node, error) {
	return generated.Node{}, f.err
}
func (f fakeQueries) CreateNode(context.Context, generated.CreateNodeParams) (generated.Node, error) {
	return generated.Node{}, f.err
}
func TestRepository(t *testing.T) {
	ctx := context.Background()
	r := New(fakeQueries{}, nil)
	_, _, e := r.Credentials(ctx, "x")
	require.NoError(t, e)
	require.NoError(t, r.CreateUser(ctx, "x", "hash"))
	require.NoError(t, r.SaveSession(ctx, "hash", uuid.NewString(), time.Now()))
	require.ErrorIs(t, r.SaveSession(ctx, "hash", "bad", time.Now()), domain.ErrInvalid)
	_, e = r.SessionUser(ctx, "hash")
	require.NoError(t, e)
	require.NoError(t, r.DeleteSession(ctx, "hash"))
	rows, e := r.ListNodes(ctx)
	require.NoError(t, e)
	require.Equal(t, "west", rows[0].Name)
	_, e = r.Node(ctx, uuid.NewString())
	require.NoError(t, e)
	_, e = r.Node(ctx, "bad")
	require.ErrorIs(t, e, domain.ErrInvalid)
	_, e = r.CreateNode(ctx, domain.NodeInput{})
	require.NoError(t, e)
	require.ErrorIs(t, translate(pgx.ErrNoRows), domain.ErrNotFound)
	require.ErrorIs(t, translate(&pgconn.PgError{Code: "23505"}), domain.ErrConflict)
	sentinel := errors.New("offline")
	require.ErrorIs(t, translate(sentinel), sentinel)
}
