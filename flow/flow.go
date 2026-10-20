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
// Out of scope (YAGNI): no FanIn/FanOut channels, no pipelines, no
// Retry/backoff; those are later slices if ever.
package flow

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
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
// through a WaitGroup on all paths, and no lock
// is held while transform runs.
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
	var failureOnce sync.Once
	var firstFailure error
	for runner := 0; runner < runners; runner++ {
		workGroup.Add(1)
		go func() {
			defer workGroup.Done()
			for position := range positions {
				select {
				case <-taskContext.Done():
					continue
				default:
				}
				transformed, transformFailure := transform(taskContext, inputs[position])
				if transformFailure != nil {
					failureOnce.Do(func() {
						firstFailure = transformFailure
						cancelTask()
					})
					return
				}
				ordered[position] = transformed
			}
		}()
	}
	workGroup.Wait()
	if firstFailure != nil {
		return nil, firstFailure
	}
	return ordered, nil
}

// ForEach applies consume to every input with at most workerCount concurrent
// runners, unordered and without results; the first consume failure cancels
// the rest and wins as the returned failure, nil when all succeed. A
// workerCount of zero or less falls back to runtime.NumCPU. Empty inputs
// return nil without running consume. An already-canceled active context
// returns promptly with active.Err() and starts no consume step.
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
		transformed, transformFailure := transform(taskContext, inputs[position])
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
