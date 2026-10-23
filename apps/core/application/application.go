// Package application is Core's supported integration surface for the unified CLI.
// It aliases shared services rather than duplicating domain or persistence behavior.
package application

import (
	"github.com/velora/panel/internal/app"
	"github.com/velora/panel/internal/domain"
	"github.com/velora/panel/internal/modules/auth"
	"github.com/velora/panel/internal/platform/redact"
)

type Engine = app.Engine
type User = domain.User
type AccountInput = domain.AccountInput

var Open = app.Open
var Migrate = app.Migrate
var Redact = redact.Error
var ValidateAccount = auth.ValidateAccount
var ValidatePassword = auth.ValidatePassword
