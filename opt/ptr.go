package opt

// To wraps value in a fresh pointer.
//
// The result is always live and independent: mutating through it never
// touches the caller's variable. The standard library offers no scalar
// value-to-pointer helper; taking &value aliases the caller's variable.
func To[T any](value T) *T {
	return &value
}

// Deref reads the pointer, reporting presence.
//
// A nil pointer yields the zero value and false. This is the OrNull shape:
// check found before trusting the value. Use ValueOr with Some when the
// caller already holds an Option instead of a raw pointer. The standard
// library has no nil-safe pointer read.
func Deref[T any](ptr *T) (T, bool) {
	if ptr == nil {
		var zero T
		return zero, false
	}
	return *ptr, true
}

// Coalesce returns the first live pointer, or nil when none is live.
//
// The winner is aliased, not copied: mutating through the result mutates the
// winning input. An empty call yields nil. The standard library has no
// first-live-pointer helper.
func Coalesce[T any](ptrs ...*T) *T {
	for _, ptr := range ptrs {
		if ptr != nil {
			return ptr
		}
	}
	return nil
}
