// Package app wires infrastructure. Business rules remain in the tested services.
package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/velora/panel/internal/config"
	"github.com/velora/panel/internal/database"
	"github.com/velora/panel/internal/database/generated"
	"github.com/velora/panel/internal/domain"
	"github.com/velora/panel/internal/modules/auth"
	"github.com/velora/panel/internal/modules/nodes"
	"github.com/velora/panel/internal/platform/redact"
	httptransport "github.com/velora/panel/internal/transport/http"
	"github.com/velora/panel/migrations"
)

type Engine struct {
	Pool  *pgxpool.Pool
	Repo  *database.Repository
	Users *auth.Accounts
	get   func(string) string
}

func Open(ctx context.Context, get func(string) string) (*Engine, error) {
	url, e := config.DatabaseURL(get, false)
	if e != nil {
		return nil, e
	}
	cfg, e := pgxpool.ParseConfig(url)
	if e != nil {
		return nil, fmt.Errorf("database configuration: %s", redact.Error(e, url))
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		return nil, fmt.Errorf("database pool: %s", redact.Error(e, url))
	}
	bounded, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if e = pool.Ping(bounded); e != nil {
		pool.Close()
		return nil, fmt.Errorf("database connection: %s", redact.Error(e, url))
	}
	repo := database.New(generated.New(pool), database.Transactions(pool))
	return &Engine{Pool: pool, Repo: repo, Users: auth.NewAccounts(repo), get: get}, nil
}
func (e *Engine) Close() { e.Pool.Close() }
func (e *Engine) CheckSchema(ctx context.Context) error {
	ok, err := generated.New(e.Pool).SchemaInitialized(ctx)
	if err != nil {
		return fmt.Errorf("schema check: %s", redact.Error(err, e.get("DATABASE_URL")))
	}
	if !ok.Bool {
		return domain.ErrSchema
	}
	return nil
}
func (e *Engine) Serve(ctx context.Context, out io.Writer) error {
	cfg, err := config.Load(e.get)
	if err != nil {
		return err
	}
	if err = e.CheckSchema(ctx); err != nil {
		return err
	}
	var level slog.Level
	if err = level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return err
	}
	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level}))
	if cfg.RedisEnabled {
		options, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			return fmt.Errorf("redis configuration: %s", redact.Error(err, cfg.RedisURL))
		}
		client := redis.NewClient(options)
		defer func() { _ = client.Close() }()
		bounded, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = client.Ping(bounded).Err()
		cancel()
		if err != nil {
			logger.Warn("optional Redis unavailable")
		} else {
			logger.Info("optional Redis connected")
		}
	}
	server := httptransport.New(cfg, auth.New(e.Repo), nodes.New(e.Repo), e.Pool.Ping, logger, e.Users)
	failures := make(chan error, 1)
	go func() { failures <- server.Listen(cfg.Address, fiber.ListenConfig{DisableStartupMessage: true}) }()
	logger.Info("Core started", "address", cfg.Address)
	select {
	case err = <-failures:
		return err
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("Core shutting down")
	return server.ShutdownWithContext(shutdown)
}
func Migrate(ctx context.Context, get func(string) string, action string, out io.Writer) error {
	url, e := config.DatabaseURL(get, true)
	if e != nil {
		return e
	}
	cfg, e := pgx.ParseConfig(url)
	if e != nil {
		return fmt.Errorf("migration configuration: %s", redact.Error(e, url))
	}
	cfg.ConnectTimeout = 5 * time.Second
	db := stdlib.OpenDB(*cfg)
	defer func() { _ = db.Close() }()
	provider, e := goose.NewProvider(goose.DialectPostgres, db, migrations.Files)
	if e != nil {
		return e
	}
	bounded, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	switch action {
	case "up":
		_, e = provider.Up(bounded)
	case "down":
		_, e = provider.Down(bounded)
	case "version":
		var version int64
		version, e = provider.GetDBVersion(bounded)
		if e == nil {
			_, _ = fmt.Fprintln(out, version)
		}
	case "status":
		var status []*goose.MigrationStatus
		status, e = provider.Status(bounded)
		if e == nil {
			for _, item := range status {
				_, _ = fmt.Fprintf(out, "%d %s\n", item.Source.Version, item.State)
			}
		}
	default:
		return fmt.Errorf("unknown migration action: use up, down, status or version")
	}
	if e != nil {
		return fmt.Errorf("migration %s failed: %s", action, redact.Error(e, url))
	}
	return nil
}
