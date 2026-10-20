// Package result provides a typed Result[T] for operations that yield either a success or a failure.
//
// Use Ok to wrap a success and Fail to wrap a failure. Unwrap splits both
// paths explicitly, Must unwraps a known success (it panics only on programmer
// error, never for control flow), and Map/MapValue transform the success side
// while carrying the failure through untouched.
package result

import (
	"errors"
	"fmt"
)

// Result holds either a success of type T or a failure.
// Never use the zero value: construct with Ok or Fail.
type Result[T any] struct {
	success T
	failure error
	isOk    bool
}

// errZeroResult is reported when Unwrap meets a zero Result, which carries no
// failure to explain itself. It only arises from bypassing Ok and Fail.
var errZeroResult = errors.New("axiom result: zero Result carries no failure; construct with Ok or Fail")

// Ok wraps success as a successful Result.
func Ok[T any](success T) Result[T] {
	return Result[T]{success: success, isOk: true}
}

// Fail wraps failure as a failed Result.
// A nil failure is replaced with a descriptive error so failures stay explicit.
func Fail[T any](failure error) Result[T] {
	if failure == nil {
		failure = fmt.Errorf("axiom result: nil failure wrapped as error")
	}
	return Result[T]{failure: failure}
}

// IsOk reports whether current holds a success.
func (current Result[T]) IsOk() bool {
	return current.isOk
}

// Unwrap returns the success and the failure. Exactly one is meaningful:
// success when IsOk is true, failure otherwise. A zero Result reports an
// error instead of a nil failure, so failures stay explicit.
func (current Result[T]) Unwrap() (T, error) {
	if !current.isOk {
		if current.failure == nil {
			return current.success, errZeroResult
		}
		return current.success, current.failure
	}
	return current.success, nil
}

// Must returns the success, panicking only when current holds a failure.
// Use Must solely for values known at construction time (tests, static seeds).
func Must[T any](current Result[T]) T {
	success, failure := current.Unwrap()
	if failure != nil {
		panic(fmt.Sprintf("axiom result: Must on failure: %v", failure))
	}
	return success
}

// Map transforms the success side with transform, carrying failures through.
// Pure function form; see MapValue for the chainable Go 1.27 generic method.
func Map[T any, Transformed any](current Result[T], transform func(T) Transformed) Result[Transformed] {
	if !current.isOk {
		return Fail[Transformed](current.failure)
	}
	return Ok(transform(current.success))
}

// MapValue transforms the success side, carrying failures through.
// Go 1.27 generic method: the method declares its own type params, so calls
// chain left to right: Ok(2).MapValue(triple).MapValue(describe).
func (current Result[T]) MapValue[Transformed any](transform func(T) Transformed) Result[Transformed] {
	return Map(current, transform)
}

// MapFailure transforms the failure side with transform, carrying success through.
// The Rust map_err precedent: success never runs transform. A nil transform
// outcome flows into Fail, whose nil guard replaces it with a descriptive error.
func MapFailure[T any](current Result[T], transform func(error) error) Result[T] {
	if current.isOk {
		return current
	}
	return Fail[T](transform(current.failure))
}

// AndThen binds the held success through transform, which returns the next Result.
// Failure short-circuits: transform never runs and the failure carries through.
func AndThen[T any, Transformed any](current Result[T], transform func(T) Result[Transformed]) Result[Transformed] {
	if !current.isOk {
		return Fail[Transformed](current.failure)
	}
	return transform(current.success)
}

// OrElse recovers a failure through handleFailure, carrying success through untouched.
// The Rust or_else precedent: success never runs handleFailure.
func OrElse[T any](current Result[T], handleFailure func(error) Result[T]) Result[T] {
	if current.isOk {
		return current
	}
	return handleFailure(current.failure)
}

// UnwrapOr returns the held success, or fallback when current holds a failure.
func UnwrapOr[T any](current Result[T], fallback T) T {
	if !current.isOk {
		return fallback
	}
	return current.success
}

// Try runs attempt and wraps its value in Ok, converting a panic into Fail
// with fmt.Errorf("axiom result: recovered panic: %v", payload).
// The Scala Try precedent: Try is for boundary conversion of panicking code
// (parsers, callbacks, legacy calls) into an explicit Result, not for control
// flow — never use it to steer ordinary branching. A nil attempt panics when
// called and is recovered like any other panic, yielding the same Fail shape.
func Try[T any](attempt func() T) (outcome Result[T]) {
	defer func() {
		if payload := recover(); payload != nil {
			outcome = Fail[T](fmt.Errorf("axiom result: recovered panic: %v", payload))
		}
	}()
	return Ok(attempt())
}
