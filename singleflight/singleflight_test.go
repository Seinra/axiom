package singleflight_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/seinra/axiom/singleflight"
)

func TestCallConcurrentHammerSharesSingleExecution(t *testing.T) {
	t.Parallel()
	const workerCount = 32
	const expectedOutcome = 42
	const holdOpen = 50 * time.Millisecond
	var flight singleflight.Group[string, int]
	var executions atomic.Int64
	starter := make(chan struct{})
	outcomes := make([]int, workerCount)
	sharedFlags := make([]bool, workerCount)
	var writeGuard sync.Mutex
	var waiterGroup sync.WaitGroup
	for workerIdx := 0; workerIdx < workerCount; workerIdx++ {
		waiterGroup.Add(1)
		go func(position int) {
			defer waiterGroup.Done()
			<-starter
			outcome, failure, isShared := flight.Call("hammer", func() (int, error) {
				executions.Add(1)
				time.Sleep(holdOpen)
				return expectedOutcome, nil
			})
			if failure != nil {
				t.Errorf("worker %d: failure = %v, want nil", position, failure)
				return
			}
			writeGuard.Lock()
			outcomes[position] = outcome
			sharedFlags[position] = isShared
			writeGuard.Unlock()
		}(workerIdx)
	}
	close(starter)
	waiterGroup.Wait()
	if got := executions.Load(); got != 1 {
		t.Fatalf("executions = %d, want 1", got)
	}
	for workerIdx, outcome := range outcomes {
		if outcome != expectedOutcome {
			t.Fatalf("outcomes[%d] = %d, want %d", workerIdx, outcome, expectedOutcome)
		}
	}
	sharedCount := 0
	for _, isShared := range sharedFlags {
		if isShared {
			sharedCount++
		}
	}
	if sharedCount < workerCount-1 {
		t.Fatalf("sharedCount = %d, want at least %d", sharedCount, workerCount-1)
	}
}

func TestCallDistinctKeysRunIndependently(t *testing.T) {
	t.Parallel()
	requestKeys := []string{"north", "south", "east", "west"}
	var flight singleflight.Group[string, int]
	var executions atomic.Int64
	outcomes := make([]int, len(requestKeys))
	var writeGuard sync.Mutex
	var waiterGroup sync.WaitGroup
	for position := range requestKeys {
		waiterGroup.Add(1)
		go func(slot int) {
			defer waiterGroup.Done()
			outcome, failure, _ := flight.Call(requestKeys[slot], func() (int, error) {
				executions.Add(1)
				time.Sleep(20 * time.Millisecond)
				return slot*100 + 7, nil
			})
			if failure != nil {
				t.Errorf("slot %d: failure = %v, want nil", slot, failure)
				return
			}
			writeGuard.Lock()
			outcomes[slot] = outcome
			writeGuard.Unlock()
		}(position)
	}
	waiterGroup.Wait()
	if got := executions.Load(); got != int64(len(requestKeys)) {
		t.Fatalf("executions = %d, want %d", got, len(requestKeys))
	}
	for slot, outcome := range outcomes {
		if outcome != slot*100+7 {
			t.Fatalf("outcomes[%d] = %d, want %d", slot, outcome, slot*100+7)
		}
	}
}

func TestCallErrorPropagatesToWaiters(t *testing.T) {
	t.Parallel()
	const workerCount = 8
	boomFailure := errors.New("axiom singleflight: boom")
	var flight singleflight.Group[string, int]
	starter := make(chan struct{})
	failures := make([]error, workerCount)
	sharedFlags := make([]bool, workerCount)
	var writeGuard sync.Mutex
	var waiterGroup sync.WaitGroup
	for workerIdx := 0; workerIdx < workerCount; workerIdx++ {
		waiterGroup.Add(1)
		go func(position int) {
			defer waiterGroup.Done()
			<-starter
			_, failure, isShared := flight.Call("fragile", func() (int, error) {
				time.Sleep(30 * time.Millisecond)
				return 0, boomFailure
			})
			writeGuard.Lock()
			failures[position] = failure
			sharedFlags[position] = isShared
			writeGuard.Unlock()
		}(workerIdx)
	}
	close(starter)
	waiterGroup.Wait()
	for position, failure := range failures {
		if !errors.Is(failure, boomFailure) {
			t.Fatalf("failures[%d] = %v, want boomFailure", position, failure)
		}
	}
	sharedCount := 0
	for _, isShared := range sharedFlags {
		if isShared {
			sharedCount++
		}
	}
	if sharedCount < workerCount-1 {
		t.Fatalf("sharedCount = %d, want at least %d", sharedCount, workerCount-1)
	}
}

