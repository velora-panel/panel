package commands

import (
	"context"
	"io"

	"github.com/velora/panel/application"
)

type Lifecycle interface {
	Close()
	CheckSchema(context.Context) error
	Serve(context.Context, io.Writer) error
}
type AccountService interface {
	Create(context.Context, application.User, application.AccountInput) (application.User, error)
	List(context.Context, application.User) ([]application.User, error)
	OperatorUpdate(context.Context, string, bool) error
	IssueReset(context.Context, application.User, string) (string, error)
	Reset(context.Context, string, string) error
}
type CredentialRepository interface {
	Credentials(context.Context, string) (application.User, string, error)
}
type Runtime struct {
	Lifecycle
	Users AccountService
	Repo  CredentialRepository
}

func (r Runtime) Create(ctx context.Context, input application.AccountInput) error {
	_, e := r.Users.Create(ctx, application.User{Role: "admin"}, input)
	return e
}
func (r Runtime) List(ctx context.Context, admins bool) ([]application.User, error) {
	users, e := r.Users.List(ctx, application.User{Role: "admin"})
	if e != nil {
		return nil, e
	}
	result := make([]application.User, 0, len(users))
	for _, u := range users {
		if !admins || u.Role == "admin" {
			result = append(result, u)
		}
	}
	return result, nil
}
func (r Runtime) SetDisabled(ctx context.Context, email string, disabled bool) error {
	u, _, e := r.Repo.Credentials(ctx, email)
	if e != nil {
		return e
	}
	return r.Users.OperatorUpdate(ctx, u.ID, disabled)
}
func (r Runtime) Reset(ctx context.Context, email, password string) error {
	u, _, e := r.Repo.Credentials(ctx, email)
	if e != nil {
		return e
	}
	token, e := r.Users.IssueReset(ctx, application.User{Role: "admin"}, u.ID)
	if e != nil {
		return e
	}
	return r.Users.Reset(ctx, token, password)
}
