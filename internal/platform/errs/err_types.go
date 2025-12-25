package errs

import "errors"

var ErrNilCause = errors.New("nil error cause")

var (
	Aborted            = ErrorType{"aborted"}
	AlreadyExists      = ErrorType{"already_exists"}
	EditConflict       = ErrorType{"edit_conflict"}
	FailedPrecondition = ErrorType{"failed_precondition"}
	NotFound           = ErrorType{"not_found"}
	Internal           = ErrorType{"internal"}
	InvalidArgument    = ErrorType{"invalid_argument"}
	PermissionDenied   = ErrorType{"permission_denied"}
	TooManyRequests    = ErrorType{"too_many_requests"}
	Unauthenticated    = ErrorType{"unauthenticated"}
	Unknown            = ErrorType{"unknown"}
)

type ErrorType struct {
	t string
}

func (e ErrorType) String() string {
	return e.t
}

func (e ErrorType) Equal(e2 ErrorType) bool {
	return e.t == e2.t
}
