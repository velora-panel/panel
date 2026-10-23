package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/velora/panel/internal/domain"
)

type Repository interface {
	Credentials(context.Context, string) (domain.User, string, error)
	SaveSession(context.Context, string, string, string, time.Time) error
	SessionUser(context.Context, string) (domain.User, error)
	DeleteSession(context.Context, string) error
}
type Service struct {
	repo  Repository
	dummy string
}

func New(repo Repository) *Service { return &Service{repo: repo, dummy: HashPassword(NewToken())} }
func ValidCredentials(email, password string) bool {
	addr, e := mail.ParseAddress(email)
	return e == nil && addr.Address == email && len(email) <= 254 && len(password) >= 12 && len(password) <= 128
}
func (s *Service) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !ValidCredentials(email, password) {
		return domain.User{}, "", domain.ErrUnauthorized
	}
	user, hash, err := s.repo.Credentials(ctx, email)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, "", err
		}
		hash = s.dummy
	}
	valid := VerifyPassword(hash, password)
	if err != nil || !valid || user.Disabled {
		return domain.User{}, "", domain.ErrUnauthorized
	}
	token := NewToken()
	if err = s.repo.SaveSession(ctx, TokenHash(token), user.ID, hash, time.Now().Add(12*time.Hour)); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, "", domain.ErrUnauthorized
		}
		return domain.User{}, "", err
	}
	return user, token, nil
}
func (s *Service) Current(ctx context.Context, token string) (domain.User, error) {
	decoded, decodeErr := base64.RawURLEncoding.DecodeString(token)
	if len(token) != 43 || decodeErr != nil || len(decoded) != 32 {
		return domain.User{}, domain.ErrUnauthorized
	}
	user, err := s.repo.SessionUser(ctx, TokenHash(token))
	if errors.Is(err, domain.ErrNotFound) || user.Disabled {
		return domain.User{}, domain.ErrUnauthorized
	}
	return user, err
}
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.repo.DeleteSession(ctx, TokenHash(token))
}
