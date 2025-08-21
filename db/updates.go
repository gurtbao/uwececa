package db

import (
	"fmt"
	"strings"
)

func BuildUpdate(updates []UpdateData) (string, []any) {
	keys := make([]string, len(updates))
	values := make([]any, len(updates))

	for i := range updates {
		keys[i] = updates[i].Sql()
		values[i] = updates[i].Data()
	}

	var clause string
	if len(keys) != 0 {
		clause = " set " + strings.Join(keys, ", ")
	}

	return clause, values
}

func Updates(updates ...UpdateData) []UpdateData {
	return updates
}

type UpdateData struct {
	key  string
	data any
}

func (u UpdateData) Sql() string {
	return fmt.Sprintf("%s = ?", u.key)
}

func (u UpdateData) Data() any {
	return u.data
}

func Update(key string, data any) UpdateData {
	return UpdateData{
		key:  key,
		data: data,
	}
}
