package nodes

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/velora/panel/internal/domain"
)

type Repository interface {
	ListNodes(context.Context) ([]domain.Node, error)
	Node(context.Context, string) (domain.Node, error)
	CreateNode(context.Context, domain.NodeInput) (domain.Node, error)
}
type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context) ([]domain.Node, error) {
	return s.repo.ListNodes(ctx)
}
func (s *Service) Get(ctx context.Context, id string) (domain.Node, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.Node{}, domain.ErrInvalid
	}
	return s.repo.Node(ctx, id)
}
func (s *Service) Create(ctx context.Context, user domain.User, input domain.NodeInput) (domain.Node, error) {
	if user.Role != "admin" {
		return domain.Node{}, domain.ErrForbidden
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 80 || utf8.RuneCountInString(input.Description) > 500 {
		return domain.Node{}, domain.ErrInvalid
	}
	return s.repo.CreateNode(ctx, input)
}
