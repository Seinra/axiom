# di — one lazy singleton per name

You register a constructor with `Provide` and resolve the shared singleton
with `Resolve`. The constructor runs once on first `Resolve`; later calls
get the cached value.

**Version:** `github.com/seinra/axiom`, Go 1.27.
**Prereqs:** `import "github.com/seinra/axiom/di"`. Share the `*Injector`
from `New`; never copy it after use.

## Quick path

1. Create the container: `injector := di.New()`.
2. Register: `di.Provide(injector, "greeting", func(*di.Injector) (string, error) { return "hello", nil })`.
3. Resolve: `greeting, err := di.Resolve[string](injector, "greeting")`.

## Which function?

The package doc (`go doc ./di`) is the source; this table mirrors its
"Which function?" table:

| When you want to ... | Reach for ... |
|---|---|
| Create an empty container | `New` (`Injector`) |
| Register a lazy constructor | `Provide` |
| Resolve the shared singleton (retries on fail) | `Resolve` |
| Resolve or panic (tests, static seeds) | `MustResolve` |
| Swap the constructor and drop the cache | `Replace` |
| Match duplicate / missing / in-flight | `ErrDuplicateService` / `ErrServiceNotFound` / `ErrBuildInFlight` |

## Example

From `di/example_test.go` (`ExampleResolve`):

```go fragment: excerpt of ExampleResolve — full runnable at di/example_test.go
injector := di.New()
_ = di.Provide(injector, "greeting", func(*di.Injector) (string, error) { return "hello", nil })
greeting, _ := di.Resolve[string](injector, "greeting")
fmt.Println(greeting)
// Output: hello
```

Run it: `go test ./di/ -run ExampleResolve -v`.

## Contracts and errors

| Contract | What it means for you |
|---|---|
| One type per name | A service name maps to exactly one `Service` type. A second `Provide` with the same name reports `ErrDuplicateService` (match with `errors.Is`) and keeps the first registration. |
| Failing builds retry | A constructor failure propagates without caching, so the next `Resolve` retries. A panicking build resets the entry too — every caller still panics with the same payload, and the next `Resolve` after the storm retries. |
| `Replace` swaps constructor and drops cache | The next `Resolve` rebuilds with the new constructor. `Replace` on an unknown name reports `ErrServiceNotFound`; on a `nil` injector the same sentinel. |
| `Replace` refuses mid-build swaps | If a build is in flight for the name, `Replace` changes nothing and returns `ErrBuildInFlight` — neither the constructor pointer nor the cache is touched, so in-flight waiters still receive the original build. |
| Concurrent `Resolve` shares one build | Concurrent callers for one service wait on a single constructor run; the table lock guards only lookup and publish, never the run. But cycles are a programmer error: a constructor that (transitively) resolves its own service never settles, and the container performs no cycle detection. |
| `MustResolve` is for tests and seeds only | It panics on a missing name or failed build — programmer error, never control flow. |

The type-erased table stays inside the package: every crossing is gated by
a generic `Resolve` returning `T` or an error, the single sanctioned
exception to the no-`any` rule.

## Next step

Recipe (c) in `docs/cookbook.md` wires a `repo.MemoryRepository` behind
`di` end to end. For per-key duplicate suppression rather than named
singletons, see `docs/guides/singleflight.md`.
