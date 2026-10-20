// Package httpx provides typed Echo adapters over a single JSON envelope.
//
// No middleware, router helpers, or validation here — validation is the valid
// package's job at the handler edge.
package httpx

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v5"

	"github.com/seinra/axiom/repo"
)

// Response is the single JSON envelope for success and failure replies.
// Exactly one of Data/Error is meaningful per response: success sets Data
// with Error empty, failure sets Error with Data omitted.
type Response[Body any] struct {
	Data  Body   `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// Bind parses the current request into a fresh payload of type Body.
// Binding failures return the Echo failure wrapped with the target type name
// for teachable messages.
func Bind[Body any](httpContext *echo.Context) (Body, error) {
	var payload Body
	if bindFailure := httpContext.Bind(&payload); bindFailure != nil {
		var empty Body
		return empty, fmt.Errorf("axiom httpx: bind %T: %w", payload, bindFailure)
	}
	return payload, nil
}

// JSON writes envelope holding body with status via context JSON encoding.
func JSON[Body any](httpContext *echo.Context, status int, body Body) error {
	return httpContext.JSON(status, Response[Body]{Data: body})
}

// Fail writes an envelope holding failure text with status.
// A nil failure becomes a descriptive message, never an empty envelope.
func Fail(httpContext *echo.Context, status int, failure error) error {
	if failure == nil {
		failure = errors.New("axiom httpx: Fail needs a non-nil failure")
	}
	return httpContext.JSON(status, Response[any]{Error: failure.Error()})
}

// ListResponse carries a page of records with its page cursor and total count.
type ListResponse[Record any] struct {
	Items []Record  `json:"items"`
	Page  repo.Page `json:"page"`
	Total int       `json:"total"`
}
