package comm

import (
	"errors"
)

var (
	ErrEmptyID            = errors.New("empty id")
	ErrInvalidTime        = errors.New("invalid time")
	ErrInvalidHour        = errors.New("invalid hour")
	ErrInvalidMinute      = errors.New("invalid minute")
	ErrInvalidSecond      = errors.New("invalid seconds")
	ErrInvalidMillisecond = errors.New("invalid milliseconds")
)
