package reduction

import (
	"fmt"
	"net/http"
)

// StatusError is an error with the HTTP status it answers. A reducer, Init
// or Refresh may return one (wrapped or not); the mux answers its status
// instead of 500.
type StatusError interface {
	error
	HTTPStatus() int
}

// FieldError says an action's field is wrong: 400, with the field named.
type FieldError struct {
	Field   string
	Message string
}

func (e *FieldError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *FieldError) HTTPStatus() int { return http.StatusBadRequest }

// Invalid is a FieldError. field may be empty for a request wrong as a whole.
func Invalid(field, message string) error {
	return &FieldError{Field: field, Message: message}
}

// NotAllowedError says the caller may not do this: 403.
type NotAllowedError struct {
	Message string
}

func (e *NotAllowedError) Error() string { return e.Message }

func (e *NotAllowedError) HTTPStatus() int { return http.StatusForbidden }

// NotAllowed is a NotAllowedError.
func NotAllowed(message string) error {
	return &NotAllowedError{Message: message}
}

type noKeyError struct{}

func (noKeyError) Error() string   { return "no reduction key in context" }
func (noKeyError) HTTPStatus() int { return http.StatusUnauthorized }

// ErrNoKey is Get's and Set's answer when ctx names no key (see WithKey).
var ErrNoKey error = noKeyError{}
