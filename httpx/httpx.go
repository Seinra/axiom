// Package httpx provides typed Echo adapters over a single JSON envelope.
//
// No middleware, router helpers, or validation here — validation is the valid
// package's job at the handler edge. Reach for net/http or Echo directly for
// transport; reach for this package when handlers speak the single JSON envelope.
//
// Which function?
//
// | When you want to ...                | Reach for ...       |
// |--------------------------------------|---------------------|
// | Parse the request body               | Decode              |
// | Reply with a success envelope        | WriteJSON (Response) |
// | Reply with a failure envelope        | WriteError (Response) |
// | Reply validation failures as 422     | ValidationWriteError |
// | Name one field failure (matchable)  | FieldFailure (+ ErrValidationFailed) |
// | Shape a paged listing payload        | ListResponse        |
package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/seinra/axiom/repo"
)

// Response is the single JSON envelope for success and failure replies.
// Exactly one of Data/Error is meaningful per response: success sets Data
// with Error empty, failure sets Error with Data omitted.
//
// Contract: Data carries `omitempty`, so a scalar zero Body (0, "", false)
// encodes with no data key — read a keyless `{}` success as "zero value",
// never as failure. Prefer struct, slice, or map Bodies: a zero struct
// still emits `"data":{}`.
type Response[Body any] struct {
	Data  Body   `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// Decode parses the current request into a fresh payload of type Body.
// Decode failures return the Echo failure wrapped with the target type name
// for teachable messages.
func Decode[Body any](httpContext *echo.Context) (Body, error) {
	var payload Body
	if decodeFailure := httpContext.Bind(&payload); decodeFailure != nil {
		var empty Body
		return empty, fmt.Errorf("axiom httpx: decode %T: %w", payload, decodeFailure)
	}
	return payload, nil
}

// WriteJSON writes envelope holding body with status via context JSON encoding.
func WriteJSON[Body any](httpContext *echo.Context, status int, body Body) error {
	return httpContext.JSON(status, Response[Body]{Data: body})
}

// WriteError writes an envelope holding failure text with status.
// A nil failure becomes a descriptive message, never an empty envelope.
func WriteError(httpContext *echo.Context, status int, failure error) error {
	if failure == nil {
		failure = errors.New("axiom httpx: WriteError needs a non-nil failure")
	}
	return httpContext.JSON(status, Response[any]{Error: failure.Error()})
}

// ErrValidationFailed is the sentinel every FieldFailure wraps. Match
// per-field failures with errors.Is(failure, ErrValidationFailed) and
// recover the field with errors.As into FieldFailure.
var ErrValidationFailed = errors.New("axiom httpx: validation failed")

// FieldFailure reports one field's validation failure. Field names the
// rejected field and Reason explains the failure in plain words; Cause
// optionally carries the underlying error. Error renders the human text the
// 422 envelope carries; Unwrap exposes both the sentinel and the cause, so
// errors.Is and errors.As match per-field where the stdlib contract allows
// (in-process values — the HTTP envelope itself carries only text).
type FieldFailure struct {
	Field  string
	Reason string
	Cause  error
}

// Error renders the field failure as human text for the 422 envelope.
func (failure FieldFailure) Error() string {
	if failure.Cause == nil {
		return fmt.Sprintf("axiom httpx: field %q invalid: %s", failure.Field, failure.Reason)
	}
	return fmt.Sprintf("axiom httpx: field %q invalid: %s: %v", failure.Field, failure.Reason, failure.Cause)
}

// Unwrap exposes the sentinel plus the cause (when present) for errors.Is
// and errors.As traversal.
func (failure FieldFailure) Unwrap() []error {
	if failure.Cause == nil {
		return []error{ErrValidationFailed}
	}
	return []error{ErrValidationFailed, failure.Cause}
}

// ValidationWriteError writes validator failures as a 422 envelope with the
// failure messages joined by "; ". Nil entries inside failures are
// skipped. Empty or all-nil failures write a descriptive guard message,
// never an empty envelope. Per-field identity stays caller-side: match
// failures with errors.Is/As (for example FieldFailure) before rendering —
// the envelope carries only text.
func ValidationWriteError(httpContext *echo.Context, failures []error) error {
	messages := make([]string, 0, len(failures))
	for _, failure := range failures {
		if failure == nil {
			continue
		}
		messages = append(messages, failure.Error())
	}
	if len(messages) == 0 {
		return WriteError(httpContext, http.StatusUnprocessableEntity, errors.New("axiom httpx: ValidationWriteError needs at least one failure"))
	}
	return WriteError(httpContext, http.StatusUnprocessableEntity, errors.New(strings.Join(messages, "; ")))
}

// ListResponse carries a page of records with its page cursor and total count.
type ListResponse[Record any] struct {
	Items []Record  `json:"items"`
	Page  repo.Page `json:"page"`
	Total int       `json:"total"`
}
