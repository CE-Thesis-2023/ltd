package custerror

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type CustomError struct {
	Msg  string
	Code uint32
}

func (e *CustomError) String() string {
	return fmt.Sprintf("%d - %s", e.Code, e.Msg)
}

func (e *CustomError) Error() string {
	return e.Msg
}

func (e *CustomError) Is(err error) bool {
	x, ok := err.(*CustomError)
	if !ok {
		return false
	}
	return x.Code == e.Code
}

func (e *CustomError) Parent() *CustomError {
	switch e {
	case ErrorAlreadyExists:
		return ErrorAlreadyExists
	case ErrorInvalidArgument:
		return ErrorInvalidArgument
	case ErrorFailedPrecondition:
		return ErrorFailedPrecondition
	case ErrorNotFound:
		return ErrorNotFound
	case ErrorPermissionDenied:
		return ErrorPermissionDenied
	case ErrorTimeout:
		return ErrorTimeout
	case ErrorTooManyRequest:
		return ErrorTooManyRequest
	case ErrorUnavailable:
		return ErrorUnavailable
	case ErrorUnimplemented:
		return ErrorUnimplemented
	default:
		return ErrorInternal
	}
}

func (e *CustomError) Fiber(ctx *fiber.Ctx) error {
	return ctx.
		Status(int(e.Code)).
		JSON(e.Response())
}

func (e *CustomError) Response() *ErrorResponse {
	return &ErrorResponse{
		Message: e.Msg,
		Code:    e.Code,
	}
}

type ErrorResponse struct {
	Message string `json:"message" msgpack:"message"`
	Code    uint32 `json:"code" msgpack:"code"`
}

func NewError(msg string, code uint32) *CustomError {
	return &CustomError{
		Msg:  msg,
		Code: code,
	}
}

func fromBase(err *CustomError, msg string) error {
	return &CustomError{
		Msg:  fmt.Sprintf("%s: %s", err.Msg, msg),
		Code: err.Code,
	}
}

var (
	ErrorNotFound           = NewError("Not found", http.StatusNotFound)
	ErrorAlreadyExists      = NewError("Already exists", http.StatusConflict)
	ErrorPermissionDenied   = NewError("Permission denied", http.StatusForbidden)
	ErrorInternal           = NewError("Internal error", http.StatusInternalServerError)
	ErrorInvalidArgument    = NewError("Invalid argument", http.StatusBadRequest)
	ErrorFailedPrecondition = NewError("Precondition unsatisfied", http.StatusFailedDependency)
	ErrorTooManyRequest     = NewError("Too many requests", http.StatusTooManyRequests)
	ErrorUnimplemented      = NewError("Unimplemented", http.StatusNotImplemented)
	ErrorUnavailable        = NewError("Unavailable", http.StatusServiceUnavailable)
	ErrorTimeout            = NewError("Timed out", http.StatusRequestTimeout)
)

func FormatNotFound(msg string, args ...interface{}) error {
	return fromBase(ErrorNotFound, fmt.Sprintf(msg, args...))
}

func FormatAlreadyExists(msg string, args ...interface{}) error {
	return fromBase(ErrorAlreadyExists, fmt.Sprintf(msg, args...))
}

func FormatPermissionDenied(msg string, args ...interface{}) error {
	return fromBase(ErrorPermissionDenied, fmt.Sprintf(msg, args...))
}

func FormatInternalError(msg string, args ...interface{}) error {
	return fromBase(ErrorInternal, fmt.Sprintf(msg, args...))
}

func FormatInvalidArgument(msg string, args ...interface{}) error {
	return fromBase(ErrorInvalidArgument, fmt.Sprintf(msg, args...))
}

func FormatFailedPrecondition(msg string, args ...interface{}) error {
	return fromBase(ErrorFailedPrecondition, fmt.Sprintf(msg, args...))
}

func FormatTooManyRequests(msg string, args ...interface{}) error {
	return fromBase(ErrorTooManyRequest, fmt.Sprintf(msg, args...))
}

func FormatUnimplemented(msg string, args ...interface{}) error {
	return fromBase(ErrorUnimplemented, fmt.Sprintf(msg, args...))
}

func FormatUnavailable(msg string, args ...interface{}) error {
	return fromBase(ErrorUnavailable, fmt.Sprintf(msg, args...))
}

func FormatTimeout(msg string, args ...interface{}) error {
	return fromBase(ErrorTimeout, fmt.Sprintf(msg, args...))
}
