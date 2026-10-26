# Foundations — the generic mechanism, from concrete to generic

**Version:** Axiom v0.2.0 on Go 1.27.
**Prereqs:** Go toolchain. No generics knowledge — this chapter teaches it.

After this chapter you read any generic signature in Axiom: what `[T any]`
declares, what a constraint promises, when the compiler fills in the types,
why some APIs spell their types out, and where the language stops
(no generic methods on interfaces — frozen decision, taught as lesson, not lore).

> API source: `go doc` on any package below; doc comments are the reference, this chapter is the bridge. Every runnable line here is executed by `TestFoundationsMechanism` in [../foundations_test.go](../foundations_test.go) — same outputs.

## Concept

Generics remove repetition, not types. Write the concrete version first;
when the same shape repeats for another type, name the varying part:

```go
func doubleInts(values []int) []int
// becomes
func doubleAll[Element any](values []Element) []Element
```

`[Element any]` declares a *type parameter*: a placeholder the caller fills
with a real type, once per call. Axiom's rule (stdlib-first): reach for a
type parameter only when the stdlib cannot say it — `slices`, `maps`, `cmp`
and `iter` own the concrete shapes, Axiom adds the missing generic ones
([result/result.go](../../result/result.go),
[slice/slice.go](../../slice/slice.go)).

## Steps

1. Declare the parameter where the type varies: `func Min[Element cmp.Ordered](elements []Element) (Element, bool)` ([slice/slice.go](../../slice/slice.go)).
2. Constrain what the body needs — nothing more: `any` accepts everything, `comparable` allows `==`/`!=` (and map keys), `cmp.Ordered` allows ordering, a custom interface lists exactly the methods used.
3. Let the compiler infer from value arguments; spell types out only where inference has nothing to read (zero-argument constructors like `opt.None[string]()`).
4. Put behavior on values, never on missing ones: read absence with the OrNull shape (`(T, bool)`), `opt.Option`, or `result.Result` — never branch on a nil map lookup you could have typed.

## Type parameters: declare once, fill per call

A type parameter list lives in brackets after the function name. Each call
fills every parameter with a real type:

```go
func NewMemoryRepository[Record any, Identity comparable](keyOf func(Record) Identity) *MemoryRepository[Record, Identity]
```

([repo/repo.go](../../repo/repo.go)). `Record any` says "any type at all";
`Identity comparable` says "only types supporting `==`/`!=`", which is what a
map key needs. Constructors declare the most parameters because they create
from nothing; methods reuse their receiver's.

## Constraints: promise only what the body uses

| Constraint | Promises | Axiom use |
| --- | --- | --- |
| `any` | nothing; accepts every type | payloads carried untouched (`Record any`) |
| `comparable` | `==`/`!=`; usable as map key | identities, set elements, map keys (`Identity comparable`) |
| `cmp.Ordered` | ordering (`<`, sorting) | `Min`/`Max` over ordered elements, `SortBy` over ordered keys ([slice/slice.go](../../slice/slice.go)) |
| custom interface | exactly the listed methods | narrow behavior contracts, never kitchen-sink |

`comparable` is a compile-time check with one runtime edge: comparing two
interface values panics if their dynamic type is uncomparable (a slice in an
`any`). Axiom keeps `comparable` at map keys and identities, where the
dynamics stay comparable by construction.

## Inference: read from the values, all or nothing

When every parameter appears in a value argument, the call site stays clean:

```go
doubled := axiomslice.MapSlice([]int{1, 2, 3}, func(n int) int { return n * 2 })
```

No type arguments written — `Element` and `Transformed` come from the slice
and the callback. Two rules, no exceptions:

- Either the compiler infers **all** parameters from the value arguments, or you write **all** of them explicitly. No partial lists.
- Zero-argument constructors give inference nothing to read, so they spell it out: `opt.None[string]()`, `opt.Some` needs no help because the value carries the type.

## Generic methods — and the one place they stop

Methods may declare their **own** type parameters, so calls chain left to right:

```go
func (current Result[T]) MapValue[Transformed any](transform func(T) Transformed) Result[Transformed]
```

