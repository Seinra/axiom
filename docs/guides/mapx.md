# mapx — reshape builtin maps without a wrapper

> Version: Axiom on Go 1.27 · Prereqs: Go toolchain, generics basics (`[T]`), `go test` green.
> API source: `go doc ./mapx` ([mapx/mapx.go](../../mapx/mapx.go)). This guide summarizes; the doc comments are the reference.

## Concept

`mapx` adds small typed helpers over builtin maps (`KeyBy`, `Filter`,
`MapValues`, `Merge`, `Pick`, `Omit`, `Invert`, …). It is deliberately
function-only: no wrapper type, no methods. Catalyst and the stdlib pass
builtin maps everywhere (tallies, envelopes), and interop beats chaining here.
All helpers are pure — inputs are never mutated, and nil inputs yield empty
non-nil maps.

> [!NOTE]
> Collisions resolve last-wins: `KeyBy` keeps the last element per key and
> `MapKeys` keeps the last write. When every key must survive a flip, reach for
> `Invert` — it groups keys by shared value so nothing is silently dropped.

> [!NOTE]
> Helpers are pure and nil-tolerant: inputs are never mutated, and nil inputs
> yield empty non-nil maps you can read immediately.

**When to use the stdlib directly:** for `Keys`, `Values`, `Collect`, and
`Clone`, reach for the stdlib `maps` package — this package does not repeat
them. Reach for `mapx` for the reshape shapes: indexing a slice by key,
filtering entries, remapping keys or values, merging with collision handling,
and picking/omitting/inverting.

## Steps

1. Index: `KeyBy` keys a slice (last wins on collision); `GroupByMapped`
   groups transformed values per key.
2. Narrow: `Filter` / `FilterKeys` / `FilterValues` keep matching entries;
   `Pick` keeps only listed keys; `Omit` drops listed keys (its complement).
3. Remap: `MapValues` transforms values keeping keys; `MapKeys` remaps keys
   (last write wins on collision — never a silent merge).
4. Combine and flip: `Merge` unions two maps with your `combine` resolving
   collisions; `Invert` groups keys by shared value so nothing is dropped.
5. Read: `ValueOr` reads with an explicit fallback (the `getOrDefault` shape).

## Which function?

Resumed from the package table — see `go doc ./mapx` for the full version.

| When you want to ... | Reach for ... |
|---|---|
| Index a slice by key | `KeyBy` |
| Group values by key | `GroupByMapped` |
| Keep only matching entries | `Filter` |
| Keep only matching keys | `FilterKeys` |
| Keep only matching values | `FilterValues` |
| Transform every value | `MapValues` |
| Remap every key | `MapKeys` |
| Combine two maps | `Merge` |
| Keep only listed keys | `Pick` |
| Drop listed keys | `Omit` |
| Flip values into key groups | `Invert` |
| Read with an explicit fallback | `ValueOr` |

## Example

Runnable — adapted from `ExampleKeyBy` in [mapx/example_test.go](../../mapx/example_test.go)
(same behavior; the record type is defined locally so the snippet stands
alone). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/mapx"
)

type lesson struct {
	slug  string
	level int
}

func main() {
	lessons := []lesson{{slug: "a", level: 1}, {slug: "b", level: 2}}
	bySlug := mapx.KeyBy(lessons, func(entry lesson) string { return entry.slug })
	fmt.Println(mapx.ValueOr(bySlug, "a", lesson{}).level)
	fmt.Println(mapx.ValueOr(bySlug, "missing", lesson{level: -1}).level)
	titles := mapx.MapValues(bySlug, func(entry lesson) int { return entry.level * 10 })
	fmt.Println(mapx.ValueOr(titles, "b", -1))
}
```

Expected output:

```text
1
-1
20
```

## Presence and absence

Map reads answer presence with the comma-ok idiom, and `ValueOr` folds it into
one call: present returns the held value, absent returns your fallback — no
wrapper, no cause. That matches the README OrNull ladder: a loose value that
may be missing stays a direct read; reach for `opt.Option` when the absent
branch must be chained or transformed, and `result.Result` when the caller
needs to know *why* a key is missing.

## Next step

- Neighbor: [slice](slice.md) — `GroupBySlice` / `Tally` produce the maps `mapx` reshapes.
- [Cookbook](../cookbook.md): the recipes layer indexes domain slices with `KeyBy`, then `Pick`/`Omit` shapes the response envelope.
