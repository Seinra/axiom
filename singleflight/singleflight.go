// Package singleflight collapses concurrent duplicate executions sharing one
// key into a single run, broadcasting the shared outcome to every waiter.
//
// Use a Group for one key space. Concurrent Call invocations with equal keys share
// one execution: the first caller runs while duplicates wait, then every
// waiter receives the same outcome and failure. Sequential calls re-run;
// entries live only while executing.
//
// No CallChan, no sharding, no batching: YAGNI. Reach for a plain call
// directly when keys never overlap concurrently; there is nothing to deduplicate.
//
// Which function?
//
// | When you want to ...                          | Reach for ... |
// |------------------------------------------------|---------------|
// | Share one execution across same-key callers    | Group.Call    |
// | Drop the in-flight entry so next Call re-runs  | Group.Forget  |
// | Hold one key space (zero value ready)          | Group         |
package singleflight

import "sync"

// callEntry tracks one in-flight execution and its broadcast outcome.
// The leader writes outcome, failure, and panic state before closing doneCh;
// waiters read only after doneCh closes, so the close carries happens-before.
type callEntry[Value any] struct {
	doneCh       chan struct{}
	outcome      Value
	failure      error
	panicPayload any
	didPanic     bool
}

// Group holds the in-flight executions for one key space.
// The zero value is ready to use; the map initializes lazily on first Call.
// Use through a pointer so concurrent callers share the same table.
type Group[Key comparable, Value any] struct {
	guard    sync.Mutex
	inflight map[Key]*callEntry[Value]
}

// Call runs execute for requestKey, sharing one execution across concurrent
// callers with equal keys. The leader runs execute and reports isShared
// false; duplicates wait and receive the same outcome and failure with
// isShared true. Sequential calls re-run: singleflight dedups concurrency,
// it is NOT a cache.
// If execute panics, every waiter panics with the same value.
// Call never holds the guard mutex while running execute.
func (flight *Group[Key, Value]) Call(requestKey Key, execute func() (Value, error)) (outcome Value, failure error, isShared bool) {
	flight.guard.Lock()
	if flight.inflight == nil {
		flight.inflight = make(map[Key]*callEntry[Value])
	}
	if pending, duplicated := flight.inflight[requestKey]; duplicated {
		flight.guard.Unlock()
		return flight.awaitDuplicate(pending)
	}
	pending := &callEntry[Value]{doneCh: make(chan struct{})}
	flight.inflight[requestKey] = pending
	flight.guard.Unlock()
	return flight.settleExecution(requestKey, pending, execute)
}

// Forget drops the in-flight entry for requestKey, so the next Call re-runs
// even while earlier waiters still hold the dropped execution.
func (flight *Group[Key, Value]) Forget(requestKey Key) {
	flight.guard.Lock()
	defer flight.guard.Unlock()
	delete(flight.inflight, requestKey)
}

// awaitDuplicate waits for the leader broadcast, then replays it.
// A panicking leader re-panics here with the same payload.
func (flight *Group[Key, Value]) awaitDuplicate(pending *callEntry[Value]) (Value, error, bool) {
	<-pending.doneCh
	if pending.didPanic {
		panic(pending.panicPayload)
	}
	return pending.outcome, pending.failure, true
}

// settleExecution runs execute without holding the guard, records the
// outcome, then publishes. A panic is recorded so publishOutcome can wake
// waiters before re-panicking the leader with the same payload.
func (flight *Group[Key, Value]) settleExecution(requestKey Key, pending *callEntry[Value], execute func() (Value, error)) (outcome Value, failure error, isShared bool) {
	defer flight.publishOutcome(requestKey, pending)
	defer func() {
		if payload := recover(); payload != nil {
			pending.panicPayload = payload
			pending.didPanic = true
		}
	}()
	outcome, failure = execute()
	pending.outcome = outcome
	pending.failure = failure
	return outcome, failure, false
}

// publishOutcome removes a completed entry, wakes waiters, and re-panics
// the leader when execute panicked. Forget may already have replaced the
// entry, so removal applies only while the table still holds this exact one.
func (flight *Group[Key, Value]) publishOutcome(requestKey Key, pending *callEntry[Value]) {
	flight.guard.Lock()
	if current, held := flight.inflight[requestKey]; held && current == pending {
		delete(flight.inflight, requestKey)
	}
	close(pending.doneCh)
	flight.guard.Unlock()
	if pending.didPanic {
		panic(pending.panicPayload)
	}
}
