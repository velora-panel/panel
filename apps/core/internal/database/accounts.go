package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/velora/panel/internal/database/generated"
	"github.com/velora/panel/internal/domain"
)

type Beginner interface {
	Begin(context.Context) (pgx.Tx, error)
}
type Transaction func(context.Context, func(generated.Querier) error) error

func Transactions(db Beginner) Transaction {
	return func(ctx context.Context, work func(generated.Querier) error) error {
		tx, e := db.Begin(ctx)
		if e != nil {
			return e
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if e = work(generated.New(tx)); e != nil {
			return e
		}
		return tx.Commit(ctx)
	}
}
func (r *Repository) InsertAccount(ctx context.Context, input domain.AccountInput, hash string) (domain.User, error) {
	row, e := r.queries.CreateUser(ctx, generated.CreateUserParams{Email: input.Email, PasswordHash: hash, Role: input.Role})
	return user(row), translate(e)
}
func (r *Repository) Account(ctx context.Context, id string) (domain.User, string, error) {
	parsed, e := uuid.Parse(id)
	if e != nil {
		return domain.User{}, "", domain.ErrInvalid
	}
	row, e := r.queries.UserByID(ctx, parsed)
	return user(row), row.PasswordHash, translate(e)
}
func (r *Repository) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, e := r.queries.ListUsers(ctx)
	result := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		result = append(result, user(row))
	}
	return result, translate(e)
}
func (r *Repository) ChangeAccount(ctx context.Context, id string, input domain.AccountUpdate) (domain.User, error) {
	parsed, e := uuid.Parse(id)
	if e != nil {
		return domain.User{}, domain.ErrInvalid
	}
	var result domain.User
	e = r.transact(ctx, func(q generated.Querier) error {
		if err := q.LockAccounts(ctx); err != nil {
			return err
		}
		row, err := q.UserByID(ctx, parsed)
		if err != nil {
			return err
		}
		if row.Role == "admin" && !row.Disabled && (input.Role != "admin" || input.Disabled) {
			count, err := q.CountActiveAdmins(ctx)
			if err != nil {
				return err
			}
			if count <= 1 {
				return domain.ErrLastAdmin
			}
		}
		row, err = q.UpdateAccount(ctx, generated.UpdateAccountParams{ID: parsed, Role: input.Role, Disabled: input.Disabled})
		if err != nil {
			return err
		}
		if err = q.RevokeSessions(ctx, generated.RevokeSessionsParams{UserID: parsed, TokenHash: ""}); err != nil {
			return err
		}
		if err = q.DeleteResets(ctx, parsed); err != nil {
			return err
		}
		result = user(row)
		return nil
	})
	return result, translate(e)
}
func (r *Repository) ChangePassword(ctx context.Context, id, oldHash, newHash, keep string) error {
	parsed, e := uuid.Parse(id)
	if e != nil {
		return domain.ErrInvalid
	}
	return translate(r.transact(ctx, func(q generated.Querier) error {
		if _, err := q.UpdatePassword(ctx, generated.UpdatePasswordParams{ID: parsed, PasswordHash: newHash, PasswordHash_2: oldHash}); err != nil {
			return err
		}
		if err := q.DeleteResets(ctx, parsed); err != nil {
			return err
		}
		return q.RevokeSessions(ctx, generated.RevokeSessionsParams{UserID: parsed, TokenHash: keep})
	}))
}
func (r *Repository) ChangeProfile(ctx context.Context, id, email, hash string) (domain.User, error) {
	parsed, e := uuid.Parse(id)
	if e != nil {
		return domain.User{}, domain.ErrInvalid
	}
	row, e := r.queries.UpdateProfile(ctx, generated.UpdateProfileParams{ID: parsed, Email: email, PasswordHash: hash})
	return user(row), translate(e)
}
func (r *Repository) RevokeAll(ctx context.Context, id string) error {
	parsed, e := uuid.Parse(id)
	if e != nil {
		return domain.ErrInvalid
	}
	return translate(r.queries.RevokeSessions(ctx, generated.RevokeSessionsParams{UserID: parsed, TokenHash: ""}))
}
func (r *Repository) SaveReset(ctx context.Context, id, hash string, expires time.Time) error {
	parsed, e := uuid.Parse(id)
	if e != nil {
		return domain.ErrInvalid
	}
	return translate(r.transact(ctx, func(q generated.Querier) error {
		if e := q.LockAccounts(ctx); e != nil {
			return e
		}
		if e := q.DeleteResets(ctx, parsed); e != nil {
			return e
		}
		return q.CreateReset(ctx, generated.CreateResetParams{TokenHash: hash, UserID: parsed, ExpiresAt: expires})
	}))
}
func (r *Repository) ConsumeReset(ctx context.Context, hash, passwordHash string) error {
	e := r.transact(ctx, func(q generated.Querier) error {
		id, e := q.ConsumeReset(ctx, hash)
		if e != nil {
			return e
		}
		if _, e = q.ResetPassword(ctx, generated.ResetPasswordParams{ID: id, PasswordHash: passwordHash}); e != nil {
			return e
		}
		if e = q.DeleteResets(ctx, id); e != nil {
			return e
		}
		return q.RevokeSessions(ctx, generated.RevokeSessionsParams{UserID: id, TokenHash: ""})
	})
	if e == pgx.ErrNoRows {
		return domain.ErrUnauthorized
	}
	return translate(e)
}
