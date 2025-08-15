package comm

import (
	"encoding/json"
	"testing"
	"time"
)

var _ json.Marshaler = DateTime(0)
var _ json.Unmarshaler = (*DateTime)(nil)

// 实现 ITime 接口的测试结构体
type testITime struct {
	t time.Time
}

func (ti testITime) Time() time.Time {
	return ti.t
}

func TestAsDateTime(t *testing.T) {
	now := time.Now()
	nowDateTime := NewDateTimeFromTime(now)

	tests := []struct {
		name     string
		input    interface{}
		expected DateTime
		ok       bool
	}{
		{"time.Time", now, nowDateTime, true},
		{"*time.Time", &now, nowDateTime, true},
		{"DateTime", nowDateTime, nowDateTime, true},
		{"*DateTime", &nowDateTime, nowDateTime, true},
		{"ITime", testITime{now}, nowDateTime, true},
		{"int64", int64(1630000000), DateTime(1630000000), true},
		{"valid RFC3339 string", "2021-08-27T12:00:00Z", NewDateTimeFromTime(time.Date(2021, 8, 27, 12, 0, 0, 0, time.UTC)), true},
		{"valid DateTime string", "2021-08-27 12:00:00", NewDateTimeFromTime(time.Date(2021, 8, 27, 12, 0, 0, 0, time.UTC)), true},
		{"invalid string", "invalid", DateTime(0), false},
		{"unsupported type", 3.14, DateTime(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := AsDateTime(tt.input)
			if ok != tt.ok {
				t.Errorf("AsDateTime(%v) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if result != tt.expected {
				t.Errorf("AsDateTime(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
