package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
	"github.com/velora/panel/internal/config"
	"github.com/velora/panel/internal/domain"
)

type fakeAuth struct {
	err  error
	role string
}

func (f *fakeAuth) Login(context.Context, string, string) (domain.User, string, error) {
	return domain.User{Role: f.role}, "token", f.err
}
func (f *fakeAuth) Current(_ context.Context, token string) (domain.User, error) {
	if token == "" {
		return domain.User{}, domain.ErrUnauthorized
	}
	return domain.User{Role: f.role}, f.err
}
func (f *fakeAuth) Logout(context.Context, string) error { return f.err }

type fakeNodes struct{ err error }

func (f *fakeNodes) List(context.Context) ([]domain.Node, error) {
	return []domain.Node{}, f.err
}
func (f *fakeNodes) Get(context.Context, string) (domain.Node, error) {
	return domain.Node{}, f.err
}
func (f *fakeNodes) Create(_ context.Context, u domain.User, _ domain.NodeInput) (domain.Node, error) {
	if u.Role != "admin" {
		return domain.Node{}, domain.ErrForbidden
	}
	return domain.Node{}, f.err
}
func TestHTTP(t *testing.T) {
	a := &fakeAuth{role: "admin"}
	s := &fakeNodes{}
	var pingErr error
	app := New(config.Config{PublicURL: "https://console.test", SecureCookie: true}, a, s, func(context.Context) error { return pingErr }, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	request := func(method, path, body, origin, token string) int {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		if token != "" {
			r.Header.Set("Cookie", cookieName+"="+token)
		}
		response, e := app.Test(r)
		require.NoError(t, e)
		defer func() { require.NoError(t, response.Body.Close()) }()
		b, e := io.ReadAll(response.Body)
		require.NoError(t, e)
		require.NotEmpty(t, response.Header.Get("X-Request-ID"))
		require.NotContains(t, string(b), "secret database")
		return response.StatusCode
	}
	require.Equal(t, 200, request("GET", "/api/v1/health/live", "", "", ""))
	require.Equal(t, 200, request("GET", "/api/v1/health/ready", "", "", ""))
	pingErr = errors.New("secret database")
	require.Equal(t, 503, request("GET", "/api/v1/health/ready", "", "", ""))
	require.Equal(t, 401, request("GET", "/api/v1/nodes", "", "", ""))
	require.Equal(t, 403, request("POST", "/api/v1/auth/login", "{}", "https://evil.test", ""))
	r := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader("{}"))
	r.Header.Set("Origin", "https://console.test")
	res, e := app.Test(r)
	require.NoError(t, e)
	require.Equal(t, 400, res.StatusCode)
	require.NoError(t, res.Body.Close())
	require.Equal(t, 400, request("POST", "/api/v1/auth/login", "{", "https://console.test", ""))
	require.Equal(t, 200, request("POST", "/api/v1/auth/login", "{}", "https://console.test", ""))
	a.err = domain.ErrUnauthorized
	require.Equal(t, 401, request("POST", "/api/v1/auth/login", "{}", "https://console.test", ""))
	a.err = nil
	require.Equal(t, 200, request("GET", "/api/v1/auth/me", "", "", "token"))
	require.Equal(t, 204, request("POST", "/api/v1/auth/logout", "{}", "https://console.test", "token"))
	require.Equal(t, 200, request("GET", "/api/v1/nodes", "", "", "token"))
	require.Equal(t, 200, request("GET", "/api/v1/nodes/id", "", "", "token"))
	require.Equal(t, 400, request("POST", "/api/v1/nodes", "{", "https://console.test", "token"))
	require.Equal(t, 201, request("POST", "/api/v1/nodes", "{}", "https://console.test", "token"))
	a.role = "user"
	require.Equal(t, 403, request("POST", "/api/v1/nodes", "{}", "https://console.test", "token"))
	a.role = "admin"
	for _, err := range []error{domain.ErrNotFound, domain.ErrConflict, errors.New("secret database")} {
		s.err = err
		require.GreaterOrEqual(t, request("GET", "/api/v1/nodes", "", "", "token"), 400)
		require.GreaterOrEqual(t, request("GET", "/api/v1/nodes/id", "", "", "token"), 400)
		require.GreaterOrEqual(t, request("POST", "/api/v1/nodes", "{}", "https://console.test", "token"), 400)
	}
	require.Equal(t, 404, request("GET", "/missing", "", "", "token"))
	require.Equal(t, 404, request("HEAD", "/missing", "", "", "token"))
	require.Equal(t, 404, request("OPTIONS", "/missing", "", "", "token"))
	for range 6 {
		request("POST", "/api/v1/auth/login", "{}", "https://console.test", "")
	}
	require.Equal(t, 429, request("POST", "/api/v1/auth/login", "{}", "https://console.test", ""))
}
func TestLogoutError(t *testing.T) {
	s := Server{auth: &fakeAuth{err: errors.New("db")}}
	app := fiber.New()
	app.Post("/", s.logout)
	r, e := app.Test(httptest.NewRequest("POST", "/", nil))
	require.NoError(t, e)
	require.Equal(t, 500, r.StatusCode)
	require.NoError(t, r.Body.Close())
}
