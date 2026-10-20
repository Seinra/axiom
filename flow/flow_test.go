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
