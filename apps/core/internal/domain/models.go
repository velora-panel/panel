package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalid      = errors.New("invalid input")
	ErrUnauthorized = errors.New("authentication required")
	ErrForbidden    = errors.New("permission denied")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("already exists")
)

type User struct {
	Disabled bool   `json:"disabled"`
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
type Node struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
type NodeInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
