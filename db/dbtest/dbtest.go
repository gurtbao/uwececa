package dbtest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/db"
)

func GetTestDB(t testing.TB) *db.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "db.sqlite3")

	d, err := db.New(dbPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := d.Close()
		t.Logf("error closing db: %v", err)
	})

	return d
}
