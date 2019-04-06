package logs

type prefixedLogger struct {
	base   Logger
	prefix string
}

func (l *prefixedLogger) Debugf(format string, args ...interface{}) {
	l.base.Debugf(l.prefix+format, args...)
}

func (l *prefixedLogger) Infof(format string, args ...interface{}) {
	l.base.Infof(l.prefix+format, args...)
}

func (l *prefixedLogger) Warningf(format string, args ...interface{}) {
	l.base.Warningf(l.prefix+format, args...)
}

func (l *prefixedLogger) Errorf(format string, args ...interface{}) {
	l.base.Errorf(l.prefix+format, args...)
}

func (l *prefixedLogger) Fatalf(format string, args ...interface{}) {
	l.base.Fatalf(l.prefix+format, args...)
}

func NewPrefixedLogger(base Logger, prefix string) Logger {
	return &prefixedLogger{
		base:   base,
		prefix: prefix,
	}
}
