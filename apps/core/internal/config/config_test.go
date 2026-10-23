package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	base := map[string]string{"DATABASE_URL": "postgres://localhost/velora", "PUBLIC_URL": "http://localhost:5173", "APP_ENV": "development"}
	tests := []struct {
		name, key, value string
		valid            bool
	}{
		{"defaults", "", "", true}, {"bad database", "DATABASE_URL", "https://localhost", false}, {"malformed database", "DATABASE_URL", "://", false}, {"bad origin", "PUBLIC_URL", "http://localhost/path", false}, {"malformed origin", "PUBLIC_URL", "://", false}, {"production", "APP_ENV", "production", false}, {"bad bool", "COOKIE_SECURE", "nope", false}, {"valid bool", "COOKIE_SECURE", "true", true}, {"bad port", "HTTP_PORT", "zero", false}, {"port low", "HTTP_PORT", "0", false}, {"port high", "HTTP_PORT", "65536", false}, {"port", "HTTP_PORT", "9000", true}, {"host", "HTTP_HOST", "0.0.0.0", true}, {"log", "LOG_LEVEL", "DEBUG", true}, {"bad log", "LOG_LEVEL", "VERBOSE", false}, {"redis missing", "REDIS_ENABLED", "true", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{}
			for k, v := range base {
				env[k] = v
			}
			env[tt.key] = tt.value
			c, e := Load(func(k string) string { return env[k] })
			if tt.valid {
				require.NoError(t, e)
				require.NotEmpty(t, c.Address)
			} else {
				require.Error(t, e)
			}
		})
	}
	base["REDIS_ENABLED"] = "true"
	base["REDIS_URL"] = "redis://localhost:6379"
	base["APP_ENV"] = "production"
	base["PUBLIC_URL"] = "https://velora.example"
	base["COOKIE_SECURE"] = "true"
	c, e := Load(func(k string) string { return base[k] })
	require.NoError(t, e)
	require.True(t, c.RedisEnabled)
}

func TestDatabaseURL(t *testing.T) {
	values := map[string]string{"DATABASE_URL": "postgres://runtime@localhost/db", "MIGRATION_DATABASE_URL": "postgres://owner@localhost/db"}
	get := func(k string) string { return values[k] }
	v, e := DatabaseURL(get, false)
	require.NoError(t, e)
	require.Equal(t, values["DATABASE_URL"], v)
	v, e = DatabaseURL(get, true)
	require.NoError(t, e)
	require.Equal(t, values["MIGRATION_DATABASE_URL"], v)
	delete(values, "MIGRATION_DATABASE_URL")
	v, e = DatabaseURL(get, true)
	require.NoError(t, e)
	require.Equal(t, values["DATABASE_URL"], v)
	for _, bad := range []string{"%", "postgres:///db", "https://localhost/db"} {
		values["DATABASE_URL"] = bad
		_, e = DatabaseURL(get, false)
		require.Error(t, e)
	}
}
