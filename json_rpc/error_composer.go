package json_rpc

type ErrorComposer interface {
	MakeError(code int, err error) ServerError
}

type defaultErrorComposer struct {
	errors map[int]string
	module int
}

func (e *defaultErrorComposer) MakeError(code int, err error) ServerError {
	v, ok := e.errors[code]
	if !ok {
		v = "UNKNOWN ERROR."
	}
	return MakeError(e.module, code, v, err)
}

func NewDefaultErrorComposer(module int, errors map[int]string) ErrorComposer {
	return &defaultErrorComposer{
		errors: errors,
		module: module,
	}
}
