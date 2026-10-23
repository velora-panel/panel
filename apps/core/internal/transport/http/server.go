package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/google/uuid"
	"github.com/velora/panel/internal/config"
	"github.com/velora/panel/internal/domain"
	"github.com/velora/panel/internal/modules/auth"
	"github.com/velora/panel/internal/modules/nodes"
)

const cookieName = "velora_session"

type AuthService interface {
	Login(context.Context, string, string) (domain.User, string, error)
	Current(context.Context, string) (domain.User, error)
	Logout(context.Context, string) error
}
type NodeService interface {
	List(context.Context) ([]domain.Node, error)
	Get(context.Context, string) (domain.Node, error)
	Create(context.Context, domain.User, domain.NodeInput) (domain.Node, error)
}

var _ AuthService = (*auth.Service)(nil)
var _ NodeService = (*nodes.Service)(nil)

type Server struct {
	auth  AuthService
	nodes NodeService
	cfg   config.Config
}

func New(cfg config.Config, authentication AuthService, nodeService NodeService, ping func(context.Context) error, logger *slog.Logger, accounts AccountService) *fiber.App {
	s := Server{auth: authentication, nodes: nodeService, cfg: cfg}
	app := fiber.New(fiber.Config{AppName: "Velora Core", BodyLimit: 16 * 1024, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, ErrorHandler: errorHandler})
	app.Use(func(c fiber.Ctx) error {
		c.Set("X-Request-ID", uuid.NewString())
		c.Set("Cache-Control", "no-store")
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		c.SetContext(ctx)
		start := time.Now()
		err := c.Next()
		if err != nil {
			err = errorHandler(c, err)
		}
		logger.Info("http request", "request_id", c.GetRespHeader("X-Request-ID"), "method", c.Method(), "status", c.Response().StatusCode(), "duration", time.Since(start))
		return err
	})
	app.Use(recover.New(), helmet.New())
	app.Get("/api/v1/health/live", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/api/v1/health/ready", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		if ping(ctx) != nil {
			return fiber.NewError(503, "Not ready")
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})
	// All browser mutations require our configured origin, including login (login CSRF).
	app.Use(func(c fiber.Ctx) error {
		if c.Method() != "GET" && c.Method() != "HEAD" && c.Method() != "OPTIONS" {
			if c.Get("Origin") != cfg.PublicURL {
				return domain.ErrForbidden
			}
			if c.Get("Content-Type") != "application/json" {
				return domain.ErrInvalid
			}
		}
		return c.Next()
	})
	app.Post("/api/v1/auth/login", limiter.New(limiter.Config{Max: 5, Expiration: time.Minute, LimitReached: func(_ fiber.Ctx) error { return fiber.ErrTooManyRequests }}), s.login)
	secured := app.Group("/api/v1", s.requireUser)
	secured.Get("/auth/me", func(c fiber.Ctx) error { return c.JSON(c.Locals("user")) })
	secured.Post("/auth/logout", s.logout)
	secured.Get("/nodes", func(c fiber.Ctx) error {
		result, err := s.nodes.List(c.Context())
		if err != nil {
			return err
		}
		return c.JSON(result)
	})
	secured.Get("/nodes/:id", func(c fiber.Ctx) error {
		result, err := s.nodes.Get(c.Context(), c.Params("id"))
		if err != nil {
			return err
		}
		return c.JSON(result)
	})
	secured.Post("/nodes", s.createNode)
	registerAccounts(secured, accounts)
	app.Use(func(_ fiber.Ctx) error { return fiber.ErrNotFound })
	return app
}
func errorHandler(c fiber.Ctx, err error) error {
	status, code, message := 500, "INTERNAL_ERROR", "An internal error occurred."
	switch {
	case errors.Is(err, domain.ErrInvalid):
		status, code, message = 400, "VALIDATION_ERROR", "The request is invalid."
		var validation domain.ValidationError
		if errors.As(err, &validation) {
			message = validation.Message
		}
	case errors.Is(err, domain.ErrUnauthorized):
		status, code, message = 401, "UNAUTHORIZED", "Authentication is required."
	case errors.Is(err, domain.ErrForbidden):
		status, code, message = 403, "FORBIDDEN", "Permission denied."
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = 404, "NOT_FOUND", "The record was not found."
	case errors.Is(err, domain.ErrLastAdmin):
		status, code, message = 409, "LAST_ADMIN", "The last active administrator cannot be disabled or demoted."
	case errors.Is(err, domain.ErrConflict):
		status, code, message = 409, "CONFLICT", "A record with that name already exists."
	default:
		var e *fiber.Error
		if errors.As(err, &e) {
			status = e.Code
			code = "HTTP_ERROR"
			message = "The request could not be completed."
		}
	}
	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": message, "details": fiber.Map{}}, "request_id": c.GetRespHeader("X-Request-ID")})
}
func (s Server) cookie(c fiber.Ctx, value string, expired bool) {
	age := 43200
	expiry := time.Now().Add(12 * time.Hour)
	if expired {
		age = -1
		expiry = time.Unix(1, 0)
	}
	c.Cookie(&fiber.Cookie{Name: cookieName, Value: value, Path: "/", HTTPOnly: true, Secure: s.cfg.SecureCookie, SameSite: "Strict", MaxAge: age, Expires: expiry})
}
func (s Server) login(c fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if json.Unmarshal(c.Body(), &input) != nil {
		return domain.ErrInvalid
	}
	user, token, err := s.auth.Login(c.Context(), input.Email, input.Password)
	if err != nil {
		return err
	}
	s.cookie(c, token, false)
	return c.JSON(user)
}
func (s Server) requireUser(c fiber.Ctx) error {
	user, err := s.auth.Current(c.Context(), c.Cookies(cookieName))
	if err != nil {
		return err
	}
	c.Locals("user", user)
	return c.Next()
}
func (s Server) logout(c fiber.Ctx) error {
	if err := s.auth.Logout(c.Context(), c.Cookies(cookieName)); err != nil {
		return err
	}
	s.cookie(c, "", true)
	return c.SendStatus(204)
}
func (s Server) createNode(c fiber.Ctx) error {
	var input domain.NodeInput
	if json.Unmarshal(c.Body(), &input) != nil {
		return domain.ErrInvalid
	}
	result, err := s.nodes.Create(c.Context(), c.Locals("user").(domain.User), input)
	if err != nil {
		return err
	}
	return c.Status(201).JSON(result)
}
