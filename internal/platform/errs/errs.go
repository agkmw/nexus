package errs

import (
	"errors"
	"fmt"
	"runtime"
)

type ErrorInfo map[string]any

type Error struct {
	typ   ErrorType
	cause error
	data  ErrorInfo
	file  string
	fn    string
}

func New(t ErrorType, cause error, data ErrorInfo) *Error {
	if t.t == "" {
		t = Unknown
	}

	if cause == nil {
		cause = ErrNilCause
	}

	pc, file, line, _ := runtime.Caller(2)

	return &Error{
		typ:   t,
		cause: cause,
		data:  data,
		file:  file,
		fn:    fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), line),
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf(
		"[%s] %s:%s: %v",
		e.typ, e.file, e.fn, e.cause,
	)
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Type() ErrorType {
	return e.typ
}

func (e *Error) Data() ErrorInfo {
	return e.data
}

func (e *Error) Location() (file, fn string) {
	return e.file, e.fn
}

// =============================================================================

func Is(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

func IsType(err error, t ErrorType) bool {
	var e *Error
	if !errors.As(err, &e) {
		return false
	}

	return e.typ == t
}

func Get(err error) (*Error, bool) {
	var e *Error
	if !errors.As(err, &e) {
		return nil, false
	}

	return e, true
}
