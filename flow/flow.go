// Package flow provides bounded parallel combinators with errgroup semantics.
//
// Map runs a transform over inputs with at most workerCount concurrent
// runners and reports results in input order. ForEach applies a consume step
// for side effects without collecting results. Race runs the same transform
// and returns the first success, canceling its siblings; reach for Race when
// any single success is enough and input order does not matter. The first
// transform failure cancels its Map and ForEach siblings through context
// cancellation and wins as the returned failure, matching errgroup parity:
// one failure in, that failure out, siblings stopped. A Race with every
// transform failed instead reports the zero Output plus the first failure.
//
// Worker sizing follows a documented convention, not magic: a workerCount
// of zero or less falls back to runtime.NumCPU.
//
// Transform functions receive the cancellable context and must honor it so
// cancellation stays prompt; siblings stop starting new work once the first
// failure cancels the shared context:
//
//	transform := func(active context.Context, candidate int) (string, error) {
//	    select {
//	    case <-active.Done():
//	        return "", active.Err()
//	    case reading := <-loadAsync(candidate):
//	        return reading, nil
//	    }
//	}
//
// Out of scope (YAGNI): no FanIn/FanOut channels; those are later slices if
// ever. Reach for plain loops or golang.org/x/sync/errgroup directly for
// single-goroutine or unshaped work. Retry landed as the bounded-retry slice.
//
// Which function?
//
// | When you want to ...                                   | Reach for ...     |
// |---------------------------------------------------------|-------------------|
// | Transform every input, keep input order                  | Map               |
// | Run side effects without results                         | ForEach           |
// | Take the first success, cancel siblings                  | Race              |
// | Match the empty-Race sentinel                            | ErrNoRaceInputs   |
// | Transform a sequence, keep stream order                  | MapSeq            |
// | ForEach over a sequence, result dropped                  | EachSeq           |
// | Transform a sequence, keep stream order, collect Results | MapSeqResults     |
// | Retry a fallible step, bounded attempts                | Retry             |
// | Retry for a value, bounded attempts                    | RetryWithResult   |
package flow

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"runtime"
	"sync"

	"github.com/seinra/axiom/result"
)

// normalizeWorkerCount resolves runners for inputCount items: a workerCount
// of zero or less falls back to runtime.NumCPU, and runners never exceed
// inputCount.
func normalizeWorkerCount(workerCount int, inputCount int) int {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}
	if workerCount > inputCount {
		workerCount = inputCount
	}
	return workerCount
}

// runShared carries the mutable outcome across Map runners. The first failure
// and the first panic payload latch once; a latched panic re-panics the
// caller after every runner joins.
type runShared struct {
	failureOnce  sync.Once
	firstFailure error
	panicOnce    sync.Once
	firstPanic   any
}

// callTransform runs transform with a panic boundary: a panic latches through
// onPanic, cancels siblings via the caller's cancel, and reports didPanic
// instead of unwinding the runner goroutine (an escaping panic would take the
// whole process down). The zero transform result on the panic path is
// discarded: callers must check didPanic before touching the other returns.
// panic(nil) arrives as *PanicNilError (Go 1.21+) and re-panics as such.
func callTransform[Input any, Output any](taskContext context.Context, candidate Input, transform func(context.Context, Input) (Output, error), onPanic func(payload any)) (transformed Output, transformFailure error, didPanic bool) {
	defer func() {
		if payload := recover(); payload != nil {
			onPanic(payload)
			didPanic = true
		}
	}()
	transformed, transformFailure = transform(taskContext, candidate)
	return transformed, transformFailure, false
}

// mapRunner drains positions through transform, writing each success to its
// own ordered slot. The first failure and the first panic latch once; either
// stops this runner, and cancellation drains the rest. Every path joins
// workGroup.
func mapRunner[Input any, Output any](taskContext context.Context, positions <-chan int, inputs []Input, transform func(context.Context, Input) (Output, error), ordered []Output, shared *runShared, cancelTask context.CancelFunc, workGroup *sync.WaitGroup) {
	defer workGroup.Done()
	// Drained positions stay zero-valued, but a drain implies a latched failure
	// or panic, so Map never returns a holed slice: failure returns nil and a
	// latched panic re-panics in the caller.
	for position := range positions {
		select {
		case <-taskContext.Done():
			continue
		default:
		}
		transformed, transformFailure, didPanic := callTransform(taskContext, inputs[position], transform, func(payload any) {
			latchFirstPanic(&shared.panicOnce, &shared.firstPanic, payload, cancelTask)
		})
		if didPanic {
			return
		}
		if transformFailure != nil {
			shared.failureOnce.Do(func() {
				shared.firstFailure = transformFailure
				cancelTask()
			})
			return
		}
		ordered[position] = transformed
	}
}

