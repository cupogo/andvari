package oid

import (
	"strings"
)

type StringSlice []string

func (ss StringSlice) String() string {
	return strings.Join(ss, ",")
}

func (ss StringSlice) Decode() (a OIDs, err error) {
	for _, s := range ss {
		var id OID
		if _, id, err = Parse(s); err != nil {
			return
		}
		a = append(a, id)
	}

	return
}

// 以逗号分隔的 ids
type OIDsStr string

func (s OIDsStr) Slice() StringSlice {
	return strings.Split(string(s), ",")
}

func (s OIDsStr) Decode() (a OIDs, err error) {
	return s.Slice().Decode()
}

func (s OIDsStr) Valid() bool {
	return len(s) > 0
}

func (s OIDsStr) String() string {
	return string(s)
}
