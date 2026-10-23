package http

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/velora/panel/internal/config"
	"github.com/velora/panel/internal/domain"
)

type fakeAccounts struct{ err, reauth error }

func (f *fakeAccounts) Create(context.Context, domain.User, domain.AccountInput) (domain.User, error) {
	return domain.User{}, f.err
}
func (f *fakeAccounts) List(context.Context, domain.User) ([]domain.User, error) {
	return []domain.User{}, f.err
}
func (f *fakeAccounts) Get(context.Context, domain.User, string) (domain.User, error) {
	return domain.User{}, f.err
}
func (f *fakeAccounts) Update(context.Context, domain.User, string, domain.AccountUpdate) (domain.User, error) {
	return domain.User{}, f.err
}
func (f *fakeAccounts) Password(context.Context, domain.User, string, domain.PasswordInput) error {
	return f.err
}
func (f *fakeAccounts) Profile(context.Context, domain.User, domain.ProfileInput) (domain.User, error) {
	return domain.User{}, f.err
}
func (f *fakeAccounts) Revoke(context.Context, domain.User, string, string) error { return f.err }
func (f *fakeAccounts) Reauthenticate(context.Context, domain.User, string) error { return f.reauth }
func TestAccountRoutes(t *testing.T) {
	for _, tc := range []struct {
		method, path string
		ok           int
	}{{"GET", "/users", 200}, {"GET", "/users/id", 200}, {"POST", "/users", 201}, {"PATCH", "/users/id", 200}, {"POST", "/users/id/revoke", 204}, {"PATCH", "/auth/password", 204}, {"PATCH", "/auth/profile", 200}} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			a := &fakeAuth{role: "admin"}
			accounts := &fakeAccounts{}
			app := New(config.Config{PublicURL: "https://console.test"}, a, &fakeNodes{}, func(context.Context) error { return nil }, slog.New(slog.NewTextHandler(io.Discard, nil)), accounts)
			request := func(body, token string) int {
				r := httptest.NewRequest(tc.method, "/api/v1"+tc.path, strings.NewReader(body))
				r.Header.Set("Origin", "https://console.test")
				r.Header.Set("Content-Type", "application/json")
				if token != "" {
					r.Header.Set("Cookie", cookieName+"="+token)
				}
				res, e := app.Test(r)
				require.NoError(t, e)
				require.NoError(t, res.Body.Close())
				return res.StatusCode
			}
			require.Equal(t, 401, request("{}", ""))
			require.Equal(t, tc.ok, request("{}", "token"))
			accounts.err = domain.Invalid("Invalid account details.")
			require.Equal(t, 400, request("{}", "token"))
			accounts.err = domain.ErrLastAdmin
			require.Equal(t, 409, request("{}", "token"))
			accounts.err = nil
			if tc.method != "GET" {
				require.Equal(t, 400, request("{", "token"))
				require.Equal(t, 400, request("{} {}", "token"))
				require.Equal(t, 400, request(`{"unexpected":true}`, "token"))
			}
			if tc.method == "POST" && tc.path == "/users" {
				accounts.reauth = domain.ErrUnauthorized
				require.Equal(t, 401, request("{}", "token"))
			}
			if strings.HasPrefix(tc.path, "/users") {
				a.role = "user"
				require.Equal(t, 403, request("{}", "token"))
			}
		})
	}
}
func TestSensitiveRateLimit(t *testing.T) {
	app := New(config.Config{PublicURL: "https://console.test"}, &fakeAuth{role: "admin"}, &fakeNodes{}, func(context.Context) error { return nil }, slog.New(slog.NewTextHandler(io.Discard, nil)), &fakeAccounts{})
	for i := 0; i < 11; i++ {
		r := httptest.NewRequest("PATCH", "/api/v1/auth/password", strings.NewReader("{}"))
		r.Header.Set("Origin", "https://console.test")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Cookie", cookieName+"=token")
		res, e := app.Test(r)
		require.NoError(t, e)
		if i == 10 {
			require.Equal(t, 429, res.StatusCode)
		}
		require.NoError(t, res.Body.Close())
	}
}
