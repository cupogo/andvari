package utils

import (
	"strconv"
	"strings"
)

func EnsureArgs(n int, args ...any) bool {
	if len(args) < n {
		return false
	}
	for _, v := range args {
		if v == 0 || v == "" {
			return false
		}
	}
	return true
}

func ParseInts(s string) ([]int, bool) {
	if len(s) == 0 {
		return nil, false
	}
	out, err := StringsToInts(SliceRidZero(strings.Split(s, ",")))
	return out, err == nil && len(out) > 0
}

func ParseStrs(s string) ([]string, bool) {
	out := strings.Split(s, ",")
	out = SliceRidZero(out)
	return out, len(out) > 0
}

// StringsToInts convert []string to []int
func StringsToInts(sa []string) ([]int, error) {
	si := make([]int, len(sa))
	for i := 0; i < len(sa); i++ {
		v, err := strconv.Atoi(sa[i])
		if err != nil {
			return si, err
		}
		si[i] = v
	}
	return si, nil
}

func SliceRidZero[T comparable](in []T) []T {
	var zero T
	out := make([]T, 0, len(in))
	for _, v := range in {
		if v != zero {
			out = append(out, v)
		}
	}

	return out
}

func ParseBool(s string) (r bool, ok bool) {
	switch s {
	case "1", "t", "T", "true", "TRUE", "True", "y", "Y", "yes", "Yes", "YES":
		return true, true
	case "0", "f", "F", "false", "FALSE", "False", "n", "N", "no", "No", "NO":
		return false, true
	}
	return
}