// waiterAttachWindow bounds the waiter's attach (a map lookup plus a channel
// wait, microsecond-scale) before Forget runs. Not a rendezvous: black-box
// callers cannot observe attach, so a generous fixed window stands in; a
// miss only flakes the run, quarantined per VALIDATION_GATES, never a wrong
// verdict. A true rendezvous would need a library hook, declined: no API
// surface exists for tests alone.
const waiterAttachWindow = 50 * time.Millisecond

func TestForgetDuringFlightForcesRerun(t *testing.T) {
	t.Parallel()
	var flight singleflight.Group[string, int]
	var executions atomic.Int64
	firstRelease := make(chan struct{})
	produce := func() (int, error) {
		turnNumber := executions.Add(1)
		if turnNumber == 1 {
			<-firstRelease
			return 11, nil
		}
		return 22, nil
	}
	leaderDone := make(chan int, 1)
	go func() {
		outcome, _, _ := flight.Call("lonely", produce)
		leaderDone <- outcome
	}()
	awaitFirstExecution(t, &executions)
	waiterDone := make(chan int, 1)
	go func() {
		outcome, _, _ := flight.Call("lonely", produce)
		waiterDone <- outcome
	}()
	time.Sleep(waiterAttachWindow)
	flight.Forget("lonely")
	freshOutcome, failure, _ := flight.Call("lonely", produce)
	if failure != nil {
		t.Fatalf("fresh failure = %v, want nil", failure)
	}
	close(firstRelease)
	leaderOutcome := <-leaderDone
	waiterOutcome := <-waiterDone
	if got := executions.Load(); got != 2 {
		t.Fatalf("executions = %d, want 2", got)
	}
	if freshOutcome != 22 {
		t.Fatalf("freshOutcome = %d, want 22", freshOutcome)
	}
	if leaderOutcome != 11 || waiterOutcome != 11 {
		t.Fatalf("leaderOutcome = %d waiterOutcome = %d, want both 11", leaderOutcome, waiterOutcome)
	}
}

func TestCallSequentialCallsRerun(t *testing.T) {
	t.Parallel()
	var flight singleflight.Group[string, int]
	var executions atomic.Int64
	produce := func() (int, error) {
		return int(executions.Add(1)), nil
	}
	firstOutcome, firstFailure, firstShared := flight.Call("chain", produce)
	if firstFailure != nil {
		t.Fatalf("first failure = %v, want nil", firstFailure)
	}
	secondOutcome, secondFailure, secondShared := flight.Call("chain", produce)
	if secondFailure != nil {
		t.Fatalf("second failure = %v, want nil", secondFailure)
	}
	if got := executions.Load(); got != 2 {
		t.Fatalf("executions = %d, want 2", got)
	}
	if firstOutcome != 1 || secondOutcome != 2 {
		t.Fatalf("outcomes = %d, %d, want 1, 2", firstOutcome, secondOutcome)
	}
	if firstShared || secondShared {
		t.Fatalf("shared = %v, %v, want false, false", firstShared, secondShared)
	}
}

func TestCallPanicPropagatesToWaiters(t *testing.T) {
	t.Parallel()
	const workerCount = 8
	const blastPayload = "axiom singleflight: blast"
	var flight singleflight.Group[string, int]
	var executions atomic.Int64
	starter := make(chan struct{})
	recovered := make([]any, workerCount)
	var writeGuard sync.Mutex
	var waiterGroup sync.WaitGroup
	for workerIdx := 0; workerIdx < workerCount; workerIdx++ {
		waiterGroup.Add(1)
		go func(position int) {
			defer waiterGroup.Done()
			defer func() {
				writeGuard.Lock()
				recovered[position] = recover()
				writeGuard.Unlock()
			}()
			<-starter
			_, _, _ = flight.Call("volatile", func() (int, error) {
				executions.Add(1)
				time.Sleep(50 * time.Millisecond)
				panic(blastPayload)
			})
		}(workerIdx)
	}
	close(starter)
	waiterGroup.Wait()
	if got := executions.Load(); got != 1 {
		t.Fatalf("executions = %d, want 1", got)
	}
	for position, payload := range recovered {
		if payload != blastPayload {
			t.Fatalf("recovered[%d] = %v, want %q", position, payload, blastPayload)
		}
	}
}