// Map runs transform over inputs with at most workerCount concurrent
// runners and returns results in input order, even when later items finish
// first: each runner writes only its own positions, and positions read back
// in input sequence after every runner joins.
//
// A workerCount of zero or less falls back to runtime.NumCPU. Empty inputs
// yield an empty non-nil slice. An already-canceled active context returns
// promptly with active.Err() and starts no transform.
//
// The first transform failure cancels its siblings via context cancellation
// and Map returns that failure (errgroup parity); sibling transforms must
// honor the context, as shown in the package example, for prompt shutdown.
// Cancelled siblings stop starting new work: runners that observe the
// cancellation drain their remaining positions without invoking transform,
// so unstarted inputs never run after the first failure. A transform already
// in flight still runs to completion (Go cannot preempt it), and a transform
// that ignores the context runs to completion too — extra work stays bounded
// by the runner count, never by the input length. Every spawned runner joins
// through a WaitGroup on all paths, and no lock is held while transform runs.
//
// A panicking transform never escapes its runner goroutine: the first panic
// payload latches, siblings cancel through the same context path as failures,
// every runner joins, and Map re-panics the original payload in the caller —
// the Rayon/PLINQ consensus (in-flight work completes, new work stops, panics
// propagate, never silent). A latched panic wins over a latched failure:
// programmer bugs stay loud.
func Map[Input any, Output any](active context.Context, inputs []Input, workerCount int, transform func(context.Context, Input) (Output, error)) ([]Output, error) {
	if active.Err() != nil {
		return nil, active.Err()
	}
	if len(inputs) == 0 {
		return []Output{}, nil
	}
	runners := normalizeWorkerCount(workerCount, len(inputs))
	taskContext, cancelTask := context.WithCancel(active)
	defer cancelTask()
	ordered := make([]Output, len(inputs))
	positions := make(chan int, len(inputs))
	for position := range inputs {
		positions <- position
	}
	close(positions)
	var workGroup sync.WaitGroup
	shared := &runShared{}
	for range runners {
		workGroup.Add(1)
		go mapRunner(taskContext, positions, inputs, transform, ordered, shared, cancelTask, &workGroup)
	}
	workGroup.Wait()
	if shared.firstPanic != nil {
		panic(shared.firstPanic)
	}
	if shared.firstFailure != nil {
		return nil, shared.firstFailure
	}
	return ordered, nil
}

// ForEach applies consume to every input with at most workerCount concurrent
// runners, unordered and without results; the first consume failure cancels
// the rest and wins as the returned failure, nil when all succeed. A
// workerCount of zero or less falls back to runtime.NumCPU. Empty inputs
// return nil without running consume. An already-canceled active context
// returns promptly with active.Err() and starts no consume step. ForEach
// inherits the Map panic policy: a panicking consume re-panics its payload
// in the caller after every runner joins.
func ForEach[Input any](active context.Context, inputs []Input, workerCount int, consume func(context.Context, Input) error) error {
	_, mapFailure := Map(active, inputs, workerCount, func(taskContext context.Context, candidate Input) (struct{}, error) {
		consumeFailure := consume(taskContext, candidate)
		return struct{}{}, consumeFailure
	})
	return mapFailure
}

// ErrNoRaceInputs reports a Race call with no inputs: with nothing to race
// there is no winner, so Race returns the zero Output plus this sentinel
// wrapped for errors.Is matching instead of a silent zero value.
var ErrNoRaceInputs = errors.New("axiom flow: Race needs at least one input")

// raceShared carries the mutable Race outcome across runners. The champion
// fields change under outcomeMutex; the first failure latches once.
// Canceling while holding outcomeMutex is a documented design call:
// context.CancelFunc never blocks and never re-enters this mutex, so the
// critical section stays a flag flip plus a non-blocking broadcast.
type raceShared[Output any] struct {
	outcomeMutex  sync.Mutex
	champion      Output
	championFound bool
	failureOnce   sync.Once
	firstFailure  error
	panicOnce     sync.Once
	firstPanic    any
}

