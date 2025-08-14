package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"uwece.ca/app/db"
)

func TestFilterConditionWorksForSlice(t *testing.T) {
	t.Parallel()

	filter := db.FilterIn("id", []string{"1234", "12"})

	require.Equal(t, "id in (?, ?)", filter.Condition())
}

func TestFilterConditionRespectsUint8Slice(t *testing.T) {
	t.Parallel()

	filter := db.FilterEq("id", []byte("hello"))

	require.Equal(t, "id = ?", filter.Condition())
}

func TestFilterConditionAlwaysFalseOnEmptyArgs(t *testing.T) {
	t.Parallel()

	filter := db.FilterEq("id", []string{})

	require.Equal(t, "1 = 0", filter.Condition())
}

func TestFilterConditionCorrectForEQ(t *testing.T) {
	t.Parallel()

	filter := db.FilterEq("id", 12345)

	require.Equal(t, "id = ?", filter.Condition())
}

func TestBuildWhere(t *testing.T) {
	t.Parallel()

	where, args := db.BuildWhere([]db.Filter{
		db.FilterEq("id", 1),
		db.FilterIn("id", []int{2, 3, 4}),
	})

	require.Equal(t, " where id = ? and id in (?, ?, ?)", where)
	require.Equal(t, []any{1, 2, 3, 4}, args)
}

func TestBuildWhereReturnsEmptyWhenFiltersEmpty(t *testing.T) {
	t.Parallel()

	where, args := db.BuildWhere([]db.Filter{})

	require.Equal(t, "", where)
	require.Empty(t, args)
}
