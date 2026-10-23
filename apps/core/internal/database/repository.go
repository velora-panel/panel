package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/velora/panel/internal/database/generated"
	"github.com/velora/panel/internal/domain"
)

type Repository struct {
	queries  generated.Querier
	transact Transaction
}

func New(queries generated.Querier, transact Transaction) *Repository {
	return &Repository{queries: queries, transact: transact}
}
func translate(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == "23505" {
		return domain.ErrConflict
	}
	return err
}
func user(row generated.User) domain.User {
	return domain.User{ID: row.ID.String(), Email: row.Email, Role: row.Role, Disabled: row.Disabled}
}
func node(row generated.Node) domain.Node {
	return domain.Node{ID: row.ID.String(), Name: row.Name, Description: row.Description, Status: row.Status, CreatedAt: row.CreatedAt}
}
func (r *Repository) Credentials(ctx context.Context, email string) (domain.User, string, error) {
	row, err := r.queries.UserByEmail(ctx, email)
	return user(row), row.PasswordHash, translate(err)
}
func (r *Repository) CreateUser(ctx context.Context, email, hash string) error {
	_, err := r.queries.CreateUser(ctx, generated.CreateUserParams{Email: email, PasswordHash: hash, Role: "admin"})
	return translate(err)
}
func (r *Repository) SaveSession(ctx context.Context, hash, id, passwordHash string, expires time.Time) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrInvalid
	}
	_, err = r.queries.CreateSession(ctx, generated.CreateSessionParams{TokenHash: hash, ID: parsed, ExpiresAt: expires, PasswordHash: passwordHash})
	return translate(err)
}
func (r *Repository) SessionUser(ctx context.Context, hash string) (domain.User, error) {
	row, err := r.queries.SessionUser(ctx, hash)
	return user(row), translate(err)
}
func (r *Repository) DeleteSession(ctx context.Context, hash string) error {
	return translate(r.queries.DeleteSession(ctx, hash))
}
func (r *Repository) ListNodes(ctx context.Context) ([]domain.Node, error) {
	rows, err := r.queries.ListNodes(ctx)
	result := make([]domain.Node, 0, len(rows))
	for _, row := range rows {
		result = append(result, node(row))
	}
	return result, translate(err)
}
func (r *Repository) Node(ctx context.Context, id string) (domain.Node, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return domain.Node{}, domain.ErrInvalid
	}
	row, err := r.queries.NodeByID(ctx, parsed)
	return node(row), translate(err)
}
func (r *Repository) CreateNode(ctx context.Context, input domain.NodeInput) (domain.Node, error) {
	row, err := r.queries.CreateNode(ctx, generated.CreateNodeParams{Name: input.Name, Description: input.Description})
	return node(row), translate(err)
}
