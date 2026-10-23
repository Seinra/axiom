package flow

import (
	"context"
	"fmt"
	"time"
)

// Retry runs work until it succeeds, up to attempts tries in total.
//
// A nil failure stops the loop and reports nil: later attempts never run.
// When every attempt fails, the last failure wins. Attempts below one and a
// negative backoff panic: both are programmer errors, matching the
// size-below-zero panic policy of Chunked and Take. A zero backoff retries
// immediately.
//
// Cancellation is honored before each attempt and while waiting between
// them: a cancelled context reports its own error and work stops starting.
// A panicking work aborts the loop and the panic propagates — panics are
// fail-fast here, never retried, matching Race's panic broadcast. The
// standard library has no bounded retry.
func Retry(active context.Context, attempts int, backoff time.Duration, work func(context.Context) error) error {
	if attempts < 1 {
		panic(fmt.Sprintf("axiom flow: Retry needs attempts >= 1, got %d", attempts))
	}
	if backoff < 0 {
		panic(fmt.Sprintf("axiom flow: Retry needs backoff >= 0, got %s", backoff))
	}
	_, failure := RetryWithResult(active, attempts, backoff, func(taskContext context.Context) (struct{}, error) {
		return struct{}{}, work(taskContext)
	})
	return failure
}

// RetryWithResult runs work until it succeeds, up to attempts tries in total,
// returning the successful value.
//
// Semantics match Retry: nil failure stops with the value, exhaustion
// reports the zero value plus the last failure, attempts below one and a
// negative backoff panic, cancellation reports the context error, and a
// panicking work propagates without retry. The standard library has no
// bounded retry.
func RetryWithResult[Output any](active context.Context, attempts int, backoff time.Duration, work func(context.Context) (Output, error)) (Output, error) {
	if attempts < 1 {
		panic(fmt.Sprintf("axiom flow: RetryWithResult needs attempts >= 1, got %d", attempts))
	}
	if backoff < 0 {
		panic(fmt.Sprintf("axiom flow: RetryWithResult needs backoff >= 0, got %s", backoff))
	}
	var vacant Output
	for attempt := 1; ; attempt++ {
		select {
		case <-active.Done():
			return vacant, active.Err()
		default:
		}
		if outcome, failure := work(active); failure == nil {
			return outcome, nil
		} else if attempt >= attempts {
			return vacant, failure
		}
		select {
		case <-active.Done():
			return vacant, active.Err()
		case <-time.After(backoff):
		}
	}
}
