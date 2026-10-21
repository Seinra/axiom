# Axiom cookbook — three standalone recipes

Each recipe is a standalone program: copy it into an empty module with a
`replace` to your Axiom checkout, `go run` it, and compare with the
expected output.

**Version:** `github.com/seinra/axiom`, Go 1.27.
**Prereqs for every recipe:** Go 1.27, a local Axiom checkout, and a
consumer module wired with:

```go.mod
// go.mod of the consumer
require github.com/seinra/axiom v0.0.0
replace github.com/seinra/axiom => ../Axiom
```

## (a) The ladder of absence — OrNull → opt → result

You hold a loose value that may be missing (your own code: `(T, bool)`),
an absence to chain or transform (`opt.Option`), and a failure with a
cause (`result.Result`). Climb the ladder only as far as the caller needs.

```go
package main

import (
	"errors"
	"fmt"

	"github.com/seinra/axiom/opt"
	"github.com/seinra/axiom/result"
	axiomslice "github.com/seinra/axiom/slice"
)

func main() {
	lessons := []string{"generics-intro", "slices-deep", "maps-fast"}
	// Rung 1: (T, bool) OrNull at your own boundary.
	slug, found := axiomslice.FindSlice(lessons, func(s string) bool { return s == "slices-deep" })
	fmt.Println(slug, found)
	missing, found := axiomslice.FindSlice(lessons, func(s string) bool { return s == "nope" })
	fmt.Println(missing == "", found)
	// Rung 2: opt.Option to chain or transform the absence.
	maybe := opt.Some(slug)
	fmt.Println(maybe.ValueOr("nobody"))
	fmt.Println(opt.None[string]().ValueOr("nobody"))
	// Rung 3: result.Result when absence becomes a failure with a cause.
	outcome := opt.ToResult(opt.None[string](), errors.New("lesson not found"))
	fmt.Println(result.UnwrapOr(outcome, "fallback-lesson"))
}
```

Expected output:

```text
slices-deep true
true false
slices-deep
nobody
fallback-lesson
```

> [!WARNING]
> Climb only as far as the caller needs: never trust the value without the `bool`, never use `Option` for failures with causes, and never use `Result` for plain absence.

See also `docs/guides/valid.md` for what happens after the value arrives.

## (b) First success wins — Race over mirrors

You query several mirrors and take whichever answers first. The winner
cancels its siblings through context cancellation; a dead mirror's failure
never stops the race.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/seinra/axiom/flow"
)

func main() {
	mirrors := []string{"slow-mirror", "fast-mirror", "dead-mirror"}
	winner, err := flow.Race(context.Background(), mirrors, 3, func(ctx context.Context, mirror string) (string, error) {
		switch mirror {
		case "fast-mirror":
			time.Sleep(10 * time.Millisecond)
			return "lesson from " + mirror, nil
		case "slow-mirror":
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(5 * time.Second):
				return "lesson from " + mirror, nil
			}
		default:
			return "", errors.New("mirror unreachable: " + mirror)
		}
	})
	fmt.Println(winner, err == nil)
}
```

Expected output:

```text
lesson from fast-mirror true
```

> [!NOTE]
> The slow branch selects on `ctx.Done()`: honoring the context is what makes sibling shutdown prompt. When every mirror fails, `Race` returns the zero value plus the first failure in time.

See
`docs/guides/flow.md` for the `Map`/`ForEach` failure-wins contract and
the latch-first panic policy.

## (c) Wired store — di + repo MemoryRepository

You hold a repository behind a named singleton: `Provide` the constructor
once, `Resolve` the shared store from any caller, and every caller sees
the same records.

```go
package main

import (
	"fmt"

	"github.com/seinra/axiom/di"
	"github.com/seinra/axiom/repo"
)

type lesson struct {
	Slug  string
	Title string
}

func lessonKeyOf(candidate lesson) string { return candidate.Slug }

func main() {
	injector := di.New()
	_ = di.Provide(injector, "lessons", func(*di.Injector) (*repo.MemoryRepository[lesson, string], error) {
		return repo.NewMemoryRepository(lessonKeyOf), nil
	})
	lessons, err := di.Resolve[*repo.MemoryRepository[lesson, string]](injector, "lessons")
	if err != nil {
		panic(err)
	}
	_ = lessons.Store(lesson{Slug: "generics-intro", Title: "Intro to Generics"})
	fetched, err := lessons.Fetch("generics-intro")
	if err != nil {
		panic(err)
	}
	fmt.Println(fetched.Title)
	again, _ := di.Resolve[*repo.MemoryRepository[lesson, string]](injector, "lessons")
	fmt.Println(again == lessons)
}
```

Expected output:

```text
Intro to Generics
true
```

> [!NOTE]
> The final `true` proves the singleton: the second `Resolve` returns the same store, not a rebuild. Swap the constructor in tests with `Replace` (one type per name still applies).

See `docs/guides/di.md` and
`docs/guides/repo.md` for the contracts behind each half.
