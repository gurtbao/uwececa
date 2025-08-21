package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/db"
)

func TestUpdatesWork(t *testing.T) {
	t.Parallel()

	updates := []db.UpdateData{db.Update("verified_at", 1), db.Update("hi", 1)}

	keys, values := db.BuildUpdate(updates)

	require.Equal(t, " set verified_at = ?, hi = ?", keys)
	require.Equal(t, values, []any{1, 1})
}
