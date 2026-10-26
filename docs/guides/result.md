# result — success or failure, with the cause attached

**Version:** Axiom v0.2.0 on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`), `go test` green.
> API source: `go doc ./result` ([result/result.go](../../result/result.go)). This guide summarizes; the doc comments are the reference.

## Concept

`result.Result[T]` holds either a success (`Ok`) or a failure (`Fail`, always
with a non-nil cause — a nil failure is replaced with a descriptive error).
Transform the success side while the failure carries through untouched; split
both paths explicitly with `Unwrap`.

> [!WARNING]
> Never construct the zero value — build with `Ok` or `Fail`. `Unwrap` on a zero
> `Result` reports an error instead of a nil failure, so a bypassed constructor
> surfaces as a failure you did not write.

**When to use the stdlib directly:** plain `(T, error)` returns at function
boundaries are idiomatic Go and stay that way. Reach for `result` when a success
travels through several transforms (`MapValue`/`AndThen`), recovers (`OrElse`), or
crosses a boundary where panicking code must become a value (`Try`).

## Steps

1. Wrap: `Ok(value)` holds success; `Fail(err)` holds failure.
2. Ask or read: `IsOk` checks; `Unwrap` splits both paths; `UnwrapOr` /
   `UnwrapOrElse` read with a fallback; `Must` reads a *known* success
   (tests, static seeds — never use it for control flow).

> [!WARNING]
> `Must` panics on failure. Use it only for values known at construction time
> (tests, static seeds) — never to steer branches.
3. Transform: `MapValue` converts the success side; `MapFailure`
   converts the failure side; failures carry through untouched either way.
4. Chain and recover: `AndThen` feeds success into the next `Result`;
   `Flatten` collapses one nesting level; `OrElse` recovers from failure.
5. Boundaries: `Try` converts panicking code (parsers, callbacks, legacy
   calls) into a `Result` at the boundary — never for steering branches.

> [!NOTE]
> `Try` turns a panic into `Fail`, and a nil attempt panics when called —
> recovered like any other panic. Use it for boundary conversion only.

## Which function?

Resumed from the package table — see `go doc ./result` for the full version.

| When you want to ... | Reach for ... |
|---|---|
| Wrap a success | `Ok` |
| Wrap a failure | `Fail` |
| Ask which side is held | `IsOk` |
| Split both paths explicitly | `Unwrap` |
| Read a known success (tests, seeds) | `Must` |
| Read with an explicit fallback | `UnwrapOr` |
| Read with a lazy fallback | `UnwrapOrElse` |
| Transform the success side | `MapValue` |
| Transform the failure side | `MapFailure` |
| Chain to the next Result | `AndThen` |
| Collapse one nesting level | `Flatten` |
| Recover from a failure | `OrElse` |
| Convert panicking code at a boundary | `Try` |

## Example

Runnable — adapted from `ExampleOk` in [result/example_test.go](../../result/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"errors"
	"fmt"

	"github.com/seinra/axiom/result"
)

func main() {
	doubled := result.Ok(21).MapValue(func(value int) int { return value * 2 })
	success, _ := doubled.Unwrap()
	fmt.Println(success)
	failed := result.Fail[int](errors.New("boom"))
	fmt.Println(result.UnwrapOr(failed, -1))
}
```

Expected output:

```text
42
-1
```

## Presence and absence

`Result` pairs presence with a cause: `Ok` holds the value, `Fail` explains why
there is none — and the explanation is never nil. Contrast with the neighbors:
`opt.Option` models absence *without* a cause, `either.Either` holds one of two
plain values with no success/failure meaning, and the `(T, bool)` OrNull shape
reports presence with no explanation at all. Pick `result` when the caller must
know *why* there is no value; pick `opt` when "absent" is the whole story.

## Next step

- Neighbor: [opt](opt.md) — absence without a cause; convert with `ToResult` only at the point a cause becomes known.
- [Cookbook](../cookbook.md): the recipes layer chains `AndThen`/`OrElse` across parse → validate → persist pipelines.
