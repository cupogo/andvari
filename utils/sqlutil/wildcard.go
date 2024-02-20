package sqlutil

import (
	"os"
	"strings"

	"github.com/cupogo/andvari/utils"
)

var (
	cleaningReplacer = strings.NewReplacer("%", "", "--", "", ";", "", "'", "", "\"", "")
	wildcardReplacer = strings.NewReplacer("*", "%", "?", "_")

	allowLeftWildcard bool
)

func init() {
	if s, ok := os.LookupEnv("DB_ALLOW_LEFT_WILDCARD"); ok && len(s) > 0 {
		allowLeftWildcard, _ = utils.ParseBool(s)
	}
}

// CleanWildcard 清除字串中的无效SQL字符，并去除开头的通配符
func CleanWildcard(s string, opt ...bool) string {
	s = cleaningReplacer.Replace(s)
	fulike := len(opt) > 0 && opt[0]
	if !allowLeftWildcard && !fulike {
		s = strings.TrimLeftFunc(s, func(c rune) bool {
			return c == '*' || c == '_' || c == '?'
		})
	}

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
func MendValue(v string, opt ...bool) (cv string) {
	cv = CleanWildcard(v, opt...)
	if !strings.HasSuffix(cv, "%") && !strings.HasSuffix(cv, "_") {
		cv = cv + "%"
	}
	if allowLeftWildcard || len(opt) > 0 && opt[0] {
		if !strings.HasPrefix(cv, "%") && !strings.HasPrefix(cv, "_") {
			cv = "%" + cv
		}
	}

	return
}
