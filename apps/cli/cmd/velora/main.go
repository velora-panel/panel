package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/velora/panel/application"
	"github.com/velora/panel/cli/internal/commands"
	"golang.org/x/term"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	runner := commands.Runner{Out: os.Stdout, Err: os.Stderr, Input: commands.Input{In: os.Stdin, Out: os.Stderr, Terminal: term.IsTerminal(int(os.Stdin.Fd())), Password: func() (string, error) { b, e := term.ReadPassword(int(os.Stdin.Fd())); return string(b), e }}, Open: func(ctx context.Context) (commands.Backend, error) {
		engine, e := application.Open(ctx, os.Getenv)
		if e != nil {
			return nil, e
		}
		return commands.Runtime{Lifecycle: engine, Users: engine.Users, Repo: engine.Repo}, nil
	}, Migrate: func(ctx context.Context, action string, out io.Writer) error {
		return application.Migrate(ctx, os.Getenv, action, out)
	}}
	if e := runner.Run(ctx, os.Args[1:]); e != nil {
		fmt.Fprintln(os.Stderr, application.Redact(e, os.Getenv("DATABASE_URL"), os.Getenv("MIGRATION_DATABASE_URL"), os.Getenv("REDIS_URL")))
		os.Exit(1)
	}
}
