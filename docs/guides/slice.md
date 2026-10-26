# slice — eager helpers over slices, stdlib first

**Version:** Axiom v0.2.0 on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`), `go test` green.
> API source: `go doc ./slice` ([slice/slice.go](../../slice/slice.go)). This guide summarizes; the doc comments are the reference.

## Concept

`slice` adds small typed helpers (`MapSlice`, `FilterSlice`, `ReduceSlice`,
`FindSlice`, `GroupBySlice`, …) layered on the standard library (`slices`,
`maps`, `cmp`, `iter`) rather than replacing it. All helpers are pure: inputs
are never mutated.

**When to use the stdlib directly:** whenever it already names the operation —
`slices.Contains`, `slices.Sort`, `slices.Collect` and friends stay the default.
Reach for this package when a lo-style `Map`/`Filter`/`Reduce`/`Find`/`GroupBy`
reads clearer than a hand-rolled loop, and for the grouping, chunking, and
folding shapes the stdlib leaves to you.

## Steps

1. Transform and keep: `MapSlice` / `MapIndexed` convert every element;
   `FilterSlice` / `FilterIndexed` keep the matching ones (never nil).
2. Fold: `ReduceSlice` folds to one value; `Sum` adds; `Average` takes the mean.
3. Find and compare: `FindSlice`, `Min` / `Max`, `MinBy` / `MaxBy` return the
   OrNull shape — check the `bool` before trusting the value (see below).
4. Reshape: `GroupBySlice` / `Tally` bucket elements; `FlatMap` / `Flatten`
   join; `Chunked` / `Windows` cut fixed pieces (see the size callout below).
5. Combine and order: `Zip` / `Unzip` (see `Pair`), `Distinct`, `Partition`,
   `Intersperse`, `SortBy` (stable copy, input untouched), `RunningFold`.

> [!WARNING]
> `Chunked` and `Windows` panic on size <= 0 — a non-positive size is a
> programmer error, matching `slices.Chunk`. Validate sizes at the call site.

> [!NOTE]
> `MapSlice`, `FilterSlice`, and friends never return nil: empty input yields
> an empty slice, so you range the outcome directly without a nil check.

## Which function?

Resumed from the package table — see `go doc ./slice` for the full version.

| When you want to ... | Reach for ... |
|---|---|
| Transform every element | `MapSlice` / `MapIndexed` |
| Keep only matching elements | `FilterSlice` / `FilterIndexed` |
| Fold to one value | `ReduceSlice` / `Sum` |
| Find the first match | `FindSlice` |
| Smallest / largest, plain or by key | `Min` / `Max` / `MinBy` / `MaxBy` |
| Mean of numbers | `Average` |
| Bucket elements by key | `GroupBySlice` / `Tally` |
| Map each element to a slice, then join | `FlatMap` / `Flatten` |
| Fixed chunks / sliding windows | `Chunked` / `Windows` |
| Pair two slices / split pairs back | `Zip` / `Unzip` (see `Pair`) |
| Drop later duplicates | `Distinct` |
| Split into matching plus rest | `Partition` |
| Insert a separator between elements | `Intersperse` |
| Sort a copy by key, ties keep input order | `SortBy` |
| Running folds element by element | `RunningFold` |
| First element, or rest without it | `Head` / `Tail` / `Initial` |

## Example

Runnable — adapted from `ExampleMapSlice` in [slice/example_test.go](../../slice/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/slice"
)

func main() {
	doubled := slice.MapSlice([]int{1, 2, 3}, func(value int) int { return value * 2 })
	fmt.Println(doubled)
	halved := slice.FilterSlice(doubled, func(value int) bool { return value%4 == 0 })
	fmt.Println(halved)
	total := slice.ReduceSlice(doubled, 0, func(accumulation, value int) int { return accumulation + value })
	fmt.Println(total)
}
```

Expected output:

```text
[2 4 6]
[4]
12
```

## Presence and absence

This package's OrNull is the `(value, bool)` return on `FindSlice`, `Min`,
`Max`, `MinBy`, `MaxBy`, and `Average`: the `bool` reports presence and the
value is the zero value when absent. Empty input or no match is a normal
outcome here, never an error — check the flag, e.g.
`if found, ok := slice.FindSlice(xs, match); ok { … }`. This is the same canon
stated in the README guide index: a loose value that may be missing stays a
bare `(T, bool)`; wrap it in `opt.Option` when absence must be chained or
transformed, and in `result.Result` when the caller needs a failure cause.

## Next step

- Neighbor: [iterx](iterx.md) — the lazy counterpart; same pipeline shapes over `iter.Seq`, nothing runs until collected.
- [Cookbook](../cookbook.md): the recipes layer composes find → transform → fold over concrete domain slices.
