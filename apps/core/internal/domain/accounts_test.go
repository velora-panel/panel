package domain
import("testing";"github.com/stretchr/testify/require")
func TestAccountPolicy(t *testing.T){err:=Invalid("Invalid email.");require.Equal(t,"Invalid email.",err.Error());require.ErrorIs(t,err,ErrInvalid);require.NoError(t,RequireAdmin(User{Role:"admin"}));require.ErrorIs(t,RequireAdmin(User{Role:"user"}),ErrForbidden);require.ErrorIs(t,RequireAdmin(User{Role:"admin",Disabled:true}),ErrForbidden)}
