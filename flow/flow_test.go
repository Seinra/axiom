package flow_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/seinra/axiom/flow"
)

const (
	testTimeout  = 10 * time.Second
	staggerStep  = 50 * time.Millisecond
	slowDelay    = 500 * time.Millisecond
	outputFactor = 10
)

var errFastFailure = errors.New("fast failure")

var errSlowFailure = errors.New("slow failure")

func contextWithTestTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, testTimeout)
}

func TestMapKeepsInputOrderDespiteStaggeredCompletion(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{0, 1, 2, 3, 4, 5}
	lastPosition := len(inputs) - 1
	transform := func(taskContext context.Context, candidate int) (int, error) {
		stagger := time.Duration(lastPosition-candidate) * staggerStep
		staggerTimer := time.NewTimer(stagger)
		defer staggerTimer.Stop()
		select {
		case <-taskContext.Done():
			return 0, taskContext.Err()
		case <-staggerTimer.C:
			return candidate * outputFactor, nil
		}
	}
	ordered, mapFailure := flow.Map(active, inputs, 3, transform)
	if mapFailure != nil {
		t.Fatalf("Map failure = %v, want nil", mapFailure)
	}
	expected := []int{0, 10, 20, 30, 40, 50}
	if !reflect.DeepEqual(ordered, expected) {
		t.Fatalf("Map ordered = %v, want %v", ordered, expected)
	}
}

func TestMapFirstFailureCancelsSiblings(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	workerTotal := 4
	inputTotal := 20
	inputs := make([]int, inputTotal)
	for position := range inputs {
		inputs[position] = position
	}
	var started atomic.Int64
	transform := func(taskContext context.Context, candidate int) (int, error) {
		started.Add(1)
		if candidate == 0 {
			return 0, errFastFailure
		}
		slowTimer := time.NewTimer(slowDelay)
		defer slowTimer.Stop()
		select {
		case <-taskContext.Done():
			return 0, taskContext.Err()
		case <-slowTimer.C:
			return candidate, nil
		}
	}
	began := time.Now()
	_, mapFailure := flow.Map(active, inputs, workerTotal, transform)
	cancelLatency := time.Since(began)
	if !errors.Is(mapFailure, errFastFailure) {
		t.Fatalf("Map failure = %v, want %v", mapFailure, errFastFailure)
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("Map started = %d, want <= %d", observed, workerTotal)
	}
	noCancelFloor := 2 * time.Second
	if cancelLatency >= noCancelFloor {
		t.Fatalf("Map cancel latency = %v, want < %v", cancelLatency, noCancelFloor)
	}
}

func TestEachCollectsEverySideEffect(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{1, 2, 3, 4, 5, 6, 7, 8}
	var consumeMutex sync.Mutex
	consumed := make([]int, 0, len(inputs))
	consume := func(taskContext context.Context, candidate int) error {
		select {
		case <-taskContext.Done():
			return taskContext.Err()
		default:
		}
		consumeMutex.Lock()
		consumed = append(consumed, candidate*outputFactor)
		consumeMutex.Unlock()
		return nil
	}
	if eachFailure := flow.Each(active, inputs, 3, consume); eachFailure != nil {
		t.Fatalf("Each failure = %v, want nil", eachFailure)
	}
	consumeMutex.Lock()
	defer consumeMutex.Unlock()
	if len(consumed) != len(inputs) {
		t.Fatalf("Each consumed count = %d, want %d", len(consumed), len(inputs))
	}
	seen := make(map[int]bool, len(inputs))
	for _, observed := range consumed {
		seen[observed] = true
	}
	for _, candidate := range inputs {
		if !seen[candidate*outputFactor] {
			t.Fatalf("Each missing input %d in %v", candidate, consumed)
		}
	}
}

func TestEachFirstFailureCancelsSiblings(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	workerTotal := 4
	inputTotal := 20
	inputs := make([]int, inputTotal)
	for position := range inputs {
		inputs[position] = position
	}
	var started atomic.Int64
	consume := func(taskContext context.Context, candidate int) error {
		started.Add(1)
		if candidate == 0 {
			return errFastFailure
		}
		slowTimer := time.NewTimer(slowDelay)
		defer slowTimer.Stop()
		select {
		case <-taskContext.Done():
			return taskContext.Err()
		case <-slowTimer.C:
			return nil
		}
	}
	if eachFailure := flow.Each(active, inputs, workerTotal, consume); !errors.Is(eachFailure, errFastFailure) {
		t.Fatalf("Each failure = %v, want %v", eachFailure, errFastFailure)
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("Each started = %d, want <= %d", observed, workerTotal)
	}
}

