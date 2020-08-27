package logs

import "log"

const (
	log_debug_prefix   = "\033[1;37m[DEBUG]\033[0m "
	log_info_prefix    = "[INFO] "
	log_warning_prefix = "\033[1;33m[WARNING]\033[0m "
	log_error_prefix   = "\033[1;31m[ERROR]\033[0m "
	log_fatal_prefix   = "\033[1;34m[FATAL]\033[0m "
)

type stdLogger struct {
}

func (l *stdLogger) Debugf(format string, args ...interface{}) {
	log.Printf(log_debug_prefix+format, args...)
}

func (l *stdLogger) Infof(format string, args ...interface{}) {
	log.Printf(log_info_prefix+format, args...)
}

func (l *stdLogger) Warningf(format string, args ...interface{}) {
	log.Printf(log_warning_prefix+format, args...)
}

func (l *stdLogger) Errorf(format string, args ...interface{}) {
	log.Printf(log_error_prefix+format, args...)
}

func (l *stdLogger) Fatalf(format string, args ...interface{}) {
	log.Printf(log_fatal_prefix+format, args...)
}

func NewStdLogger() Logger {
	return &stdLogger{}
}
