package utils

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
