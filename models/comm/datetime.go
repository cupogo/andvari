package comm

import (
	"encoding/json"
	"time"
)

type ITime interface {
	Time() time.Time
}

func AsTime(ts any) (time.Time, bool) {
	switch v := ts.(type) {
	case time.Time:
		return v, true
	case ITime:
		return v.Time(), true
	case int64:
		return time.UnixMilli(v), true
	default:
		return time.Time{}, false
	}
}

type TimeValue interface {
	IsZero() bool
	Format(layout string) string
}

func AsDateTime(tv any) (DateTime, bool) {
	switch t := tv.(type) {
	case time.Time:
		return NewDateTimeFromTime(t), true
	case *time.Time:
		return NewDateTimeFromTime(*t), true
	case DateTime:
		return t, true
	case *DateTime:
		return *t, true
	case ITime:
		return NewDateTimeFromTime(t.Time()), true
	case int64:
		return DateTime(t), true
	case string:
		if vt, err := time.Parse(time.RFC3339, t); err == nil {
			return NewDateTimeFromTime(vt), true
		}
		if vt, err := time.Parse(time.DateTime, t); err == nil {
			return NewDateTimeFromTime(vt), true
		}
		return 0, false
	default:
		return 0, false
	}
}

// copy and optimize from go.mongodb.org/mongo-driver/bson/primitive

// DateTime represents the BSON datetime value.
type DateTime int64

var _ json.Marshaler = DateTime(0)
var _ json.Unmarshaler = (*DateTime)(nil)
func (d DateTime) String() string {
	if d == 0 {
		return ""
	}
	return d.Time().String()
}

func (d DateTime) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}


// MarshalJSON marshal to time type.
func (d DateTime) MarshalJSON() ([]byte, error) {
	if d == 0 {
		return []byte{'"', '"'}, nil
	}
	return d.Time().MarshalJSON()
}

// UnmarshalJSON creates a primitive.DateTime from a JSON string.
func (d *DateTime) UnmarshalJSON(data []byte) error {
	// Ignore "null" to keep parity with the time.Time type and the standard library. Decoding "null" into a non-pointer
	// DateTime field will leave the field unchanged. For pointer values, the encoding/json will set the pointer to nil
	// and will not defer to the UnmarshalJSON hook.
	if len(data) == 0 || string(data) == "null" || string(data) == "\"\"" {
		return nil
	}

	if string(data) == "O" || string(data) == "\"0\"" {
		return nil
	}

	var tempTime time.Time
	if err := tempTime.UnmarshalJSON(data); err != nil {
		return err
	}

	*d = NewDateTimeFromTime(tempTime)
	return nil
}

// Time returns the date as a time type.
func (d DateTime) Time() time.Time {
	return time.UnixMilli(int64(d))
}

// NewDateTimeFromTime creates a new DateTime from a Time.
func NewDateTimeFromTime(t time.Time) DateTime {
	return DateTime(t.UnixMilli())
}

func (d DateTime) IsZero() bool {
	return d == 0
}

func (d DateTime) Format(layout string) string {
	return d.Time().Format(layout)
}
