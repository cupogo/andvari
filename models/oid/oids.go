package oid

import (
	"slices"
	"strings"
)

type OIDs []OID

func (z OIDs) String() string {
	a := make(StringSlice, len(z))
	for i := range z {
		a[i] = z[i].String()
	}
	return a.String()
}

func (z OIDs) Has(id OID) bool {
	return slices.Contains(z, id)
}

func (z OIDs) Remove(id OID) (r OIDs) {
	for _, v := range z {
		if v != id {
			r = append(r, v)
		}
	}
	return
}

func (z OIDs) ToJSON() string {
	var b strings.Builder
	b.WriteByte('[')
	for i, s := range z {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('"')
		b.WriteString(s.String())
		b.WriteByte('"')
	}
	b.WriteByte(']')
	return b.String()
}

type StringSlice []string

func (ss StringSlice) String() string {
	return strings.Join(ss, ",")
}

func (ss StringSlice) Decode() (OIDs, error) {
	if len(ss) == 0 {
		return nil, ErrEmptyOID
	}
	a := make(OIDs, len(ss))
	for i := range ss {
		if _, id, err := Parse(ss[i]); err != nil {
			return nil, err
		} else {
			a[i] = id
		}

	}

	return a, nil
}

// 以逗号分隔的 ids
type OIDsStr string

func (s OIDsStr) Slice() StringSlice {
	return strings.Split(string(s), ",")
}

func (s OIDsStr) Decode() (OIDs, error) {
	return s.Slice().Decode()
}

func (s OIDsStr) Valid() bool {
	return len(s) > 0
}

func (s OIDsStr) String() string {
	return string(s)
}

// Vals for gencode
func (s OIDsStr) Vals() OIDs {
	ids, err := s.Slice().Decode()
	if err != nil {
		return nil
	}
	return ids
}

func ParseOIDs(s string) (OIDs, bool) {
	if len(s) == 0 {
		return nil, false
	}
	ids, err := OIDsStr(s).Decode()
	if err != nil {
		return nil, false
	}
	return ids, true
}
