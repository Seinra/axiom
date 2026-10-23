// Package task provides a deferred cancelable Task returning a result.Result:
// async done right in one teachable unit.
//
// Spawn starts compute eagerly in its own goroutine and hands back a Task
// handle; Await joins that goroutine and reports its single shared outcome.
// Every Await — from N concurrent goroutines, before or after completion —
// receives the same Result, so one background computation fans out without
// re-running. A compute panic becomes Fail with the exact result.Try shape,
// turning panics into data at the spawn boundary instead of crashing the
// process. Canceling the spawn context makes every Await report
// Fail(active.Err()) after the goroutine joins, so no goroutine ever leaks:
// every spawned computation is joined, every panic becomes data, every
// cancellation is observed, every result is shared.
//
// The standard library already offers goroutines, channels, and errgroup:
// what task adds is the cancel-plus-panic-plus-leak contract in one place.
// It composes with result.Try (the sync panic boundary) and flow (bounded
// fan-out over tasks, lazy ordered streams of Results).
//
// Which primitive?
//
// | When you want to ...               | Reach for ...      |
// |------------------------------------|--------------------|
// | Background single result, joinable | task.Spawn / Await |
// | Sync panic boundary, no goroutine  | result.Try         |
// | Parallel fan-out, ordered results  | flow.Map           |
// | Lazy ordered stream of Results     | flow.MapSeqResults |
// | Dedup concurrent same-key runs     | singleflight.Call  |
package task

import (
	"context"

	"github.com/seinra/axiom/result"
)

// Task is one deferred cancelable computation returning a result.Result.
//
// The zero value is not usable: Await panics on it, so a Task always comes
// from Spawn. Values copy safely — every copy shares the same done channel
// and outcome cell, so copies awaited anywhere report the same Result.
// Do not construct a Task by hand; its fields stay opaque on purpose.
type Task[Output any] struct {
	active  context.Context
	done    chan struct{}
	outcome *result.Result[Output]
}

// Spawn starts compute in its own goroutine before returning and reports the
// run through the returned Task: eager by contract, so compute runs even
// when Await never comes. The goroutine carries the spawn context — compute
// must honor active.Done for prompt shutdown, the same duty flow transforms
// owe their task context. Only the context plus result packages back this:
// close(done) is the join barrier, so no mutex is needed.
//
// A compute panic never escapes its goroutine: the panic boundary delegates
// to result.Try, so the Await outcome is Fail with the exact Try shape
// ("axiom result: recovered panic: ...", %w for error payloads and %v
// otherwise) and errors.Is/errors.As keep matching through it. A compute
// failure becomes Fail carrying that failure unwrapped.
//
// A pre-canceled active context starts no goroutine and never runs compute;
// Await then reports Fail(active.Err()). Canceling active mid-run keeps the
// join: Await waits for compute to return and then reports
// Fail(active.Err()), so cancellation is observed and no goroutine leaks.
//
// Spawn panics when compute is nil or active is nil: both are programmer
// errors, loud at the spawn site.
//
// Out of scope (YAGNI): no Then/Map combinators, no timeouts beyond context
// deadlines, no TaskGroup (reach for flow or errgroup), no tunable buffers,
// no metrics.
func Spawn[Output any](active context.Context, compute func(context.Context) (Output, error)) Task[Output] {
	if compute == nil {
		panic("axiom task: Spawn needs a non-nil compute")
	}
	if active == nil {
		panic("axiom task: Spawn needs a non-nil context")
	}
	settled := new(result.Result[Output])
	done := make(chan struct{})
	future := Task[Output]{active: active, done: done, outcome: settled}
	if active.Err() != nil {
		*settled = result.Fail[Output](active.Err())
		close(done)
		return future
	}
	go func() {
		defer close(done)
		// The panic boundary delegates to result.Try so the recovered shape
		// cannot drift from Try; Flatten collapses Result[Result[Output]].
		*settled = result.Flatten(result.Try(func() result.Result[Output] {
			value, failure := compute(active)
			if failure != nil {
				return result.Fail[Output](failure)
			}
			return result.Ok(value)
		}))
	}()
	return future
}

// Await joins the spawned goroutine and reports its outcome: success as Ok,
// compute failure as Fail, compute panic as Fail with the Try shape. Every
// Await on the same Task reports the same Result — concurrent waiters share
// one execution the way singleflight duplicates share one leader, except the
// run always happens: eager by contract, nothing to deduplicate.
//
// Once active is done, Await reports Fail(active.Err()) instead of the
// compute outcome: cancellation is observed, never half-yielded. Await
// always joins first, so returning never strands the goroutine.
//
// Await panics on the zero Task: Spawn a task first.
func (future Task[Output]) Await() result.Result[Output] {
	if future.done == nil || future.outcome == nil {
		panic("axiom task: Await on zero Task; Spawn a task first")
	}
	<-future.done
	if failure := future.active.Err(); failure != nil {
		return result.Fail[Output](failure)
	}
	return *future.outcome
}