([result/result.go](../../result/result.go)). `MapValue` reuses the receiver's
`T` and introduces `Transformed` for the mapping — one chain, no annotation.

The limit is interfaces: **an interface method must not declare type
parameters** — a language rule, not a style choice. So Axiom's interfaces
stay non-generic in their methods and put variation in the interface itself:

```go
type Repository[Record any, Identity comparable] interface {
	Store(candidate Record) error
	Fetch(identity Identity) (Record, error)
	// ...
}
```

([repo/repo.go](../../repo/repo.go)). Polymorphism over behavior never asks
for per-call type arguments; genericity lives in functions and concrete
types. That is frozen decision 1, taught here instead of whispered in an
architecture note.

## Zero values: each package states its rule

Go gives every type a zero value; Axiom decides per package whether it is
usable. Read the package doc — never guess:

| Package | Zero value rule |
| --- | --- |
| `result` | Never construct it: `Unwrap` on a zero `Result` reports `errZeroResult` ("construct with Ok or Fail"). Build with `Ok`/`Fail` ([result/result.go](../../result/result.go)). |
| `opt` | Zero `Option` is absent (same as `None`): `ValueOr` reads through it. |
| `orderedmap` / `stack` / `queue` | Ready without `New`: declare and use directly; only `NewLRU` demands construction, because capacity has no useful zero. |
| `task` | Not usable: `Await` on a zero `Task` panics — `Spawn` a task first. |

## Which mechanism?

| When you hold ... | Reach for ... |
| --- | --- |
| a value of unknown type to carry through | a type parameter constrained to `any` |
| values to compare or key a map by | a `comparable` parameter |
| values to order, min, or clamp | a `cmp.Ordered` parameter |
| behavior shared across many types | a function with type params, or an interface whose *methods* take none |
| absence without a cause | `opt.None[T]()` (explicit: nothing to infer from) |
| absence with a cause | `result.Fail[T](err)` |
| a computation still running | `task.Spawn` + `Await` — the join is the mechanism from [flow and task](flow.md) |

## Example

Runnable — executed as `TestFoundationsMechanism` in
[../foundations_test.go](../foundations_test.go) (same outputs). Save as
`main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/opt"
	"github.com/seinra/axiom/result"
	axiomslice "github.com/seinra/axiom/slice"
)

func main() {
	// Inference from value arguments: no type list written.
	doubled := axiomslice.MapSlice([]int{1, 2, 3}, func(n int) int { return n * 2 })
	fmt.Println(doubled)

	// Explicit where inference has nothing to read: zero-arg constructor.
	fmt.Println(opt.None[string]().ValueOr("nobody"))

	// Constraint-governed call: Min works because int is cmp.Ordered.
	smallest, found := axiomslice.Min([]int{3, 1, 2})
	fmt.Println(smallest, found)

	// Generic method with its own type params, chained left to right.
	chained := result.Ok(21).MapValue(func(value int) int { return value * 2 })
	success, _ := chained.Unwrap()
	fmt.Println(success)

	// Zero values follow each package's rule, never a guess.
	var emptyOpt opt.Option[int]
	fmt.Println(emptyOpt.ValueOr(0))
	var zeroResult result.Result[int]
	_, failure := zeroResult.Unwrap()
	fmt.Println(failure != nil)
}
```

Expected output:

```text
[2 4 6]
nobody
1 true
42
0
true
```

## Checklist

- [ ] You can name which parameter each value argument fills in a generic call.
- [ ] You can say when explicit type arguments are required (zero-argument constructors).
- [ ] You can state frozen decision 1 in one sentence (interface methods take no type params; variation lives in the interface or the function).
- [ ] You can state your package's zero-value rule before constructing it.

## Next step

- Run it: [quickstart](../quickstart.md) — one typed pipeline, five minutes.
- Apply it per family: [result](result.md), [opt](opt.md), [slice](slice.md) — every guide assumes this chapter.
- Teach the limit: frozen decision 1 above is the whole lesson — interfaces dispatch, functions generalize.
