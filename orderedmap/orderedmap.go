// Package orderedmap provides an insertion-ordered Map[Key, Value] and an
// LRU[Key, Value] eviction cache on top of the same ordering idea.
//
// The standard library offers the unordered map (iteration order is
// unspecified and must not be relied upon) and container/list (untyped,
// doubly linked, no key lookup): an ordered map — key lookup plus stable
// insertion order — is a recurring shape every caller re-derives with a
// backing map plus a keys slice. Map pins it once: Put and Get behave like
// the builtin map, while Keys and Values replay first-insertion order, and
// the zero value is ready without New. LRU pins the caching lesson on top:
// a fixed-capacity holder where every hit refreshes recency and entries
// beyond capacity evict least-recently-used first.
//
// Which type?
//
// | When you want to ...                          | Reach for ... |
// |-----------------------------------------------|-----------------|
// | Remember insertion order alongside lookup     | Map             |
// | Bound memory with least-recently-used eviction | LRU            |
// | Build an empty ordered map                    | New             |
// | Build a bounded recency cache                 | NewLRU          |
// | Add or replace an entry                       | Put             |
// | Read an entry (LRU: refreshes recency)        | Get             |
// | Remove an entry                               | Delete (Map)    |
// | Count held entries                            | Len             |
// | Ask whether nothing is held                   | IsEmpty (Map)   |
// | Replay keys or values in insertion order      | Keys, Values    |
//
// The (value, found) return on Get is this package's OrNull: found reports
// presence and value is the zero value when absent. A missing key is a
// normal outcome here, never an error and never a panic; for failures with
// causes, see the result package, and for a wrapped maybe, see opt. The only
// panic in this package is NewLRU with a capacity below 1: a capacity is a
// programmer decision fixed at construction, so a non-positive one is a
// programmer error, never a runtime condition.
//
// The zero Map is ready: declare var ordered Map[string, int] and Put
// directly. Methods initialize the nil backing on first use. (LRU has no
// useful zero value: capacity is required, so NewLRU is the only way in.)
// Do not copy a Map or LRU value after use; always use the pointer from New
// or NewLRU (or the address of a zero Map), because the slice, map, and
// list headers share their backing storage.
//
// Complexity: Map Put and Get run in amortized O(1) on the backing map;
// Delete runs in O(n) on the live keys because the surviving keys slide to
// close the gap. LRU Get and Put run in O(1) on the backing map plus the
// recency list. Tests pin eviction and recency order, never timing.
//
// This package is not safe for concurrent use; F3 owns concurrency.
package orderedmap

import "container/list"

// Map holds entries in first-insertion order: Put and Get behave like the
// builtin map, while Keys and Values replay insertion order. An upsert keeps
// the key's first-insertion position; Delete of a missing key is a no-op,
// and a re-Put after Delete takes the back.
// Do not copy a Map value after use; keep and share the pointer from New.
type Map[Key comparable, Value any] struct {
	values    map[Key]Value
	keys      []Key
	positions map[Key]int
}

// New returns an empty ordered map ready to Put.
// Do not copy the returned pointer's value; share the pointer itself.
func New[Key comparable, Value any]() *Map[Key, Value] {
	return &Map[Key, Value]{}
}

// Put adds the key with value, or replaces the value when the key already
// exists. An upsert keeps the key's first-insertion position.
func (ordered *Map[Key, Value]) Put(key Key, value Value) {
	if ordered.values == nil {
		ordered.values = make(map[Key]Value)
		ordered.positions = make(map[Key]int)
	}
	if _, present := ordered.values[key]; present {
		ordered.values[key] = value
		return
	}
	ordered.positions[key] = len(ordered.keys)
	ordered.keys = append(ordered.keys, key)
	ordered.values[key] = value
}

// Get returns the value held for key with found true. A missing key yields
// (zero, false): never an error, never a panic.
func (ordered *Map[Key, Value]) Get(key Key) (Value, bool) {
	value, found := ordered.values[key]
	return value, found
}

