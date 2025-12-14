package errs

import (
	"fmt"
	"runtime"
)

type ServerError struct {
	Code     ErrCode
	LogMsg   error // for internal logging
	ResMsg   error // for client response
	FileName string
	FuncName string
}

func NewServerError(code ErrCode, logMsg error, resMsg error) *ServerError {
	pc, file, line, _ := runtime.Caller(1)

	return &ServerError{
		Code:     code,
		LogMsg:   logMsg,
		ResMsg:   resMsg,
		FileName: fmt.Sprintf("%s:%d", file, line),
		FuncName: runtime.FuncForPC(pc).Name(),
	}
}

func NewServerErrorf(code ErrCode, logMsg error, format string, v ...any) *ServerError {
	pc, file, line, _ := runtime.Caller(1)

	return &ServerError{
		Code:     code,
		LogMsg:   logMsg,
		ResMsg:   fmt.Errorf(format, v...),
		FileName: fmt.Sprintf("%s:%d", file, line),
		FuncName: runtime.FuncForPC(pc).Name(),
	}
}

func (se *ServerError) Error() string {
	return se.ResMsg.Error()
}

// =============================================================================

type ClientError struct {
	Code     ErrCode
	Data     any
	ResMsg   error
	FileName string
	FuncName string
}

func NewClientError(code ErrCode, data any, resMsg error) *ClientError {
	pc, file, line, _ := runtime.Caller(1)

	return &ClientError{
		Code:     code,
		Data:     data,
		ResMsg:   resMsg,
		FileName: fmt.Sprintf("%s:%d", file, line),
		FuncName: runtime.FuncForPC(pc).Name(),
	}
}

func NewClientErrorf(code ErrCode, data any, format string, v ...any) *ClientError {
	pc, file, line, _ := runtime.Caller(1)

	return &ClientError{
		Code:     code,
		Data:     data,
		ResMsg:   fmt.Errorf(format, v...),
		FileName: fmt.Sprintf("%s:%d", file, line),
		FuncName: runtime.FuncForPC(pc).Name(),
	}
}

func (ce *ClientError) Error() string {
	return ce.ResMsg.Error()
}
