# singleflight — one execution per key, shared with every waiter

You collapse concurrent duplicate calls sharing one key into a single run.
The first caller executes while duplicates wait, then every waiter receives
the same outcome and failure.

You reach for this shape to suppress duplicate work under concurrency: one execution per key serves every same-key waiter, and entries live only while executing — there is no cache. The package ships no `CallChan`, sharding, or batching by design.

When not to use it: keys that never overlap concurrently carry nothing to deduplicate — call `execute` directly.

**Version:** Axiom v0.2.0, Go 1.27.
**Prereqs:** `import "github.com/seinra/axiom/singleflight"`. Share one
`Group` per key space through a pointer; the zero value is ready to use.

## Quick path

1. Hold one `Group[Key, Value]` per key space (zero value ready).
2. Call `group.Call(key, execute)` from every concurrent caller.
3. Read the `isShared` flag to know whether you ran or waited.

## Which function?

The package doc (`go doc ./singleflight`) is the source; this table mirrors
its "Which function?" table:

| When you want to ... | Reach for ... |
|---|---|
| Share one execution across same-key callers | `Group.Call` |
| Drop the in-flight entry so next `Call` re-runs | `Group.Forget` |
| Inspect calls, shared waiters, and panics | `Group.Stats` |
| Hold one key space (zero value ready) | `Group` |

## Example

From `singleflight/example_test.go` (`ExampleGroup_Call`):

```go fragment: excerpt of ExampleGroup_Call — full runnable at singleflight/example_test.go
var flight singleflight.Group[string, string]
executions := 0
outcome, _, _ := flight.Call("greeting", func() (string, error) {
	executions++
	return "hello", nil
})
fmt.Println(outcome)
fmt.Println(executions)
// Output:
// hello
// 1
```

Run it: `go test ./singleflight/ -run ExampleGroup_Call -v`.

> [!WARNING]
> `isShared` reports sharing, never presence: `false` means you ran `execute`, `true` means you received the broadcast. Check `failure` before touching `outcome`.

> [!WARNING]
> Deduplication is concurrency-only, not a cache: entries live only while executing, so sequential `Call`s with the same key re-run `execute` every time.

## Contracts and errors

| Contract | What it means for you |
|---|---|
| `isShared` reports sharing, never presence | The third return of `Call` tells you whether you were the leader (`false`, you ran `execute`) or a waiter (`true`, you received the broadcast). It is **not** an OrNull-style presence flag — never read it as "value found". Always check `failure` before touching `outcome`. |
| Deduplication is concurrency-only, not a cache | Entries live only while executing. Sequential `Call`s with the same key re-run `execute` every time. When keys never overlap concurrently there is nothing to deduplicate — call `execute` directly. |
| One outcome for leader and waiters | Duplicates receive the same outcome *and* the same failure the leader produced. If `execute` panics, every waiter panics with the same payload. |
| `Forget` forces a re-run | `Forget(key)` drops the in-flight entry so the next `Call` runs fresh, even while earlier waiters still hold the dropped execution. |
| No extras by design | No `CallChan`, no sharding, no batching (YAGNI). `Call` never holds its internal lock while running `execute`. |

## Observability

`Group.Stats` returns a point-in-time `Stats{Calls, Shared, Panics}` snapshot:
every `Call` invocation, duplicates that received the leader broadcast, and
leader executions that panicked (waiter re-panics never add). In serving it
tells you whether the key space actually collapses — rising `Shared` means
deduplication is working, while `Panics` flags a poisoned execution fanning
out to every waiter.

```go
var flight singleflight.Group[string, string]
_, _, _ = flight.Call("greeting", func() (string, error) { return "hello", nil })
stats := flight.Stats()
fmt.Println(stats.Calls, stats.Shared, stats.Panics)
// Output:
// 1 0 0
```

## Next step

When you need bounded parallel transforms instead of duplicate suppression,
see `docs/guides/flow.md`. For named singleton services rather than
per-key executions, see `docs/guides/di.md`.
