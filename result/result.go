// Package result provides a typed Result[T] for operations that yield either a success or a failure.
//
// Use Ok to wrap a success and Fail to wrap a failure. Unwrap splits both
// paths explicitly, Must unwraps a known success (it panics only on programmer
// error, never for control flow), and MapValue transforms the success side
// while carrying the failure through untouched.
//
// Which function?
//
// | When you want to ...                | Reach for ...      |
// |-------------------------------------|--------------------|
// | Wrap a success                      | Ok                 |
// | Wrap a failure                      | Fail               |
// | Ask which side is held              | IsOk               |
// | Split both paths explicitly         | Unwrap             |
// | Read a known success (tests, seeds) | Must               |
// | Read with an explicit fallback      | UnwrapOr           |
// | Read with a lazy fallback           | UnwrapOrElse       |
// | Transform the success side          | MapValue           |
// | Transform the failure side          | MapFailure         |
// | Chain to the next Result            | AndThen            |
// | Collapse one nesting level          | Flatten            |
// | Recover from a failure              | OrElse             |
// | Convert panicking code at a boundary| Try                |
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
//
// Which panicker?
//
// | You hold ...                  | Reach for ...                          |
// |-------------------------------|------------------------------------------|
// | A Result with a known success | Must                                     |
// | A service name in a container | di.MustResolve                           |
// | Panicking code at a boundary  | Try, which returns a Result, never panic |
func Must[T any](current Result[T]) T {
	success, failure := current.Unwrap()
	if failure != nil {
		panic(fmt.Sprintf("axiom result: Must on failure: %v", failure))
	}
	return success
}

// MapValue transforms the success side, carrying failures through.
// Go 1.27 generic method: the method declares its own type params, so calls
// chain left to right: Ok(2).MapValue(triple).MapValue(describe).
func (current Result[T]) MapValue[Transformed any](transform func(T) Transformed) Result[Transformed] {
	if !current.isOk {
		return Fail[Transformed](current.failure)
	}
	return Ok(transform(current.success))
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

// Flatten collapses one nesting level: Fail yields Fail carrying the outer
// failure, Ok(inner) yields inner. A zero outer Result yields Fail with an
// explicit error, matching MapValue on a zero Result.
// For example, Flatten(Ok(Ok(3))) holds 3 while Flatten(Ok(Fail(cause))) holds cause.
func Flatten[Outcome any](current Result[Result[Outcome]]) Result[Outcome] {
	if !current.isOk {
		return Fail[Outcome](current.failure)
	}
	return current.success
}

// UnwrapOr returns the held success, or fallback when current holds a failure.
func UnwrapOr[T any](current Result[T], fallback T) T {
	if !current.isOk {
		return fallback
	}
	return current.success
}

// UnwrapOrElse returns the held success, or fallback(failure) when current holds a failure.
// Fallback runs lazily: it never runs when current holds a success, mirroring
// opt.ValueOrElse for the failure channel. A zero Result carries no failure, so
// fallback receives errZeroResult instead of nil and failures stay explicit.
// For example, UnwrapOrElse(Fail(cause), describe) yields describe(cause)
// while UnwrapOrElse(Ok(4), describe) yields 4 without running describe.
func UnwrapOrElse[Outcome any](current Result[Outcome], fallback func(error) Outcome) Outcome {
	if current.isOk {
		return current.success
	}
	if current.failure == nil {
		return fallback(errZeroResult)
	}
	return fallback(current.failure)
}

// Try runs attempt and wraps its value in Ok, converting a panic into Fail.
// An error payload is carried with %w, so errors.Is and errors.As match it
// with its message preserved ("axiom result: recovered panic: <payload>");
// a non-error payload is rendered with %v under the same message shape.
// The Scala Try precedent: Try is for boundary conversion of panicking code
// (parsers, callbacks, legacy calls) into an explicit Result, not for control
// flow — never use it to steer ordinary branching. A nil attempt panics when
// called and is recovered like any other panic, yielding the same Fail shape.
func Try[T any](attempt func() T) (outcome Result[T]) {
	defer func() {
		if payload := recover(); payload != nil {
			if payloadError, isError := payload.(error); isError {
				outcome = Fail[T](fmt.Errorf("axiom result: recovered panic: %w", payloadError))
				return
			}
			outcome = Fail[T](fmt.Errorf("axiom result: recovered panic: %v", payload))
		}
	}()
	return Ok(attempt())
}
