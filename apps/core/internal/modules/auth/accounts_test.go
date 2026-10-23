package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/velora/panel/internal/domain"
)

type accountsRepo struct {
	AccountRepository
	u           domain.User
	hash        string
	err         error
	saved, keep string
	expires     time.Time
}

func (f *accountsRepo) Account(context.Context, string) (domain.User, string, error) {
	return f.u, f.hash, f.err
}
func (f *accountsRepo) InsertAccount(_ context.Context, i domain.AccountInput, h string) (domain.User, error) {
	f.hash = h
	f.u.Email = i.Email
	return f.u, f.err
}
func (f *accountsRepo) ListUsers(context.Context) ([]domain.User, error) {
	return []domain.User{f.u}, f.err
}
func (f *accountsRepo) ChangeAccount(_ context.Context, _ string, i domain.AccountUpdate) (domain.User, error) {
	f.u.Role = i.Role
	f.u.Disabled = i.Disabled
	return f.u, f.err
}
func (f *accountsRepo) ChangePassword(_ context.Context, _ string, old, next, keep string) error {
	if old != f.hash {
		return domain.ErrConflict
	}
	f.hash = next
	f.keep = keep
	return f.err
}
func (f *accountsRepo) ChangeProfile(_ context.Context, _ string, email, _ string) (domain.User, error) {
	f.u.Email = email
	return f.u, f.err
}
func (f *accountsRepo) RevokeAll(context.Context, string) error { f.keep = "revoked"; return f.err }
func (f *accountsRepo) SaveReset(_ context.Context, _ string, hash string, expires time.Time) error {
	f.saved = hash
	f.expires = expires
	return f.err
}
func (f *accountsRepo) ConsumeReset(_ context.Context, hash, next string) error {
	if hash != f.saved || time.Now().After(f.expires) {
		return domain.ErrUnauthorized
	}
	f.saved = ""
	f.hash = next
	return f.err
}
func TestAccounts(t *testing.T) {
	ctx := context.Background()
	admin := domain.User{ID: uuid.NewString(), Role: "admin"}
	ordinary := domain.User{ID: uuid.NewString(), Role: "user"}
	f := &accountsRepo{u: admin, hash: HashPassword("current password")}
	a := NewAccounts(f)
	for _, email := range []string{"bad", "Name <a@example.com>", strings.Repeat("a", 250) + "@b.com"} {
		require.ErrorIs(t, ValidateEmail(email), domain.ErrInvalid)
	}
	require.Error(t, ValidatePassword("short"))
	require.Error(t, ValidatePassword(strings.Repeat("a", 129)))
	require.Error(t, ValidateAccount("a@b.com", "long password", "owner"))
	require.Error(t, ValidateAccount("a@b.com", "short", "user"))
	for _, actor := range []domain.User{ordinary, {Role: "admin", Disabled: true}} {
		_, e := a.Create(ctx, actor, domain.AccountInput{})
		require.ErrorIs(t, e, domain.ErrForbidden)
		_, e = a.List(ctx, actor)
		require.ErrorIs(t, e, domain.ErrForbidden)
		_, e = a.Get(ctx, actor, admin.ID)
		require.ErrorIs(t, e, domain.ErrForbidden)
		_, e = a.Update(ctx, actor, admin.ID, domain.AccountUpdate{})
		require.ErrorIs(t, e, domain.ErrForbidden)
		require.ErrorIs(t, a.Revoke(ctx, actor, admin.ID, ""), domain.ErrForbidden)
		_, e = a.IssueReset(ctx, actor, admin.ID)
		require.ErrorIs(t, e, domain.ErrForbidden)
	}
	_, e := a.Create(ctx, admin, domain.AccountInput{})
	require.Error(t, e)
	u, e := a.Create(ctx, admin, domain.AccountInput{Email: " A@B.COM ", Password: "current password", Role: "user"})
	require.NoError(t, e)
	require.Equal(t, "a@b.com", u.Email)
	require.True(t, VerifyPassword(f.hash, "current password"))
	users, e := a.List(ctx, admin)
	require.NoError(t, e)
	require.Len(t, users, 1)
	_, e = a.Get(ctx, admin, admin.ID)
	require.NoError(t, e)
	_, e = a.Get(ctx, admin, ordinary.ID)
	require.NoError(t, e)
	_, e = a.Get(ctx, admin, "bad")
	require.Error(t, e)
	_, e = a.Update(ctx, admin, "bad", domain.AccountUpdate{})
	require.Error(t, e)
	_, e = a.Update(ctx, admin, admin.ID, domain.AccountUpdate{Role: "owner"})
	require.Error(t, e)
	_, e = a.Update(ctx, admin, admin.ID, domain.AccountUpdate{Role: "user", CurrentPassword: "wrong"})
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	_, e = a.Update(ctx, admin, admin.ID, domain.AccountUpdate{Role: "user", CurrentPassword: "current password"})
	require.NoError(t, e)
	require.Equal(t, "user", f.u.Role)
	require.Error(t, a.OperatorUpdate(ctx, "bad", true))
	require.NoError(t, a.OperatorUpdate(ctx, admin.ID, true))
	require.True(t, f.u.Disabled)
	require.Error(t, a.Password(ctx, admin, "token", domain.PasswordInput{NewPassword: "short"}))
	require.ErrorIs(t, a.Password(ctx, admin, "token", domain.PasswordInput{CurrentPassword: "wrong", NewPassword: "new password here"}), domain.ErrUnauthorized)
	require.NoError(t, a.Password(ctx, admin, "token", domain.PasswordInput{CurrentPassword: "current password", NewPassword: "new password here"}))
	require.Equal(t, TokenHash("token"), f.keep)
	require.True(t, VerifyPassword(f.hash, "new password here"))
	_, e = a.Profile(ctx, admin, domain.ProfileInput{Email: "bad"})
	require.Error(t, e)
	_, e = a.Profile(ctx, admin, domain.ProfileInput{Email: "new@example.com", CurrentPassword: "wrong"})
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	u, e = a.Profile(ctx, admin, domain.ProfileInput{Email: "NEW@example.com", CurrentPassword: "new password here"})
	require.NoError(t, e)
	require.Equal(t, "new@example.com", u.Email)
	require.Error(t, a.Revoke(ctx, admin, "bad", ""))
	require.ErrorIs(t, a.Revoke(ctx, admin, admin.ID, "wrong"), domain.ErrUnauthorized)
	require.NoError(t, a.Revoke(ctx, admin, admin.ID, "new password here"))
	require.Equal(t, "revoked", f.keep)
	_, e = a.IssueReset(ctx, admin, "bad")
	require.Error(t, e)
	_, e = a.IssueReset(ctx, admin, admin.ID)
	require.ErrorIs(t, e, domain.ErrForbidden)
	f.u.Disabled = false
	token, e := a.IssueReset(ctx, admin, admin.ID)
	require.NoError(t, e)
	require.Equal(t, TokenHash(token), f.saved)
	require.NotEqual(t, token, f.saved)
	require.WithinDuration(t, time.Now().Add(15*time.Minute), f.expires, time.Second)
	require.Error(t, a.Reset(ctx, "bad", "new password here"))
	require.Error(t, a.Reset(ctx, token, "short"))
	require.NoError(t, a.Reset(ctx, token, "reset password"))
	require.True(t, VerifyPassword(f.hash, "reset password"))
	require.ErrorIs(t, a.Reset(ctx, token, "reset password"), domain.ErrUnauthorized)
	token, e = a.IssueReset(ctx, admin, admin.ID)
	require.NoError(t, e)
	f.expires = time.Now().Add(-time.Second)
	require.ErrorIs(t, a.Reset(ctx, token, "reset password"), domain.ErrUnauthorized)
	f.err = errors.New("database unavailable")
	require.ErrorIs(t, a.Reauthenticate(ctx, admin, "x"), f.err)
	require.ErrorIs(t, a.OperatorUpdate(ctx, admin.ID, true), f.err)
	require.ErrorIs(t, a.Password(ctx, admin, "token", domain.PasswordInput{NewPassword: "long password"}), f.err)
	_, e = a.Profile(ctx, admin, domain.ProfileInput{Email: "a@b.com"})
	require.ErrorIs(t, e, f.err)
	_, e = a.IssueReset(ctx, admin, admin.ID)
	require.ErrorIs(t, e, f.err)
}

type resetFailure struct{ *accountsRepo }

func (f resetFailure) SaveReset(context.Context, string, string, time.Time) error {
	return domain.ErrConflict
}
func TestResetStorageFailure(t *testing.T) {
	admin := domain.User{ID: uuid.NewString(), Role: "admin"}
	a := NewAccounts(resetFailure{&accountsRepo{u: admin}})
	_, e := a.IssueReset(context.Background(), admin, admin.ID)
	require.ErrorIs(t, e, domain.ErrConflict)
}
func TestDisabledAndMalformedSessions(t *testing.T) {
	f := &fakeRepo{user: domain.User{Disabled: true}, hash: HashPassword("long password")}
	s := New(f)
	_, _, e := s.Login(context.Background(), "a@b.com", "long password")
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	for _, token := range []string{NewToken(), strings.Repeat("!", 43), "AA"} {
		_, e = s.Current(context.Background(), token)
		require.ErrorIs(t, e, domain.ErrUnauthorized)
	}
}
