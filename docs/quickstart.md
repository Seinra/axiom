# Ship your first typed pipeline in under 5 minutes

You point a scratch module at Axiom, run one typed pipeline, and watch it print `42` and `[2 4 6]`.

> **Version:** Axiom is unreleased — no tagged version yet; these steps track the working tree.
> **Prerequisites:** Go 1.27 in your `PATH` (`go version` reports `go1.27.x`).

## Quick path

### 1. Point a scratch module at Axiom

Axiom is a private library, so you link it with `replace`, not `go get`:

```bash
mkdir axiom-hello && cd axiom-hello
go mod init axiom-hello
```

Add these lines to your `go.mod`:

```go.mod
require github.com/seinra/axiom v0.0.0

replace github.com/seinra/axiom => ../Axiom
```

> [!NOTE]
> Point the `replace` at your local Axiom checkout — Axiom is unreleased, so
> there is no tagged version to `go get`. Adjust `../Axiom` to the path where
> you cloned it.

### 2. Run your first typed pipeline

Save this as `main.go`. It runs as-is — no edits needed:

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/result"
	axiomslice "github.com/seinra/axiom/slice"
)

func main() {
	current := result.Ok(21)
	doubled := current.MapValue(func(success int) int { return success * 2 })
	fmt.Println(result.UnwrapOr(doubled, 0))

	numbers := axiomslice.MapSlice([]int{1, 2, 3}, func(success int) int { return success * 2 })
	fmt.Println(numbers)
}
```

Then fetch dependencies and run it:

```bash
go mod tidy
go run .
```

Expected output:

```text
42
[2 4 6]
```

### 3. Verify the library

From your Axiom checkout, run:

```bash
go test ./...
```

Expect `ok` across all packages.

## Details

| Topic | Behavior |
|-------|----------|
| `result.Ok` | Wraps a success value as a `result.Result`; you build with `Ok`/`Fail`, never the zero value. |
| `MapValue` | Transforms the success side and carries failures through untouched; it declares its own type params, so calls chain left to right. |
| `result.UnwrapOr` | Reads the success side with an explicit fallback when the result holds a failure. |
| `axiomslice.MapSlice` | Applies a transform to every element, preserving order. |

## Checklist

- [ ] `go run .` prints `42` and `[2 4 6]` with no edits to `main.go`
- [ ] `go test ./...` from the Axiom checkout reports `ok` for every package
- [ ] You can name what `MapValue` does to the failure side (nothing — it carries it through)

## Next step

Pick the family that matches what you are holding and read its guide with `go doc`:

| If you are holding ... | Start with |
|---|---|
| a maybe-present value | `go doc ./opt` |
| a success-or-failure outcome | `go doc ./result` |
| two typed sides (left / right) | `go doc ./either` |
| a slice to transform, find, or fold | `go doc ./slice` |
| distinct values to combine or compare | `go doc ./set` |
| a builtin map to reshape | `go doc ./mapx` |
| a lazy `iter.Seq` pipeline | `go doc ./iterx` |
| predicates to combine (And / Or / Not) | `go doc ./valid` |
| one call to run once per key | `go doc ./singleflight` |
| objects to reuse, not rebuild | `go doc ./pool` |
| named singleton services | `go doc ./di` |
| records addressed by identity | `go doc ./repo` |
| the same over GORM | `go doc ./repogorm` |
| an Echo JSON envelope | `go doc ./httpx` |
| a bounded concurrent map or first-success race | `go doc ./flow` |

The full package map lives in [README.md](../README.md).
