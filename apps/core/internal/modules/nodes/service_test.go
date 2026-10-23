package nodes

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/velora/panel/internal/domain"
)

type fakeRepo struct{ input domain.NodeInput }

func (f *fakeRepo) ListNodes(context.Context) ([]domain.Node, error) {
	return []domain.Node{}, nil
}
func (f *fakeRepo) Node(_ context.Context, id string) (domain.Node, error) {
	return domain.Node{ID: id}, nil
}
func (f *fakeRepo) CreateNode(_ context.Context, input domain.NodeInput) (domain.Node, error) {
	f.input = input
	return domain.Node{Name: input.Name}, nil
}
func TestService(t *testing.T) {
	ctx := context.Background()
	r := &fakeRepo{}
	s := New(r)
	_, e := s.List(ctx)
	require.NoError(t, e)
	_, e = s.Get(ctx, "bad")
	require.ErrorIs(t, e, domain.ErrInvalid)
	_, e = s.Get(ctx, "00000000-0000-0000-0000-000000000001")
	require.NoError(t, e)
	_, e = s.Create(ctx, domain.User{Role: "user"}, domain.NodeInput{Name: "x"})
	require.ErrorIs(t, e, domain.ErrForbidden)
	for _, input := range []domain.NodeInput{{Name: " "}, {Name: strings.Repeat("x", 81)}, {Name: "x", Description: strings.Repeat("x", 501)}} {
		_, e = s.Create(ctx, domain.User{Role: "admin"}, input)
		require.ErrorIs(t, e, domain.ErrInvalid)
	}
	_, e = s.Create(ctx, domain.User{Role: "admin"}, domain.NodeInput{Name: " West ", Description: " Rack 1 "})
	require.NoError(t, e)
	require.Equal(t, "West", r.input.Name)
	require.Equal(t, "Rack 1", r.input.Description)
}
