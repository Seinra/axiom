# stack — last in, first out, with empty as a normal outcome

**Version:** Axiom v0.2.0 on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`) — or [foundations](foundations.md) first.

> API source: `go doc ./stack` ([stack/stack.go](../../stack/stack.go)). This guide summarizes; the doc comments are the reference.

## Concept

`stack.Stack[Element]` holds a LIFO pile over a slice: undo histories,
depth-first frontiers, nested scopes. The standard library offers
`container/list` (untyped, doubly linked) and no generic stack — LIFO over a
slice is a one-liner, but every caller re-derives empty handling. `Stack`
pins it once: empty `Pop` and `Peek` yield `(zero, false)` instead of
panicking, the zero value is ready without `New`, and `Push`/`Pop` run in
amortized O(1) on the tail of the backing slice.

The `(element, held)` return on `Pop` and `Peek` is this package's OrNull:
`held` reports presence and `element` is the zero value when absent. An empty
stack is a normal outcome here, never an error and never a panic; for failures
with causes, see the [result](result.md) package, and for a wrapped maybe,
see [opt](opt.md).

## Steps

1. Build: `stack.New[string]()` — or declare `var container stack.Stack[int]` and use it directly; the zero value is ready.
2. Push work on top: `container.Push("parse")`.
3. Peek without disturbing: `top, held := container.Peek()` — `held` is false on empty, never a panic.
4. Pop from the top down: `next, held := container.Pop()` until `IsEmpty()`.
5. Share the pointer from `New` (or the address of a zero `Stack`); never copy a `Stack` value after use, because the slice header shares its backing array. Draining retains backing capacity for reuse — use a fresh `Stack` to release memory.

## Which method?

Resumed from the package table — see `go doc ./stack` for the full version.

| When you want to ... | Reach for ... |
| --- | --- |
| Build an empty stack | `New` |
| Add an element on top | `Push` |
| Remove and return the top element | `Pop` |
| Read the top element without removing | `Peek` |
| Count held elements | `Len` |
| Ask whether the stack holds nothing | `IsEmpty` |

## Example

Runnable — adapted from `ExampleNew` in [stack/example_test.go](../../stack/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/stack"
)

func main() {
	container := stack.New[string]()
	container.Push("first")
	container.Push("top")
	top, _ := container.Peek()
	fmt.Println(top)
	fmt.Println(container.Len())
	for !container.IsEmpty() {
		next, _ := container.Pop()
		fmt.Println(next)
	}
}
```

Expected output:

```text
top
2
top
first
```

## Presence and absence

`Stack` reports absence as `(T, false)` at the call site — no wrapper, no
error. It pairs with its twin: [queue](queue.md) is the same contract in FIFO
order (LIFO undoes, FIFO serves). When absence needs a cause, step up to
[result](result.md); when it needs a wrapper, step sideways to
[opt](opt.md).

## Next step

- Twin: [queue](queue.md) — same OrNull contract, front-to-back order.
- [Cookbook](../cookbook.md): work-pile patterns that drain a `Stack` to empty.
