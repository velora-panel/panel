package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/velora/panel/internal/database/generated"
	"github.com/velora/panel/internal/domain"
)

type accountQueries struct {
	generated.Querier
	row   generated.User
	count int64
	fail  string
	err   error
	keep  string
}

func (f *accountQueries) e(step string) error {
	if f.fail == step {
		return f.err
	}
	return nil
}
func (f *accountQueries) CreateUser(_ context.Context, p generated.CreateUserParams) (generated.User, error) {
	f.row.Email = p.Email
	return f.row, f.e("create")
}
func (f *accountQueries) UserByID(context.Context, uuid.UUID) (generated.User, error) {
	return f.row, f.e("get")
}
func (f *accountQueries) ListUsers(context.Context) ([]generated.User, error) {
	return []generated.User{f.row}, f.e("list")
}
func (f *accountQueries) LockAccounts(context.Context) error { return f.e("lock") }
func (f *accountQueries) CountActiveAdmins(context.Context) (int64, error) {
	return f.count, f.e("count")
}
func (f *accountQueries) UpdateAccount(context.Context, generated.UpdateAccountParams) (generated.User, error) {
	return f.row, f.e("update")
}
func (f *accountQueries) RevokeSessions(_ context.Context, p generated.RevokeSessionsParams) error {
	f.keep = p.TokenHash
	return f.e("revoke")
}
func (f *accountQueries) DeleteResets(context.Context, uuid.UUID) error { return f.e("delete") }
func (f *accountQueries) UpdatePassword(context.Context, generated.UpdatePasswordParams) (generated.User, error) {
	return f.row, f.e("password")
}
func (f *accountQueries) UpdateProfile(context.Context, generated.UpdateProfileParams) (generated.User, error) {
	return f.row, f.e("profile")
}
func (f *accountQueries) CreateReset(context.Context, generated.CreateResetParams) error {
	return f.e("reset")
}
func (f *accountQueries) ConsumeReset(context.Context, string) (uuid.UUID, error) {
	return f.row.ID, f.e("consume")
}
func (f *accountQueries) ResetPassword(context.Context, generated.ResetPasswordParams) (generated.User, error) {
	return f.row, f.e("resetPassword")
}
func TestAccountRepository(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	f := &accountQueries{row: generated.User{ID: id, Role: "admin", PasswordHash: "hash"}, count: 2, err: errors.New("database failure")}
	r := New(f, func(ctx context.Context, work func(generated.Querier) error) error { return work(f) })
	u, e := r.InsertAccount(ctx, domain.AccountInput{Email: "a@b.com"}, "hash")
	require.NoError(t, e)
	require.Equal(t, "a@b.com", u.Email)
	_, _, e = r.Account(ctx, "bad")
	require.ErrorIs(t, e, domain.ErrInvalid)
	u, h, e := r.Account(ctx, id.String())
	require.NoError(t, e)
	require.Equal(t, id.String(), u.ID)
	require.Equal(t, "hash", h)
	rows, e := r.ListUsers(ctx)
	require.NoError(t, e)
	require.Len(t, rows, 1)
	_, e = r.ChangeAccount(ctx, "bad", domain.AccountUpdate{})
	require.ErrorIs(t, e, domain.ErrInvalid)
	for _, step := range []string{"lock", "get", "count", "update", "revoke", "delete", ""} {
		f.fail = step
		_, e = r.ChangeAccount(ctx, id.String(), domain.AccountUpdate{Role: "user"})
		if step == "" {
			require.NoError(t, e)
		} else {
			require.ErrorIs(t, e, f.err)
		}
	}
	f.count = 1
	_, e = r.ChangeAccount(ctx, id.String(), domain.AccountUpdate{Role: "user"})
	require.ErrorIs(t, e, domain.ErrLastAdmin)
	_, e = r.ChangeAccount(ctx, id.String(), domain.AccountUpdate{Role: "admin"})
	require.NoError(t, e)
	require.ErrorIs(t, r.ChangePassword(ctx, "bad", "a", "b", "c"), domain.ErrInvalid)
	for _, step := range []string{"password", "delete", "revoke", ""} {
		f.fail = step
		e = r.ChangePassword(ctx, id.String(), "old", "new", "keep")
		if step == "" {
			require.NoError(t, e)
			require.Equal(t, "keep", f.keep)
		} else {
			require.ErrorIs(t, e, f.err)
		}
	}
	_, e = r.ChangeProfile(ctx, "bad", "a@b.com", "hash")
	require.ErrorIs(t, e, domain.ErrInvalid)
	_, e = r.ChangeProfile(ctx, id.String(), "a@b.com", "hash")
	require.NoError(t, e)
	require.ErrorIs(t, r.RevokeAll(ctx, "bad"), domain.ErrInvalid)
	require.NoError(t, r.RevokeAll(ctx, id.String()))
	require.Empty(t, f.keep)
	require.ErrorIs(t, r.SaveReset(ctx, "bad", "hash", time.Now()), domain.ErrInvalid)
	for _, step := range []string{"lock", "delete", "reset", ""} {
		f.fail = step
		e = r.SaveReset(ctx, id.String(), "hash", time.Now())
		if step == "" {
			require.NoError(t, e)
		} else {
			require.ErrorIs(t, e, f.err)
		}
	}
	for _, step := range []string{"consume", "resetPassword", "delete", "revoke", ""} {
		f.fail = step
		e = r.ConsumeReset(ctx, "hash", "password")
		if step == "" {
			require.NoError(t, e)
		} else {
			require.ErrorIs(t, e, f.err)
		}
	}
	f.fail = "consume"
	f.err = pgx.ErrNoRows
	require.ErrorIs(t, r.ConsumeReset(ctx, "hash", "password"), domain.ErrUnauthorized)
}

type fakeBegin struct {
	tx  pgx.Tx
	err error
}

func (f fakeBegin) Begin(context.Context) (pgx.Tx, error) { return f.tx, f.err }

type fakeTx struct {
	pgx.Tx
	committed, rolled bool
	err               error
}

func (f *fakeTx) Commit(context.Context) error   { f.committed = true; return f.err }
func (f *fakeTx) Rollback(context.Context) error { f.rolled = true; return nil }
func TestTransactions(t *testing.T) {
	ctx := context.Background()
	failure := errors.New("transaction failed")
	require.ErrorIs(t, Transactions(fakeBegin{err: failure})(ctx, nil), failure)
	tx := &fakeTx{}
	require.ErrorIs(t, Transactions(fakeBegin{tx: tx})(ctx, func(generated.Querier) error { return failure }), failure)
	require.False(t, tx.committed)
	require.True(t, tx.rolled)
	tx = &fakeTx{}
	require.NoError(t, Transactions(fakeBegin{tx: tx})(ctx, func(generated.Querier) error { return nil }))
	require.True(t, tx.committed)
	require.True(t, tx.rolled)
	tx = &fakeTx{err: failure}
	require.ErrorIs(t, Transactions(fakeBegin{tx: tx})(ctx, func(generated.Querier) error { return nil }), failure)
}
