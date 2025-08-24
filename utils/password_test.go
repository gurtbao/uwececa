package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/utils"
)

func TestPasswordHashing(t *testing.T) {
	t.Parallel()

	password := "hello_123453124"

	hash := utils.HashPassword(password)

	require.True(t, utils.MustVerifyPassword(password, hash))
}
