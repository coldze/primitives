package custom_error

type ErrorBuilder interface {
	MakeErrorf(format string, args ...interface{}) CustomError
	NewErrorf(subError CustomError, format string, args ...interface{}) CustomError
}
