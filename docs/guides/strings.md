# strings — words and cuts counted in runes, never bytes

**Version:** Axiom on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`) — or [foundations](foundations.md) first.

> API source: `go doc ./strings` ([strings/strings.go](../../strings/strings.go)). This guide summarizes; the doc comments are the reference.

## Concept

Byte slicing tears multibyte text apart: `"héllo"[0:2]` is not two
characters. Every `strings` helper counts runes, never bytes, so a cut or a
chunk always lands on a character boundary. When you need this package and
the standard library in one file, alias the import:

```go
import axiomstrings "github.com/seinra/axiom/strings"
```

Case converters live in `x/text` and stay out of this package on purpose —
this family splits and cuts, it never changes case.

## Steps

1. Split on whitespace: `strings.Words("  go   rust\tzig\n")` keeps
   `[go rust zig]`; blank or empty input yields an empty non-nil slice, and
   multibyte words stay whole.
2. Chunk by runes: `strings.ChunkString("héllo", 2)` gives `[hé ll o]` —
   size below one panics, matching the size policy of `Chunked` and `Take`.
3. Cut to a budget: `strings.Truncate("héllo", 2)` is `"hé"` with no suffix
   added; text within budget returns unchanged, and a negative budget
   panics. Presentation (ellipses and the like) belongs to the caller.

## Which function?

Resumed from the package table — see `go doc ./strings` for the full version.

| When you want to ... | Reach for ... |
| --- | --- |
| Split on whitespace into words | `Words` |
| Split into fixed-rune chunks | `ChunkString` |
| Cut to a rune budget, no suffix | `Truncate` |

## Example

Runnable — adapted from `ExampleWords`, `ExampleChunkString`, and
`ExampleTruncate` in [strings/example_test.go](../../strings/example_test.go)
(same behavior). Save as `main.go`, with
`replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/strings"
)

func main() {
	fmt.Println(strings.Words("  go   rust\tzig\n"))
	fmt.Println(strings.ChunkString("héllo", 2))
	fmt.Println(strings.Truncate("héllo", 2))
}
```

Expected output:

```text
[go rust zig]
[hé ll o]
hé
```

## Presence and absence

`Words` reports absence as an empty non-nil slice — iterate it directly, no
nil check. `Truncate` always returns usable text, never an error. When
absence needs a flag, step sideways to [opt](opt.md); the empty slice here
is a normal outcome, not a missing value.

## Next step

- Twin: [mathx](mathx.md) — the same "no stdlib home" story for integers: Euclid plus overflow-safe multiples.
- [Cookbook](../cookbook.md): text-shaping patterns that rejoin `Words` output.
