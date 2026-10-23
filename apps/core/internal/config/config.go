package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type Config struct {
	Address, DatabaseURL, PublicURL, RedisURL, LogLevel string
	SecureCookie, RedisEnabled                          bool
}

func Load(get func(string) string) (Config, error) {
	c := Config{DatabaseURL: get("DATABASE_URL"), PublicURL: get("PUBLIC_URL"), RedisURL: get("REDIS_URL"), LogLevel: get("LOG_LEVEL")}
	if c.LogLevel == "" {
		c.LogLevel = "INFO"
	}
	if c.LogLevel != "INFO" && c.LogLevel != "DEBUG" && c.LogLevel != "WARN" && c.LogLevel != "ERROR" {
		return c, fmt.Errorf("LOG_LEVEL must be DEBUG, INFO, WARN or ERROR")
	}
	u, err := url.Parse(c.DatabaseURL)
	if err != nil || u.Host == "" || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return c, fmt.Errorf("DATABASE_URL must be a PostgreSQL URL")
	}
	u, err = url.Parse(c.PublicURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return c, fmt.Errorf("PUBLIC_URL must be an HTTP origin without a path")
	}
	for name, target := range map[string]*bool{"COOKIE_SECURE": &c.SecureCookie, "REDIS_ENABLED": &c.RedisEnabled} {
		if value := get(name); value != "" {
			b, e := strconv.ParseBool(value)
			if e != nil {
				return c, fmt.Errorf("%s must be boolean", name)
			}
			*target = b
		}
	}
	if get("APP_ENV") != "development" && (!c.SecureCookie || u.Scheme != "https") {
		return c, fmt.Errorf("production requires HTTPS PUBLIC_URL and COOKIE_SECURE=true")
	}
	if c.RedisEnabled {
		r, e := url.Parse(c.RedisURL)
		if e != nil || r.Host == "" || (r.Scheme != "redis" && r.Scheme != "rediss") {
			return c, fmt.Errorf("REDIS_URL required when Redis is enabled")
		}
	}
	host, port := get("HTTP_HOST"), get("HTTP_PORT")
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "8080"
	}
	n, e := strconv.Atoi(port)
	if e != nil || n < 1 || n > 65535 {
		return c, fmt.Errorf("HTTP_PORT must be 1..65535")
	}
	c.Address = net.JoinHostPort(host, port)
	return c, nil
}

// DatabaseURL loads only the configuration needed by operational commands.
func DatabaseURL(get func(string) string, migration bool) (string, error) {
	value := get("DATABASE_URL")
	if migration && get("MIGRATION_DATABASE_URL") != "" {
		value = get("MIGRATION_DATABASE_URL")
	}
	u, e := url.Parse(value)
	if e != nil || u.Host == "" || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return "", fmt.Errorf("Set a valid PostgreSQL DATABASE_URL (or MIGRATION_DATABASE_URL for migrations)")
	}
	return value, nil
}
