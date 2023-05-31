package zlog

import "context"

// Logger 日志记录器
type Logger interface {

	// Debug uses fmt.Sprint to construct and log a message. // deprecated
	Debug(args ...any)
	// Info uses fmt.Sprint to construct and log a message. // deprecated
	Info(args ...any)
	// Warn uses fmt.Sprint to construct and log a message. // deprecated

	// Debugf uses fmt.Sprintf to log a templated message.
	Debugf(template string, args ...any)
	// Infof uses fmt.Sprintf to log a templated message.
	Infof(template string, args ...any)
	// Warnf uses fmt.Sprintf to log a templated message.
	Warnf(template string, args ...any)
	// Errorf uses fmt.Sprintf to log a templated message.
	Errorf(template string, args ...any)
	// Panicf uses fmt.Sprintf to log a templated message, then panics.
	Panicf(template string, args ...any)
	// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
	Fatalf(template string, args ...any)

	// Debugw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	//
	// When debug-level logging is disabled, this is much faster than
	//  s.With(keysAndValues).Debug(msg)
	Debugw(msg string, keysAndValues ...any)
	// Infow logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Infow(msg string, keysAndValues ...any)
	// Warnw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Warnw(msg string, keysAndValues ...any)
	// Errorw logs a message with some additional context. The variadic key-value
	// pairs are treated as they are in With.
	Errorw(msg string, keysAndValues ...any)
	// Panicw logs a message with some additional context, then panics. The
	// variadic key-value pairs are treated as they are in With.
	Panicw(msg string, keysAndValues ...any)
	// Fatalw logs a message with some additional context, then calls os.Exit. The
	// variadic key-value pairs are treated as they are in With.
	Fatalw(msg string, keysAndValues ...any)
}

type LoggerX interface {
	Logger
	// Infow logs a message with some additional context.
	InfowContext(ctx context.Context, msg string, keysAndValues ...any)
	// Warnw logs a message with some additional context.
	WarnwContext(ctx context.Context, msg string, keysAndValues ...any)
	// Errorw logs a message with some additional context.
	ErrorwContext(ctx context.Context, msg string, keysAndValues ...any)
	// Panicw logs a message with some additional context, then panics.
	PanicwContext(ctx context.Context, msg string, keysAndValues ...any)
}
