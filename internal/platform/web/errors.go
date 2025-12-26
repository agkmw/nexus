package web

import (
	"context"
	"net/http"

	"github.com/agkmw/reddit-clone/internal/platform/errs"
)

// NOTE: Add request_id field or not?

var httpStatus = map[errs.ErrorType]int{
	errs.Aborted:            http.StatusConflict,
	errs.AlreadyExists:      http.StatusConflict,
	errs.EditConflict:       http.StatusConflict,
	errs.FailedPrecondition: http.StatusPreconditionFailed,
	errs.FailedValidation:   http.StatusUnprocessableEntity,
	errs.NotFound:           http.StatusNotFound,
	errs.Internal:           http.StatusInternalServerError,
	errs.InvalidArgument:    http.StatusBadRequest,
	errs.PermissionDenied:   http.StatusForbidden,
	errs.TooManyRequests:    http.StatusTooManyRequests,
	errs.Unauthenticated:    http.StatusUnauthorized,
	errs.Unknown:            http.StatusInternalServerError,
}

func ServerErrorResponse(ctx context.Context, w http.ResponseWriter) error {
	msg := "the server encountered a problem and could not process your request"
	return ErrorResponse(ctx, w, errs.Internal, msg)
}

func NotFoundResponse(ctx context.Context, w http.ResponseWriter) error {
	msg := "the requested resource was not found"
	return ErrorResponse(ctx, w, errs.NotFound, msg)
}

func BadRequestResponse(ctx context.Context, w http.ResponseWriter, data errs.ErrorInfo) error {
	msg := "the request body contains invalid JSON"
	return ErrorResponseWithData(ctx, w, errs.InvalidArgument, msg, data)
}

func FailedValidationResponse(ctx context.Context, w http.ResponseWriter, data errs.ErrorInfo) error {
	msg := "the request failed validation"
	return ErrorResponseWithData(ctx, w, errs.FailedValidation, msg, data)
}

func RateLimitExceeded(ctx context.Context, w http.ResponseWriter) error {
	msg := "too many requests, please try again later"
	return ErrorResponse(ctx, w, errs.TooManyRequests, msg)
}

func EditConflictResponse(ctx context.Context, w http.ResponseWriter) error {
	msg := "unable to modify the resource due to an edit conflict, please try again"
	return ErrorResponse(ctx, w, errs.EditConflict, msg)
}

func AlreadyExistsResponse(ctx context.Context, w http.ResponseWriter) error {
	msg := "the resource already exists"
	return ErrorResponse(ctx, w, errs.AlreadyExists, msg)
}

func InvalidCredentialsResponse(ctx context.Context, w http.ResponseWriter) error {
	msg := "invalid authentication credentials"
	return ErrorResponse(ctx, w, errs.Unauthenticated, msg)
}

func InvalidAuthenticationTokenResponse(ctx context.Context, w http.ResponseWriter) error {
	w.Header().Set("WWW-Authenticate", "Bearer")

	msg := "invalid or missing authentication token"
	return ErrorResponse(ctx, w, errs.Unauthenticated, msg)
}

func AuthenticationRequiredResponse(ctx context.Context, w http.ResponseWriter) error {
	msg := "authentication is required to access this resource"
	return ErrorResponse(ctx, w, errs.Unauthenticated, msg)
}

func InactiveAccountResponse(ctx context.Context, w http.ResponseWriter) error {
	msg := "your account must be activated to access this resource"
	return ErrorResponse(ctx, w, errs.PermissionDenied, msg)
}

func NotPermittedResponse(ctx context.Context, w http.ResponseWriter) error {
	msg := "you do not have the permissions to access this resource"
	return ErrorResponse(ctx, w, errs.PermissionDenied, msg)
}

func ErrorResponse(
	ctx context.Context,
	w http.ResponseWriter,
	errType errs.ErrorType,
	message string,
) error {
	return errorResponse(ctx, w, errType, message, nil)
}

func ErrorResponseWithData(
	ctx context.Context,
	w http.ResponseWriter,
	errType errs.ErrorType,
	message string,
	data errs.ErrorInfo,
) error {
	return errorResponse(ctx, w, errType, message, data)
}

func errorResponse(
	ctx context.Context,
	w http.ResponseWriter,
	errType errs.ErrorType,
	message string,
	data errs.ErrorInfo,
) error {
	status, ok := httpStatus[errType]
	if !ok {
		status = http.StatusInternalServerError
	}

	env := Envelope{
		"code":    errType.String(),
		"message": message,
	}

	if data != nil {
		env["data"] = data
	}

	return Encode(ctx, w, status, env)
}
