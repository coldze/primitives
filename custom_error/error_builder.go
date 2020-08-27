package custom_error

type ErrorBuilder interface {
	MakeErrorf(format string, args ...interface{}) CustomError
	WrapErrorf(subError CustomError, format string, args ...interface{}) CustomError
}
