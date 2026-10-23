# task — one background computation, joined and shared

**Version:** Axiom on Go 1.27.
**Prereqs:** Go toolchain, generics basics (`[T]`) — or [foundations](foundations.md) first.

> API source: `go doc ./task` ([task/task.go](../../task/task.go)). This guide summarizes; the doc comments are the reference.

## Concept

`task.Task[Output]` is one deferred cancelable computation returning a
`result.Result`: async done right in one teachable unit. `Spawn` starts the
compute eagerly in its own goroutine and hands back a handle; `Await` joins
that goroutine and reports its single shared outcome. Every `Await` — from
any number of goroutines, before or after completion — receives the same
`Result`, so one background computation fans out without re-running. A
compute panic becomes `Fail` with the exact `result.Try` shape, turning
panics into data instead of crashing the process. Canceling the spawn context
makes every `Await` report `Fail(active.Err())` after the goroutine joins,
so no goroutine ever leaks: every spawned computation is joined, every panic
becomes data, every cancellation is observed, every result is shared.

The standard library already offers goroutines, channels, and errgroup: what
`task` adds is the cancel-plus-panic-plus-leak contract in one place. It
composes with `result.Try` (the sync panic boundary) and [flow](flow.md)
(bounded fan-out over tasks, lazy ordered streams of results).

## Steps

1. Spawn eager work: `future := task.Spawn(ctx, func(taskCtx context.Context) (Out, error) { … })` — compute starts before `Spawn` returns.
2. Await the shared outcome: `outcome := future.Await()` — joins first, then reports `Ok`, compute failure as `Fail`, or compute panic as `Fail` with the `Try` shape. Call it N times from N goroutines; all see the same `Result`.
3. Cancel to stop: cancel the spawn context and every `Await` reports `Fail(ctx.Err())` after the join. Honor `taskCtx.Done()` inside compute for prompt shutdown.
4. Never run on nothing: a pre-canceled context starts no goroutine — `Await` reports `Fail` at once. Never `Await` a zero `Task` (it panics) and never pass a nil compute or nil context to `Spawn` (both panic: programmer errors, loud at the spawn site).
5. Share the handle freely: copies share the same outcome cell, and `Await` is safe for concurrent use.

## Which primitive?

Resumed from the package table — see `go doc ./task` for the full version.

| When you want to ... | Reach for ... |
| --- | --- |
| Background single result, joinable | `task.Spawn` / `Await` |
| Sync panic boundary, no goroutine | `result.Try` |
| Parallel fan-out, ordered results | `flow.Map` |
| Lazy ordered stream of Results | `flow.MapSeqResults` |
| Dedup concurrent same-key runs | `singleflight.Call` |

## Example

Runnable — adapted from `ExampleSpawn` and `ExampleTask_Await` in
[task/example_test.go](../../task/example_test.go)
(same behavior). Save as `main.go`, with `replace github.com/seinra/axiom => <path-to-axiom>`:

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/seinra/axiom/task"
)

func main() {
	future := task.Spawn(context.Background(), func(_ context.Context) (int, error) {
		return 40 + 2, nil
	})
	success, failure := future.Await().Unwrap()
	fmt.Println(success, failure)

	compute := func(_ context.Context) (int, error) {
		return 0, errors.New("boom")
	}
	shared := task.Spawn(context.Background(), compute)
	_, first := shared.Await().Unwrap()
	_, second := shared.Await().Unwrap()
	fmt.Println(first, second)
}
```

Expected output:

```text
42 <nil>
boom boom
```

## Presence and absence

`Task` has no OrNull of its own — `result.Result` is the absence-and-failure
channel. A still-running computation is simply an `Await` not yet returned;
a canceled one is `Fail(ctx.Err())`; a missing result is a `Fail` with the
cause attached. When absence needs no cause at the call site, read it with
[opt](opt.md); when the boundary is synchronous and spawns nothing, reach for
[result.Try](result.md).

## Next step

- Fan-out: [flow](flow.md) — bounded parallel maps over many inputs, where one `Task` is a single background result.
- Panic boundary: [result](result.md) — `Try` is the sync version of the spawn boundary.
