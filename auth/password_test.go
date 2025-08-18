package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/auth"
)

func TestPasswordHashing(t *testing.T) {
	t.Parallel()

	password := "hello_123453124"

	hash := auth.HashPassword(password)

	require.True(t, auth.MustVerifyPassword(password, hash))
}
