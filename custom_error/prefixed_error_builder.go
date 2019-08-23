package custom_error

import "fmt"

type prefixedCustomError struct {
	prefix string
}

func (p *prefixedCustomError) MakeErrorf(format string, args ...interface{}) CustomError {
	return newError(nil, fmt.Errorf(p.prefix+format, args...), 2)
}

func (p *prefixedCustomError) NewErrorf(subError CustomError, format string, args ...interface{}) CustomError {
	return newTypedError(subError.GetType(), subError, fmt.Errorf(p.prefix+format, args...), 2)
}

func NewPrefixedErrorBuilder(prefix string) ErrorBuilder {
	return &prefixedCustomError{
		prefix: prefix,
	}
}
