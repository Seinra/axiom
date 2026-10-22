package flow_test

import (
	"context"
	"errors"
	"iter"
	"reflect"
	"runtime"
	"slices"
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

func TestForEachCollectsEverySideEffect(t *testing.T) {
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
	if forEachFailure := flow.ForEach(active, inputs, 3, consume); forEachFailure != nil {
		t.Fatalf("ForEach failure = %v, want nil", forEachFailure)
	}
	consumeMutex.Lock()
	defer consumeMutex.Unlock()
	if len(consumed) != len(inputs) {
		t.Fatalf("ForEach consumed count = %d, want %d", len(consumed), len(inputs))
	}
	seen := make(map[int]bool, len(inputs))
	for _, observed := range consumed {
		seen[observed] = true
	}
	for _, candidate := range inputs {
		if !seen[candidate*outputFactor] {
			t.Fatalf("ForEach missing input %d in %v", candidate, consumed)
		}
	}
}

func TestForEachFirstFailureCancelsSiblings(t *testing.T) {
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
	if forEachFailure := flow.ForEach(active, inputs, workerTotal, consume); !errors.Is(forEachFailure, errFastFailure) {
		t.Fatalf("ForEach failure = %v, want %v", forEachFailure, errFastFailure)
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("ForEach started = %d, want <= %d", observed, workerTotal)
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
		if forEachFailure := flow.ForEach(active, inputs, workerCount, consume); forEachFailure != nil {
			t.Fatalf("ForEach(%d) failure = %v, want nil", workerCount, forEachFailure)
		}
		if observed := consumed.Load(); observed != int64(len(inputs)) {
			t.Fatalf("ForEach(%d) consumed = %d, want %d", workerCount, observed, len(inputs))
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
	if forEachFailure := flow.ForEach(active, []int{1, 2, 3}, 2, neverConsumes); !errors.Is(forEachFailure, context.Canceled) {
		t.Fatalf("ForEach failure = %v, want %v", forEachFailure, context.Canceled)
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
		t.Error("ForEach consume ran on empty inputs")
		return nil
	}
	if forEachFailure := flow.ForEach(active, []int{}, 2, neverConsumes); forEachFailure != nil {
		t.Fatalf("ForEach failure = %v, want nil", forEachFailure)
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
	// Champion-latch contract: the fast champion claims the win once, and
	// slower context-ignoring siblings that finish afterward must not
	// overwrite it. Runners exit after their first success, so starts stay
	// bounded by the runner count while in-flight context-ignoring work
	// still runs to completion and unstarted work is skipped via the drain.
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	workerTotal := 4
	inputTotal := 20
	championCandidate := 0
	// championDelay lets every runner start yet stays strictly below
	// siblingDelay, so the champion always wins first and siblings arrive
	// late enough to exercise the latch.
	championDelay := 2 * staggerStep
	siblingDelay := slowDelay
	inputs := make([]int, inputTotal)
	for position := range inputs {
		inputs[position] = position
	}
	var started atomic.Int64
	transform := func(taskContext context.Context, candidate int) (int, error) {
		started.Add(1)
		if candidate == championCandidate {
			// Brief head start for the remaining runners; still strictly
			// first, so sibling completions below become late latch claims.
			time.Sleep(championDelay)
			return championCandidate, nil
		}
		// Ignores taskContext on purpose: proves the champion latch holds
		// and unstarted work is still skipped (via the drain) while
		// in-flight work completes.
		time.Sleep(siblingDelay)
		return candidate * outputFactor, nil
	}
	began := time.Now()
	winner, raceFailure := flow.Race(active, inputs, workerTotal, transform)
	raceLatency := time.Since(began)
	if raceFailure != nil {
		t.Fatalf("Race failure = %v, want nil", raceFailure)
	}
	if winner != championCandidate {
		t.Fatalf("Race winner = %v, want %v", winner, championCandidate)
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("Race started = %d, want <= %d", observed, workerTotal)
	}
	// Prompt return: well above siblingDelay so in-flight work can finish,
	// well below the test timeout so hanging on siblings would fail.
	winLatencyCeiling := 2 * time.Second
	if raceLatency >= winLatencyCeiling {
		t.Fatalf("Race latency = %v, want < %v", raceLatency, winLatencyCeiling)
	}
}

var errPanicSentinel = errors.New("panic sentinel")

func assertPanicPayload(panicked any, want error, t *testing.T) {
	t.Helper()
	payload, holds := panicked.(error)
	if !holds || !errors.Is(payload, want) {
		t.Fatalf("panic = %v, want sentinel %v", panicked, want)
	}
}

func TestMapPanicPropagatesPayload(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{1, 2, 3}
	ran := make(map[int]bool)
	transform := func(taskContext context.Context, candidate int) (int, error) {
		ran[candidate] = true
		if candidate == 2 {
			panic(errPanicSentinel)
		}
		return candidate * outputFactor, nil
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Map(active, inputs, 1, transform)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
	if ran[3] {
		t.Fatal("Map ran input 3 after panic, want stop")
	}
}

func TestMapPanicCancelsSiblings(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	gate := make(chan struct{}) // never closed: progress comes only from cancellation
	sawCancel := make(chan struct{}, 1)
	inputs := []int{0, 1}
	transform := func(taskContext context.Context, candidate int) (string, error) {
		if candidate == 1 {
			panic(errPanicSentinel)
		}
		select {
		case <-gate:
			return "gated", nil
		case <-taskContext.Done():
			sawCancel <- struct{}{}
			return "", taskContext.Err()
		}
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Map(active, inputs, 2, transform)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
	select {
	case <-sawCancel:
	default:
		t.Fatal("Map sibling did not observe cancellation after panic")
	}
}

func TestRaceAllPanicRepanicsPayload(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{1, 2}
	transform := func(taskContext context.Context, candidate int) (int, error) {
		panic(errPanicSentinel)
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Race(active, inputs, 1, transform)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
}

func TestRacePanicBeforeSuccessPoisons(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{1, 2}
	transform := func(taskContext context.Context, candidate int) (int, error) {
		if candidate == 1 {
			panic(errPanicSentinel)
		}
		return candidate * outputFactor, nil
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Race(active, inputs, 1, transform)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
}

func TestRaceSuccessBeforePanicWins(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{1, 2}
	transform := func(taskContext context.Context, candidate int) (int, error) {
		if candidate == 2 {
			panic(errPanicSentinel)
		}
		return candidate * outputFactor, nil
	}
	winner, raceFailure := flow.Race(active, inputs, 1, transform)
	if raceFailure != nil {
		t.Fatalf("Race failure = %v, want nil", raceFailure)
	}
	if winner != outputFactor {
		t.Fatalf("Race winner = %v, want %v", winner, outputFactor)
	}
}

func TestForEachPanicPropagatesPayload(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{1, 2, 3}
	consume := func(taskContext context.Context, candidate int) error {
		if candidate == 2 {
			panic(errPanicSentinel)
		}
		return nil
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_ = flow.ForEach(active, inputs, 1, consume)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
}

var errPanicOther = errors.New("other panic")

func TestMapConcurrentPanicsLatchOne(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	var gate sync.Mutex
	release := make(chan struct{})
	arrived := 0
	barrier := func() {
		gate.Lock()
		arrived++
		if arrived == 2 {
			close(release)
		}
		gate.Unlock()
		<-release
	}
	ran := make(map[int]bool)
	var ranGuard sync.Mutex
	transform := func(taskContext context.Context, candidate int) (int, error) {
		barrier()
		ranGuard.Lock()
		ran[candidate] = true
		ranGuard.Unlock()
		if candidate == 1 {
			panic(errPanicSentinel)
		}
		panic(errPanicOther)
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Map(active, []int{1, 2}, 2, transform)
		return nil
	}()
	payload, holds := panicked.(error)
	if !holds || (!errors.Is(payload, errPanicSentinel) && !errors.Is(payload, errPanicOther)) {
		t.Fatalf("Map panic = %v, want one of the two sentinels", panicked)
	}
	ranGuard.Lock()
	defer ranGuard.Unlock()
	if !ran[1] || !ran[2] {
		t.Fatalf("Map ran = %v, want both inputs attempted", ran)
	}
}

func TestMapPanicWinsOverFailure(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	gate := make(chan struct{}) // never closed: failure path resolves only via cancellation
	inputs := []int{0, 1}
	transform := func(taskContext context.Context, candidate int) (string, error) {
		if candidate == 1 {
			panic(errPanicSentinel)
		}
		select {
		case <-gate:
			return "gated", nil
		case <-taskContext.Done():
			return "", taskContext.Err()
		}
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Map(active, inputs, 2, transform)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
}

func TestRaceFailureThenPanicPoisons(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{1, 2}
	calls := 0
	transform := func(taskContext context.Context, candidate int) (int, error) {
		calls++
		if candidate == 1 {
			return 0, errFastFailure
		}
		panic(errPanicSentinel)
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Race(active, inputs, 1, transform)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
	if calls != 2 {
		t.Fatalf("Race calls = %d, want 2 (failure keeps racing into the panic)", calls)
	}
}

func TestRaceChampionThenPanicPoisons(t *testing.T) {
	// Poison contract: a panic latched while a champion is already claimed
	// still re-panics in the caller, even though the winner stands. The gate
	// makes the interleaving deterministic: the panicker is guaranteed
	// in flight (sleeping) when the champion lands, so scheduling cannot
	// drain the panic away; only the duration depends on slowDelay.
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	started := make(chan struct{}) // closed when the panicker goes in flight
	inputs := []int{1, 2}
	transform := func(taskContext context.Context, candidate int) (int, error) {
		if candidate == 2 {
			close(started)
			time.Sleep(slowDelay)
			panic(errPanicSentinel)
		}
		<-started // champion waits until the panicker is in flight
		return candidate * outputFactor, nil
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Race(active, inputs, 2, transform)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
}

func TestMapNilTransformPanics(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	// Observed contract: a nil transform is a programmer error that panics
	// (nil function call inside the runner latch), never a skip or an error
	// return. Empty inputs short-circuit before any transform runs.
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Map[int, int](active, []int{1, 2, 3}, 2, nil)
		return nil
	}()
	if panicked == nil {
		t.Fatal("Map with nil transform did not panic, want panic (programmer error, never silent)")
	}
	ordered, mapFailure := flow.Map[int, int](active, []int{}, 2, nil)
	if mapFailure != nil {
		t.Fatalf("Map empty with nil transform failure = %v, want nil", mapFailure)
	}
	if ordered == nil || len(ordered) != 0 {
		t.Fatalf("Map empty with nil transform = %v, want empty non-nil slice", ordered)
	}
}

func TestForEachNilConsumePanics(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	// Observed contract: ForEach wraps consume in the Map runner, so a nil
	// consume panics the same way a nil Map transform does. Empty inputs
	// return nil without running consume.
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_ = flow.ForEach(active, []int{1, 2, 3}, 2, nil)
		return nil
	}()
	if panicked == nil {
		t.Fatal("ForEach with nil consume did not panic, want panic (programmer error, never silent)")
	}
	if forEachFailure := flow.ForEach(active, []int{}, 2, nil); forEachFailure != nil {
		t.Fatalf("ForEach empty with nil consume failure = %v, want nil", forEachFailure)
	}
}

func TestRaceNilTransformPanics(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	// Observed contract: a nil transform panics (latched poison, never a
	// silent zero). Empty inputs short-circuit to ErrNoRaceInputs before any
	// transform runs.
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.Race[int, int](active, []int{1, 2, 3}, 2, nil)
		return nil
	}()
	if panicked == nil {
		t.Fatal("Race with nil transform did not panic, want panic (programmer error, never silent)")
	}
	_, raceFailure := flow.Race[int, int](active, []int{}, 2, nil)
	if !errors.Is(raceFailure, flow.ErrNoRaceInputs) {
		t.Fatalf("Race empty with nil transform failure = %v, want ErrNoRaceInputs", raceFailure)
	}
}

func TestMapSeqKeepsInputOrderDespiteStaggeredCompletion(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{0, 1, 2, 3, 4, 5}
	lastPosition := len(inputs) - 1
	sequence := slices.Values(inputs)
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
	ordered, mapFailure := flow.MapSeq(active, sequence, 3, transform)
	if mapFailure != nil {
		t.Fatalf("MapSeq failure = %v, want nil", mapFailure)
	}
	expected := []int{0, 10, 20, 30, 40, 50}
	if !reflect.DeepEqual(ordered, expected) {
		t.Fatalf("MapSeq ordered = %v, want %v", ordered, expected)
	}
}

func TestMapSeqNilSequenceYieldsEmpty(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	// A nil iter.Seq would panic if ranged over, so MapSeq guards it and
	// yields empty without spawning any goroutine.
	var sequence iter.Seq[int]
	if sequence != nil {
		t.Fatal("test setup: sequence is not nil, want nil seq guard pin")
	}
	neverRuns := func(taskContext context.Context, candidate int) (int, error) {
		t.Error("MapSeq transform ran on nil sequence")
		return candidate, nil
	}
	ordered, mapFailure := flow.MapSeq(active, sequence, 2, neverRuns)
	if mapFailure != nil {
		t.Fatalf("MapSeq failure = %v, want nil", mapFailure)
	}
	if ordered == nil {
		t.Fatal("MapSeq ordered = nil, want empty non-nil slice")
	}
	if len(ordered) != 0 {
		t.Fatalf("MapSeq ordered len = %d, want 0", len(ordered))
	}
}

func TestMapSeqEmptySequenceYieldsEmpty(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	neverRuns := func(taskContext context.Context, candidate int) (int, error) {
		t.Error("MapSeq transform ran on empty sequence")
		return candidate, nil
	}
	ordered, mapFailure := flow.MapSeq(active, slices.Values([]int{}), 2, neverRuns)
	if mapFailure != nil {
		t.Fatalf("MapSeq failure = %v, want nil", mapFailure)
	}
	if ordered == nil {
		t.Fatal("MapSeq ordered = nil, want empty non-nil slice")
	}
	if len(ordered) != 0 {
		t.Fatalf("MapSeq ordered len = %d, want 0", len(ordered))
	}
}

func TestMapSeqPreCanceledStartsNothing(t *testing.T) {
	active, cancelTask := context.WithCancel(context.Background())
	cancelTask()
	var started atomic.Int64
	counting := func(taskContext context.Context, candidate int) (int, error) {
		started.Add(1)
		return candidate, nil
	}
	ordered, mapFailure := flow.MapSeq(active, slices.Values([]int{1, 2, 3}), 2, counting)
	if !errors.Is(mapFailure, context.Canceled) {
		t.Fatalf("MapSeq failure = %v, want %v", mapFailure, context.Canceled)
	}
	if ordered != nil {
		t.Fatalf("MapSeq ordered = %v, want nil on canceled context", ordered)
	}
	if observed := started.Load(); observed != 0 {
		t.Fatalf("MapSeq started = %d, want 0", observed)
	}
}

func TestMapSeqMidRunCancelYieldsError(t *testing.T) {
	active, cancelTask := context.WithCancel(context.Background())
	var calls atomic.Int64
	transform := func(taskContext context.Context, candidate int) (int, error) {
		if calls.Add(1) == 1 {
			cancelTask()
		}
		return candidate, nil
	}
	ordered, mapFailure := flow.MapSeq(active, slices.Values([]int{1, 2, 3, 4}), 1, transform)
	if !errors.Is(mapFailure, context.Canceled) {
		t.Fatalf("MapSeq failure = %v, want %v", mapFailure, context.Canceled)
	}
	if ordered != nil {
		t.Fatalf("MapSeq ordered = %v, want nil on mid-run cancel (never partial success)", ordered)
	}
}

func TestMapSeqFirstFailureStopsUnstartedWorkOverLongSequence(t *testing.T) {
	// Fail-fast contract over a long bounded sequence: the head failure
	// cancels the feeder and siblings, so unstarted tail work never runs and
	// the sequence still terminates (no hang on the remaining tail).
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	workerTotal := 4
	longInputTotal := 5000
	longSequence := func(yield func(int) bool) {
		for candidate := 0; candidate < longInputTotal; candidate++ {
			if !yield(candidate) {
				return
			}
		}
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
	ordered, mapFailure := flow.MapSeq(active, longSequence, workerTotal, transform)
	cancelLatency := time.Since(began)
	if !errors.Is(mapFailure, errFastFailure) {
		t.Fatalf("MapSeq failure = %v, want %v", mapFailure, errFastFailure)
	}
	if ordered != nil {
		t.Fatalf("MapSeq ordered len = %d, want nil on failure", len(ordered))
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("MapSeq started = %d, want <= %d", observed, workerTotal)
	}
	noCancelFloor := 2 * time.Second
	if cancelLatency >= noCancelFloor {
		t.Fatalf("MapSeq cancel latency = %v, want < %v", cancelLatency, noCancelFloor)
	}
}

func TestMapSeqInfiniteSequenceFailsFast(t *testing.T) {
	// Unbounded-input lesson pin: an endless sequence still terminates when
	// the head transform fails promptly, because the failure cancels the
	// feeder mid-send. The endless generator is abandoned, never drained.
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	workerTotal := 4
	endless := func(yield func(int) bool) {
		for candidate := 0; ; candidate++ {
			if !yield(candidate) {
				return
			}
		}
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
	_, mapFailure := flow.MapSeq(active, endless, workerTotal, transform)
	cancelLatency := time.Since(began)
	if !errors.Is(mapFailure, errFastFailure) {
		t.Fatalf("MapSeq failure = %v, want %v", mapFailure, errFastFailure)
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("MapSeq started = %d, want <= %d", observed, workerTotal)
	}
	noCancelFloor := 2 * time.Second
	if cancelLatency >= noCancelFloor {
		t.Fatalf("MapSeq cancel latency = %v, want < %v", cancelLatency, noCancelFloor)
	}
}

func TestMapSeqSkipsUnstartedWorkWhenTransformIgnoresContext(t *testing.T) {
	// Drain contract, Seq edition: runners that observe cancellation stop
	// starting new work even when the transform ignores the context, so
	// in-flight work stays bounded by the runner count.
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	workerTotal := 4
	longInputTotal := 200
	longSequence := func(yield func(int) bool) {
		for candidate := 0; candidate < longInputTotal; candidate++ {
			if !yield(candidate) {
				return
			}
		}
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
	_, mapFailure := flow.MapSeq(active, longSequence, workerTotal, ignoring)
	cancelLatency := time.Since(began)
	if !errors.Is(mapFailure, errFastFailure) {
		t.Fatalf("MapSeq failure = %v, want %v", mapFailure, errFastFailure)
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("MapSeq started = %d, want <= %d", observed, workerTotal)
	}
	noHangCeiling := 5 * time.Second
	if cancelLatency >= noHangCeiling {
		t.Fatalf("MapSeq cancel latency = %v, want < %v", cancelLatency, noHangCeiling)
	}
}

func TestMapSeqWorkerCountDefaultsWhenNotPositive(t *testing.T) {
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
		ordered, mapFailure := flow.MapSeq(active, slices.Values(inputs), workerCount, identity)
		if mapFailure != nil {
			t.Fatalf("MapSeq(%d) failure = %v, want nil", workerCount, mapFailure)
		}
		expected := []int{10, 20, 30, 40, 50, 60, 70, 80}
		if !reflect.DeepEqual(ordered, expected) {
			t.Fatalf("MapSeq(%d) ordered = %v, want %v", workerCount, ordered, expected)
		}
	}
}

func TestMapSeqPanicCancelsSiblings(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	gate := make(chan struct{}) // never closed: progress comes only from cancellation
	sawCancel := make(chan struct{}, 1)
	sequence := slices.Values([]int{0, 1})
	transform := func(taskContext context.Context, candidate int) (string, error) {
		if candidate == 1 {
			panic(errPanicSentinel)
		}
		select {
		case <-gate:
			return "gated", nil
		case <-taskContext.Done():
			sawCancel <- struct{}{}
			return "", taskContext.Err()
		}
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.MapSeq(active, sequence, 2, transform)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
	select {
	case <-sawCancel:
	default:
		t.Fatal("MapSeq sibling did not observe cancellation after panic")
	}
}

func TestMapSeqPanicWinsOverFailure(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	gate := make(chan struct{}) // never closed: failure path resolves only via cancellation
	sequence := slices.Values([]int{0, 1})
	transform := func(taskContext context.Context, candidate int) (string, error) {
		if candidate == 1 {
			panic(errPanicSentinel)
		}
		select {
		case <-gate:
			return "gated", nil
		case <-taskContext.Done():
			return "", taskContext.Err()
		}
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_, _ = flow.MapSeq(active, sequence, 2, transform)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
}

func TestEachSeqCollectsEverySideEffect(t *testing.T) {
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
	if eachFailure := flow.EachSeq(active, slices.Values(inputs), 3, consume); eachFailure != nil {
		t.Fatalf("EachSeq failure = %v, want nil", eachFailure)
	}
	consumeMutex.Lock()
	defer consumeMutex.Unlock()
	if len(consumed) != len(inputs) {
		t.Fatalf("EachSeq consumed count = %d, want %d", len(consumed), len(inputs))
	}
	seen := make(map[int]bool, len(inputs))
	for _, observed := range consumed {
		seen[observed] = true
	}
	for _, candidate := range inputs {
		if !seen[candidate*outputFactor] {
			t.Fatalf("EachSeq missing input %d in %v", candidate, consumed)
		}
	}
}

func TestEachSeqFirstFailureCancelsSiblings(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	workerTotal := 4
	inputTotal := 20
	longSequence := func(yield func(int) bool) {
		for candidate := 0; candidate < inputTotal; candidate++ {
			if !yield(candidate) {
				return
			}
		}
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
	if eachFailure := flow.EachSeq(active, longSequence, workerTotal, consume); !errors.Is(eachFailure, errFastFailure) {
		t.Fatalf("EachSeq failure = %v, want %v", eachFailure, errFastFailure)
	}
	if observed := started.Load(); observed > int64(workerTotal) {
		t.Fatalf("EachSeq started = %d, want <= %d", observed, workerTotal)
	}
}

func TestEachSeqPreCanceledStartsNothing(t *testing.T) {
	active, cancelTask := context.WithCancel(context.Background())
	cancelTask()
	var started atomic.Int64
	counting := func(taskContext context.Context, candidate int) error {
		started.Add(1)
		return nil
	}
	if eachFailure := flow.EachSeq(active, slices.Values([]int{1, 2, 3}), 2, counting); !errors.Is(eachFailure, context.Canceled) {
		t.Fatalf("EachSeq failure = %v, want %v", eachFailure, context.Canceled)
	}
	if observed := started.Load(); observed != 0 {
		t.Fatalf("EachSeq started = %d, want 0", observed)
	}
}

func TestEachSeqEmptyAndNilSequencesSucceed(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	neverConsumes := func(taskContext context.Context, candidate int) error {
		t.Error("EachSeq consume ran on empty sequence")
		return nil
	}
	if eachFailure := flow.EachSeq(active, slices.Values([]int{}), 2, neverConsumes); eachFailure != nil {
		t.Fatalf("EachSeq empty failure = %v, want nil", eachFailure)
	}
	var nilSequence iter.Seq[int]
	if eachFailure := flow.EachSeq(active, nilSequence, 2, neverConsumes); eachFailure != nil {
		t.Fatalf("EachSeq nil failure = %v, want nil", eachFailure)
	}
}

func TestEachSeqPanicPropagatesPayload(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	sequence := slices.Values([]int{1, 2, 3})
	consume := func(taskContext context.Context, candidate int) error {
		if candidate == 2 {
			panic(errPanicSentinel)
		}
		return nil
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_ = flow.EachSeq(active, sequence, 1, consume)
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
}

func TestMapSeqResultsLazyUntilPulled(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	var transformCalls atomic.Int64
	var generatorPulls atomic.Int64
	sequence := func(yield func(int) bool) {
		generatorPulls.Add(1)
		if !yield(1) {
			return
		}
		if !yield(2) {
			return
		}
	}
	transform := func(taskContext context.Context, candidate int) (int, error) {
		transformCalls.Add(1)
		return candidate * outputFactor, nil
	}
	streamed := flow.MapSeqResults(active, sequence, 2, transform)
	if observed := transformCalls.Load(); observed != 0 {
		t.Fatalf("MapSeqResults transform calls = %d, want 0 before first pull", observed)
	}
	if observed := generatorPulls.Load(); observed != 0 {
		t.Fatalf("MapSeqResults generator pulls = %d, want 0 before first pull", observed)
	}
	collected := slices.Collect(streamed)
	if len(collected) != 2 {
		t.Fatalf("MapSeqResults collected count = %d, want 2", len(collected))
	}
	for _, outcome := range collected {
		if _, failure := outcome.Unwrap(); failure != nil {
			t.Fatalf("MapSeqResults outcome failure = %v, want nil", failure)
		}
	}
	if observed := transformCalls.Load(); observed != 2 {
		t.Fatalf("MapSeqResults transform calls = %d, want 2 after collect", observed)
	}
}

func TestMapSeqResultsKeepsOrderUnderSkew(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{0, 1, 2, 3, 4, 5}
	lastPosition := len(inputs) - 1
	sequence := slices.Values(inputs)
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
	position := 0
	for outcome := range flow.MapSeqResults(active, sequence, 3, transform) {
		candidate, failure := outcome.Unwrap()
		if failure != nil {
			t.Fatalf("MapSeqResults outcome %d failure = %v, want nil", position, failure)
		}
		expected := position * outputFactor
		if candidate != expected {
			t.Fatalf("MapSeqResults outcome %d = %d, want %d", position, candidate, expected)
		}
		position++
	}
	if position != len(inputs) {
		t.Fatalf("MapSeqResults yielded = %d, want %d", position, len(inputs))
	}
}

func TestMapSeqResultsCollectAllContinuesAfterFail(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{0, 1, 2, 3, 4}
	sequence := slices.Values(inputs)
	var completed atomic.Int64
	transform := func(taskContext context.Context, candidate int) (int, error) {
		defer completed.Add(1)
		if candidate == 2 {
			return 0, errFastFailure
		}
		return candidate * outputFactor, nil
	}
	collected := slices.Collect(flow.MapSeqResults(active, sequence, 3, transform))
	if len(collected) != len(inputs) {
		t.Fatalf("MapSeqResults collected = %d, want %d (every input yields one Result)", len(collected), len(inputs))
	}
	if observed := completed.Load(); observed != int64(len(inputs)) {
		t.Fatalf("MapSeqResults completed = %d, want %d (Fail never cancels siblings)", observed, len(inputs))
	}
	for position, outcome := range collected {
		candidate, failure := outcome.Unwrap()
		if inputs[position] == 2 {
			if !errors.Is(failure, errFastFailure) {
				t.Fatalf("MapSeqResults outcome %d failure = %v, want %v", position, failure, errFastFailure)
			}
			continue
		}
		if failure != nil {
			t.Fatalf("MapSeqResults outcome %d failure = %v, want nil", position, failure)
		}
		if expected := inputs[position] * outputFactor; candidate != expected {
			t.Fatalf("MapSeqResults outcome %d = %d, want %d", position, candidate, expected)
		}
	}
}

func TestMapSeqResultsPreCanceledYieldsEmpty(t *testing.T) {
	active, cancelTask := context.WithCancel(context.Background())
	cancelTask()
	var started atomic.Int64
	counting := func(taskContext context.Context, candidate int) (int, error) {
		started.Add(1)
		return candidate, nil
	}
	baseline := runtime.NumGoroutine()
	count := 0
	for range flow.MapSeqResults(active, slices.Values([]int{1, 2, 3}), 2, counting) {
		count++
	}
	if count != 0 {
		t.Fatalf("MapSeqResults pre-canceled yielded = %d, want 0", count)
	}
	if observed := started.Load(); observed != 0 {
		t.Fatalf("MapSeqResults pre-canceled started = %d, want 0", observed)
	}
	// Without spawning, the goroutine count must not grow: poll briefly to
	// rule out a slow spawn racing the range above.
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > baseline && time.Now().Before(deadline) {
		runtime.Gosched()
	}
	if observed := runtime.NumGoroutine(); observed > baseline {
		t.Fatalf("MapSeqResults pre-canceled goroutines = %d, want <= baseline %d", observed, baseline)
	}
}

func TestMapSeqResultsNilSequenceYieldsEmpty(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	var nilSequence iter.Seq[int]
	if nilSequence != nil {
		t.Fatal("test setup: sequence is not nil, want nil seq guard pin")
	}
	neverRuns := func(taskContext context.Context, candidate int) (int, error) {
		t.Error("MapSeqResults transform ran on nil sequence")
		return candidate, nil
	}
	baseline := runtime.NumGoroutine()
	count := 0
	for range flow.MapSeqResults(active, nilSequence, 2, neverRuns) {
		count++
	}
	if count != 0 {
		t.Fatalf("MapSeqResults nil yielded = %d, want 0", count)
	}
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > baseline && time.Now().Before(deadline) {
		runtime.Gosched()
	}
	if observed := runtime.NumGoroutine(); observed > baseline {
		t.Fatalf("MapSeqResults nil goroutines = %d, want <= baseline %d", observed, baseline)
	}
}

func TestMapSeqResultsEarlyBreakJoinsWorkers(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	gate := make(chan struct{})
	var started atomic.Int64
	transform := func(taskContext context.Context, candidate int) (int, error) {
		started.Add(1)
		if candidate < 2 {
			return candidate, nil
		}
		select {
		case <-gate:
			return candidate, nil
		case <-taskContext.Done():
			return 0, taskContext.Err()
		}
	}
	longSequence := func(yield func(int) bool) {
		for candidate := 0; candidate < 500; candidate++ {
			if !yield(candidate) {
				return
			}
		}
	}
	baseline := runtime.NumGoroutine()
	consumed := 0
	for outcome := range flow.MapSeqResults(active, longSequence, 4, transform) {
		_ = outcome
		consumed++
		if consumed == 2 {
			break
		}
	}
	close(gate)
	if consumed != 2 {
		t.Fatalf("MapSeqResults consumed = %d, want 2 before break", consumed)
	}
	// Workers blocked in transform observe cancellation (break cancels the
	// run) and join: poll until the baseline returns, fail on deadline.
	deadline := time.Now().Add(5 * time.Second)
	for runtime.NumGoroutine() > baseline && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if observed := runtime.NumGoroutine(); observed > baseline {
		t.Fatalf("MapSeqResults goroutines = %d, want <= baseline %d after break", observed, baseline)
	}
}

func TestMapSeqResultsActiveCancelStopsStream(t *testing.T) {
	active, cancelTask := context.WithCancel(context.Background())
	defer cancelTask()
	transform := func(taskContext context.Context, candidate int) (int, error) {
		select {
		case <-taskContext.Done():
			return 0, taskContext.Err()
		case <-time.After(slowDelay):
			return candidate, nil
		}
	}
	longSequence := func(yield func(int) bool) {
		for candidate := 0; candidate < 500; candidate++ {
			if !yield(candidate) {
				return
			}
		}
	}
	baseline := runtime.NumGoroutine()
	go func() {
		time.Sleep(staggerStep)
		cancelTask()
	}()
	consumed := 0
	for range flow.MapSeqResults(active, longSequence, 4, transform) {
		consumed++
	}
	deadline := time.Now().Add(5 * time.Second)
	for runtime.NumGoroutine() > baseline && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if observed := runtime.NumGoroutine(); observed > baseline {
		t.Fatalf("MapSeqResults goroutines = %d, want <= baseline %d after cancel", observed, baseline)
	}
}

func TestMapSeqResultsPanicRepanicsInPuller(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	sequence := slices.Values([]int{1, 2, 3})
	transform := func(taskContext context.Context, candidate int) (int, error) {
		if candidate == 2 {
			panic(errPanicSentinel)
		}
		select {
		case <-taskContext.Done():
			return 0, taskContext.Err()
		default:
			return candidate * outputFactor, nil
		}
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		for range flow.MapSeqResults(active, sequence, 2, transform) {
		}
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
}

func TestMapSeqResultsPanicWinsOverFail(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	gate := make(chan struct{})
	var gateOnce sync.Once
	sequence := slices.Values([]int{0, 1})
	transform := func(taskContext context.Context, candidate int) (string, error) {
		if candidate == 1 {
			gateOnce.Do(func() { close(gate) })
			panic(errPanicSentinel)
		}
		select {
		case <-gate:
			return "", errFastFailure
		case <-taskContext.Done():
			return "", taskContext.Err()
		}
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		for range flow.MapSeqResults(active, sequence, 2, transform) {
		}
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
}

func TestMapSeqResultsGeneratorPanicPropagates(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	baseline := runtime.NumGoroutine()
	sequence := func(yield func(int) bool) {
		if !yield(1) {
			return
		}
		panic(errPanicSentinel)
	}
	transform := func(taskContext context.Context, candidate int) (int, error) {
		select {
		case <-taskContext.Done():
			return 0, taskContext.Err()
		default:
			return candidate, nil
		}
	}
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		for range flow.MapSeqResults(active, sequence, 2, transform) {
		}
		return nil
	}()
	assertPanicPayload(panicked, errPanicSentinel, t)
	deadline := time.Now().Add(5 * time.Second)
	for runtime.NumGoroutine() > baseline && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if observed := runtime.NumGoroutine(); observed > baseline {
		t.Fatalf("MapSeqResults generator-panic goroutines = %d, want <= baseline %d", observed, baseline)
	}
}

func TestMapSeqResultsWorkerCountDefaultsWhenNotPositive(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{1, 2, 3, 4}
	identity := func(taskContext context.Context, candidate int) (int, error) {
		select {
		case <-taskContext.Done():
			return 0, taskContext.Err()
		default:
			return candidate * outputFactor, nil
		}
	}
	for _, workerCount := range []int{0, -3} {
		collected := slices.Collect(flow.MapSeqResults(active, slices.Values(inputs), workerCount, identity))
		if len(collected) != len(inputs) {
			t.Fatalf("MapSeqResults(%d) collected = %d, want %d", workerCount, len(collected), len(inputs))
		}
		for position, outcome := range collected {
			candidate, failure := outcome.Unwrap()
			if failure != nil {
				t.Fatalf("MapSeqResults(%d) outcome %d failure = %v, want nil", workerCount, position, failure)
			}
			if expected := inputs[position] * outputFactor; candidate != expected {
				t.Fatalf("MapSeqResults(%d) outcome %d = %d, want %d", workerCount, position, candidate, expected)
			}
		}
	}
}

func TestMapSeqResultsLargeStreamStaysOrdered(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputTotal := 2000
	longSequence := func(yield func(int) bool) {
		for candidate := 0; candidate < inputTotal; candidate++ {
			if !yield(candidate) {
				return
			}
		}
	}
	identity := func(taskContext context.Context, candidate int) (int, error) {
		select {
		case <-taskContext.Done():
			return 0, taskContext.Err()
		default:
			return candidate, nil
		}
	}
	position := 0
	for outcome := range flow.MapSeqResults(active, longSequence, 8, identity) {
		candidate, failure := outcome.Unwrap()
		if failure != nil {
			t.Fatalf("MapSeqResults outcome %d failure = %v, want nil", position, failure)
		}
		if candidate != position {
			t.Fatalf("MapSeqResults outcome %d = %d, want %d", position, candidate, position)
		}
		position++
	}
	if position != inputTotal {
		t.Fatalf("MapSeqResults yielded = %d, want %d", position, inputTotal)
	}
}

func TestMapSeqResultsExactlyOneResultPerInput(t *testing.T) {
	active, cancelTask := contextWithTestTimeout(context.Background())
	defer cancelTask()
	inputs := []int{0, 1, 2, 3, 4, 5, 6, 7}
	collected := slices.Collect(flow.MapSeqResults(active, slices.Values(inputs), 3, func(taskContext context.Context, candidate int) (int, error) {
		if candidate%3 == 0 {
			return 0, errSlowFailure
		}
		return candidate, nil
	}))
	if len(collected) != len(inputs) {
		t.Fatalf("MapSeqResults collected = %d, want %d", len(collected), len(inputs))
	}
	okCount := 0
	failCount := 0
	for _, outcome := range collected {
		if outcome.IsOk() {
			okCount++
			continue
		}
		failCount++
	}
	if okCount == 0 || failCount == 0 {
		t.Fatalf("MapSeqResults ok = %d fail = %d, want both present", okCount, failCount)
	}
}
