package comm

import (
	"bytes"
	"errors"
	"fmt"
)

type Gender uint8

const (
	GnUnknown Gender = iota // 0
	Male                    // 1
	Female                  // 2
	GnOther                 // 3
)

var ErrEmptyGender = errors.New("empty gender")
var genderKeys = []string{"unknown", "male", "famale", "other"}

// fmt.Stringer
func (this Gender) String() string {
	if this >= GnUnknown && this <= GnOther {
		return genderKeys[this]
	}
	return "unknown"
}

func ParseGender(s string) (g Gender, err error) {
	if len(s) == 0 {
		return GnUnknown, ErrEmptyGender
	}
	r := bytes.Runes(bytes.TrimLeft([]byte(s), "\""))
	if len(r) == 0 {
		return GnUnknown, ErrEmptyGender
	}
	switch c := r[0]; c {
	case 'm', 'M', '1', '男':
		g = Male
	case 'f', 'F', '2', '女':
		g = Female
	case 'o', 'O', '3':
		g = GnOther
	case 'u', 'U', '0':
		g = GnUnknown
	default:
		g = GnUnknown
		err = fmt.Errorf("invalid gender '%s'", s)
	}
	return
}
