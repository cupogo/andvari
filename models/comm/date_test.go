package comm

import (
	"testing"
	"time"
)

type date struct {
	year  int
	month int
	day   int
}

func (d date) Time() time.Time {
	return time.Date(d.year, time.Month(d.month), d.day, 0, 0, 0, 0, time.UTC)
}

func TestNewDate(t *testing.T) {
	tests := []struct {
		name         string
		year         int
		month        int
		day          int
		expectedDate Date
	}{
		{"	one year", 1901, 0, 0, Date(-32)},
		{"	one day", 1901, 0, 1, Date(-31)},
		{"	one month", 1901, 1, 0, Date(-1)}, // Assuming January
		{"	zero start", 1901, 1, 1, Date(0)},
		{"	two year", 1904, 0, 0, Date(1063)}, // Non-leap year
		{"	present 1", 2023, 10, 23, Date(44855)},
		{"	present 2", 2023, 10, 28, Date(44860)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := date{tt.year, tt.month, tt.day}.Time()
			got := NewDate(tt.year, tt.month, tt.day)
			t.Logf("date %s(%d) from %s, dur %s", got.String(), got, dt.Format(time.DateOnly), dt.Sub(zeroTime))
			if got != tt.expectedDate {
				t.Errorf("NewDate(%d, %d, %d) = %d; want %d", tt.year, tt.month, tt.day, got, tt.expectedDate)
			}
		})
	}
}

func TestNewDateFromTime(t *testing.T) {
	tests := []struct {
		name         string
		time         time.Time
		expectedDate Date
	}{
		{"from base time", zeroTime, Date(0)},
		{"one day later", zeroTime.AddDate(0, 0, 1), Date(1)},
		{"two days later", zeroTime.AddDate(0, 0, 2), Date(2)},
		{"100 days later", zeroTime.AddDate(0, 0, 100), Date(100)},
		{"500 days later", zeroTime.AddDate(0, 0, 500), Date(500)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDateFromTime(tt.time)
			if got != tt.expectedDate {
				t.Errorf("NewDateFromTime(%v) = %v; want %v", tt.time, got, tt.expectedDate)
			}
		})
	}
}

func TestDate_Time(t *testing.T) {
	baseTime := zeroTime
	_, offset := baseTime.Zone()
	tests := []struct {
		name string
		date Date
		want time.Time
	}{
		{"basic test for day 1", 1, baseTime.AddDate(0, 0, 1)},
		{"test for day 2", 2, baseTime.AddDate(0, 0, 2)},
		{"zero date", zeroDate - 1, baseNormalizeTime(-24*time.Hour, offset)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.date.Time(); !got.Equal(tt.want) {
				t.Errorf("Date.Time() = %v, want %v", got, tt.want)
			}
		})
	}
}

func baseNormalizeTime(duration time.Duration, offset int) time.Time {
	return zeroTime.Add(duration - time.Duration(offset)*time.Second)
}

func TestDate_Age(t *testing.T) {
	currentTime := time.Date(2023, 3, 25, 0, 0, 0, 0, time.UTC)
	timeNow := func() time.Time { return currentTime }

	tests := []struct {
		name    string
		date    Date
		mockNow func() time.Time
		wantAge int
	}{
		{"born today", NewDateFromTime(currentTime), timeNow, 0},
		{"30 years ago", NewDateFromTime(currentTime.AddDate(-30, 0, 0)), timeNow, 30},
		{"before birthday", NewDateFromTime(timeNow().AddDate(-20, 0, -1)), timeNow, 20},
		{"after birthday", NewDateFromTime(timeNow().AddDate(-25, 0, 1)), timeNow, 24},
		{"future date", NewDateFromTime(timeNow().AddDate(10, 0, 0)), timeNow, -10}, // Time travelers!
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Logf("date: %s", tt.date)
			nowFunc = tt.mockNow
			if got := tt.date.Age(); got != tt.wantAge {
				t.Errorf("%v Date.Age() = %d, want %d", tt.date.Time(), got, tt.wantAge)
			}
		})
	}
}

func TestDate_String_Format_Marshal_Unmarshal(t *testing.T) {
	zeroTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		date         Date
		stringOutput string
		format       string
		formattedOut string
		marshalOut   string
	}{
		{"Beginning of epoch", Date(0), "2000-01-01", "January 01, 2006", "January 01, 2000", "2000-01-01"},
		{"One day later", Date(1), "2000-01-02", "Jan 02, 2006", "Jan 02, 2000", "2000-01-02"},
		{"One month later", Date(31), "2000-02-01", "Jan 02, 2006 Monday", "Feb 01, 2000 Tuesday", "2000-02-01"},
		{"Leap year test", Date(366), "2001-01-01", "2006", "2001", "2001-01-01"},
		{"Random typical date", Date(2789), "2007-08-21", "January 01 2006", "August 08 2007", "2007-08-21"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.date.String(); got != tt.stringOutput {
				t.Errorf("Date.String() = %v, want %v", got, tt.stringOutput)
			}

			if got := tt.date.Format(tt.format); got != tt.formattedOut {
				t.Errorf("Date.Format(%v) = %v, want %v", tt.format, got, tt.formattedOut)
			}

			marshaled, err := tt.date.MarshalText()
			if err != nil {
				t.Errorf("Date.MarshalText() unexpected error: %v", err)
			}
			if string(marshaled) != tt.marshalOut {
				t.Errorf("Date.MarshalText() = %v, want %v", string(marshaled), tt.marshalOut)
			}
			marshaled = append(marshaled, []byte("T16:00:00.000Z")...)

			var unmarshaled Date
			if err := unmarshaled.UnmarshalText(marshaled); err != nil {
				t.Errorf("Date.UnmarshalText() unexpected error: %v", err)
			}
			if !unmarshaled.Equal(tt.date) {
				t.Errorf("Date.UnmarshalText() got %v, want %v", unmarshaled, tt.date)
			}
		})
	}
}