// latchFirstPanic records the first panic payload once and cancels siblings.
// Shared by Map and Race runners so the latch-and-cancel pairing cannot drift
// between them.
func latchFirstPanic(panicOnce *sync.Once, firstPanic *any, payload any, cancelTask context.CancelFunc) {
	panicOnce.Do(func() {
		*firstPanic = payload
	})
	cancelTask()
}

// raceRunner drains positions until the first success cancels the task
// context. Failures latch the first one and keep racing; a success claims
// the championship once and stops this runner, while concurrent successes
// that lose the claim return without overwriting. The pre-transform Done
// check is best effort: a win landing after the check still leaves that
// transform in flight, bounded by the runner count. Every path joins
// workGroup.
func raceRunner[Input any, Output any](taskContext context.Context, positions <-chan int, inputs []Input, transform func(context.Context, Input) (Output, error), shared *raceShared[Output], cancelTask context.CancelFunc, workGroup *sync.WaitGroup) {
	defer workGroup.Done()
	for position := range positions {
		select {
		case <-taskContext.Done():
			continue
		default:
		}
		transformed, transformFailure, didPanic := callTransform(taskContext, inputs[position], transform, func(payload any) {
			latchFirstPanic(&shared.panicOnce, &shared.firstPanic, payload, cancelTask)
		})
		if didPanic {
			return
		}
		if transformFailure != nil {
			shared.failureOnce.Do(func() {
				shared.firstFailure = transformFailure
			})
			continue
		}
		shared.outcomeMutex.Lock()
		if !shared.championFound {
			shared.champion = transformed
			shared.championFound = true
			cancelTask()
		}
		shared.outcomeMutex.Unlock()
		return
	}
}

// Race runs transform over inputs with at most workerCount concurrent
// runners and returns the first success in time, canceling its siblings via
// context cancellation: one outcome in, siblings stopped, matching the
// errgroup parity Map documents. Use Race when any single success is enough
// and input order does not matter; use Map when every transformed value is
// needed in input order.
//
// A workerCount of zero or less falls back to runtime.NumCPU. Empty inputs
// yield the zero Output plus ErrNoRaceInputs wrapped, never a silent zero.
// An already-canceled active context returns promptly with active.Err() and
// starts no transform.
//
// The first success in time wins, even when an earlier input position is
// still running: under concurrent completions "in time" means first to claim
// the championship, not strict wall-clock finish order. Runners that observe
// the win drain their remaining positions without invoking transform, so no
// runner starts new work after observing the win. A transform already
// in flight still runs to completion (Go cannot preempt it), including one
// that passed its Done check before the win landed, and a transform that
// ignores the context runs to completion too — extra work stays bounded by
// the runner count, never by the input length.
//
// Transform failures never cancel siblings: runners keep racing until the
// first success or until every input fails. When every transform fails, Race
// returns the zero Output plus the first failure in time. Sibling transforms
// must honor the context, as shown in the package example, for prompt
// shutdown after the win. Every spawned runner joins through a WaitGroup on
// all paths, and no lock is held while transform runs.
//
// Panics follow the Map policy: the first payload latches, siblings cancel,
// and the caller re-panics the original payload. A *latched* panic poisons
// even a claimed championship; inputs never started after a win never run,
// so an unstarted panic never fires. A latched panic poisons the Race:
// programmer bugs never hide behind a lucky winner.
func Race[Input any, Output any](active context.Context, inputs []Input, workerCount int, transform func(context.Context, Input) (Output, error)) (winner Output, failure error) {
	if active.Err() != nil {
		var empty Output
		return empty, active.Err()
	}
	if len(inputs) == 0 {
		var empty Output
		return empty, fmt.Errorf("axiom flow: Race with no inputs: %w", ErrNoRaceInputs)
	}
	runners := normalizeWorkerCount(workerCount, len(inputs))
	taskContext, cancelTask := context.WithCancel(active)
	defer cancelTask()
	positions := make(chan int, len(inputs))
	for position := range inputs {
		positions <- position
	}
	close(positions)
	var workGroup sync.WaitGroup
	shared := &raceShared[Output]{}
	for range runners {
		workGroup.Add(1)
		go raceRunner(taskContext, positions, inputs, transform, shared, cancelTask, &workGroup)
	}
	workGroup.Wait()
	// Poison beats champion: see Race's godoc panic policy.
	if shared.firstPanic != nil {
		panic(shared.firstPanic)
	}
	if shared.championFound {
		return shared.champion, nil
	}
	if shared.firstFailure != nil {
		var empty Output
		return empty, shared.firstFailure
	}
	var empty Output
	return empty, taskContext.Err()
}

