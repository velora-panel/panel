package redact

import (
	"net/url"
	"regexp"
	"strings"
)

var urls = regexp.MustCompile(`(?i)(?:postgres(?:ql)?|rediss?)://[^\s]+`)
var fields = regexp.MustCompile(`(?i)(password|token|secret|cookie|authorization)\s*[=:]\s*(?:"[^"]*"|'[^']*'|[^\s;,]+)`)

// Error retains operational diagnostics while removing configured and labeled secrets.
func Error(err error, secrets ...string) string {
	if err == nil {
		return ""
	}
	message := fields.ReplaceAllString(err.Error(), "$1=[REDACTED]")
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		message = strings.ReplaceAll(message, secret, "[REDACTED]")
		if u, e := url.Parse(secret); e == nil && u.User != nil {
			if p, ok := u.User.Password(); ok && p != "" {
				message = strings.ReplaceAll(message, p, "[REDACTED]")
			}
		}
	}
	message = urls.ReplaceAllString(message, "[REDACTED URL]")
	return message
}