func TestStatsSingleCallSnapshot(t *testing.T) {
	t.Parallel()
	var flight singleflight.Group[string, int]
	if fresh := flight.Stats(); fresh.Calls != 0 || fresh.Shared != 0 || fresh.Panics != 0 {
		t.Fatalf("fresh Stats = %+v, want zero value", fresh)
	}
	outcome, failure, isShared := flight.Call("solo", func() (int, error) {
		return 7, nil
	})
	if failure != nil {
		t.Fatalf("failure = %v, want nil", failure)
	}
	if outcome != 7 || isShared {
		t.Fatalf("outcome = %d isShared = %v, want 7 false", outcome, isShared)
	}
	if stats := flight.Stats(); stats.Calls != 1 || stats.Shared != 0 || stats.Panics != 0 {
		t.Fatalf("Stats = %+v, want {Calls:1 Shared:0 Panics:0}", stats)
	}
}

func TestStatsConcurrentSameKeyCountsShared(t *testing.T) {
	t.Parallel()
	const workerCount = 8
	const expectedOutcome = 7
	var flight singleflight.Group[string, int]
	var executions atomic.Int64
	starter := make(chan struct{})
	release := make(chan struct{})
	sharedResults := make(chan bool, workerCount)
	var waiterGroup sync.WaitGroup
	for workerIdx := 0; workerIdx < workerCount; workerIdx++ {
		waiterGroup.Add(1)
		go func(position int) {
			defer waiterGroup.Done()
			<-starter
			outcome, failure, isShared := flight.Call("shared-key", func() (int, error) {
				executions.Add(1)
				<-release
				return expectedOutcome, nil
			})
			if failure != nil {
				t.Errorf("worker %d: failure = %v, want nil", position, failure)
				return
			}
			if outcome != expectedOutcome {
				t.Errorf("worker %d: outcome = %d, want %d", position, outcome, expectedOutcome)
				return
			}
			sharedResults <- isShared
		}(workerIdx)
	}
	close(starter)
	awaitFirstExecution(t, &executions)
	time.Sleep(waiterAttachWindow)
	close(release)
	waiterGroup.Wait()
	close(sharedResults)
	if got := executions.Load(); got != 1 {
		t.Fatalf("executions = %d, want 1", got)
	}
	reported := 0
	flagShared := 0
	for isShared := range sharedResults {
		reported++
		if isShared {
			flagShared++
		}
	}
	if reported != workerCount {
		t.Fatalf("reported = %d, want %d", reported, workerCount)
	}
	stats := flight.Stats()
	if stats.Calls != workerCount {
		t.Fatalf("Stats.Calls = %d, want %d", stats.Calls, workerCount)
	}
	if stats.Shared == 0 {
		t.Fatalf("Stats.Shared = 0, want at least 1")
	}
	if stats.Shared != uint64(flagShared) {
		t.Fatalf("Stats.Shared = %d, isShared flags count = %d, want match", stats.Shared, flagShared)
	}
	if stats.Panics != 0 {
		t.Fatalf("Stats.Panics = %d, want 0", stats.Panics)
	}
}

func TestStatsPanicCountsExecutionOnce(t *testing.T) {
	t.Parallel()
	const workerCount = 8
	const blastPayload = "axiom singleflight: blast"
	var flight singleflight.Group[string, int]
	var executions atomic.Int64
	starter := make(chan struct{})
	release := make(chan struct{})
	recovered := make([]any, workerCount)
	var writeGuard sync.Mutex
	var waiterGroup sync.WaitGroup
	for workerIdx := 0; workerIdx < workerCount; workerIdx++ {
		waiterGroup.Add(1)
		go func(position int) {
			defer waiterGroup.Done()
			defer func() {
				writeGuard.Lock()
				recovered[position] = recover()
				writeGuard.Unlock()
			}()
			<-starter
			_, _, _ = flight.Call("volatile", func() (int, error) {
				executions.Add(1)
				<-release
				panic(blastPayload)
			})
		}(workerIdx)
	}
	close(starter)
	awaitFirstExecution(t, &executions)
	time.Sleep(waiterAttachWindow)
	close(release)
	waiterGroup.Wait()
	if got := executions.Load(); got != 1 {
		t.Fatalf("executions = %d, want 1", got)
	}
	for position, payload := range recovered {
		if payload != blastPayload {
			t.Fatalf("recovered[%d] = %v, want %q", position, payload, blastPayload)
		}
	}
	stats := flight.Stats()
	if stats.Calls != workerCount {
		t.Fatalf("Stats.Calls = %d, want %d", stats.Calls, workerCount)
	}
	if stats.Panics != 1 {
		t.Fatalf("Stats.Panics = %d, want 1", stats.Panics)
	}
}

func awaitFirstExecution(testingT *testing.T, executions *atomic.Int64) {
	testingT.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for executions.Load() < 1 {
		if time.Now().After(deadline) {
			testingT.Fatal("leader never started execution")
		}
		time.Sleep(time.Millisecond)
	}
}
