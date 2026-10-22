// Package stack provides a generic LIFO Stack[Element] over a slice.
//
// The standard library offers container/list (untyped, doubly linked) and no
// generic stack: LIFO over a slice is a one-liner, but every caller re-derives
// empty handling. Stack pins it once: empty Pop and Peek yield (zero, false)
// instead of panicking, the zero value is ready without New, and Push/Pop run
// in amortized O(1) on the tail of the backing slice.
//
// Which method?
//
// | When you want to ...                  | Reach for ... |
// |---------------------------------------|---------------|
// | Build an empty stack                  | New           |
// | Add an element on top                 | Push          |
// | Remove and return the top element     | Pop           |
// | Read the top element without removing | Peek          |
// | Count held elements                   | Len           |
// | Ask whether the stack holds nothing   | IsEmpty       |
//
// The (element, held) return on Pop and Peek is this package's OrNull: held
// reports presence and element is the zero value when absent. An empty stack
// is a normal outcome here, never an error and never a panic; for failures
// with causes, see the result package, and for a wrapped maybe, see opt.
//
// The zero Stack is ready: declare var container Stack[int] and Push directly.
// Methods mutate only the receiver's own backing — there is no aliasing
// surface at all, since constructors take no slices. Do not copy a Stack
// value after use; always use the pointer from New (or the address of a zero
// Stack), because the slice header shares its backing array.
package stack

// Stack holds elements in last-in, first-out order.
// Do not copy a Stack value after use; keep and share the pointer from New.
type Stack[Element any] struct {
	elements []Element
}

// New returns an empty stack ready to Push.
// Do not copy the returned pointer's value; share the pointer itself.
func New[Element any]() *Stack[Element] {
	return &Stack[Element]{}
}

// Push adds element on top of the stack.
func (current *Stack[Element]) Push(element Element) {
	current.elements = append(current.elements, element)
}

// Pop removes the top element and returns it with held true.
// An empty stack yields (zero, false): never an error, never a panic.
// The vacated slot is zeroed so a popped element stays collectable.
// Draining retains backing capacity for reuse; use a fresh Stack to release memory.
func (current *Stack[Element]) Pop() (Element, bool) {
	if len(current.elements) == 0 {
		var empty Element
		return empty, false
	}
	top := current.elements[len(current.elements)-1]
	var empty Element
	current.elements[len(current.elements)-1] = empty
	current.elements = current.elements[:len(current.elements)-1]
	return top, true
}

// Peek returns the top element with held true, leaving the stack untouched.
// An empty stack yields (zero, false): never an error, never a panic.
func (current *Stack[Element]) Peek() (Element, bool) {
	if len(current.elements) == 0 {
		var empty Element
		return empty, false
	}
	return current.elements[len(current.elements)-1], true
}

// Len counts the elements currently held.
func (current *Stack[Element]) Len() int {
	return len(current.elements)
}

// IsEmpty reports whether the stack holds no elements.
func (current *Stack[Element]) IsEmpty() bool {
	return len(current.elements) == 0
}
