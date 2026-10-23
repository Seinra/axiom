package task_test

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/seinra/axiom/result"
	"github.com/seinra/axiom/task"
)

const (
	taskTestTimeout = 10 * time.Second
	taskAwaiters    = 32
)

func contextWithTaskTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, taskTestTimeout)
}

func waitForTaskSignal(t *testing.T, active context.Context, signal <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-signal:
	case <-active.Done():
		t.Fatalf("task %s did not happen before test timeout", what)
	}
}

func waitForGoroutineBaseline(t *testing.T, baseline int, what string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for runtime.NumGoroutine() > baseline && time.Now().Before(deadline) {
		runtime.Gosched()
	}
	if observed := runtime.NumGoroutine(); observed > baseline {
		t.Fatalf("task %s goroutines = %d, want <= baseline %d", what, observed, baseline)
	}
}

func TestSpawnRunsComputeEagerlyWithoutAwait(t *testing.T) {
	active, cancelTask := contextWithTaskTimeout(context.Background())
	defer cancelTask()
	started := make(chan struct{})
	future := task.Spawn(active, func(_ context.Context) (int, error) {
		close(started)
		return 42, nil
	})
	waitForTaskSignal(t, active, started, "compute start")
	success, failure := future.Await().Unwrap()
	if failure != nil {
		t.Fatalf("Spawn eager Await failure = %v, want nil", failure)
	}
	if success != 42 {
		t.Fatalf("Spawn eager Await success = %d, want 42", success)
	}
}

func TestAwaitSharesSuccessAcrossConcurrentAwaiters(t *testing.T) {
	active, cancelTask := contextWithTaskTimeout(context.Background())
	defer cancelTask()
	started := make(chan struct{})
	release := make(chan struct{})
	future := task.Spawn(active, func(taskContext context.Context) (int, error) {
		close(started)
		select {
		case <-release:
			return 42, nil
		case <-taskContext.Done():
			return 0, taskContext.Err()
		}
	})
	waitForTaskSignal(t, active, started, "compute start")
	outcomes := make(chan int, taskAwaiters)
	var awaitGroup sync.WaitGroup
	for range taskAwaiters {
		awaitGroup.Add(1)
		go func() {
			defer awaitGroup.Done()
			success, failure := future.Await().Unwrap()
			if failure != nil {
				t.Errorf("Await shared success failure = %v, want nil", failure)
				return
			}
			outcomes <- success
		}()
	}
	close(release)
	awaitGroup.Wait()
	close(outcomes)
	shared := 0
	for success := range outcomes {
		shared++
		if success != 42 {
			t.Fatalf("Await shared success = %d, want 42", success)
		}
	}
	if shared != taskAwaiters {
		t.Fatalf("Await shared successes = %d, want %d", shared, taskAwaiters)
	}
}

func TestAwaitSharesFailureAcrossConcurrentAwaiters(t *testing.T) {
	active, cancelTask := contextWithTaskTimeout(context.Background())
	defer cancelTask()
	sentinel := errors.New("shared stop")
	started := make(chan struct{})
	release := make(chan struct{})
	future := task.Spawn(active, func(taskContext context.Context) (int, error) {
		close(started)
		select {
		case <-release:
			return 0, sentinel
		case <-taskContext.Done():
			return 0, taskContext.Err()
		}
	})
	waitForTaskSignal(t, active, started, "compute start")
	failures := make(chan error, taskAwaiters)
	var awaitGroup sync.WaitGroup
	for range taskAwaiters {
		awaitGroup.Add(1)
		go func() {
			defer awaitGroup.Done()
			_, failure := future.Await().Unwrap()
			if failure == nil {
				t.Errorf("Await shared failure = nil, want %v", sentinel)
				return
			}
			failures <- failure
		}()
	}
	close(release)
	awaitGroup.Wait()
	close(failures)
	shared := 0
	for failure := range failures {
		shared++
		if !errors.Is(failure, sentinel) {
			t.Fatalf("Await shared failure = %v, want errors.Is match for sentinel", failure)
		}
	}
	if shared != taskAwaiters {
		t.Fatalf("Await shared failures = %d, want %d", shared, taskAwaiters)
	}
}

func TestSpawnComputeErrorBecomesSharedFailure(t *testing.T) {
	active, cancelTask := contextWithTaskTimeout(context.Background())
	defer cancelTask()
	sentinel := errors.New("compute stop")
	future := task.Spawn(active, func(_ context.Context) (int, error) {
		return 0, sentinel
	})
	_, first := future.Await().Unwrap()
	if !errors.Is(first, sentinel) {
		t.Fatalf("Await compute failure = %v, want errors.Is match for sentinel", first)
	}
	_, second := future.Await().Unwrap()
	if !errors.Is(second, sentinel) {
		t.Fatalf("Await second failure = %v, want errors.Is match for sentinel", second)
	}
}

