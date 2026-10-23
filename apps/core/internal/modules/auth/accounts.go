package auth

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/velora/panel/internal/domain"
)

// AccountRepository owns atomic credential/session changes and last-admin protection.
type AccountRepository interface {
	Credentials(context.Context, string) (domain.User, string, error)
	Account(context.Context, string) (domain.User, string, error)
	InsertAccount(context.Context, domain.AccountInput, string) (domain.User, error)
	ListUsers(context.Context) ([]domain.User, error)
	ChangeAccount(context.Context, string, domain.AccountUpdate) (domain.User, error)
	ChangePassword(context.Context, string, string, string, string) error
	ChangeProfile(context.Context, string, string, string) (domain.User, error)
	RevokeAll(context.Context, string) error
	SaveReset(context.Context, string, string, time.Time) error
	ConsumeReset(context.Context, string, string) error
}
type Accounts struct{ repo AccountRepository }

func NewAccounts(repo AccountRepository) *Accounts { return &Accounts{repo: repo} }
func ValidateEmail(email string) error {
	a, e := mail.ParseAddress(email)
	if e != nil || a.Address != email || len(email) > 254 {
		return domain.Invalid("Enter a valid email address.")
	}
	return nil
}
func ValidatePassword(password string) error {
	if len(password) < 12 {
		return domain.Invalid("Password must contain at least 12 characters.")
	}
	if len(password) > 128 {
		return domain.Invalid("Password must not exceed 128 bytes.")
	}
	return nil
}
func ValidateAccount(email, password, role string) error {
	if e := ValidateEmail(email); e != nil {
		return e
	}
	if e := ValidatePassword(password); e != nil {
		return e
	}
	if role != "admin" && role != "user" {
		return domain.Invalid("Role must be admin or user.")
	}
	return nil
}
func validID(id string) error {
	if _, e := uuid.Parse(id); e != nil {
		return domain.Invalid("Invalid user ID.")
	}
	return nil
}
func (a *Accounts) Create(ctx context.Context, actor domain.User, input domain.AccountInput) (domain.User, error) {
	if e := domain.RequireAdmin(actor); e != nil {
		return domain.User{}, e
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if e := ValidateAccount(input.Email, input.Password, input.Role); e != nil {
		return domain.User{}, e
	}
	return a.repo.InsertAccount(ctx, input, HashPassword(input.Password))
}
func (a *Accounts) List(ctx context.Context, actor domain.User) ([]domain.User, error) {
	if e := domain.RequireAdmin(actor); e != nil {
		return nil, e
	}
	return a.repo.ListUsers(ctx)
}
func (a *Accounts) Get(ctx context.Context, actor domain.User, id string) (domain.User, error) {
	if actor.ID != id {
		if e := domain.RequireAdmin(actor); e != nil {
			return domain.User{}, e
		}
	}
	if e := validID(id); e != nil {
		return domain.User{}, e
	}
	u, _, e := a.repo.Account(ctx, id)
	return u, e
}
func (a *Accounts) Reauthenticate(ctx context.Context, actor domain.User, password string) error {
	_, hash, e := a.repo.Account(ctx, actor.ID)
	if e != nil {
		return e
	}
	if !VerifyPassword(hash, password) {
		return domain.ErrUnauthorized
	}
	return nil
}
func (a *Accounts) Update(ctx context.Context, actor domain.User, id string, input domain.AccountUpdate) (domain.User, error) {
	if e := domain.RequireAdmin(actor); e != nil {
		return domain.User{}, e
	}
	if e := validID(id); e != nil {
		return domain.User{}, e
	}
	if input.Role != "admin" && input.Role != "user" {
		return domain.User{}, domain.Invalid("Role must be admin or user.")
	}
	if e := a.Reauthenticate(ctx, actor, input.CurrentPassword); e != nil {
		return domain.User{}, e
	}
	return a.repo.ChangeAccount(ctx, id, input)
}

// OperatorUpdate is only exposed through the privileged local CLI, never an HTTP route.
func (a *Accounts) OperatorUpdate(ctx context.Context, id string, disabled bool) error {
	if e := validID(id); e != nil {
		return e
	}
	u, _, e := a.repo.Account(ctx, id)
	if e != nil {
		return e
	}
	_, e = a.repo.ChangeAccount(ctx, id, domain.AccountUpdate{Role: u.Role, Disabled: disabled})
	return e
}
func (a *Accounts) Password(ctx context.Context, actor domain.User, token string, input domain.PasswordInput) error {
	if e := ValidatePassword(input.NewPassword); e != nil {
		return e
	}
	_, hash, e := a.repo.Account(ctx, actor.ID)
	if e != nil {
		return e
	}
	if !VerifyPassword(hash, input.CurrentPassword) {
		return domain.ErrUnauthorized
	}
	return a.repo.ChangePassword(ctx, actor.ID, hash, HashPassword(input.NewPassword), TokenHash(token))
}
func (a *Accounts) Profile(ctx context.Context, actor domain.User, input domain.ProfileInput) (domain.User, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if e := ValidateEmail(input.Email); e != nil {
		return domain.User{}, e
	}
	_, hash, e := a.repo.Account(ctx, actor.ID)
	if e != nil {
		return domain.User{}, e
	}
	if !VerifyPassword(hash, input.CurrentPassword) {
		return domain.User{}, domain.ErrUnauthorized
	}
	return a.repo.ChangeProfile(ctx, actor.ID, input.Email, hash)
}
func (a *Accounts) Revoke(ctx context.Context, actor domain.User, id, password string) error {
	if e := domain.RequireAdmin(actor); e != nil {
		return e
	}
	if e := validID(id); e != nil {
		return e
	}
	if e := a.Reauthenticate(ctx, actor, password); e != nil {
		return e
	}
	return a.repo.RevokeAll(ctx, id)
}

// Reset issuance is privileged. No public reset endpoint exists without delivery infrastructure.
func (a *Accounts) IssueReset(ctx context.Context, actor domain.User, id string) (string, error) {
	if e := domain.RequireAdmin(actor); e != nil {
		return "", e
	}
	if e := validID(id); e != nil {
		return "", e
	}
	u, _, e := a.repo.Account(ctx, id)
	if e != nil {
		return "", e
	}
	if u.Disabled {
		return "", domain.ErrForbidden
	}
	token := NewToken()
	if e = a.repo.SaveReset(ctx, id, TokenHash(token), time.Now().Add(15*time.Minute)); e != nil {
		return "", e
	}
	return token, nil
}
func (a *Accounts) Reset(ctx context.Context, token, password string) error {
	if len(token) != 43 {
		return domain.ErrUnauthorized
	}
	if e := ValidatePassword(password); e != nil {
		return e
	}
	return a.repo.ConsumeReset(ctx, TokenHash(token), HashPassword(password))
}
