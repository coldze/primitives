package logs

import "context"

type LoggerGetter func(ctx context.Context) Logger

type loggerKey struct {
	ID string
}

var ctxLoggerKey loggerKey
var defaultLogger Logger

func SetLogger(ctx context.Context, logger Logger) context.Context {
	if ctx == nil {
		return context.WithValue(context.Background(),ctxLoggerKey, logger)
	}
	return context.WithValue(ctx, ctxLoggerKey, logger)
}

func GetLogger(ctx context.Context) Logger {
	if ctx == nil {
		return defaultLogger
	}
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