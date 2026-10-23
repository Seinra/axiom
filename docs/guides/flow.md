# flow — bounded parallel map, or the first success wins

You run a transform over inputs with at most `workerCount` concurrent
runners. `Map` returns every result in input order; `ForEach` runs side
effects without results; `Race` returns the first success and cancels its
siblings.

You reach for this shape for bounded parallel work with errgroup parity: `Map` keeps input order while the first failure cancels siblings, and `Race` takes the first success in time when any single win is enough. Worker sizing is a convention, not magic (`workerCount <= 0` falls back to `runtime.NumCPU`).

When not to use it: single-goroutine work stays a plain loop, and unshaped concurrency stays `golang.org/x/sync/errgroup` directly — reach for `flow` when you need bounded runners with ordered results or a first-success race.

**Version:** `github.com/seinra/axiom`, Go 1.27.
**Prereqs:** `import "github.com/seinra/axiom/flow"`. Transforms receive a
cancellable context and must honor it, or cancellation stays late.

## Quick path

1. Write the transform: `func(ctx context.Context, in int) (string, error)`.
2. Every value in order: `flow.Map(ctx, inputs, workers, transform)`.
3. Any single success: `flow.Race(ctx, inputs, workers, transform)`.
4. Side effects only: `flow.ForEach(ctx, inputs, workers, consume)`.

## Which function?

The package doc (`go doc ./flow`) is the source; this table mirrors its
"Which function?" table:

| When you want to ... | Reach for ... |
|---|---|
| Transform every input, keep input order | `Map` |
| Run side effects without results | `ForEach` |
| Take the first success, cancel siblings | `Race` |
| Match the empty-`Race` sentinel | `ErrNoRaceInputs` |
| Retry a fallible step, bounded attempts | `Retry` |
| Retry for a value, bounded attempts | `RetryWithResult` |

## Example

From `flow/example_test.go` (`ExampleMap` — the ordered transform):

```go fragment: excerpt of ExampleMap — full runnable at flow/example_test.go
doubled, _ := flow.Map(context.Background(), []int{1, 2, 3}, 2, func(_ context.Context, candidate int) (int, error) {
	return candidate * 2, nil
})
fmt.Println(doubled)
// Output: [2 4 6]
```

Run it: `go test ./flow/ -run ExampleMap -v`. Recipe (b) in
`docs/cookbook.md` shows the `Race` first-success path end to end.

> [!WARNING]
> A transform already in flight still runs to completion — Go cannot preempt it, and a transform that ignores the context runs on too. Honor the context inside your transform (the package-doc `select` pattern) for prompt shutdown; extra work stays bounded by the runner count.

> [!WARNING]
> In `Race`, failures never cancel siblings — runners keep racing until the first success or until every input fails. A latched panic wins over a latched failure and poisons even a claimed championship.

## Contracts and errors

| Contract | What it means for you |
|---|---|
| Workers are a convention, not magic | `workerCount <= 0` falls back to `runtime.NumCPU`; runners never exceed `len(inputs)`. Empty `Map`/`ForEach` inputs succeed without running anything; empty `Race` inputs return the zero output plus `ErrNoRaceInputs` (match with `errors.Is`), never a silent zero. |
| First failure wins in `Map`/`ForEach` | The first transform failure cancels its siblings through context cancellation and is the returned failure (errgroup parity). Runners stop *starting* new work; a transform already in flight still runs to completion — extra work stays bounded by the runner count. Honor the context inside your transform (see the package-doc select pattern) for prompt shutdown. |
| `Race`: first success in time wins | Failures never cancel siblings — runners keep racing until the first success or until every input fails. All-failed `Race` returns the zero output plus the first failure in time. Use `Race` when any single success is enough and input order does not matter; use `Map` when you need every value in order. |
| Panics stay loud (latch-first) | A panicking transform never escapes its runner: the first panic payload latches, siblings cancel, every runner joins, and the caller re-panics the original payload. A latched panic wins over a latched failure, and in `Race` it poisons even a claimed championship — programmer bugs never hide behind a lucky winner. |
| No channels, pipelines, or fan-out | No FanIn/FanOut channels or pipelines; bounded retry lives in `flow.Retry`/`flow.RetryWithResult`. For single-goroutine work reach for plain loops or `golang.org/x/sync/errgroup` directly. |

An already-canceled context returns promptly with `active.Err()` and
starts no transform.

## Next step

Run recipe (b) in `docs/cookbook.md` for the `Race` first-success shape.
When the problem is duplicate suppression rather than parallel transform,
see `docs/guides/singleflight.md`.
