package custom_error

import (
	"fmt"
	"runtime"
)

type ErrorType int

type ErrorDetails struct {
	Line int
	File string
}

type CustomError interface {
	fmt.Stringer
	error
	Unwrap() error
	GetError() error
	GetSubError() CustomError
	GetDetails() ErrorDetails
	GetType() ErrorType
}

type customErrorImpl struct {
	subError CustomError
	details  ErrorDetails
	err      error
}

func (c *customErrorImpl) GetError() error {
	return c.err
}

func (c *customErrorImpl) Unwrap() error {
	if c.subError == nil {
		return c.err
	}
	return c.subError
}

func (c *customErrorImpl) GetSubError() CustomError {
	return c.subError
}

func (c *customErrorImpl) GetDetails() ErrorDetails {
	return c.details
}

func (c *customErrorImpl) GetType() ErrorType {
	return ErrorType(0)
}

func (c *customErrorImpl) String() string {
	thisErr := fmt.Sprintf("%v:%v\n\tError: %v\n\n", c.details.File, c.details.Line, c.err)
	if c.subError != nil {
		thisErr += c.subError.String()
	}
	return thisErr
}

func (c *customErrorImpl) Error() string {
	return c.String()
}

type typedCustomErrorImpl struct {
	customErrorImpl
	errType ErrorType
}

func (c *typedCustomErrorImpl) GetType() ErrorType {
	return c.errType
}

func newError(subError CustomError, err error, skip int) *customErrorImpl {
	details := ErrorDetails{}
	_, details.File, details.Line, _ = runtime.Caller(skip)
	return &customErrorImpl{
		subError: subError,
		details:  details,
		err:      err,
	}
}

func newTypedError(errType ErrorType, subError CustomError, err error, skip int) CustomError {
	return &typedCustomErrorImpl{
		customErrorImpl: *newError(subError, err, skip),
		errType:         errType,
	}
}

func MakeError(err error) CustomError {
	return newError(nil, err, 2)
}

func MakeErrorf(format string, args ...interface{}) CustomError {
	return newError(nil, fmt.Errorf(format, args...), 2)
}

func WrapError(subError CustomError, err error) CustomError {
	return newTypedError(subError.GetType(), subError, err, 3)
}

func WrapErrorf(subError CustomError, format string, args ...interface{}) CustomError {
	return newTypedError(subError.GetType(), subError, fmt.Errorf(format, args...), 3)
}

func NewTypedError(errType ErrorType, subError CustomError, err error) CustomError {
	return newTypedError(errType, subError, err, 3)
}

func MakeTypedError(errType ErrorType, err error) CustomError {
	return newTypedError(errType, nil, err, 3)
}

func MakeTypedErrorf(errType ErrorType, format string, args ...interface{}) CustomError {
	return newTypedError(errType, nil, fmt.Errorf(format, args...), 3)
}

func NewTypedErrorf(errType ErrorType, subError CustomError, format string, args ...interface{}) CustomError {
	return newTypedError(errType, subError, fmt.Errorf(format, args...), 3)
}