func TestMapSkipsUnstartedWorkAfterFailure(t *testing.T) {
	// Drain contract: after the first failure, runners stop starting new
	// work even when the transform ignores the context. In-flight
	// context-ignoring transforms still run to completion, so extra work is
	// bounded by the runner count and Map returns without hanging.
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	workerTotal := 4
	inputTotal := 50
	inputs := make([]int, inputTotal)
	for position := range inputs {
		inputs[position] = position
	}
	var started atomic.Int64
	ignoring := func(taskContext context.Context, candidate int) (int, error) {
		started.Add(1)
		if candidate == 0 {
			return 0, errFastFailure
		}
		// Ignores taskContext on purpose: proves unstarted work is still
		// skipped (via the drain) while in-flight work completes.
		time.Sleep(slowDelay)
		return candidate, nil
	}
	began := time.Now()
	_, mapFailure := flow.Map(active, inputs, workerTotal, ignoring)
	cancelLatency := time.Since(began)
	if !errors.Is(mapFailure, errFastFailure) {
		t.Fatalf("Map failure = %v, want %v", mapFailure, errFastFailure)
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("Map started = %d, want <= %d", observed, workerTotal)
	}
	noHangCeiling := 5 * time.Second
	if cancelLatency >= noHangCeiling {
		t.Fatalf("Map cancel latency = %v, want < %v", cancelLatency, noHangCeiling)
	}
}

func TestWorkerCountDefaultsWhenNotPositive(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{1, 2, 3, 4, 5, 6, 7, 8}
	identity := func(taskContext context.Context, candidate int) (int, error) {
		select {
		case <-taskContext.Done():
			return 0, taskContext.Err()
		default:
			return candidate * outputFactor, nil
		}
	}
	for _, workerCount := range []int{0, -3} {
		ordered, mapFailure := flow.Map(active, inputs, workerCount, identity)
		if mapFailure != nil {
			t.Fatalf("Map(%d) failure = %v, want nil", workerCount, mapFailure)
		}
		expected := []int{10, 20, 30, 40, 50, 60, 70, 80}
		if !reflect.DeepEqual(ordered, expected) {
			t.Fatalf("Map(%d) ordered = %v, want %v", workerCount, ordered, expected)
		}
		var consumed atomic.Int64
		consume := func(taskContext context.Context, candidate int) error {
			select {
			case <-taskContext.Done():
				return taskContext.Err()
			default:
				consumed.Add(1)
				return nil
			}
		}
		if eachFailure := flow.Each(active, inputs, workerCount, consume); eachFailure != nil {
			t.Fatalf("Each(%d) failure = %v, want nil", workerCount, eachFailure)
		}
		if observed := consumed.Load(); observed != int64(len(inputs)) {
			t.Fatalf("Each(%d) consumed = %d, want %d", workerCount, observed, len(inputs))
		}
	}
}

func TestPreCanceledContextReturnsCanceled(t *testing.T) {
	active, cancelTask := context.WithCancel(context.Background())
	cancelTask()
	neverRuns := func(taskContext context.Context, candidate int) (int, error) {
		return candidate, nil
	}
	began := time.Now()
	_, mapFailure := flow.Map(active, []int{1, 2, 3}, 2, neverRuns)
	mapLatency := time.Since(began)
	if !errors.Is(mapFailure, context.Canceled) {
		t.Fatalf("Map failure = %v, want %v", mapFailure, context.Canceled)
	}
	promptCeiling := 2 * time.Second
	if mapLatency >= promptCeiling {
		t.Fatalf("Map cancel latency = %v, want < %v", mapLatency, promptCeiling)
	}
	neverConsumes := func(taskContext context.Context, candidate int) error {
		return nil
	}
	if eachFailure := flow.Each(active, []int{1, 2, 3}, 2, neverConsumes); !errors.Is(eachFailure, context.Canceled) {
		t.Fatalf("Each failure = %v, want %v", eachFailure, context.Canceled)
	}
}

