package timeutil

import (
	"sync"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	// 设置测试用例的本地时区为UTC，确保测试结果一致
	time.Local = time.UTC

	tests := []struct {
		name     string
		timeStr  string
		wantTime time.Time
		wantErr  bool
	}{
		// 空字符串测试
		{
			name:    "Empty string",
			timeStr: "",
			wantErr: true,
		},
		{
			name:    "Whitespace only",
			timeStr: "   ",
			wantErr: true,
		},

		// ISO 8601 格式测试
		{
			name:     "ISO 8601 with timezone",
			timeStr:  "2024-04-01T15:04:05+08:00",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.FixedZone("", 8*60*60)),
			wantErr:  false,
		},
		{
			name:     "ISO 8601 UTC",
			timeStr:  "2024-04-01T15:04:05Z",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "ISO 8601 without timezone",
			timeStr:  "2024-04-01T15:04:05",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "ISO 8601 variant with space",
			timeStr:  "2024-04-01 15:04:05",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},

		// 日期格式测试
		{
			name:     "ISO 8601 date only",
			timeStr:  "2024-04-01",
			wantTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Asian date format",
			timeStr:  "2024/04/01",
			wantTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "European date format (DD/MM/YYYY)",
			timeStr:  "01/04/2024",
			wantTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		// {
		// 	name:     "US date format (MM/DD/YYYY)",
		// 	timeStr:  "04/01/2024",
		// 	wantTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		// 	wantErr:  false,
		// },
		{
			name:     "European date format with dots",
			timeStr:  "01.04.2024",
			wantTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Compact date format",
			timeStr:  "20240401",
			wantTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},

		// 带时间的日期格式测试
		{
			name:     "Date with hour and minute",
			timeStr:  "2024-04-01 15:04",
			wantTime: time.Date(2024, 4, 1, 15, 4, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Asian date with time",
			timeStr:  "2024/04/01 15:04:05",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "European date with time",
			timeStr:  "01/04/2024 15:04:05",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		// {
		// 	name:     "US date with time",
		// 	timeStr:  "04/01/2024 15:04:05",
		// 	wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
		// 	wantErr:  false,
		// },

		// 12小时制格式测试
		{
			name:     "ISO date with 12-hour format (PM)",
			timeStr:  "2024-04-01 3:04:05 PM",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "ISO date with 12-hour format (AM)",
			timeStr:  "2024-04-01 3:04:05 AM",
			wantTime: time.Date(2024, 4, 1, 3, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "US date with 12-hour format",
			timeStr:  "04/01/2024 3:04:05 PM",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},

		// RFC 格式测试
		{
			name:     "RFC 3339",
			timeStr:  "2024-04-01T15:04:05Z",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "RFC 1123",
			timeStr:  "Mon, 01 Apr 2024 15:04:05 GMT",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "RFC 822",
			timeStr:  "01 Apr 24 15:04 GMT",
			wantTime: time.Date(2024, 4, 1, 15, 4, 0, 0, time.UTC),
			wantErr:  false,
		},

		// 带毫秒/微秒/纳秒的格式测试
		{
			name:     "Date time with milliseconds",
			timeStr:  "2024-04-01 15:04:05.123",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 123000000, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Date time with microseconds",
			timeStr:  "2024-04-01 15:04:05.123456",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 123456000, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Date time with nanoseconds",
			timeStr:  "2024-04-01 15:04:05.123456789",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 123456789, time.UTC),
			wantErr:  false,
		},

		// 紧凑格式测试
		{
			name:     "Compact datetime format",
			timeStr:  "20240401150405",
			wantTime: time.Date(2024, 4, 1, 15, 4, 5, 0, time.UTC),
			wantErr:  false,
		},

		// Unix时间戳测试
		{
			name:     "Unix timestamp (seconds)",
			timeStr:  "1712073845", // 2024-04-01 15:04:05 UTC
			wantTime: time.Date(2024, 4, 2, 16, 4, 5, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Unix timestamp (milliseconds)",
			timeStr:  "1712073845123", // 2024-04-01 15:04:05.123 UTC
			wantTime: time.Date(2024, 4, 2, 16, 4, 5, 123000000, time.UTC),
			wantErr:  false,
		},

		// 错误格式测试
		{
			name:    "Invalid format",
			timeStr: "2024-13-01",
			wantErr: true,
		},
		{
			name:    "Random string",
			timeStr: "not a date",
			wantErr: true,
		},
		{
			name:    "Invalid timestamp",
			timeStr: "99999999999999999999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, err := ParseTime(tt.timeStr)

			// 检查错误情况
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 如果期望有错误，不需要检查时间值
			if tt.wantErr {
				return
			}

			// 检查解析的时间是否符合预期
			if !gotTime.Equal(tt.wantTime) {
				t.Errorf("ParseTime() = %v, want %v", gotTime, tt.wantTime)
			}
		})
	}
}

// 测试边界情况
func TestParseTimeEdgeCases(t *testing.T) {
	// 设置测试用例的本地时区为UTC，确保测试结果一致
	time.Local = time.UTC

	tests := []struct {
		name     string
		timeStr  string
		wantTime time.Time
		wantErr  bool
	}{
		{
			name:     "Very old date",
			timeStr:  "1800-01-01",
			wantTime: time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Future date",
			timeStr:  "2100-12-31",
			wantTime: time.Date(2100, 12, 31, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Leap year date",
			timeStr:  "2024-02-29",
			wantTime: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:    "Non-leap year February 29",
			timeStr: "2023-02-29",
			wantErr: true,
		},
		{
			name:     "Extreme timezone",
			timeStr:  "2024-04-01T12:00:00+14:00",
			wantTime: time.Date(2024, 4, 1, 12, 0, 0, 0, time.FixedZone("", 14*60*60)),
			wantErr:  false,
		},
		{
			name:     "Negative timezone",
			timeStr:  "2024-04-01T12:00:00-12:00",
			wantTime: time.Date(2024, 4, 1, 12, 0, 0, 0, time.FixedZone("", -12*60*60)),
			wantErr:  false,
		},
		{
			name:     "Unix epoch start",
			timeStr:  "0",
			wantTime: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:    "Negative timestamp",
			timeStr: "-1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, err := ParseTime(tt.timeStr)

			// 检查错误情况
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 如果期望有错误，不需要检查时间值
			if tt.wantErr {
				return
			}

			// 检查解析的时间是否符合预期
			if !gotTime.Equal(tt.wantTime) {
				t.Errorf("ParseTime() = %v, want %v", gotTime, tt.wantTime)
			}
		})
	}
}

// 测试性能
func BenchmarkParseTime(b *testing.B) {
	testCases := []string{
		"2024-04-01T15:04:05Z",          // ISO 8601
		"2024-04-01 15:04:05",           // 常见格式
		"01/04/2024",                    // 欧洲日期格式
		"1712073845",                    // Unix时间戳
		"Mon, 01 Apr 2024 15:04:05 GMT", // RFC 1123
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 轮流测试不同格式
		testCase := testCases[i%len(testCases)]
		_, _ = ParseTime(testCase)
	}
}

// 测试并发安全性
func TestParseTimeConcurrent(t *testing.T) {
	testCases := []string{
		"2024-04-01T15:04:05Z",
		"2024-04-01 15:04:05",
		"01/04/2024",
		"1712073845",
		"Mon, 01 Apr 2024 15:04:05 GMT",
	}

	const numGoroutines = 10
	const iterationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				testCase := testCases[(id+j)%len(testCases)]
				_, err := ParseTime(testCase)
				if err != nil {
					t.Errorf("Goroutine %d: ParseTime(%q) failed: %v", id, testCase, err)
				}
			}
		}(i)
	}

	wg.Wait()
}