// seqSlot carries one streamed sequence value with its stream position, so
// runners write each success back to its own ordered slot and positions read
// back in sequence order after every goroutine joins.
type seqSlot[Input any] struct {
	position  int
	candidate Input
}

// seqFeeder drains sequence into positions, reserving one ordered slot per
// streamed value before offering it: the reservation order is the stream
// order, so runners write successes back in sequence order. Every offer
// selects on taskContext.Done, so a canceled run never wedges the feeder.
// The positions channel closes on every path so runners always join.
// The six-parameter split mirrors feeder/runner ownership (feeder appends, runners store by index); one call site, so no options struct per YAGNI.
func seqFeeder[Input any, Output any](taskContext context.Context, sequence iter.Seq[Input], positions chan<- seqSlot[Input], orderedGuard *sync.Mutex, ordered *[]Output, workGroup *sync.WaitGroup) {
	defer workGroup.Done()
	defer close(positions)
	for candidate := range sequence {
		orderedGuard.Lock()
		position := len(*ordered)
		var zero Output
		*ordered = append(*ordered, zero)
		orderedGuard.Unlock()
		// Reserved before offered: on cancellation MapSeq returns (nil, err), so a reserved-but-unoffered slot never surfaces as a hole.
		offer := seqSlot[Input]{position: position, candidate: candidate}
		select {
		case <-taskContext.Done():
			return
		case positions <- offer:
		}
	}
}

// seqRunner drains streamed slots through transform, writing each success to
// its own ordered slot under orderedGuard: the feeder is the only appender
// while runners only store by index, so no lock is held while transform
// runs. Mirrors mapRunner: the first failure and the first panic latch once,
// either stops this runner, and cancellation drains the rest so unstarted
// inputs never run after the first failure. Every path joins workGroup.
func seqRunner[Input any, Output any](taskContext context.Context, positions <-chan seqSlot[Input], transform func(context.Context, Input) (Output, error), orderedGuard *sync.Mutex, ordered *[]Output, shared *runShared, cancelTask context.CancelFunc, workGroup *sync.WaitGroup) {
	defer workGroup.Done()
	for slot := range positions {
		select {
		case <-taskContext.Done():
			continue
		default:
		}
		transformed, transformFailure, didPanic := callTransform(taskContext, slot.candidate, transform, func(payload any) {
			latchFirstPanic(&shared.panicOnce, &shared.firstPanic, payload, cancelTask)
		})
		if didPanic {
			return
		}
		if transformFailure != nil {
			shared.failureOnce.Do(func() {
				shared.firstFailure = transformFailure
				cancelTask()
			})
			return
		}
		orderedGuard.Lock()
		(*ordered)[slot.position] = transformed
		orderedGuard.Unlock()
	}
}

