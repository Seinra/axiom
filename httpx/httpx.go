// Package httpx provides typed Echo adapters over a single JSON envelope.
//
// No middleware, router helpers, or validation here — validation is the valid
// package's job at the handler edge.
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

// ValidationWriteError writes validator failures as a 422 envelope with the
// failure messages joined by "; ". Nil entries inside failures are
// skipped. Empty or all-nil failures write a descriptive guard message,
// never an empty envelope.
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
