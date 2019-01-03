package logs

import "context"

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

type LoggerGetter func(ctx context.Context) Logger

type loggerKey struct {
	ID string
}

var ctxLoggerKey loggerKey
var defaultLogger Logger

func SetLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey, logger)
}

func GetLogger(ctx context.Context) Logger {
	v := ctx.Value(ctxLoggerKey)
	if v == nil {
		return defaultLogger
	}
	res, ok := v.(Logger)
	if !ok {
		return defaultLogger
	}
	return res
}

func init() {
	defaultLogger = NewStdLogger()
	ctxLoggerKey = loggerKey{
		ID: "logger",
	}
}
