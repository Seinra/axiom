// Package pool provides a typed Pool[Item] for transient reuse of buffers.
//
// The pool may drop items at any GC; it is for transient reuse (buffers),
// NOT a cache and NOT an object pool with guarantees. Never assert Get
// returns the exact Put item: assert validity (a usable recycled or fresh
// value) and that the maker ran at least once.
//
// Use New to build a pool with a maker, or declare a zero Pool and Get from
// it directly: the zero value returns the zero Item and never panics.
// Reach for a plain value directly when reuse buys nothing; the pool only
// pays off for transient buffers callers Put back.
//
// Which function?
//
// | When you want to ...             | Reach for ... |
// |-----------------------------------|---------------|
// | Hold recyclables; build with maker | Pool / New    |
// | Borrow a value (or zero Item)     | Pool.Get      |
// | Return a value for reuse          | Pool.Put      |
package pool

import "sync"

// Pool holds recyclable values of type Item for transient reuse.
// Do not copy the Pool value after use; always use the pointer from New
// (or a heap address of a zero Pool), because the embedded sync.Pool must
// not be copied after use.
type Pool[Item any] struct {
	backing sync.Pool
}

// New returns a pool whose empty Get calls makeItem for a fresh value.
// Do not copy the Pool value; keep and share the returned pointer.
// A nil makeItem falls back to a zero-Item maker, so Get never panics.
func New[Item any](makeItem func() Item) *Pool[Item] {
	if makeItem == nil {
		makeItem = func() Item {
			var empty Item
			return empty
		}
	}
	pooled := &Pool[Item]{}
	pooled.backing.New = func() any {
		return makeItem()
	}
	return pooled
}

// Get returns a recycled value or a fresh one from the maker.
// A zero Pool with no maker returns the zero Item instead of panicking.
func (current *Pool[Item]) Get() Item {
	recycled, held := current.backing.Get().(Item)
	if !held {
		var empty Item
		return empty
	}
	return recycled
}

// Put returns candidate for reuse. The runtime may drop it at any GC,
// so callers must never assume the next Get returns this exact value.
func (current *Pool[Item]) Put(candidate Item) {
	current.backing.Put(candidate)
}
