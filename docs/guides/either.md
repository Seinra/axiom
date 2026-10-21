# either — one of two typed sides, neither favored

> Version: Axiom on Go 1.27 · Prereqs: Go toolchain, generics basics (`[T]`), `go test` green.
> API source: `go doc ./either` ([either/either.go](../../either/either.go)). This guide summarizes; the doc comments are the reference.

## Concept

`either.Either[LeftT, RightT]` holds exactly one of two typed sides: `Left` or
`Right`. Neither side is favored — `Left` is not an error and `Right` is not a
success. That neutrality is the whole point: for success-or-failure use
`result`, for maybe-present use `opt`. Reach for `either` when both outcomes
are ordinary values, e.g. a check that yields either a lock reason or granted
content. Note the zero value reads as a `Left` holding the zero `LeftT`, so
always construct with `Left` or `Right`.

**When to use the stdlib directly:** two-value returns, a small struct with a
discriminant field, or an interface cover most branching. Reach for `either`
when you want both sides typed, the split (`Match`) enforced at one point, and
per-side transforms (`MapLeft`/`MapRight`) without touching the other side.

## Steps

1. Hold: `Left(value)` keeps the first side; `Right(value)` keeps the second.
2. Ask: `IsLeft` / `IsRight` report which side is held.
3. Split: `Match` renders both sides into one outcome — exactly one branch runs.
4. Transform one side: `MapRight` converts the right side and carries a `Left`
   through untouched; `MapLeft` mirrors it.
5. Reshape: `Swap` exchanges the sides; `ToResult` bridges into a
   `result.Result` (`Right` becomes `Ok`, `Left` becomes `Fail` via your mapping).

## Which function?

Resumed from the package table — see `go doc ./either` for the full version.

| When you want to ... | Reach for ... |
|---|---|
| Hold the first side | `Left` |
| Hold the second side | `Right` |
| Ask which side is held | `IsLeft` / `IsRight` |
| Split both sides into one outcome | `Match` |
| Transform the Right side | `MapRight` |
| Transform the Left side | `MapLeft` |
| Exchange the two sides | `Swap` |
| Bridge into result.Result | `ToResult` |

## Example

Runnable — adapted from `ExampleRight` in [either/example_test.go](../../either/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/either"
)

func main() {
	render := func(reason string) string { return "locked: " + reason }
	describe := func(packTitle string) string { return "ready: " + packTitle }
	granted := either.Right[string, string]("Frozen Depths")
	fmt.Println(either.Match(granted, render, describe))
	locked := either.Left[string, string]("season pass missing")
	fmt.Println(either.Match(locked, render, describe))
}
```

Expected output:

```text
ready: Frozen Depths
locked: season pass missing
```

## Presence and absence

`Either` has no absence: it always holds a real value on exactly one side, so
there is nothing to check before using what `Match` hands you. That sets it
apart from the neighbors: `opt.Option` may hold nothing, `result.Result` may
hold a failure instead of a value, and the `(T, bool)` OrNull shape asks you to
check the `bool` first. Pick `either` when both branches carry usable data and
the caller must handle both — not when one branch means "missing".

## Next step

- Neighbor: [result](result.md) — bridge with `ToResult` when the left side turns out to mean failure after all.
- [Cookbook](../cookbook.md): the recipes layer shows `Match` rendering divergent domain outcomes (locked vs granted, draft vs published) into one response shape.
