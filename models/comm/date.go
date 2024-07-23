package comm

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type Date int32

var (
	zeroDate = Date(0)
	zeroTime = time.Date(1901, 1, 1, 0, 0, 0, 0, time.UTC) // time to begin

	nowFunc = func() time.Time { return time.Now() }
)

func NewDate(year, month, day int) Date {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return NewDateFromTime(t)
}

func NewDateFromTime(t time.Time) Date {
	days := int32(t.Sub(zeroTime) / time.Hour / 24)
	return Date(days)
}

func Today() Date {
	return NewDateFromTime(nowFunc())
}

func (d Date) Time() time.Time {
	return zeroTime.AddDate(0, 0, int(d))
}

func (d Date) Add(days int) Date {
	return d + Date(days)
}

func (d Date) Since(other Date) int {
	return int(d - other)
}

func (d Date) After(other Date) bool {
	return d > other
}

func (d Date) Before(other Date) bool {
	return d < other
}

func (d Date) Equal(other Date) bool {
	return d == other
}

func (d Date) Age() int {
	now := nowFunc()
	birth := d.Time()
	age := now.Year() - birth.Year()

	if now.Before(birth.AddDate(age, 0, 0)) {
		age--
	}
	return age
}

func (d Date) IsZero() bool {
	return d == zeroDate
}

func (d Date) Format(layout string) string {
	t := d.Time()
	return t.Format(layout)
}

func (d Date) String() string {
	return d.Format(time.DateOnly)
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(text []byte) error {
	nd, err := ParseDate(string(text))
	if err != nil {
		return err
	}

	*d = nd
	return nil
}

func ParseDate(s string) (Date, error) {
	// truncate time
	if i := strings.IndexByte(s, 'T'); i > 0 {
		s = s[:i]
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return zeroDate, err
	}
	return NewDateFromTime(t), nil
}

func (d Date) Date() (year int, month time.Month, day int) {
	return d.Time().Date()
}

func (d Date) Weekday() time.Weekday {
	return d.Time().Weekday()
}

func (d Date) Year() int {
	return d.Time().Year()
}

func AsDate(tv any) (Date, bool) {
	switch t := tv.(type) {
	case Date:
		return t, true
	case time.Time:
		return NewDateFromTime(t), true
	case *time.Time:
		return NewDateFromTime(*t), true
	case ITime:
		return NewDateFromTime(t.Time()), true
	case int:
		return Date(t), true
	case int32:
		return Date(t), true
	case int64:
		return Date(t), true
	case float64:
		return Date(t), true
	case string:
		d, err := ParseDate(t)
		return d, err == nil
	case *string:
		d, err := ParseDate(*t)
		return d, err == nil
	default:
		return 0, false
	}
}

func (d *Date) Scan(src any) error {
	if b, ok := src.([]byte); ok {
		return d.UnmarshalText(b)
	}
	if s, ok := src.(string); ok {
		return d.UnmarshalText([]byte(s))
	}
	nd, ok := AsDate(src)
	if !ok {
		return fmt.Errorf("invalid %T(%v)", src, src)
	}
	*d = nd
	return nil
}

func (d *Date) Value() (driver.Value, error) {
	return d.String(), nil
}
