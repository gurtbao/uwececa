package db

import (
	"fmt"
	"reflect"
	"strings"
)

type Filter struct {
	key string
	arg any
	cmp string
}

func newFilter(key, cmp string, arg any) Filter {
	return Filter{
		key: key,
		arg: arg,
		cmp: cmp,
	}
}

func FilterEq(key string, arg any) Filter    { return newFilter(key, "=", arg) }
func FilterNotEq(key string, arg any) Filter { return newFilter(key, "<>", arg) }
func FilterGte(key string, arg any) Filter   { return newFilter(key, ">=", arg) }
func FilterLte(key string, arg any) Filter   { return newFilter(key, "<=", arg) }
func FilterIs(key string, arg any) Filter    { return newFilter(key, "is", arg) }
func FilterIsNot(key string, arg any) Filter { return newFilter(key, "is not", arg) }
func FilterIn(key string, arg any) Filter    { return newFilter(key, "in", arg) }

func (f Filter) Condition() string {
	rv := reflect.ValueOf(f.arg)
	kind := rv.Kind()

	if (kind == reflect.Slice && rv.Type().Elem().Kind() != reflect.Uint8) || kind == reflect.Array {
		if rv.Len() == 0 {
			return "1 = 0"
		}

		placeholders := make([]string, rv.Len())
		for i := range placeholders {
			placeholders[i] = "?"
		}

		return fmt.Sprintf("%s %s (%s)", f.key, f.cmp, strings.Join(placeholders, ", "))
	}

	return fmt.Sprintf("%s %s ?", f.key, f.cmp)
}

func (f Filter) Arg() []any {
	rv := reflect.ValueOf(f.arg)
	kind := rv.Kind()
	if (kind == reflect.Slice && rv.Type().Elem().Kind() != reflect.Uint8) || kind == reflect.Array {
		if rv.Len() == 0 {
			return nil
		}

		out := make([]any, rv.Len())
		for i := range rv.Len() {
			out[i] = rv.Index(i).Interface()
		}
		return out
	}

	return []any{f.arg}
}
