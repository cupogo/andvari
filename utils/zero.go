package utils

import (
	"reflect"
)

func IsNil(v any) bool {
	if v == nil {
		return true
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(v).IsNil()
	}
	return false
}

func IsZero(v any) bool {
	if i, ok := v.(interface{ IsZero() bool }); ok {
		return i.IsZero()
	}
	return IsNil(v) || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}
