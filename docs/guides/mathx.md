# mathx — primes and divisors over plain ints

**Version:** Axiom on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`) — or [foundations](foundations.md) first.

> API source: `go doc ./mathx` ([mathx/mathx.go](../../mathx/mathx.go)). This guide summarizes; the doc comments are the reference.

## Concept

`mathx` answers small number-theory questions over plain `int`: is it prime,
what divides both, what multiple do they share. The standard library offers
primality only via `big.Int.ProbablyPrime` (big-number machinery for a
yes-or-no question) and GCD only for big integers — there is no scalar
version of either, and `slices` has no number theory at all. These are
deterministic, pure functions: same inputs, same outputs, nothing to mutate.

The lesson is Euclid's: remainders strictly decrease, so the algorithm
always terminates, and the division in `Lcm` runs before the multiplication,
so pairs whose naive product overflows still fit when the true multiple
does — the same divide-before-multiply discipline as overflow-safe
arithmetic elsewhere.

## Steps

1. Test primality: `mathx.IsPrime(7)` reports true; anything below 2
   (negatives, zero, one) reports false — there is no divisor story to tell.
2. Take the common divisor: `mathx.Gcd(12, 18)` is `6`; the result is always
   non-negative, and `Gcd(0, 0)` is `0` by convention.
3. Take the common multiple: `mathx.Lcm(4, 6)` is `12`; either input zero
   yields zero instead of panicking.

## Which function?

Resumed from the package docs — see `go doc ./mathx` for the full version.

| When you want to ... | Reach for ... |
| --- | --- |
| Ask whether an integer is prime | `IsPrime` |
| The greatest divisor shared by two integers | `Gcd` |
| The smallest multiple shared by two integers | `Lcm` |

## Example

Runnable — adapted from `ExampleIsPrime`, `ExampleGcd`, and `ExampleLcm` in
[mathx/example_test.go](../../mathx/example_test.go) (same behavior). Save
as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/mathx"
)

func main() {
	fmt.Println(mathx.IsPrime(7))
	fmt.Println(mathx.IsPrime(9))
	fmt.Println(mathx.Gcd(12, 18))
	fmt.Println(mathx.Lcm(4, 6))
}
```

Expected output:

```text
true
false
6
12
```

## Presence and absence

Every `mathx` outcome is total: a prime test answers true or false, and
`Gcd`/`Lcm` always return a value with a documented zero convention — no
`bool` to check, no error to handle. When absence itself carries meaning,
step sideways to [opt](opt.md); when a failure needs a cause, step up to
[result](result.md).

## Next step

- Twin: [strings](strings.md) — the same "no stdlib home" story for text: runes, never bytes.
- [Cookbook](../cookbook.md): sieve patterns that generate primes with `Range`.