// MapSeq runs transform over sequence with at most workerCount concurrent
// runners and returns results in stream order, even when later items finish
// first: the feeder reserves one ordered slot per streamed value, each
// runner writes only its own slots, and slots read back in stream order
// after every goroutine joins.
//
// Input streams, output materializes: MapSeq never holds the input (values
// flow from the sequence through an unbuffered handoff, in flight bounded by
// the runner count), but the ordered result grows with the stream, so a
// successful run over an endless sequence never returns — reach for prompt
// failure plus cancellation, not success, over unbounded inputs. Success ranges
// to completion only over bounded sequences; unbounded inputs terminate
// through prompt failure plus cancellation. The standard
// library has no parallel ordered fan-in over iter.Seq; MapSeq fills that gap
// with errgroup parity instead of a new vocabulary: one failure in, that
// failure out, siblings stopped.
//
// A workerCount of zero or less falls back to runtime.NumCPU. A nil sequence
// yields an empty non-nil slice without spawning any goroutine (ranging over
// a nil iter.Seq would panic); an empty sequence yields the same. An
// already-canceled active context returns promptly with active.Err() and
// starts nothing.
//
// The first transform failure cancels its siblings AND the feeder via
// context cancellation and MapSeq returns that failure (errgroup parity);
// sibling transforms must honor the context, as shown in the package example,
// for prompt shutdown. Runners that observe the cancellation drain their
// remaining slots without invoking transform, so unstarted inputs never run
// after the first failure; the feeder selects on cancellation on every
// handoff, so it never wedges. A transform already in flight still runs to
// completion (Go cannot preempt it), and a transform that ignores the context
// runs to completion too — extra work stays bounded by the runner count,
// never by the sequence length. Every spawned goroutine (one feeder plus the
// runners) joins through a WaitGroup on all paths.
//
// A context canceled mid-run with no transform failure returns (nil, the
// context error): partially ordered output is never presented as success.
//
// A panicking transform never escapes its runner goroutine: the first panic
// payload latches, siblings cancel through the same context path as
// failures, every goroutine joins, and MapSeq re-panics the original payload
// in the caller — the Rayon/PLINQ consensus (in-flight work completes, new
// work stops, panics propagate, never silent). A latched panic wins over a
// latched failure: programmer bugs stay loud.
func MapSeq[Input any, Output any](active context.Context, sequence iter.Seq[Input], workerCount int, transform func(context.Context, Input) (Output, error)) ([]Output, error) {
	if active.Err() != nil {
		return nil, active.Err()
	}
	if sequence == nil {
		return []Output{}, nil
	}
	runners := workerCount
	if runners <= 0 {
		runners = runtime.NumCPU()
	}
	taskContext, cancelTask := context.WithCancel(active)
	defer cancelTask()
	positions := make(chan seqSlot[Input])
	var workGroup sync.WaitGroup
	shared := &runShared{}
	var orderedGuard sync.Mutex
	ordered := []Output{}
	workGroup.Add(1)
	go seqFeeder(taskContext, sequence, positions, &orderedGuard, &ordered, &workGroup)
	for range runners {
		workGroup.Add(1)
		go seqRunner(taskContext, positions, transform, &orderedGuard, &ordered, shared, cancelTask, &workGroup)
	}
	workGroup.Wait()
	if shared.firstPanic != nil {
		panic(shared.firstPanic)
	}
	if shared.firstFailure != nil {
		return nil, shared.firstFailure
	}
	if taskContext.Err() != nil {
		return nil, taskContext.Err()
	}
	return ordered, nil
}

// EachSeq applies consume to every value in sequence with at most workerCount
// concurrent runners, unordered and without collecting results; the first
// consume failure cancels the rest and wins as the returned failure, nil when
// all succeed. A workerCount of zero or less falls back to runtime.NumCPU.
// A nil or empty sequence returns nil without running consume. An
// already-canceled active context returns promptly with active.Err() and
// starts no consume step. EachSeq inherits the MapSeq fail-fast, drain, and
// panic policies: a panicking consume re-panics its payload in the caller
// after every goroutine joins.
//
// EachSeq is a fully streaming O(1) side-effect pass: it runs through MapSeq
// (the ForEach-via-Map precedent) with a zero-sized output, so nothing but
// the in-flight runners accumulates while the sequence streams.
func EachSeq[Input any](active context.Context, sequence iter.Seq[Input], workerCount int, consume func(context.Context, Input) error) error {
	_, eachFailure := MapSeq(active, sequence, workerCount, func(taskContext context.Context, candidate Input) (struct{}, error) {
		consumeFailure := consume(taskContext, candidate)
		return struct{}{}, consumeFailure
	})
	return eachFailure
}

// seqResultsReorderCapacity bounds the ordered fan-in window: at most this
// many dispatched-but-unyielded outcomes stay buffered while the stream is
// consumed. The capacity is a documented fixed constant, never tunable, so
// in-flight plus buffered work stays bounded no matter how long the input
// sequence runs. The pulling goroutine briefly holds one extra already-pulled
// candidate while making room, so the live set is capacity plus runners plus
// one lookahead, still O(1) in the sequence length.
const seqResultsReorderCapacity = 64

// seqResultsJob carries one dispatched stream value with its ordered reply
// slot. Workers write exactly one Result per job to the buffered slot, so a
// send never blocks on an abandoned consumer after cancellation or break.
type seqResultsJob[Input any, Output any] struct {
	candidate Input
	slot      chan result.Result[Output]
}

