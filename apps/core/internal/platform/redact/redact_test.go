package redact

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	require.Empty(t, Error(nil))
	s := Error(errors.New(`permission denied postgres://user:pass@localhost/db password=abc token="opaque value" cookie='sid=value' authorization=Bearer redis://:secret@localhost:6379`), "", "postgres://user:pass@localhost/db", "%bad", "plain")
	for _, v := range []string{"pass@", "abc", "opaque value", "sid=value", "Bearer", "secret@"} {
		require.NotContains(t, s, v)
	}
	require.Contains(t, s, "permission denied")
	require.Equal(t, "[REDACTED]", Error(errors.New("pass"), "postgres://user:pass@localhost/db"))
	require.Equal(t, "diagnostic", Error(errors.New("diagnostic"), "postgres://user@localhost/db"))
	require.NotContains(t, Error(errors.New("authorization: Bearer opaque-token\ncookie: sid=first; other=second")), "opaque-token")
	require.NotContains(t, Error(errors.New("cookie: sid=first; other=second")), "second")
}