func TestPreCanceledContextStartsNothing(t *testing.T) {
	active, cancelTask := context.WithCancel(context.Background())
	cancelTask()
	var started atomic.Int64
	counting := func(taskContext context.Context, candidate int) (int, error) {
		started.Add(1)
		return candidate, nil
	}
	if _, mapFailure := flow.Map(active, []int{1, 2, 3}, 2, counting); !errors.Is(mapFailure, context.Canceled) {
		t.Fatalf("Map failure = %v, want %v", mapFailure, context.Canceled)
	}
	if observed := started.Load(); observed != 0 {
		t.Fatalf("Map started = %d, want 0", observed)
	}
}

func TestEmptyInputs(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	neverRuns := func(taskContext context.Context, candidate int) (int, error) {
		t.Error("Map transform ran on empty inputs")
		return candidate, nil
	}
	ordered, mapFailure := flow.Map(active, []int{}, 2, neverRuns)
	if mapFailure != nil {
		t.Fatalf("Map failure = %v, want nil", mapFailure)
	}
	if ordered == nil {
		t.Fatal("Map ordered = nil, want empty non-nil slice")
	}
	if len(ordered) != 0 {
		t.Fatalf("Map ordered len = %d, want 0", len(ordered))
	}
	neverConsumes := func(taskContext context.Context, candidate int) error {
		t.Error("Each consume ran on empty inputs")
		return nil
	}
	if eachFailure := flow.Each(active, []int{}, 2, neverConsumes); eachFailure != nil {
		t.Fatalf("Each failure = %v, want nil", eachFailure)
	}
}

func TestRace(t *testing.T) {
	// startedCeiling bounds transforms started (-1 skips); startedExact pins
	// the exact count (-1 skips). Every row asserts prompt latency below the
	// shared ceiling, mirroring the Map cancellation proofs.
	raceCases := []struct {
		name           string
		inputs         []int
		workerCount    int
		preCanceled    bool
		delays         map[int]time.Duration
		fallbackDelay  time.Duration
		failures       map[int]error
		wantWinner     int
		wantFailure    error
		startedCeiling int64
		startedExact   int64
	}{
		{
			name:        "staggered first success wins",
			inputs:      []int{0, 1, 2, 3, 4, 5},
			workerCount: 6,
			delays: map[int]time.Duration{
				0: 5 * staggerStep,
				1: 4 * staggerStep,
				2: 3 * staggerStep,
				3: 2 * staggerStep,
				4: staggerStep,
			},
			wantWinner:     5 * outputFactor,
			startedCeiling: 6,
			startedExact:   -1,
		},
		{
			name:        "all fail returns first failure",
			inputs:      []int{0, 1, 2, 3},
			workerCount: 2,
			delays: map[int]time.Duration{
				1: staggerStep,
				2: 2 * staggerStep,
				3: 3 * staggerStep,
			},
			failures: map[int]error{
				0: errFastFailure,
				1: errSlowFailure,
				2: errSlowFailure,
				3: errSlowFailure,
			},
			wantWinner:     0,
			wantFailure:    errFastFailure,
			startedCeiling: -1,
			// Failures never cancel siblings, so every input still races.
			startedExact: 4,
		},
		{
			name:           "pre-canceled context reports promptly",
			inputs:         []int{1, 2, 3},
			workerCount:    2,
			preCanceled:    true,
			wantWinner:     0,
			wantFailure:    context.Canceled,
			startedCeiling: -1,
			startedExact:   0,
		},
		{
			name:           "empty inputs yield sentinel failure",
			inputs:         []int{},
			workerCount:    2,
			wantWinner:     0,
			wantFailure:    flow.ErrNoRaceInputs,
			startedCeiling: -1,
			startedExact:   0,
		},
		{
			name:           "fast success cancels slow siblings",
			inputs:         []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
			workerCount:    4,
			delays:         map[int]time.Duration{0: 0},
			fallbackDelay:  slowDelay,
			wantWinner:     0,
			startedCeiling: 4,
			startedExact:   -1,
		},
		{
			name:           "zero worker count falls back",
			inputs:         []int{7},
			workerCount:    0,
			wantWinner:     7 * outputFactor,
			startedCeiling: -1,
			startedExact:   1,
		},
		{
			name:           "negative worker count falls back",
			inputs:         []int{7},
			workerCount:    -3,
			wantWinner:     7 * outputFactor,
			startedCeiling: -1,
			startedExact:   1,
		},
		{
			name:           "zero value win stays a success",
			inputs:         []int{0},
			workerCount:    2,
			wantWinner:     0,
			startedCeiling: -1,
			startedExact:   1,
		},
	}
	for _, raceCase := range raceCases {
		t.Run(raceCase.name, func(t *testing.T) {
			var active context.Context
			var cancelTask context.CancelFunc
			if raceCase.preCanceled {
				active, cancelTask = context.WithCancel(context.Background())
				cancelTask()
			} else {
				active, cancelTask = contextWithTestTimeout(context.Background())
				defer cancelTask()
			}
			var started atomic.Int64
			transform := func(taskContext context.Context, candidate int) (int, error) {
				started.Add(1)
				stagger, timed := raceCase.delays[candidate]
				if !timed {
					stagger = raceCase.fallbackDelay
				}
				if stagger > 0 {
					staggerTimer := time.NewTimer(stagger)
					defer staggerTimer.Stop()
					select {
					case <-taskContext.Done():
						return 0, taskContext.Err()
					case <-staggerTimer.C:
					}
				} else {
					select {
					case <-taskContext.Done():
						return 0, taskContext.Err()
					default:
					}
				}
				if transformFailure, failed := raceCase.failures[candidate]; failed {
					return 0, transformFailure
				}
				return candidate * outputFactor, nil
			}
			began := time.Now()
			winner, raceFailure := flow.Race(active, raceCase.inputs, raceCase.workerCount, transform)
			raceLatency := time.Since(began)
			if raceCase.wantFailure == nil {
				if raceFailure != nil {
					t.Fatalf("Race failure = %v, want nil", raceFailure)
				}
			} else if !errors.Is(raceFailure, raceCase.wantFailure) {
				t.Fatalf("Race failure = %v, want %v", raceFailure, raceCase.wantFailure)
			}
			if winner != raceCase.wantWinner {
				t.Fatalf("Race winner = %v, want %v", winner, raceCase.wantWinner)
			}
			if raceCase.startedCeiling >= 0 {
				if observed := started.Load(); observed > raceCase.startedCeiling {
					t.Fatalf("Race started = %d, want <= %d", observed, raceCase.startedCeiling)
				}
			}
			if raceCase.startedExact >= 0 {
				if observed := started.Load(); observed != raceCase.startedExact {
					t.Fatalf("Race started = %d, want %d", observed, raceCase.startedExact)
				}
			}
			promptCeiling := 2 * time.Second
			if raceLatency >= promptCeiling {
				t.Fatalf("Race latency = %v, want < %v", raceLatency, promptCeiling)
			}
		})
	}
}

