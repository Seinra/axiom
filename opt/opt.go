// Package opt provides a typed Option[T] for values that may be absent.
//
// Use Some to hold a success and None for absence. ValueOr reads with an
// explicit fallback, MapOption/MapValue transform the held success while
// carrying absence through untouched. Absence is a normal outcome here, never
// an error: for failures with causes, see the result package.
//
// Which function?
//
// | When you want to ...             | Reach for ...        |
// |-------------------------------|----------------------|
// | Hold a present value          | Some                 |
// | Express absence               | None                 |
// | Ask whether a value is held   | IsSome               |
// | Read with an explicit fallback| ValueOr              |
// | Read with a lazy fallback     | ValueOrElse          |
// | Transform the held value      | MapOption / MapValue |
// | Chain to the next Option      | AndThen              |
// | Collapse one nesting level      | Flatten              |
// | Fall back to another Option   | OrElse               |
// | Keep the value only when it fits | Filter            |
// | Convert absence into a failure| ToResult (see result)|
package opt

import (
	"github.com/seinra/axiom/result"
)

// Option holds either a success of type T or no success.
type Option[T any] struct {
	success T
	isSome  bool
}

// Some holds success as a present Option.
func Some[T any](success T) Option[T] {
	return Option[T]{success: success, isSome: true}
}

// None returns an absent Option.
func None[T any]() Option[T] {
	return Option[T]{}
}

// IsSome reports whether current holds a success.
func (current Option[T]) IsSome() bool {
	return current.isSome
}

// ValueOr returns the held success, or fallback when absent.
func (current Option[T]) ValueOr(fallback T) T {
	if !current.isSome {
		return fallback
	}
	return current.success
}

// MapOption transforms the held success with transform, carrying absence through.
// Pure function form; see MapValue for the chainable Go 1.27 generic method.
func MapOption[T any, Transformed any](current Option[T], transform func(T) Transformed) Option[Transformed] {
	if !current.isSome {
		return None[Transformed]()
	}
	return Some(transform(current.success))
}

// MapValue transforms the held success, carrying absence through.
// Go 1.27 generic method: the method declares its own type params, so calls
// chain left to right: Some(5).MapValue(next).MapValue(describe).
func (current Option[T]) MapValue[Transformed any](transform func(T) Transformed) Option[Transformed] {
	return MapOption(current, transform)
}

// OrElse returns current when it holds a success, otherwise fallback.
// The first Some wins; None plus None stays None.
func OrElse[T any](current Option[T], fallback Option[T]) Option[T] {
	if current.isSome {
		return current
	}
	return fallback
}

// AndThen binds the held success through transform, which returns the next Option.
// None short-circuits: transform never runs and absence carries through.
// Function-only by design: MapValue already covers the func(T) Transformed
// shape, and Go has no overloading, so the Option-returning shape stays a
// function to keep the method set minimal.
func AndThen[T any, Transformed any](current Option[T], transform func(T) Option[Transformed]) Option[Transformed] {
	if !current.isSome {
		return None[Transformed]()
	}
	return transform(current.success)
}

// Flatten collapses one nesting level: None yields None and Some(inner) yields inner.
// Absence carries through like the other short-circuiting combinators.
// For example, Flatten(Some(Some(8))) holds 8 while Flatten(Some(None[int]())) is absent.
func Flatten[Outcome any](current Option[Option[Outcome]]) Option[Outcome] {
	if !current.isSome {
		return None[Outcome]()
	}
	return current.success
}

// Filter keeps the held success only when keep reports true, else None.
// Absence carries through without running keep.
func Filter[T any](current Option[T], keep func(T) bool) Option[T] {
	if !current.isSome {
		return None[T]()
	}
	if !keep(current.success) {
		return None[T]()
	}
	return current
}

// ValueOrElse returns the held success, or the outcome of fallback when absent.
// Fallback runs lazily: it never runs when current holds a success.
func ValueOrElse[T any](current Option[T], fallback func() T) T {
	if current.isSome {
		return current.success
	}
	return fallback()
}

// ToResult converts current into a result.Result: Some becomes Ok, None
// becomes Fail with failure. A nil failure flows into Fail, whose nil guard
// replaces it with a descriptive error, following the Rust ok_or precedent.
func ToResult[T any](current Option[T], failure error) result.Result[T] {
	if current.isSome {
		return result.Ok(current.success)
	}
	return result.Fail[T](failure)
}