// seqResultsWorker drains dispatched jobs through transform, writing exactly
// one Result per job to its own ordered slot. A transform failure becomes a
// Fail outcome and never cancels siblings (collect-all); the runner keeps
// draining. A transform panic latches once, cancels siblings, and stops this
// runner without unwinding it. Cancellation drains unstarted jobs without
// running transform. Every path joins workGroup.
func seqResultsWorker[Input any, Output any](taskContext context.Context, positions <-chan seqResultsJob[Input, Output], transform func(context.Context, Input) (Output, error), panicOnce *sync.Once, panicMutex *sync.Mutex, firstPanic *any, cancelTask context.CancelFunc, workGroup *sync.WaitGroup) {
	defer workGroup.Done()
	for offered := range positions {
		select {
		case <-taskContext.Done():
			// Abandoned slots are buffered-1 with no reader after cancel plus join, so GC owns them; exactly-once still holds for consumed inputs.
			continue
		default:
		}
		transformed, transformFailure, didPanic := callTransform(taskContext, offered.candidate, transform, func(payload any) {
			panicOnce.Do(func() {
				panicMutex.Lock()
				*firstPanic = payload
				panicMutex.Unlock()
			})
			cancelTask()
		})
		if didPanic {
			return
		}
		var outcome result.Result[Output]
		if transformFailure != nil {
			outcome = result.Fail[Output](transformFailure)
		} else {
			outcome = result.Ok[Output](transformed)
		}
		offered.slot <- outcome
	}
}

// loadSeqResultsPanic reads the latched transform panic under guard; nil
// means no worker has panicked yet. One uncontended mutex per ordered yield: the emit path already blocks on the slot receive, so a mutex keeps the latch site obvious without measurable cost.
func loadSeqResultsPanic(panicMutex *sync.Mutex, firstPanic *any) any {
	panicMutex.Lock()
	defer panicMutex.Unlock()
	return *firstPanic
}

// seqResultsEmit waits for the oldest pending slot and yields it in stream
// order. It reports shouldStop when the stream must terminate (downstream
// break or cancellation) and panicPayload when a transform panic latched
// (caller must re-panic in the pulling goroutine; panic wins over Fail).
// Emit/TryEmit/Offer share this select shape on purpose: blocking drain, non-blocking flush, and dispatch each need their own select, so no helper unifies them.
func seqResultsEmit[Output any](taskContext context.Context, pendingSlots []chan result.Result[Output], yield func(result.Result[Output]) bool, loadPanic func() any) ([]chan result.Result[Output], bool, any) {
	if panicPayload := loadPanic(); panicPayload != nil {
		return pendingSlots, true, panicPayload
	}
	select {
	case outcome := <-pendingSlots[0]:
		if panicPayload := loadPanic(); panicPayload != nil {
			return pendingSlots, true, panicPayload
		}
		if !yield(outcome) {
			return pendingSlots[1:], true, nil
		}
		return pendingSlots[1:], false, nil
	case <-taskContext.Done():
		if panicPayload := loadPanic(); panicPayload != nil {
			return pendingSlots, true, panicPayload
		}
		// Cancellation ends the stream: a ready-but-unyielded outcome is dropped, never half-yielded.
		return pendingSlots, true, nil
	}
}

// seqResultsTryEmit yields the oldest pending slot only when it is already
// ready, never blocking for it; emitted reports whether a yield happened.
// Cancellation and latched panics still stop the stream promptly.
func seqResultsTryEmit[Output any](taskContext context.Context, pendingSlots []chan result.Result[Output], yield func(result.Result[Output]) bool, loadPanic func() any) ([]chan result.Result[Output], bool, any, bool) {
	if panicPayload := loadPanic(); panicPayload != nil {
		return pendingSlots, true, panicPayload, false
	}
	select {
	case outcome := <-pendingSlots[0]:
		if panicPayload := loadPanic(); panicPayload != nil {
			return pendingSlots, true, panicPayload, false
		}
		if !yield(outcome) {
			return pendingSlots[1:], true, nil, true
		}
		return pendingSlots[1:], false, nil, true
	case <-taskContext.Done():
		if panicPayload := loadPanic(); panicPayload != nil {
			return pendingSlots, true, panicPayload, false
		}
		return pendingSlots, true, nil, false
	default:
		return pendingSlots, false, nil, false
	}
}

