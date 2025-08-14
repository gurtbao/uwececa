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
