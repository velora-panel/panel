package domain

import "fmt"

type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return e.Message }
func (e ValidationError) Unwrap() error { return ErrInvalid }
func Invalid(message string) error      { return ValidationError{Message: message} }

var ErrLastAdmin = fmt.Errorf("the last active administrator cannot be disabled or demoted: %w", ErrConflict)
var ErrSchema = fmt.Errorf("Database schema is not initialized. Run: velora migrate up")

type AccountInput struct {
	CurrentPassword string `json:"current_password"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	Role            string `json:"role"`
}
type AccountUpdate struct {
	Role            string `json:"role"`
	Disabled        bool   `json:"disabled"`
	CurrentPassword string `json:"current_password"`
}
type PasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
type ProfileInput struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
}

func RequireAdmin(user User) error {
	if user.Role != "admin" || user.Disabled {
		return ErrForbidden
	}
	return nil
}
