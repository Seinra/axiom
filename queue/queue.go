// Package queue provides a generic FIFO Queue[Element] with amortized O(1)
// Dequeue over a slice plus head offset.
//
// The standard library offers container/list (untyped, doubly linked) and no
// generic queue: FIFO over a slice is a few lines, but every caller re-derives
// empty handling and head bookkeeping. Queue pins both once: empty Dequeue
// and Peek yield (zero, false) instead of panicking, the zero value is ready
// without New, and order is preserved across interleaved Enqueue/Dequeue.
//
// Which method?
//
// | When you want to ...                  | Reach for ... |
// |---------------------------------------|---------------|
// | Build an empty queue                  | New           |
// | Add an element at the back            | Enqueue       |
// | Remove and return the front element   | Dequeue       |
// | Read the front element without removing | Peek        |
// | Count held elements                   | Len           |
// | Ask whether the queue holds nothing   | IsEmpty       |
//
// The (element, held) return on Dequeue and Peek is this package's OrNull:
// held reports presence and element is the zero value when absent. An empty
// queue is a normal outcome here, never an error and never a panic; for
// failures with causes, see the result package, and for a wrapped maybe,
// see opt.
//
// The zero Queue is ready: declare var container Queue[int] and Enqueue
// directly. Methods mutate only the receiver's own backing — there is no
// aliasing surface at all, since constructors take no slices. Do not copy a
// Queue value after use; always use the pointer from New (or the address of
// a zero Queue), because the slice header shares its backing array.
//
// Compaction: Dequeue advances a head offset instead of shifting on every
// call, so it stays amortized O(1). The consumed prefix grows until either
// the queue drains — then the backing is released and the head resets to
// zero — or the prefix reaches 64 entries and at least half the backing,
// at which point the live suffix slides to the front. Tests pin order,
// never timing.
package queue

// compactThreshold is the consumed-prefix length that, together with holding
// at least half the backing, triggers a slide of the live suffix.
// The 64-entry floor is a heuristic: slides stay rare on small queues while order stays pinned by tests, never timing.
const compactThreshold = 64

// Queue holds elements in first-in, first-out order.
// Do not copy a Queue value after use; keep and share the pointer from New.
type Queue[Element any] struct {
	elements []Element
	head     int
}

// New returns an empty queue ready to Enqueue.
// Do not copy the returned pointer's value; share the pointer itself.
func New[Element any]() *Queue[Element] {
	return &Queue[Element]{}
}

// Enqueue adds element at the back of the queue.
func (current *Queue[Element]) Enqueue(element Element) {
	current.elements = append(current.elements, element)
}

// Dequeue removes the front element and returns it with held true.
// An empty queue yields (zero, false): never an error, never a panic.
// The vacated slot is zeroed so a dequeued element stays collectable.
func (current *Queue[Element]) Dequeue() (Element, bool) {
	if current.head >= len(current.elements) {
		var empty Element
		return empty, false
	}
	front := current.elements[current.head]
	var empty Element
	current.elements[current.head] = empty
	current.head++
	current.compact()
	return front, true
}

// Peek returns the front element with held true, leaving the queue untouched.
// An empty queue yields (zero, false): never an error, never a panic.
func (current *Queue[Element]) Peek() (Element, bool) {
	if current.head >= len(current.elements) {
		var empty Element
		return empty, false
	}
	return current.elements[current.head], true
}

// Len counts the elements currently held.
func (current *Queue[Element]) Len() int {
	return len(current.elements) - current.head
}

// IsEmpty reports whether the queue holds no elements.
func (current *Queue[Element]) IsEmpty() bool {
	return current.head >= len(current.elements)
}

// compact reclaims the consumed prefix per the policy in the package doc:
// release the backing when drained, slide the live suffix to the front once
// the prefix is both long (compactThreshold) and at least half the backing.
func (current *Queue[Element]) compact() {
	if current.head >= len(current.elements) {
		current.elements = nil
		current.head = 0
		return
	}
	if current.head >= compactThreshold && current.head*2 >= len(current.elements) {
		remaining := copy(current.elements, current.elements[current.head:])
		var empty Element
		for index := remaining; index < len(current.elements); index++ {
			current.elements[index] = empty
		}
		current.elements = current.elements[:remaining]
		current.head = 0
	}
}
