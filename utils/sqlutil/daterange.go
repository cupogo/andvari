package sqlutil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	LayoutDate  = "2006-1-2"
	LayoutMonth = "2006-1"
	Separator   = "~"
	LayoutPrint = "2006-01-02T15:04"
)

var (
	reDate  = regexp.MustCompile(`^[12][0-9]{3}-?[0-9]{1,2}-?[0-9]{1,2}$`)
	reMonth = regexp.MustCompile(`^[12][0-9]{3}-[0-9]{1,2}$`)
)

func IsDate(s string) bool {
	if n := len(s); n >= 8 && n <= 10 {
		return reDate.MatchString(s)
	}
	return false
}

func IsMonth(s string) bool {
	if n := len(s); n >= 6 && n <= 7 {
		return reMonth.MatchString(s)
	}
	return false
}

// DateRange ...
type DateRange struct {
	Start time.Time
	End   time.Time

	ts  time.Time
	loc *time.Location
}

func (dr DateRange) String() string {
	return dr.Start.Format(LayoutPrint) + Separator + dr.End.Format(LayoutPrint)
}

func (dr DateRange) Interval() time.Duration {
	return dr.End.Sub(dr.Start)
}

func (dr DateRange) Previous() DateRange {
	return DateRange{
		Start: dr.Start.Add(0 - dr.Interval()),
		End:   dr.Start,
	}
}

func (dr DateRange) Location() *time.Location {
	if dr.loc != nil {
		return dr.loc
	}
	return dr.ts.Location()
}

// GetDateRange parse date during and max days
// examples:
//
//	"2019-08-12" whole day
//	"2019-08" whole month
//	"1_day" a pass day
//	"1_week" a pass week
//	"1_month" a pass 30 days
//	"1_year" a pass 365 days
//	"2019-08-12~2019-10-12" range
func GetDateRange(during string, days ...int) (*DateRange, error) {
	dr := NewDateRange(time.Now())
	if err := dr.Parse(during, days...); err != nil {
		return nil, err
	}
	return dr, nil
}

func NewDateRange(ts time.Time) *DateRange {
	return &DateRange{ts: ts, loc: ts.Location()}
}

func (dr *DateRange) Parse(during string, days ...int) (err error) {

	if 0 == len(during) || "all" == during {
		var maxDays int
		if len(days) > 0 {
			maxDays = days[0]
		}
		if maxDays < 1 {
			maxDays = 1
		}
		dr.Start = dr.ts.AddDate(0, 0, -maxDays)
		dr.End = dr.ts
		return
	}

	if IsDate(during) {
		dr.Start, err = time.ParseInLocation(LayoutDate, during, dr.Location())
		if err != nil {
			return
		}
		dr.End = dr.Start.Add(time.Hour * 24)
		return
	}

	if IsMonth(during) {
		dr.Start, err = time.ParseInLocation(LayoutMonth, during, dr.Location())
		if err != nil {
			return
		}
		year, month, day := dr.Start.Date()
		dr.End = dr.newDate(year, month+1, day)
		return
	}

	if a, b, ok := strings.Cut(during, Separator); ok {
		return dr.parse2(a, b)
	}

	return dr.parseSpec(during)
}

func (dr *DateRange) parse2(a, b string) (err error) {
	if !IsDate(a) {
		err = fmt.Errorf("invalid date %q", a)
	}
	if !IsDate(b) {
		err = fmt.Errorf("invalid date %q", b)
	}

	dr.Start, err = time.ParseInLocation(LayoutDate, a, dr.Location())
	if err != nil {
		return
	}

	dr.End, err = time.ParseInLocation(LayoutDate, b, dr.Location())
	if err != nil {
		return
	}
	dr.End = dr.End.Add(time.Hour * 24)

	return
}

func (dr *DateRange) parseSpec(during string) (err error) {

	year, month, day := dr.ts.Date()
	var unit string
	var num int
	if a, b, ok := strings.Cut(during, "_"); ok {
		unit = b

		if a == "last" {
			switch unit {
			case "day":
				dr.Start = dr.newDate(year, month, day-1)
				dr.End = dr.newDate(year, month, day)

			case "week":
				mon := WeekStart(dr.ts)
				dr.Start = mon.AddDate(0, 0, -7)
				dr.End = mon

			case "month":
				dr.Start = dr.newDate(year, month-1, 1)
				dr.End = dr.newDate(year, month, 1)

			case "year":
				dr.Start = dr.newDate(year-1, 1, 1)
				dr.End = dr.newDate(year, 1, 1)

			default:
				err = fmt.Errorf("invalid unit %q", b)
			}

			return
		}

		if a == "this" {
			switch unit {
			case "day":
				dr.Start = dr.newDate(year, month, day)
				dr.End = dr.newDate(year, month, day+1)

			case "week":
				mon := WeekStart(dr.ts)
				dr.Start = mon
				dr.End = mon.AddDate(0, 0, 7)

			case "month":
				dr.Start = dr.newDate(year, month, 1)
				dr.End = dr.newDate(year, month+1, 1)

			case "year":
				dr.Start = dr.newDate(year, 1, 1)
				dr.End = dr.newDate(year+1, 1, 1)

			default:
				err = fmt.Errorf("invalid unit %q", b)
			}

			return
		}

		num, err = strconv.Atoi(a)
		if err != nil {
			return
		}

	}

	dr.End = dr.ts
	switch unit {
	case "day", "days":
		if 0 == num {
			dr.Start = dr.ts.Truncate(24 * time.Hour)
		} else {
			dr.Start = dr.ts.AddDate(0, 0, -num)
		}
	case "week", "weeks":
		if 0 == num {
			dr.Start = WeekStart(dr.ts)
		} else {
			dr.Start = dr.ts.AddDate(0, 0, -num*7)
		}

	case "month", "months":
		if 0 == num {
			dr.Start = dr.newDate(year, month, 1)
		} else {
			dr.Start = dr.ts.AddDate(0, -num, 0)
		}

	case "year", "years":
		if 0 == num {
			dr.Start = dr.newDate(year, time.January, 1)
		} else {
			dr.Start = dr.ts.AddDate(-num, 0, 0)
		}

	default:
		err = fmt.Errorf("invalid during %q", during)
	}

	return
}

func (dr *DateRange) newDate(year int, month time.Month, day int) time.Time {
	return newDate(year, month, day, dr.Location())
}

func WeekStart(t time.Time) time.Time {
	year, month, day := t.Date()
	if wd := t.Weekday(); wd == time.Sunday {
		day = day - 6
	} else {
		day = day - int(wd) + 1
	}
	return newDate(year, month, day, t.Location())
}

func newDate(year int, month time.Month, day int, loc *time.Location) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}
