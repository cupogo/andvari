package sqlutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWeekStart(t *testing.T) {
	const startYear = 2022
	const startMonth = 9
	const startDay = 19

	for i := 0; i < 7; i++ {
		ts := time.Date(startYear, startMonth, startDay+i, 10, 20, 0, 0, time.Local)
		nt := WeekStart(ts)
		_, _, day := nt.Date()

		t.Logf("week start: %s => %s", ts.Format(LayoutPrint), nt.Format(LayoutPrint))

		assert.Equal(t, startDay, day)
		assert.Zero(t, nt.Hour())
		assert.Zero(t, nt.Minute())
	}
}

func TestDateRange(t *testing.T) {
	var dr *DateRange
	var err error
	dr, err = GetDateRange("", 5)
	assert.NoError(t, err)
	assert.NotNil(t, dr)
	assert.NotZero(t, dr.Start)
	t.Logf("dr %v", dr)

	_, err = GetDateRange("bad_days")
	assert.Error(t, err)

	_, err = GetDateRange("invalid")
	assert.Error(t, err)

	for _, during := range []string{
		"2019-07-12",
		"2019-07",
		"2019-7-6",
		"2019-7",
		"0_day",
		"3_days",
		"0_week",
		"2_week",
		"0_month",
		"2_month",
		"0_year",
		"1_year",
		"2019-07-12~2019-08-12",
	} {
		dr, err = GetDateRange(during)
		assert.NoError(t, err)
		assert.NotNil(t, dr)
		assert.NotZero(t, dr.Start)
		assert.NotZero(t, dr.End)
		assert.True(t, dr.Start.Unix() < dr.End.Unix())
		t.Logf("dr %10s => %+v", during, dr)
	}

	//日期范围的末尾精度修正
	dr, err = GetDateRange("2019-07-12~2019-08-31")
	if dr.End.Format("2006-01-02") == "2019-09-01" {
		t.Logf("dr %v", dr)
	} else {
		t.Errorf("want:\"2019-08-13\",but get:%v", dr.End.Format("2006-01-02"))
	}
}

// "last_"开头的日期测试
func TestDateRangeForLast(t *testing.T) {
	var dr *DateRange
	var err error
	dr, err = GetDateRange("last_season")
	assert.Error(t, err)
	assert.Nil(t, dr)

	for _, during := range []string{
		"last_day",
		"last_week",
		"last_month",
		"last_year",
	} {
		dr, err = GetDateRange(during)
		assert.NoError(t, err)
		assert.NotNil(t, dr)
		assert.NotZero(t, dr.Start)
		assert.NotZero(t, dr.End)
		assert.True(t, dr.Start.Unix() < dr.End.Unix())
		t.Logf("dr %10s => %+v", during, dr)
	}
}

// "this_"开头的日期测试
func TestDateRangeForThis(t *testing.T) {
	var dr *DateRange
	var err error
	dr, err = GetDateRange("this_season")
	assert.Error(t, err)
	assert.Nil(t, dr)

	for _, during := range []string{
		"this_day",
		"this_week",
		"this_month",
		"this_year",
	} {
		dr, err = GetDateRange(during)
		assert.NoError(t, err)
		assert.NotNil(t, dr)
		assert.NotZero(t, dr.Start)
		assert.NotZero(t, dr.End)
		assert.True(t, dr.Start.Unix() < dr.End.Unix())
		t.Logf("dr %10s => %+v", during, dr)
	}
}
