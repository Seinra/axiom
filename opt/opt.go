// Package opt provides a typed Option[T] for values that may be absent.
//
// Use Some to hold a success and None for absence. ValueOr reads with an
// explicit fallback, MapOption/MapValue transform the held success while
// carrying absence through untouched. Absence is a normal outcome here, never
// an error: for failures with causes, see the result package.
package opt

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
