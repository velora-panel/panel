package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/velora/panel/internal/domain"
)

type fakeRepo struct {
	user         domain.User
	hash         string
	err, saveErr error
	saved        string
	deleted      string
}

func (f *fakeRepo) Credentials(context.Context, string) (domain.User, string, error) {
	return f.user, f.hash, f.err
}
func (f *fakeRepo) SaveSession(_ context.Context, token, id string, expires time.Time) error {
	f.saved = token
	return f.saveErr
}
func (f *fakeRepo) SessionUser(context.Context, string) (domain.User, error) { return f.user, f.err }
func (f *fakeRepo) DeleteSession(_ context.Context, token string) error {
	f.deleted = token
	return f.err
}
func (f *fakeRepo) CreateUser(_ context.Context, email, hash string) error {
	f.user.Email = email
	f.hash = hash
	return f.err
}
func TestPasswords(t *testing.T) {
	hash := HashPassword("correct horse battery")
	require.True(t, VerifyPassword(hash, "correct horse battery"))
	require.False(t, VerifyPassword(hash, "wrong"))
	for _, bad := range []string{"", "$argon2id$v=18$m=65536,t=3,p=2$x$x", strings.Replace(hash, "m=65536", "m=1", 1), strings.Replace(hash, strings.Split(hash, "$")[4], "!", 1), strings.Replace(hash, strings.Split(hash, "$")[5], "!", 1), "$argon2id$v=19$m=65536,t=3,p=2$AA$AA"} {
		require.False(t, VerifyPassword(bad, "x"))
	}
	require.Len(t, NewToken(), 43)
	require.NotEqual(t, NewToken(), NewToken())
	require.Len(t, TokenHash("a"), 64)
}
func TestService(t *testing.T) {
	ctx := context.Background()
	repo := &fakeRepo{user: domain.User{ID: "id", Email: "admin@example.com", Role: "admin"}, hash: HashPassword("long password here")}
	s := New(repo)
	u, token, e := s.Login(ctx, "ADMIN@example.com", "long password here")
	require.NoError(t, e)
	require.Equal(t, repo.user, u)
	require.Equal(t, TokenHash(token), repo.saved)
	_, _, e = s.Login(ctx, "invalid", "short")
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	_, _, e = s.Login(ctx, "admin@example.com", "wrong password here")
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	repo.err = domain.ErrNotFound
	_, _, e = s.Login(ctx, "absent@example.com", "long password here")
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	repo.err = errors.New("db down")
	_, _, e = s.Login(ctx, "admin@example.com", "long password here")
	require.ErrorIs(t, e, repo.err)
	repo.err = nil
	repo.saveErr = errors.New("write error")
	_, _, e = s.Login(ctx, "admin@example.com", "long password here")
	require.ErrorIs(t, e, repo.saveErr)
	_, e = s.Current(ctx, "")
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	_, e = s.Current(ctx, token)
	require.NoError(t, e)
	repo.err = domain.ErrNotFound
	_, e = s.Current(ctx, token)
	require.ErrorIs(t, e, domain.ErrUnauthorized)
	repo.err = nil
	require.NoError(t, s.Logout(ctx, token))
	require.Equal(t, TokenHash(token), repo.deleted)
	require.ErrorIs(t, s.Bootstrap(ctx, "invalid", "short"), domain.ErrInvalid)
	require.NoError(t, s.Bootstrap(ctx, " Admin@example.com ", "another long password"))
	require.Equal(t, "admin@example.com", repo.user.Email)
	require.True(t, VerifyPassword(repo.hash, "another long password"))
	require.False(t, ValidCredentials("Name <a@example.com>", "long password here"))
	require.False(t, ValidCredentials("a@example.com", strings.Repeat("x", 129)))
}
