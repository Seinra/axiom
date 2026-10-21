# opt — a value that may be absent

**Version:** Axiom on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`), `go test` green.
> API source: `go doc ./opt` ([opt/opt.go](../../opt/opt.go)). This guide summarizes; the doc comments are the reference.

## Concept

`opt.Option[T]` holds either a present value (`Some`) or nothing (`None`).
Absence here is a normal outcome — a lookup with no match, an unset field —
never an error. If you need a *cause* for the missing value, reach for
`result.Result` instead.

**When to use the stdlib directly:** a one-off `(T, bool)` return, a nilable
pointer, or a plain zero value reads clearer when you never chain or transform
the absence. Reach for `opt` when absence flows through several steps and you
want the plumbing — map, chain, filter, fall back — without an `if` at each one.

## Steps

1. Wrap: `Some(value)` holds a value; `None[T]()` holds absence.
2. Ask or read: `IsSome` checks, `ValueOr` reads with an explicit fallback,
   `ValueOrElse` reads with a lazy fallback (runs only when absent).
3. Transform: `MapValue` converts the held value and carries absence
   through untouched.
4. Chain: `AndThen` feeds the held value into the next `Option`;
   `Flatten` collapses one nesting level; `OrElse` falls back to another
   `Option`; `Filter` keeps the value only when it fits.
5. Bridge: `ToResult` converts absence into a `result.Result` failure when the
   caller needs a cause.

> [!NOTE]
> Pass a non-nil cause to `ToResult`: a nil failure flows into `result.Fail`,
> whose nil guard replaces it with a generic descriptive error. Name the
> absence at the moment you bridge.

## Which function?

Resumed from the package table — see `go doc ./opt` for the full version.

| When you want to ... | Reach for ... |
|---|---|
| Hold a present value | `Some` |
| Express absence | `None` |
| Ask whether a value is held | `IsSome` |
| Read with an explicit fallback | `ValueOr` |
| Read with a lazy fallback | `ValueOrElse` |
| Transform the held value | `MapValue` |
| Chain to the next Option | `AndThen` |
| Collapse one nesting level | `Flatten` |
| Fall back to another Option | `OrElse` |
| Keep the value only when it fits | `Filter` |
| Convert absence into a failure | `ToResult` (see `result`) |

## Example

Runnable — adapted from `ExampleSome` in [opt/example_test.go](../../opt/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"
	"strings"

	"github.com/seinra/axiom/opt"
)

func main() {
	present := opt.Some("gopher")
	fmt.Println(present.ValueOr("nobody"))
	fmt.Println(opt.None[string]().ValueOr("nobody"))
	shouted := present.MapValue(strings.ToUpper)
	fmt.Println(shouted.ValueOr("nobody"))
}
```

Expected output:

```text
gopher
nobody
GOPHER
```

## Presence and absence

`Option` is Axiom's explicit presence wrapper: `Some` means present, `None`
means absent, and every combinator carries absence through without running your
function. Contrast with the neighbors: `result.Result` carries a failure *cause*
alongside absence, `either.Either` always holds one of two real values (no
absence at all), and the `(T, bool)` OrNull shape used by `slice` and `iterx`
reports presence without a wrapper. Pick `opt` when absence must be chained or
transformed; pick the bare `(T, bool)` when a single check at the call site is enough.

## Next step

- Neighbor: [result](result.md) — bridge with `ToResult` when absence needs a cause.
- [Cookbook](../cookbook.md): the recipes layer applies these combinators to multi-step lookups
  (parse → validate → fall back) end to end.
