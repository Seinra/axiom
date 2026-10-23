# orderedmap — insertion order you can look up, plus a recency cache

**Version:** Axiom on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`) — or [foundations](foundations.md) first.

> API source: `go doc ./orderedmap` ([orderedmap/orderedmap.go](../../orderedmap/orderedmap.go)). This guide summarizes; the doc comments are the reference.

## Concept

`orderedmap.Map[Key, Value]` is an insertion-ordered map: key lookup plus
stable first-insertion order. The standard library offers the unordered map
(iteration order is unspecified and must not be relied upon) and
`container/list` (untyped, no key lookup) — the ordered map is a recurring
shape every caller re-derives with a backing map plus a keys slice. `Map`
pins it once: `Put` and `Get` behave like the builtin map, while `Keys` and
`Values` replay first-insertion order, and the zero value is ready without
`New`. `LRU` pins the caching lesson on top: a fixed-capacity holder where
every hit refreshes recency and entries beyond capacity evict
least-recently-used first.

The `(value, found)` return on `Get` is this package's OrNull: `found`
reports presence and `value` is the zero value when absent. A missing key is a
normal outcome here, never an error and never a panic; for failures with
causes, see the [result](result.md) package, and for a wrapped maybe, see
[opt](opt.md). The only panic in this package is `NewLRU` with a capacity
below 1: capacity is fixed at construction, so a non-positive one is a
programmer error, never a runtime condition.

## Steps

1. Build: `orderedmap.New[string, int]()` — or declare `var ordered orderedmap.Map[string, int]` and `Put` directly; the zero value is ready (lazy init on first use). `NewLRU` is the only way into a cache — capacity has no useful zero.
2. Put and read like the builtin map: `ordered.Put("beta", 2)`; `value, found := ordered.Get("beta")` — a miss yields `(zero, false)`, recency untouched on an LRU miss.
3. Upsert keeps the key's first-insertion position while replacing its value; deleting a missing key is a no-op; re-adding a deleted key takes the back.
4. Replay in order: `Keys()` and `Values()` return fresh slices — mutating them never disturbs the map.
5. Refresh recency on purpose: every LRU `Get` hit and `Put` moves the key to the front; a `Put` beyond capacity evicts the back first.
6. Share the pointer from `New`/`NewLRU` (or the address of a zero `Map`); never copy a `Map` or `LRU` value after use. This package is not safe for concurrent use; [flow and task](flow.md) own concurrency.

## Which type?

Resumed from the package table — see `go doc ./orderedmap` for the full version.

| When you want to ... | Reach for ... |
| --- | --- |
| Remember insertion order alongside lookup | `Map` |
| Bound memory with least-recently-used eviction | `LRU` |
| Build an empty ordered map | `New` |
| Build a bounded recency cache | `NewLRU` |
| Add or replace an entry | `Put` |
| Read an entry (LRU: refreshes recency) | `Get` |
| Remove an entry | `Delete` (Map) |
| Count held entries | `Len` |
| Ask whether nothing is held | `IsEmpty` (Map) |
| Replay keys or values in insertion order | `Keys`, `Values` |

## Example

Runnable — adapted from `ExampleNew` and `ExampleNewLRU` in
[orderedmap/example_test.go](../../orderedmap/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/orderedmap"
)

func main() {
	ordered := orderedmap.New[string, int]()
	ordered.Put("beta", 2)
	ordered.Put("alpha", 1)
	ordered.Put("beta", 20)
	fmt.Println(ordered.Keys())
	fmt.Println(ordered.Values())

	recent := orderedmap.NewLRU[string, int](2)
	recent.Put("alpha", 1)
	recent.Put("beta", 2)
	_, _ = recent.Get("alpha")
	recent.Put("gamma", 3)
	_, betaHeld := recent.Get("beta")
	alphaValue, _ := recent.Get("alpha")
	fmt.Println(betaHeld)
	fmt.Println(alphaValue)
	fmt.Println(recent.Len())
}
```

Expected output:

```text
[beta alpha]
[20 1]
false
1
2
```

## Presence and absence

`Map` and `LRU` report absence as `(T, false)` at the call site — no wrapper,
no error. When pairs must become a map, bridge with [mapx](mapx.md); when
absence needs a cause, step up to [result](result.md); when it needs a
wrapper, step sideways to [opt](opt.md).

## Next step

- Bridge: [mapx](mapx.md) — pairs to maps and back, the ordered map's neighbor.
- [Cookbook](../cookbook.md): recency-cache patterns that read hot keys first.