// Delete removes key. A missing key is a no-op. The surviving keys slide to
// close the gap, so Delete runs in O(n) on the live keys.
func (ordered *Map[Key, Value]) Delete(key Key) {
	position, present := ordered.positions[key]
	if !present {
		return
	}
	copy(ordered.keys[position:], ordered.keys[position+1:])
	ordered.keys = ordered.keys[:len(ordered.keys)-1]
	for _, following := range ordered.keys[position:] {
		ordered.positions[following]--
	}
	delete(ordered.positions, key)
	delete(ordered.values, key)
}

// Len counts the entries currently held.
func (ordered *Map[Key, Value]) Len() int {
	return len(ordered.keys)
}

// IsEmpty reports whether the map holds no entries.
func (ordered *Map[Key, Value]) IsEmpty() bool {
	return len(ordered.keys) == 0
}

// Keys returns the keys in first-insertion order. The result is a fresh
// slice: mutating it never disturbs the map.
func (ordered *Map[Key, Value]) Keys() []Key {
	keys := make([]Key, len(ordered.keys))
	copy(keys, ordered.keys)
	return keys
}

// Values returns the values in first-insertion order. The result is a fresh
// slice: mutating it never disturbs the map.
func (ordered *Map[Key, Value]) Values() []Value {
	values := make([]Value, 0, len(ordered.keys))
	for _, key := range ordered.keys {
		values = append(values, ordered.values[key])
	}
	return values
}

// lruEntry is one cache line held as the value of the recency list.
type lruEntry[Key comparable, Value any] struct {
	key   Key
	value Value
}

// LRU holds at most capacity entries: every Get hit and every Put refreshes
// the key's recency, and a Put beyond capacity evicts the
// least-recently-used entry first.
// Do not copy an LRU value after use; keep and share the pointer from NewLRU.
type LRU[Key comparable, Value any] struct {
	capacity int
	entries  map[Key]*list.Element
	recent   list.List
}

// NewLRU returns an empty recency cache holding at most capacity entries.
// NewLRU panics when capacity is below 1: capacity is fixed at construction,
// so a non-positive one is a programmer error, never a runtime condition.
// Do not copy the returned pointer's value; share the pointer itself.
func NewLRU[Key comparable, Value any](capacity int) *LRU[Key, Value] {
	if capacity < 1 {
		panic("orderedmap: NewLRU capacity must be at least 1")
	}
	return &LRU[Key, Value]{
		capacity: capacity,
		entries:  make(map[Key]*list.Element),
	}
}

// Get returns the value held for key with found true, refreshing the key's
// recency. A missing key yields (zero, false): never an error, never a
// panic, and a miss leaves recency untouched.
func (recent *LRU[Key, Value]) Get(key Key) (Value, bool) {
	element, present := recent.entries[key]
	if !present {
		var empty Value
		return empty, false
	}
	recent.recent.MoveToFront(element)
	entry := element.Value.(*lruEntry[Key, Value])
	return entry.value, true
}

// Put adds the key with value, or replaces the value when the key already
// exists; both paths refresh the key's recency. A Put beyond capacity
// evicts the least-recently-used entry first.
func (recent *LRU[Key, Value]) Put(key Key, value Value) {
	if element, present := recent.entries[key]; present {
		element.Value.(*lruEntry[Key, Value]).value = value
		recent.recent.MoveToFront(element)
		return
	}
	recent.entries[key] = recent.recent.PushFront(&lruEntry[Key, Value]{key: key, value: value})
	if len(recent.entries) > recent.capacity {
		// Overflow implies a back element holding only *lruEntry values, so Back and the assertion below stay safe.
		back := recent.recent.Back()
		recent.recent.Remove(back)
		delete(recent.entries, back.Value.(*lruEntry[Key, Value]).key)
	}
}

// Len counts the entries currently held, never more than the capacity.
func (recent *LRU[Key, Value]) Len() int {
	return len(recent.entries)
}
