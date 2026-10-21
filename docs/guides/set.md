# set — distinct values to combine and compare

**Version:** Axiom on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`), `go test` green.
> API source: `go doc ./set` ([set/set.go](../../set/set.go)). This guide summarizes; the doc comments are the reference.

## Concept

`set.Set[Member]` holds distinct values. Seed with `Of`, start empty with
`New` (the zero value is also an empty set, ready to use). `Add`, `Remove`,
and `Clear` mutate through a pointer; queries (`Has`, `Len`, `IsEmpty`,
`Equal`) and combinations (`Union`, `Intersection`, `Difference`,
`SymmetricDifference`, `MapValue`) never mutate. `Members` snapshots come back
unordered — assert membership, never order.

> [!WARNING]
> `Members` snapshots come back in unspecified order. Assert membership, never
> order — any test or render that depends on position is broken by definition.

> [!NOTE]
> `Add`, `Remove`, and `Clear` mutate through a pointer; queries and
> combinations (`Union`, `Intersection`, `Difference`, `MapValue`, …) never
> mutate. The zero value is an empty set ready to use.

**When to use the stdlib directly:** a one-off `map[T]struct{}` with inline
add/check/delete is fine for a single function. Reach for `set` when membership
is combined (`Union`/`Intersection`/`Difference`), compared (`Equal`/
`IsSubsetOf`/`IsSupersetOf`/`Overlaps`/`IsDisjoint`), or passed around as a
typed value.

## Steps

1. Start: `New[T]()` for empty, `Of(a, b, c)` seeded (duplicates dropped).
2. Mutate: `Add` holds a member (re-adding changes nothing); `Remove` drops
   one (absent changes nothing); `Clear` empties.
3. Ask: `Has` for membership, `Len` / `IsEmpty` for size.
4. Combine: `Union` (either side), `Intersection` (both sides), `Difference`
   (current minus other), `SymmetricDifference` (exactly one side).
5. Compare and snapshot: `Equal`, `IsSubsetOf` / `IsSupersetOf`, `Overlaps`
   (shares any member — iterates the smaller side), `IsDisjoint` (its
   complement); `Members` snapshots; `MapValue` transforms every member.

## Which function?

Resumed from the package table — see `go doc ./set` for the full version.

| When you want to ... | Reach for ... |
|---|---|
| Start empty / start with values | `New` / `Of` |
| Add / drop / reset members | `Add` / `Remove` / `Clear` |
| Ask membership or size | `Has` / `Len` / `IsEmpty` |
| Merge holdings | `Union` / `Intersection` / `Difference` / `SymmetricDifference` |
| Compare holdings | `Equal` / `IsSubsetOf` / `IsSupersetOf` / `Overlaps` |
| Ask whether holdings share nothing | `IsDisjoint` |
| Snapshot or transform members | `Members` / `MapValue` |

## Example

Runnable — adapted from `ExampleOf` in [set/example_test.go](../../set/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/set"
)

func main() {
	roster := set.Of("ada", "grace")
	fmt.Println(roster.Has("ada"))
	fmt.Println(roster.Len())
	combined := roster.Union(set.Of("grace", "hedy"))
	fmt.Println(combined.Len())
	fmt.Println(combined.Has("hedy"))
}
```

Expected output:

```text
true
2
3
true
```

## Presence and absence

Membership is the presence question here, and it answers with a plain `bool`:
`Has` reports held or not, `IsEmpty` reports none held, `IsDisjoint` reports
nothing shared. There is no wrapper and no cause — absence of a member is just
`false`. Contrast with the neighbors: `opt.Option` wraps a single maybe-present
value for chaining, `result.Result` attaches a failure cause, and the `(T, bool)`
OrNull shape reports presence of a *found* value. Pick `set` when the question
is "which of these are held", not "is this one value present".

## Next step

- Neighbor: [slice](slice.md) — `Distinct` drops later duplicates without keeping a set around; [mapx](mapx.md) for the tally and index shapes sets often feed.
- [Cookbook](../cookbook.md): the recipes layer builds permission and tag sets (grant → union, revoke → difference, gate → overlap check).