func TestRaceFirstWinSticksWithContextIgnoringSiblings(t *testing.T) {
	// Champion-latch contract: a fast success claims the championship once,
	// and slower context-ignoring siblings that complete after the win
	// return without overwriting it. Every runner exits after its first
	// success, so starts stay bounded by the runner count while in-flight
	// context-ignoring transforms still run to completion.
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	workerTotal := 4
	inputTotal := 20
	inputs := make([]int, inputTotal)
	for position := range inputs {
		inputs[position] = position
	}
	var started atomic.Int64
	transform := func(taskContext context.Context, candidate int) (int, error) {
		started.Add(1)
		if candidate == 0 {
			// Short delay so every runner has started before the win lands;
			// the win is still strictly first, forcing late claims from the
			// context-ignoring siblings below.
			time.Sleep(2 * staggerStep)
			return 0, nil
		}
		// Ignores taskContext on purpose: proves the champion latch holds
		// and unstarted work is still skipped (via the drain) while
		// in-flight work completes.
		time.Sleep(slowDelay)
		return candidate * outputFactor, nil
	}
	began := time.Now()
	winner, raceFailure := flow.Race(active, inputs, workerTotal, transform)
	raceLatency := time.Since(began)
	if raceFailure != nil {
		t.Fatalf("Race failure = %v, want nil", raceFailure)
	}
	if winner != 0 {
		t.Fatalf("Race winner = %v, want 0", winner)
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("Race started = %d, want <= %d", observed, workerTotal)
	}
	promptCeiling := 2 * time.Second
	if raceLatency >= promptCeiling {
		t.Fatalf("Race latency = %v, want < %v", raceLatency, promptCeiling)
	}
}
