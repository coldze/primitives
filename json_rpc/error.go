package json_rpc

const (
	module_sh = 32
)

type Error struct {
	Code    int64       `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ServerError interface {
	GetCode() int64
	GetMessage() string
	GetData() *string
	ToError() *Error
}

type serverErrorImpl struct {
	code    int64
	message string
	data    error
}

func (e *serverErrorImpl) GetCode() int64 {
	return e.code
}

func (e *serverErrorImpl) GetMessage() string {
	return e.message
}

func (e *serverErrorImpl) GetData() *string {
	if e.data == nil {
		return nil
	}
	v := e.data.Error()
	return &v
}

func (e *serverErrorImpl) ToError() *Error {
	return &Error{
		Code:    e.GetCode(),
		Message: e.GetMessage(),
		Data:    e.GetData(),
	}
}

func MakeError(module int, errorCode int, message string, err error) ServerError {
	return &serverErrorImpl{
		code:    MakeErrorCode(module, errorCode),
		message: message,
		data:    err,
	}
}

func MakeErrorWithCode(code int64, message string, err error) ServerError {
	return &serverErrorImpl{
		code:    code,
		message: message,
		data:    err,
	}
}

func ThrowError(module int, errorCode int, message string, err error) {
	panic(MakeError(module, errorCode, message, err))
}

func MakeErrorCode(module int, errorCode int) int64 {
	return (int64(module) << module_sh) + int64(errorCode)
}