func TestSpawnErrorPanicBecomesFailureMatchingTry(t *testing.T) {
	active, cancelTask := contextWithTaskTimeout(context.Background())
	defer cancelTask()
	sentinel := errors.New("task error boom")
	future := task.Spawn(active, func(_ context.Context) (int, error) {
		panic(sentinel)
	})
	_, failure := future.Await().Unwrap()
	if failure == nil {
		t.Fatal("Await error-panic failure = nil, want recovered panic")
	}
	if !errors.Is(failure, sentinel) {
		t.Fatalf("Await error-panic failure = %v, want errors.Is match for sentinel", failure)
	}
	_, expected := result.Try(func() int { panic(sentinel) }).Unwrap()
	if failure.Error() != expected.Error() {
		t.Fatalf("Await error-panic failure = %q, want Try shape %q", failure.Error(), expected.Error())
	}
}

func TestSpawnNonErrorPanicBecomesFailureMatchingTry(t *testing.T) {
	active, cancelTask := contextWithTaskTimeout(context.Background())
	defer cancelTask()
	future := task.Spawn(active, func(_ context.Context) (int, error) {
		panic("task plain boom")
	})
	_, failure := future.Await().Unwrap()
	if failure == nil {
		t.Fatal("Await plain-panic failure = nil, want recovered panic")
	}
	if !strings.Contains(failure.Error(), "axiom result: recovered panic: task plain boom") {
		t.Fatalf("Await plain-panic failure = %q, want recovered panic with payload", failure.Error())
	}
	_, expected := result.Try(func() int { panic("task plain boom") }).Unwrap()
	if failure.Error() != expected.Error() {
		t.Fatalf("Await plain-panic failure = %q, want Try shape %q", failure.Error(), expected.Error())
	}
}

func TestAwaitAfterCancelJoinsWithoutLeaks(t *testing.T) {
	baseline := runtime.NumGoroutine()
	active, cancelTask := context.WithCancel(context.Background())
	defer cancelTask()
	computeStopped := errors.New("compute stopped")
	started := make(chan struct{})
	future := task.Spawn(active, func(taskContext context.Context) (int, error) {
		close(started)
		<-taskContext.Done()
		return 0, computeStopped
	})
	timeout, cancelTimeout := context.WithTimeout(context.Background(), taskTestTimeout)
	defer cancelTimeout()
	waitForTaskSignal(t, timeout, started, "compute start")
	cancelTask()
	_, failure := future.Await().Unwrap()
	if failure == nil {
		t.Fatal("Await canceled failure = nil, want context.Canceled")
	}
	if !errors.Is(failure, context.Canceled) {
		t.Fatalf("Await canceled failure = %v, want errors.Is context.Canceled", failure)
	}
	if errors.Is(failure, computeStopped) {
		t.Fatalf("Await canceled failure = %v, want cancellation observed, not compute outcome", failure)
	}
	waitForGoroutineBaseline(t, baseline, "cancel+join", 5*time.Second)
}

func TestSpawnPreCanceledNeverRunsCompute(t *testing.T) {
	baseline := runtime.NumGoroutine()
	active, cancelTask := context.WithCancel(context.Background())
	cancelTask()
	var started atomic.Int64
	future := task.Spawn(active, func(_ context.Context) (int, error) {
		started.Add(1)
		return 1, nil
	})
	_, failure := future.Await().Unwrap()
	if failure == nil {
		t.Fatal("Await pre-canceled failure = nil, want context.Canceled")
	}
	if !errors.Is(failure, context.Canceled) {
		t.Fatalf("Await pre-canceled failure = %v, want errors.Is context.Canceled", failure)
	}
	if observed := started.Load(); observed != 0 {
		t.Fatalf("Spawn pre-canceled started = %d, want 0", observed)
	}
	waitForGoroutineBaseline(t, baseline, "pre-canceled", 2*time.Second)
}

func TestSpawnNilComputePanics(t *testing.T) {
	defer func() {
		payload := recover()
		message, isString := payload.(string)
		if !isString || !strings.Contains(message, "axiom task: Spawn needs a non-nil compute") {
			t.Fatalf("Spawn nil compute panic = %v, want non-nil compute complaint", payload)
		}
	}()
	_ = task.Spawn[int](context.Background(), nil)
}

func TestZeroTaskAwaitPanics(t *testing.T) {
	defer func() {
		payload := recover()
		message, isString := payload.(string)
		if !isString || !strings.Contains(message, "axiom task: Await on zero Task") {
			t.Fatalf("zero Task Await panic = %v, want zero Task complaint", payload)
		}
	}()
	var future task.Task[int]
	_ = future.Await()
}
