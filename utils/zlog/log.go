package zlog

import (
	"fmt"
	syslog "log"
)

type logger struct{}

// dftLogger 默认实例
var dftLogger Logger

func init() {
	syslog.SetFlags(syslog.Ltime | syslog.Lshortfile)
	dftLogger = &logger{}
}

// Set ...
func Set(logger Logger) {
	if logger != nil {
		dftLogger = logger
	}
}

// Get ...
func Get() Logger {
	return dftLogger
}

// deprecated
func (z *logger) Debug(args ...interface{}) {
	syslog.Print(args...)
}

// deprecated
func (z *logger) Info(args ...interface{}) {
	syslog.Print(args...)
}
func (z *logger) Printf(template string, args ...interface{}) {
	syslog.Output(2, fmt.Sprintf(template, args...))
}

func (z *logger) Debugf(template string, args ...interface{}) {
	syslog.Output(2, fmt.Sprintf(template, args...))
}

func (z *logger) Infof(template string, args ...interface{}) {
	syslog.Output(2, fmt.Sprintf(template, args...))
}

func (z *logger) Warnf(template string, args ...interface{}) {
	syslog.Output(2, fmt.Sprintf(template, args...))
}

func (z *logger) Errorf(template string, args ...interface{}) {
	syslog.Output(2, fmt.Sprintf(template, args...))
}

func (z *logger) Panicf(template string, args ...interface{}) {
	syslog.Panicf(template, args...)
}

func (z *logger) Fatalf(template string, args ...interface{}) {
	syslog.Fatalf(template, args...)
}

func (z *logger) Debugw(msg string, keysAndValues ...interface{}) {
	syslog.Output(2, fmt.Sprint("DEBUG: "+msg, keysAndValues))
}

func (z *logger) Infow(msg string, keysAndValues ...interface{}) {
	syslog.Output(2, fmt.Sprint("INFO: "+msg, keysAndValues))
}

func (z *logger) Warnw(msg string, keysAndValues ...interface{}) {
	syslog.Output(2, fmt.Sprint("WARN: "+msg, keysAndValues))
}

func (z *logger) Errorw(msg string, keysAndValues ...interface{}) {
	syslog.Output(2, fmt.Sprint("ERROR: "+msg, keysAndValues))
}

func (z *logger) Panicw(msg string, keysAndValues ...interface{}) {
	syslog.Panic(msg, keysAndValues)
}

func (z *logger) Fatalw(msg string, keysAndValues ...interface{}) {
	syslog.Fatal(msg, keysAndValues)
}

// func Debug(args ...interface{}) {
// 	Get().Debug(args...)
// }

// func Info(args ...interface{}) {
// 	Get().Info(args...)
// }

// func Warn(args ...interface{}) {
// 	Get().Warn(args...)
// }

// func Error(args ...interface{}) {
// 	Get().Error(args...)
// }

// func Fatal(args ...interface{}) {
// 	Get().Fatal(args...)
// }

func Debugf(template string, args ...interface{}) {
	Get().Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	Get().Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	Get().Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	Get().Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	Get().Fatalf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	Get().Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	Get().Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	Get().Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	Get().Errorw(msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	Get().Fatalw(msg, keysAndValues...)
}
