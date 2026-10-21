# iterx — lazy pipelines over iter.Seq

**Version:** Axiom on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`), `iter.Seq` familiarity, `go test` green.
> API source: `go doc ./iterx` ([iterx/iterx.go](../../iterx/iterx.go)). This guide summarizes; the doc comments are the reference.

## Concept

`iterx` adds small lazy adapters over `iter.Seq`, complementing the eager
helpers in `slice`. The stdlib `iter` package defines the sequence protocol on
purpose without transforms; `slices.Collect` and `maps.Collect` consume
sequences back into containers. Each adapter returns a new sequence: nothing
runs until the caller ranges or collects, early termination (`break`, `Take`)
propagates upstream, and re-ranging re-runs the whole pipeline from the source.

**When to use the stdlib directly:** `slices.Values`, `slices.Collect`,
`maps.Collect`, and `iter.Pull` cover sources and sinks. Reach for `iterx` for
the middle — transform, filter, take/drop, zip, chunk — when the pipeline
should stay lazy instead of allocating each stage.

One fail-fast exception, called out below: `ChunkedSeq` and `WindowsSeq`
validate size at call time and panic on size <= 0 (a programmer error, matching
`slices.Chunk`); valid sizes stay fully lazy.

> [!WARNING]
> `ChunkedSeq` and `WindowsSeq` panic on size <= 0 at call time — the fail-fast
> check runs before any laziness kicks in. Validate sizes up front.

> [!WARNING]
> `Repeat` and `Cycle` are infinite. Realize them only through a bounded
> consumer such as `Take` — ranging one directly never terminates.

## Steps

1. Flow: `MapSeq` transforms, `FilterSeq` keeps matches, `FlatMapSeq` /
   `FlattenSeq` map-and-join, `OnEachSeq` observes without changing flow.
2. Bound: `Take` / `Drop` keep or skip the first count; `TakeWhile` /
   `DropWhile` keep or skip while a condition holds (the boundary element is
   consumed — see the doc comment).
3. Combine: `Enumerate` pairs positions, `ZipSeq` pairs two sequences (stops
   at the shorter), `ConcatSeq` runs sequences in order.
4. Generate: `Repeat` one value and `Cycle` a slice (both infinite — realize
   only through `Take`); `Unfold` grows state into values; `Generate` calls
   until stop; `FilterMapSeq` maps and keeps some.
5. Window: `ChunkedSeq` (last may be short, mirrors `slice.Chunked`),
   `WindowsSeq` (mirrors `slice.Windows`), `IntersperseSeq` slots separators,
   `DeduplicateSeq` collapses adjacent duplicates only.

> [!NOTE]
> Ranging re-runs the whole pipeline from the source, and early termination
> (`break`, `Take`) propagates upstream. Nothing runs until you range or
> collect — a built pipeline is inert.

> [!NOTE]
> `TakeWhile` consumes its boundary element from the source without yielding
> it, and `DeduplicateSeq` collapses adjacent duplicates only — distant
> duplicates still pass through.

## Which function?

Resumed from the package table — see `go doc ./iterx` for the full version.

| When you want to ... | Reach for ... |
|---|---|
| Transform elements lazily | `MapSeq` |
| Keep matching elements lazily | `FilterSeq` |
| Map and flatten in one lazy pass | `FlatMapSeq` / `FlattenSeq` |
| Keep / skip the first count | `Take` / `Drop` |
| Keep / skip while a condition holds | `TakeWhile` / `DropWhile` |
| Pair positions / pair two sequences | `Enumerate` / `ZipSeq` |
| Observe elements without changing flow | `OnEachSeq` |
| Repeat one value / cycle a slice (bound with Take) | `Repeat` / `Cycle` |
| Collapse adjacent duplicates | `DeduplicateSeq` |
| Unfold evolving state into values | `Unfold` |
| Generate values until a stop | `Generate` |
| Map and keep only some results | `FilterMapSeq` |
| Fixed chunks, last short (slice.Chunked) | `ChunkedSeq` |
| Sliding windows (slice.Windows) | `WindowsSeq` |
| Slot a separator between elements | `IntersperseSeq` |
| Concatenate sequences in order | `ConcatSeq` |

## Example

Runnable — adapted from `ExampleMapSeq` in [iterx/example_test.go](../../iterx/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"
	"slices"

	"github.com/seinra/axiom/iterx"
)

func main() {
	doubled := iterx.MapSeq(slices.Values([]int{1, 2, 3}), func(value int) int { return value * 2 })
	fmt.Println(slices.Collect(doubled))
	evens := iterx.FilterSeq(slices.Values([]int{1, 2, 3, 4}), func(value int) bool { return value%2 == 0 })
	fmt.Println(slices.Collect(evens))
	firstTwo := iterx.Take(slices.Values([]int{1, 2, 3, 4}), 2)
	fmt.Println(slices.Collect(firstTwo))
}
```

Expected output:

```text
[2 4 6]
[2 4]
[1 2]
```

## Presence and absence

`Generate`, `FilterMapSeq`, and `Unfold` take OrNull-shaped callbacks: the
`bool` reports presence — never trust the value without it. This is the same
canon as `slice`'s `(value, bool)` returns and the README guide index: a loose
value that may be missing stays a bare `bool` check; wrap it in `opt.Option`
when absence must be chained, in `result.Result` when a cause travels with it.
Note the laziness corollary: an absent (`false`) callback simply ends or skips
within the stream — no allocation, no error path.

## Next step

- Neighbor: [slice](slice.md) — the eager counterpart; `slices.Collect` realizes any pipeline back into a slice.
- [Cookbook](../cookbook.md): the recipes layer streams paged or generated inputs (`Generate`/`Unfold` + `Take`) without staging them in memory.
