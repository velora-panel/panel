package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/velora/panel/internal/domain"
)

type AccountService interface {
	Create(context.Context, domain.User, domain.AccountInput) (domain.User, error)
	List(context.Context, domain.User) ([]domain.User, error)
	Get(context.Context, domain.User, string) (domain.User, error)
	Update(context.Context, domain.User, string, domain.AccountUpdate) (domain.User, error)
	Password(context.Context, domain.User, string, domain.PasswordInput) error
	Profile(context.Context, domain.User, domain.ProfileInput) (domain.User, error)
	Revoke(context.Context, domain.User, string, string) error
	Reauthenticate(context.Context, domain.User, string) error
}

func actor(c fiber.Ctx) domain.User { return c.Locals("user").(domain.User) }
func body[T any](c fiber.Ctx) (T, error) {
	var input T
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil {
		return input, domain.ErrInvalid
	}
	var extra interface{}
	if decoder.Decode(&extra) != io.EOF {
		return input, domain.ErrInvalid
	}
	return input, nil
}
func registerAccounts(group fiber.Router, a AccountService) {
	sensitive := limiter.New(limiter.Config{Max: 10, Expiration: time.Minute, LimitReached: func(_ fiber.Ctx) error { return fiber.ErrTooManyRequests }})
	group.Patch("/auth/password", sensitive, func(c fiber.Ctx) error {
		input, e := body[domain.PasswordInput](c)
		if e != nil {
			return e
		}
		if e = a.Password(c.Context(), actor(c), c.Cookies(cookieName), input); e != nil {
			return e
		}
		return c.SendStatus(204)
	})
	group.Patch("/auth/profile", sensitive, func(c fiber.Ctx) error {
		input, e := body[domain.ProfileInput](c)
		if e != nil {
			return e
		}
		u, e := a.Profile(c.Context(), actor(c), input)
		if e != nil {
			return e
		}
		return c.JSON(u)
	})
	admins := group.Group("/users", func(c fiber.Ctx) error {
		if e := domain.RequireAdmin(actor(c)); e != nil {
			return e
		}
		return c.Next()
	})
	admins.Get("/", func(c fiber.Ctx) error {
		users, e := a.List(c.Context(), actor(c))
		if e != nil {
			return e
		}
		return c.JSON(users)
	})
	admins.Get("/:id", func(c fiber.Ctx) error {
		u, e := a.Get(c.Context(), actor(c), c.Params("id"))
		if e != nil {
			return e
		}
		return c.JSON(u)
	})
	admins.Post("/", sensitive, func(c fiber.Ctx) error {
		input, e := body[domain.AccountInput](c)
		if e != nil {
			return e
		}
		if e = a.Reauthenticate(c.Context(), actor(c), input.CurrentPassword); e != nil {
			return e
		}
		u, e := a.Create(c.Context(), actor(c), input)
		if e != nil {
			return e
		}
		return c.Status(201).JSON(u)
	})
	admins.Patch("/:id", sensitive, func(c fiber.Ctx) error {
		input, e := body[domain.AccountUpdate](c)
		if e != nil {
			return e
		}
		u, e := a.Update(c.Context(), actor(c), c.Params("id"), input)
		if e != nil {
			return e
		}
		return c.JSON(u)
	})
	admins.Post("/:id/revoke", sensitive, func(c fiber.Ctx) error {
		input, e := body[struct {
			CurrentPassword string `json:"current_password"`
		}](c)
		if e != nil {
			return e
		}
		if e = a.Revoke(c.Context(), actor(c), c.Params("id"), input.CurrentPassword); e != nil {
			return e
		}
		return c.SendStatus(204)
	})
}
