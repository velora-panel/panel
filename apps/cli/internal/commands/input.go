package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/velora/panel/application"
)

type Credentials struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Confirmation string `json:"password_confirmation"`
}
type Input struct {
	In       io.Reader
	Out      io.Writer
	Terminal bool
	Password func() (string, error)
}

// line never buffers ahead into the password prompt.
func line(in io.Reader) (string, error) {
	var result strings.Builder
	var b [1]byte
	for result.Len() < 512 {
		_, e := io.ReadFull(in, b[:])
		if e != nil {
			return "", fmt.Errorf("input ended before a complete line: %w", e)
		}
		if b[0] == '\n' {
			return strings.TrimSuffix(result.String(), "\r"), nil
		}
		result.WriteByte(b[0])
	}
	return "", fmt.Errorf("input line is too long")
}
func (i Input) Read(email string, stdin bool) (Credentials, error) {
	var c Credentials
	if stdin {
		d := json.NewDecoder(io.LimitReader(i.In, 4096))
		d.DisallowUnknownFields()
		if e := d.Decode(&c); e != nil {
			return c, fmt.Errorf("provide JSON with email, password and password_confirmation on stdin")
		}
		var extra interface{}
		if d.Decode(&extra) != io.EOF {
			return c, fmt.Errorf("stdin must contain exactly one JSON object")
		}
	} else {
		if !i.Terminal {
			return c, fmt.Errorf("interactive terminal required; use --password-stdin with a JSON object")
		}
		c.Email = email
		var e error
		if c.Email == "" {
			fmt.Fprint(i.Out, "Email: ")
			c.Email, e = line(i.In)
			if e != nil {
				return c, e
			}
		}
		fmt.Fprint(i.Out, "Password: ")
		c.Password, e = i.Password()
		fmt.Fprintln(i.Out)
		if e != nil {
			return c, fmt.Errorf("cannot read terminal password: %w", e)
		}
		fmt.Fprint(i.Out, "Confirm password: ")
		c.Confirmation, e = i.Password()
		fmt.Fprintln(i.Out)
		if e != nil {
			return c, fmt.Errorf("cannot read password confirmation: %w", e)
		}
	}
	if email != "" {
		c.Email = email
	}
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))
	if c.Password != c.Confirmation {
		return c, fmt.Errorf("Passwords do not match.")
	}
	if e := application.ValidatePassword(c.Password); e != nil {
		return c, e
	}
	return c, nil
}