// seqResultsOffer dispatches one stream value to the workers with its
// ordered reply slot, reporting shouldStop on cancellation and panicPayload
// on a latched transform panic.
func seqResultsOffer[Input any, Output any](taskContext context.Context, positions chan<- seqResultsJob[Input, Output], pendingSlots []chan result.Result[Output], candidate Input, loadPanic func() any) ([]chan result.Result[Output], bool, any) {
	if loadPanic() != nil {
		return pendingSlots, true, loadPanic()
	}
	select {
	case <-taskContext.Done():
		if panicPayload := loadPanic(); panicPayload != nil {
			return pendingSlots, true, panicPayload
		}
		return pendingSlots, true, nil
	default:
	}
	slotChannel := make(chan result.Result[Output], 1)
	offered := seqResultsJob[Input, Output]{candidate: candidate, slot: slotChannel}
	select {
	case positions <- offered:
		return append(pendingSlots, slotChannel), false, nil
	case <-taskContext.Done():
		if panicPayload := loadPanic(); panicPayload != nil {
			return pendingSlots, true, panicPayload
		}
		return pendingSlots, true, nil
	}
}

// seqResultsFlushReady yields every already-ready oldest slot without
// blocking for the next one; emitted stops at the first pending outcome.
func seqResultsFlushReady[Output any](taskContext context.Context, pendingSlots []chan result.Result[Output], yield func(result.Result[Output]) bool, loadPanic func() any) ([]chan result.Result[Output], bool, any) {
	for len(pendingSlots) > 0 {
		var shouldStop bool
		var panicPayload any
		var emitted bool
		pendingSlots, shouldStop, panicPayload, emitted = seqResultsTryEmit(taskContext, pendingSlots, yield, loadPanic)
		if panicPayload != nil {
			return pendingSlots, true, panicPayload
		}
		if shouldStop {
			return pendingSlots, true, nil
		}
		if !emitted {
			break
		}
	}
	return pendingSlots, false, nil
}

// seqResultsDrain emits every remaining pending slot in order, blocking for
// each oldest outcome until the window empties or the run stops.
func seqResultsDrain[Output any](taskContext context.Context, pendingSlots []chan result.Result[Output], yield func(result.Result[Output]) bool, loadPanic func() any) ([]chan result.Result[Output], bool, any) {
	for len(pendingSlots) > 0 {
		var shouldStop bool
		var panicPayload any
		pendingSlots, shouldStop, panicPayload = seqResultsEmit(taskContext, pendingSlots, yield, loadPanic)
		if panicPayload != nil {
			return pendingSlots, true, panicPayload
		}
		if shouldStop {
			return pendingSlots, true, nil
		}
	}
	return pendingSlots, false, nil
}

// seqResultsPump offers one stream value to the workers while keeping the
// reorder window bounded, then flushes every already-ready outcome. It
// reports shouldStop when the stream must terminate and panicPayload when a
// transform panic latched. Split from seqResultsDrive to hold the complexity
// ceiling: drive owns the range, pump owns one iteration.
func seqResultsPump[Input any, Output any](taskContext context.Context, positions chan<- seqResultsJob[Input, Output], pendingSlots []chan result.Result[Output], candidate Input, yield func(result.Result[Output]) bool, loadPanic func() any) ([]chan result.Result[Output], bool, any) {
	for len(pendingSlots) >= seqResultsReorderCapacity {
		var shouldStop bool
		var panicPayload any
		pendingSlots, shouldStop, panicPayload = seqResultsEmit(taskContext, pendingSlots, yield, loadPanic)
		if panicPayload != nil {
			return pendingSlots, true, panicPayload
		}
		if shouldStop {
			return pendingSlots, true, nil
		}
	}
	var shouldStop bool
	var panicPayload any
	pendingSlots, shouldStop, panicPayload = seqResultsOffer(taskContext, positions, pendingSlots, candidate, loadPanic)
	if panicPayload != nil {
		return pendingSlots, true, panicPayload
	}
	if shouldStop {
		return pendingSlots, true, nil
	}
	return seqResultsFlushReady(taskContext, pendingSlots, yield, loadPanic)
}

// seqResultsDrive pulls sequence on the caller's stack while workers
// transform concurrently, emitting ordered Results through the bounded
// window. It returns a transform panic payload to re-panic, nil on any
// normal termination (exhausted input, downstream break, or cancellation).
// A generator panic is never caught here: it unwinds through the caller
// after the deferred join closes the run, so it propagates on the puller's
// stack exactly as ranging the sequence directly would.
func seqResultsDrive[Input any, Output any](taskContext context.Context, sequence iter.Seq[Input], positions chan<- seqResultsJob[Input, Output], yield func(result.Result[Output]) bool, loadPanic func() any) any {
	pendingSlots := make([]chan result.Result[Output], 0, seqResultsReorderCapacity+1)
	for candidate := range sequence {
		var shouldStop bool
		var panicPayload any
		pendingSlots, shouldStop, panicPayload = seqResultsPump(taskContext, positions, pendingSlots, candidate, yield, loadPanic)
		if panicPayload != nil {
			return panicPayload
		}
		if shouldStop {
			return nil
		}
	}
	var shouldStop bool
	var panicPayload any
	_, shouldStop, panicPayload = seqResultsDrain(taskContext, pendingSlots, yield, loadPanic)
	if panicPayload != nil {
		return panicPayload
	}
	if shouldStop {
		return nil
	}
	return nil
}

