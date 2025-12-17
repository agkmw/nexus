package errs

import (
	"errors"
)

type ErrCode struct {
	value int
}

func (e ErrCode) Value() int {
	return e.value
}

var (
	Internal          = ErrCode{value: 0}
	BadRequest        = ErrCode{value: 1}
	FailedValidation  = ErrCode{value: 2}
	NotFound          = ErrCode{value: 3}
	MethodNotAllowed  = ErrCode{value: 4}
	EditConflict      = ErrCode{value: 5}
	RateLimitExceeded = ErrCode{value: 6}
	AlreadyExists     = ErrCode{value: 7}
)

var (
	InternalMsg          = errors.New("the server encountered a problem and could not process your request")
	BadRequestMsg        = errors.New("the request payload contains malformed JSON")
	FailedValidationMsg  = errors.New("the request payload failed the validation rules")
	NotFoundMsg          = errors.New("the requested resource could not be found")
	MethodNotAllowedMsg  = errors.New("the %s method is not supported for the requested resource")
	EditConflictMsg      = errors.New("unable to update the resource due to an edit conflict")
	RateLimitExceededMsg = errors.New("rate limit exceeded")
	AlreadyExistsMsg     = errors.New("the record already exists")
)
