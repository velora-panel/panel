package commands

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/velora/panel/application"
)

const Help = `Velora — Console and Core operations

velora serve
velora migrate up|down|status|version  (down requires --allow-destructive)
velora admin create|list|reset-password
velora user create|list|disable|enable
velora doctor
velora version
velora help

Account options: --email ADDRESS --role admin|user --password-stdin
Non-interactive secrets: supply JSON on stdin with email, password and password_confirmation.
Examples: pnpm velora migrate up; pnpm velora admin create
`

type Backend interface {
	Close()
	CheckSchema(context.Context) error
	Serve(context.Context, io.Writer) error
	Create(context.Context, application.AccountInput) error
	List(context.Context, bool) ([]application.User, error)
	SetDisabled(context.Context, string, bool) error
	Reset(context.Context, string, string) error
}
type Runner struct {
	Input    Input
	Out, Err io.Writer
	Open     func(context.Context) (Backend, error)
	Migrate  func(context.Context, string, io.Writer) error
}

func (r Runner) Run(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(r.Out, Help)
		return nil
	}
	if args[0] == "version" {
		fmt.Fprintln(r.Out, "velora 0.2.0")
		return nil
	}
	command := args[0]
	action := ""
	rest := args[1:]
	if command == "migrate" || command == "admin" || command == "user" {
		if len(rest) == 0 {
			return fmt.Errorf("%s requires a subcommand; run velora help", command)
		}
		action = rest[0]
		rest = rest[1:]
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(r.Err)
	email := flags.String("email", "", "account email")
	role := flags.String("role", "user", "account role")
	stdin := flags.Bool("password-stdin", false, "read credential JSON from stdin")
	allow := flags.Bool("allow-destructive", false, "permit migration rollback")
	if e := flags.Parse(rest); e != nil {
		return e
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments")
	}
	switch command {
	case "migrate":
		if action != "up" && action != "down" && action != "status" && action != "version" {
			return fmt.Errorf("unknown migration action")
		}
		if action == "down" && !*allow {
			return fmt.Errorf("rollback may destroy data; back up first and pass --allow-destructive")
		}
		return r.Migrate(ctx, action, r.Out)
	case "serve", "doctor":
	case "admin":
		if action != "create" && action != "list" && action != "reset-password" {
			return fmt.Errorf("unknown admin command")
		}
		*role = "admin"
	case "user":
		if action != "create" && action != "list" && action != "disable" && action != "enable" {
			return fmt.Errorf("unknown user command")
		}
	default:
		return fmt.Errorf("unknown command; run velora help")
	}
	backend, e := r.Open(ctx)
	if e != nil {
		return e
	}
	defer backend.Close()
	if e = backend.CheckSchema(ctx); e != nil {
		return e
	}
	switch command {
	case "serve":
		return backend.Serve(ctx, r.Out)
	case "doctor":
		fmt.Fprintln(r.Out, "Database reachable. Schema initialized. Use migrate status to inspect migration history.")
		return nil
	}
	switch action {
	case "list":
		users, e := backend.List(ctx, command == "admin")
		if e != nil {
			return e
		}
		for _, u := range users {
			fmt.Fprintf(r.Out, "%s\t%s\tdisabled=%t\n", u.Email, u.Role, u.Disabled)
		}
		return nil
	case "disable", "enable":
		if *email == "" {
			return fmt.Errorf("--email is required")
		}
		if e = backend.SetDisabled(ctx, *email, action == "disable"); e != nil {
			return e
		}
	case "create", "reset-password":
		c, e := r.Input.Read(*email, *stdin)
		if e != nil {
			return e
		}
		if e = application.ValidateAccount(c.Email, c.Password, *role); e != nil {
			return e
		}
		if action == "create" {
			e = backend.Create(ctx, application.AccountInput{Email: c.Email, Password: c.Password, Role: *role})
		} else {
			e = backend.Reset(ctx, c.Email, c.Password)
		}
		if e != nil {
			return e
		}
	}
	fmt.Fprintln(r.Out, "Account updated successfully.")
	return nil
}
