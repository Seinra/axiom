# queue — first in, first out, with empty as a normal outcome

**Version:** Axiom on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`) — or [foundations](foundations.md) first.

> API source: `go doc ./queue` ([queue/queue.go](../../queue/queue.go)). This guide summarizes; the doc comments are the reference.

## Concept

`queue.Queue[Element]` holds a FIFO line over a slice plus a head offset:
work items, breadth-first frontiers, handoffs between stages. The standard
library offers `container/list` (untyped, doubly linked) and no generic
queue — FIFO over a slice is a few lines, but every caller re-derives empty
handling and head bookkeeping. `Queue` pins both once: empty `Dequeue` and
`Peek` yield `(zero, false)` instead of panicking, the zero value is ready
without `New`, and order survives interleaved `Enqueue`/`Dequeue`.

The `(element, held)` return on `Dequeue` and `Peek` is this package's OrNull:
`held` reports presence and `element` is the zero value when absent. An empty
queue is a normal outcome here, never an error and never a panic; for failures
with causes, see the [result](result.md) package, and for a wrapped maybe,
see [opt](opt.md).

`Dequeue` advances a head offset instead of shifting on every call, so it
stays amortized O(1). The consumed prefix grows until either the queue drains
— then the backing is released and the head resets — or the prefix reaches
64 entries and at least half the backing, at which point the live suffix
slides forward. Tests pin order, never timing.

## Steps

1. Build: `queue.New[string]()` — or declare `var line queue.Queue[int]` and use it directly; the zero value is ready.
2. Enqueue at the back: `line.Enqueue("ticket")`.
3. Peek without disturbing: `front, held := line.Peek()` — `held` is false on empty, never a panic.
4. Dequeue from the front forward: `next, held := line.Dequeue()` until `IsEmpty()`.
5. Share the pointer from `New` (or the address of a zero `Queue`); never copy a `Queue` value after use, because the slice header shares its backing array.

## Which method?

Resumed from the package table — see `go doc ./queue` for the full version.

| When you want to ... | Reach for ... |
| --- | --- |
| Build an empty queue | `New` |
| Add an element at the back | `Enqueue` |
| Remove and return the front element | `Dequeue` |
| Read the front element without removing | `Peek` |
| Count held elements | `Len` |
| Ask whether the queue holds nothing | `IsEmpty` |

## Example

Runnable — adapted from `ExampleNew` in [queue/example_test.go](../../queue/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/queue"
)

func main() {
	line := queue.New[string]()
	line.Enqueue("first")
	line.Enqueue("back")
	front, _ := line.Peek()
	fmt.Println(front)
	fmt.Println(line.Len())
	for !line.IsEmpty() {
		next, _ := line.Dequeue()
		fmt.Println(next)
	}
}
```

Expected output:

```text
first
2
first
back
```

## Presence and absence

`Queue` reports absence as `(T, false)` at the call site — no wrapper, no
error. It pairs with its twin: [stack](stack.md) is the same contract in LIFO
order (LIFO undoes, FIFO serves). When absence needs a cause, step up to
[result](result.md); when it needs a wrapper, step sideways to
[opt](opt.md).

## Next step

- Twin: [stack](stack.md) — same OrNull contract, top-down order.
- [Cookbook](../cookbook.md): stage-to-stage handoffs that drain a `Queue` to empty.