// MapSeqResults runs transform over sequence with at most workerCount
// concurrent runners and streams every outcome in stream order as a
// result.Result: nth input yields nth Result, Ok on success and Fail on
// transform failure. A Fail never cancels siblings (collect-all): every
// consumed input yields exactly one Result, so ranging the stream accounts
// for the whole run — the validate-all lesson, never silent, never
// truncated.
//
// Lazy in AND lazy out: calling MapSeqResults starts nothing (a transform
// call counter stays put until the first pull), and pulling sequence values
// plus running transforms happens only while the returned iter.Seq is being
// ranged. The standard library has no lazy fallible parallel stream over
// iter.Seq (iter.Seq is infallible by construction); MapSeqResults fills
// that gap with ordered fan-in instead of a new vocabulary.
//
// A workerCount of zero or less falls back to runtime.NumCPU. A nil sequence
// yields an empty stream without spawning any goroutine (ranging over a nil
// iter.Seq would panic); an already-canceled active context likewise yields
// nothing and starts nothing.
//
// Bounded fan-out plus ordered fan-in through a bounded reorder buffer:
// dispatch plus the reorder window stay O(1) in the sequence length
// (fixed capacity seqResultsReorderCapacity, never tunable), so in-flight
// plus buffered work stays bounded while the stream is consumed. Collecting
// the stream (for example with slices.Collect) is O(n) in the sequence
// length by the caller's explicit decision: unbounded inputs stream, they
// never materialize unless the caller asks.
//
// Cancellation joins everything: active.Done() stops the feeder and the
// workers and joins every spawned goroutine, and a downstream early break
// (yield returning false) cancels the run and joins the feeder plus the
// workers before the range exits — no leaked goroutines on any path.
// Sibling transforms must honor the context for prompt shutdown; a
// transform already in flight still runs to completion (Go cannot preempt
// it), bounded by the runner count.
//
// A panicking transform never escapes its runner goroutine: the first panic
// payload latches, siblings cancel through the same context path, every
// goroutine joins, and MapSeqResults re-panics the original payload in the
// pulling goroutine. A latched panic wins over a latched Fail: programmer
// bugs stay loud. A panic inside the input generator is never caught here:
// workers join through the deferred cleanup and the payload propagates on
// the puller's stack, loud, the same as ranging the sequence directly.
//
// Out of scope (YAGNI): no fail-fast flag or mode on this symbol (fail-fast
// already exists: Map, MapSeq, EachSeq, and range-break already cancels),
// no Filter, no tunable buffers, no unbounded workers.
func MapSeqResults[Input any, Output any](active context.Context, sequence iter.Seq[Input], workerCount int, transform func(context.Context, Input) (Output, error)) iter.Seq[result.Result[Output]] {
	return func(yield func(result.Result[Output]) bool) {
		if active.Err() != nil {
			return
		}
		if sequence == nil {
			return
		}
		runners := workerCount
		if runners <= 0 {
			runners = runtime.NumCPU()
		}
		taskContext, cancelTask := context.WithCancel(active)
		defer cancelTask()
		positions := make(chan seqResultsJob[Input, Output])
		var workGroup sync.WaitGroup
		var panicOnce sync.Once
		var panicMutex sync.Mutex
		var firstPanic any
		var closeOnce sync.Once
		closePositions := func() {
			closeOnce.Do(func() { close(positions) })
		}
		defer func() {
			cancelTask()
			closePositions()
			workGroup.Wait()
		}()
		for range runners {
			workGroup.Add(1)
			go seqResultsWorker(taskContext, positions, transform, &panicOnce, &panicMutex, &firstPanic, cancelTask, &workGroup)
		}
		loadPanic := func() any { return loadSeqResultsPanic(&panicMutex, &firstPanic) }
		if panicPayload := seqResultsDrive(taskContext, sequence, positions, yield, loadPanic); panicPayload != nil {
			panic(panicPayload)
		}
	}
}
