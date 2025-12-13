package errs

import (
	"fmt"
)

type ServerError struct {
	Code   int
	LogErr error // for internal logging
	ResErr error // for client response
}

func NewServerError(code int, logErr error, resErr error) *ServerError {
	return &ServerError{
		Code:   code,
		LogErr: logErr,
		ResErr: resErr,
	}
}

func NewServerErrorf(code int, logErr error, format string, v ...any) *ServerError {
	return &ServerError{
		Code:   code,
		LogErr: logErr,
		ResErr: fmt.Errorf(format, v...),
	}
}

func (se *ServerError) Error() string {
	return se.ResErr.Error()
}

// =============================================================================

type ClientError struct {
	Code   int
	Data   any
	ResErr error
}

func NewClientError(code int, data any, resErr error) *ClientError {
	return &ClientError{
		Code:   code,
		Data:   data,
		ResErr: resErr,
	}
}

func NewClientErrorf(code int, data any, format string, v ...any) *ClientError {
	return &ClientError{
		Code:   code,
		Data:   data,
		ResErr: fmt.Errorf(format, v...),
	}
}

func (ce *ClientError) Error() string {
	return ce.ResErr.Error()
}
