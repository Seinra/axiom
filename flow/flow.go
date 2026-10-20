// Package flow provides bounded parallel combinators with errgroup semantics.
//
// Map runs a transform over inputs with at most workerCount concurrent
// runners and reports results in input order. Each applies a consume step
// for side effects without collecting results. The first transform failure
// cancels its siblings through context cancellation and wins as the
// returned failure, matching errgroup parity: one failure in, that failure
// out, siblings stopped.
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

// Each applies consume to every input with at most workerCount concurrent
// runners, unordered and without results; the first consume failure cancels
// the rest and wins as the returned failure, nil when all succeed. A
// workerCount of zero or less falls back to runtime.NumCPU. Empty inputs
// return nil without running consume. An already-canceled active context
// returns promptly with active.Err() and starts no consume step.
func Each[Input any](active context.Context, inputs []Input, workerCount int, consume func(context.Context, Input) error) error {
	_, mapFailure := Map(active, inputs, workerCount, func(taskContext context.Context, candidate Input) (struct{}, error) {
		consumeFailure := consume(taskContext, candidate)
		return struct{}{}, consumeFailure
	})
	return mapFailure
}
