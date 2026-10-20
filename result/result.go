// Package result provides a typed Result[T] for operations that yield either a success or a failure.
//
// Use Ok to wrap a success and Fail to wrap a failure. Unwrap splits both
// paths explicitly, Must unwraps a known success (it panics only on programmer
// error, never for control flow), and Map/MapValue transform the success side
// while carrying the failure through untouched.
package result

import "fmt"

// Result holds either a success of type T or a failure.
// The zero value is a failure with a nil error; prefer Ok/Fail constructors.
type Result[T any] struct {
	success T
	failure error
	isOk    bool
}

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
// success when IsOk is true, failure otherwise.
func (current Result[T]) Unwrap() (T, error) {
	if !current.isOk {
		return current.success, current.failure
	}
	return current.success, nil
}

// Must returns the success, panicking only when current holds a failure.
// Use Must solely for values known at construction time (tests, static seeds).
func Must[T any](current Result[T]) T {
	if !current.isOk {
		panic(fmt.Sprintf("axiom result: Must on failure: %v", current.failure))
	}
	return current.success
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
