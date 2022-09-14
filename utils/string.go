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
	out, err := StringsToInts(strings.Split(s, ","))
	return out, err == nil
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
