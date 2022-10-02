package sqlutil

import (
	"strings"
)

var (
	cleaningReplacer = strings.NewReplacer("%", "", "--", "", ";", "", "'", "", "\"", "")
	wildcardReplacer = strings.NewReplacer("*", "%", "?", "_")
)

// CleanWildcard 清除字串中的无效SQL字符，并去除开头的通配符
func CleanWildcard(s string) string {
	s = cleaningReplacer.Replace(s)
	s = strings.TrimLeftFunc(s, func(c rune) bool {
		return c == '*' || c == '_' || c == '?'
	})
	s = wildcardReplacer.Replace(s)

	return s
}

// StartsWith 判断字串集中是否以k开头
func StartsWith(k string, strs []string) bool {
	for _, str := range strs {
		if strings.HasPrefix(k, str) {
			return true
		}
	}
	return false
}

// ClearKV 同时替换 k,v 两个值，用于 SQL 查询
func ClearKV(k, v string) (ck string, cv string) {
	if !strings.Contains(k, "__") {
		ck = k + "__ieq"
	} else {
		ck = k
	}
	cv = MendValue(v)
	return
}

// MendValue 针对SQL查询字符值进行修补
func MendValue(v string) (cv string) {
	cv = CleanWildcard(v)
	if !strings.HasSuffix(v, "%") && !strings.HasSuffix(v, "_") {
		cv = cv + "%"
	}
	return
}
