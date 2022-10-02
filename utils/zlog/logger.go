package zlog

// Logger 日志记录器
type Logger interface {

	// Debug uses fmt.Sprint to construct and log a message. // deprecated
	Debug(args ...interface{})
	// Info uses fmt.Sprint to construct and log a message. // deprecated
	Info(args ...interface{})
	// Warn uses fmt.Sprint to construct and log a message. // deprecated

	// Debugf uses fmt.Sprintf to log a templated message.
	Debugf(template string, args ...interface{})
	// Infof uses fmt.Sprintf to log a templated message.
	Infof(template string, args ...interface{})
	// Warnf uses fmt.Sprintf to log a templated message.
	Warnf(template string, args ...interface{})
	// Errorf uses fmt.Sprintf to log a templated message.
	Errorf(template string, args ...interface{})
	// Panicf uses fmt.Sprintf to log a templated message, then panics.
	Panicf(template string, args ...interface{})
	// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
	Fatalf(template string, args ...interface{})

	// Debugw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	//
	// When debug-level logging is disabled, this is much faster than
	//  s.With(keysAndValues).Debug(msg)
	Debugw(msg string, keysAndValues ...interface{})
	// Infow logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Infow(msg string, keysAndValues ...interface{})
	// Warnw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Warnw(msg string, keysAndValues ...interface{})
	// Errorw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Errorw(msg string, keysAndValues ...interface{})
	// Panicw logs a message with some additional context, then panics. The
	// variadic key-value pairs are treated as they are in With.
	Panicw(msg string, keysAndValues ...interface{})
	// Fatalw logs a message with some additional context, then calls os.Exit. The
	// variadic key-value pairs are treated as they are in With.
	Fatalw(msg string, keysAndValues ...interface{})
}
