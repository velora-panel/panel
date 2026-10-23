package commands

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/velora/panel/application"
)

type fakeBackend struct {
	err      error
	schema   error
	closed   bool
	created  application.AccountInput
	disabled bool
	reset    string
}

func (f *fakeBackend) Close()                                 { f.closed = true }
func (f *fakeBackend) CheckSchema(context.Context) error      { return f.schema }
func (f *fakeBackend) Serve(context.Context, io.Writer) error { return f.err }
func (f *fakeBackend) Create(_ context.Context, i application.AccountInput) error {
	f.created = i
	return f.err
}
func (f *fakeBackend) List(context.Context, bool) ([]application.User, error) {
	return []application.User{{Email: "admin@example.com", Role: "admin"}}, f.err
}
func (f *fakeBackend) SetDisabled(_ context.Context, _ string, v bool) error {
	f.disabled = v
	return f.err
}
func (f *fakeBackend) Reset(_ context.Context, _, p string) error { f.reset = p; return f.err }

const credentials = `{"email":"admin@example.com","password":"long password here","password_confirmation":"long password here"}`

func TestRunner(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		args []string
		bad  bool
	}{
		{nil, false}, {[]string{"help"}, false}, {[]string{"--help"}, false}, {[]string{"version"}, false}, {[]string{"migrate"}, true}, {[]string{"unknown"}, true}, {[]string{"doctor", "--bad"}, true}, {[]string{"doctor", "--help"}, false}, {[]string{"doctor", "extra"}, true}, {[]string{"migrate", "bad"}, true}, {[]string{"migrate", "down"}, true}, {[]string{"migrate", "down", "--allow-destructive"}, false}, {[]string{"migrate", "up"}, false}, {[]string{"admin", "bad"}, true}, {[]string{"user", "bad"}, true}, {[]string{"doctor"}, false}, {[]string{"serve"}, false}, {[]string{"admin", "list"}, false}, {[]string{"user", "list"}, false}, {[]string{"user", "disable"}, true}, {[]string{"user", "disable", "--email", "a@b.com"}, false}, {[]string{"user", "enable", "--email", "a@b.com"}, false}, {[]string{"admin", "create", "--password-stdin"}, false}, {[]string{"admin", "reset-password", "--password-stdin"}, false}, {[]string{"user", "create", "--password-stdin", "--role", "bad"}, true}, {[]string{"admin", "create"}, true},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			b := &fakeBackend{}
			out := &bytes.Buffer{}
			r := Runner{Out: out, Err: out, Input: Input{In: strings.NewReader(credentials), Out: out}, Open: func(context.Context) (Backend, error) { return b, nil }, Migrate: func(context.Context, string, io.Writer) error { return nil }}
			e := r.Run(ctx, tc.args)
			if (e != nil) != tc.bad {
				t.Fatalf("error=%v, expected failure=%v", e, tc.bad)
			}
			if strings.Contains(out.String(), "long password") {
				t.Fatal("secret printed")
			}
			if len(tc.args) > 1 && tc.args[1] == "create" && !tc.bad && b.created.Role != "admin" {
				t.Fatal("admin role missing")
			}
		})
	}
	fail := errors.New("backend unavailable")
	for _, args := range [][]string{{"user", "list"}, {"user", "disable", "--email", "a@b.com"}, {"admin", "create", "--password-stdin"}, {"admin", "reset-password", "--password-stdin"}} {
		b := &fakeBackend{err: fail}
		r := Runner{Out: io.Discard, Err: io.Discard, Input: Input{In: strings.NewReader(credentials)}, Open: func(context.Context) (Backend, error) { return b, nil }}
		if !errors.Is(r.Run(ctx, args), fail) || !b.closed {
			t.Fatal("backend failure or close lost")
		}
	}
	r := Runner{Out: io.Discard, Err: io.Discard, Open: func(context.Context) (Backend, error) { return nil, fail }}
	if !errors.Is(r.Run(ctx, []string{"doctor"}), fail) {
		t.Fatal("open failure lost")
	}
	b := &fakeBackend{schema: fail}
	r.Open = func(context.Context) (Backend, error) { return b, nil }
	if !errors.Is(r.Run(ctx, []string{"doctor"}), fail) || !b.closed {
		t.Fatal("schema failure lost")
	}
}
func TestInput(t *testing.T) {
	for _, raw := range []string{"", `{}`, `{"unknown":true}`, credentials + ` {}`, strings.Replace(credentials, "long password here", "short", 2), strings.Replace(credentials, `"password_confirmation":"long password here"`, `"password_confirmation":"different"`, 1)} {
		_, e := (Input{In: strings.NewReader(raw)}).Read("", true)
		if e == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	c, e := (Input{In: strings.NewReader(credentials)}).Read(" OVERRIDE@example.com ", true)
	if e != nil || c.Email != "override@example.com" {
		t.Fatal(c.Email, e)
	}
	for _, email := range []string{"", "preset@example.com"} {
		out := &bytes.Buffer{}
		i := Input{In: strings.NewReader("ADMIN@example.com\r\n"), Out: out, Terminal: true, Password: func() (string, error) { return "long password here", nil }}
		c, e = i.Read(email, false)
		if e != nil || c.Email == "" || strings.Contains(out.String(), c.Password) {
			t.Fatal("terminal input failed or exposed secret", e)
		}
	}
	fail := errors.New("read failure")
	for _, at := range []int{1, 2} {
		n := 0
		i := Input{In: strings.NewReader("a@b.com\n"), Out: io.Discard, Terminal: true, Password: func() (string, error) {
			n++
			if n == at {
				return "", fail
			}
			return "long password here", nil
		}}
		_, e = i.Read("", false)
		if !errors.Is(e, fail) {
			t.Fatal("terminal error lost")
		}
	}
	for _, raw := range []string{"", strings.Repeat("a", 513)} {
		_, e := (Input{In: strings.NewReader(raw), Out: io.Discard, Terminal: true}).Read("", false)
		if e == nil {
			t.Fatal("invalid line accepted")
		}
	}
}

type fakeAccounts struct {
	err   error
	token string
}

func (f *fakeAccounts) Create(context.Context, application.User, application.AccountInput) (application.User, error) {
	return application.User{}, f.err
}
func (f *fakeAccounts) List(context.Context, application.User) ([]application.User, error) {
	return []application.User{{Role: "admin"}, {Role: "user"}}, f.err
}
func (f *fakeAccounts) OperatorUpdate(context.Context, string, bool) error { return f.err }
func (f *fakeAccounts) IssueReset(context.Context, application.User, string) (string, error) {
	return "opaque", f.err
}
func (f *fakeAccounts) Reset(_ context.Context, token, _ string) error { f.token = token; return f.err }

type fakeCredentials struct{ err error }

func (f *fakeCredentials) Credentials(context.Context, string) (application.User, string, error) {
	return application.User{ID: "id"}, "hash", f.err
}
func TestRuntime(t *testing.T) {
	ctx := context.Background()
	a := &fakeAccounts{}
	c := &fakeCredentials{}
	r := Runtime{Users: a, Repo: c}
	if e := r.Create(ctx, application.AccountInput{}); e != nil {
		t.Fatal(e)
	}
	for _, admin := range []bool{true, false} {
		u, e := r.List(ctx, admin)
		if e != nil || len(u) == 0 || (admin && len(u) != 1) {
			t.Fatal(u, e)
		}
	}
	if e := r.SetDisabled(ctx, "email", true); e != nil {
		t.Fatal(e)
	}
	if e := r.Reset(ctx, "email", "password"); e != nil || a.token != "opaque" {
		t.Fatal(e)
	}
	fail := errors.New("db")
	c.err = fail
	if !errors.Is(r.SetDisabled(ctx, "email", true), fail) || !errors.Is(r.Reset(ctx, "email", "password"), fail) {
		t.Fatal("credential failure lost")
	}
	c.err = nil
	a.err = fail
	if !errors.Is(r.Reset(ctx, "email", "password"), fail) {
		t.Fatal("issuance failure lost")
	}
	if _, e := r.List(ctx, true); !errors.Is(e, fail) {
		t.Fatal("list failure lost")
	}
}
