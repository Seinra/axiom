package flow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/seinra/axiom/flow"
)

func TestRetrySucceedsFirstTry(t *testing.T) {
	calls := 0
	err := flow.Retry(context.Background(), 3, time.Millisecond, func(context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("Retry = %v, want nil", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestRetrySucceedsAfterTransientFailures(t *testing.T) {
	calls := 0
	err := flow.Retry(context.Background(), 3, time.Millisecond, func(context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Retry = %v, want nil", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}

func TestRetryExhaustedReturnsLastError(t *testing.T) {
	calls := 0
	err := flow.Retry(context.Background(), 3, time.Millisecond, func(context.Context) error {
		calls++
		return errors.New("failure")
	})
	if err == nil || err.Error() != "failure" {
		t.Fatalf("Retry = %v, want last failure", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}

func TestRetrySingleAttemptCallsOnce(t *testing.T) {
	calls := 0
	_ = flow.Retry(context.Background(), 1, time.Millisecond, func(context.Context) error {
		calls++
		return errors.New("failure")
	})
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestRetryZeroAttemptsPanics(t *testing.T) {
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_ = flow.Retry(context.Background(), 0, time.Millisecond, func(context.Context) error { return nil })
		return nil
	}()
	if panicked == nil {
		t.Fatal("Retry(0) did not panic, want panic")
	}
}

func TestRetryNegativeBackoffPanics(t *testing.T) {
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_ = flow.Retry(context.Background(), 2, -time.Millisecond, func(context.Context) error { return nil })
		return nil
	}()
	if panicked == nil {
		t.Fatal("Retry(negative backoff) did not panic, want panic")
	}
}

func TestRetryCancelledContextSkipsWork(t *testing.T) {
	active, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := flow.Retry(active, 3, time.Millisecond, func(context.Context) error {
		calls++
		return nil
	})
	if calls != 0 {
		t.Fatalf("calls = %d, want 0 (cancelled before first attempt)", calls)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Retry = %v, want context.Canceled", err)
	}
}

func TestRetryCancelDuringBackoffStops(t *testing.T) {
	active, cancel := context.WithCancel(context.Background())
	calls := 0
	err := flow.Retry(active, 5, 50*time.Millisecond, func(context.Context) error {
		calls++
		if calls == 1 {
			cancel()
		}
		return errors.New("failure")
	})
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (cancelled during backoff)", calls)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Retry = %v, want context.Canceled", err)
	}
}

func TestRetryPanicPropagatesWithoutRetry(t *testing.T) {
	calls := 0
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_ = flow.Retry(context.Background(), 3, time.Millisecond, func(context.Context) error {
			calls++
			panic("work blew up")
		})
		return nil
	}()
	if panicked == nil {
		t.Fatal("Retry did not propagate work panic, want panic")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (panic aborts, never retried)", calls)
	}
}

func TestRetryWithResultSucceeds(t *testing.T) {
	got, err := flow.RetryWithResult(context.Background(), 3, time.Millisecond, func(context.Context) (string, error) {
		return "gopher", nil
	})
	if err != nil {
		t.Fatalf("RetryWithResult err = %v, want nil", err)
	}
	if got != "gopher" {
		t.Fatalf("RetryWithResult = %q, want %q", got, "gopher")
	}
}

func TestRetryWithResultExhaustedYieldsZeroAndLastError(t *testing.T) {
	calls := 0
	got, err := flow.RetryWithResult(context.Background(), 2, time.Millisecond, func(context.Context) (string, error) {
		calls++
		return "", errors.New("down")
	})
	if err == nil || err.Error() != "down" {
		t.Fatalf("RetryWithResult err = %v, want last failure", err)
	}
	if got != "" {
		t.Fatalf("RetryWithResult = %q, want zero value", got)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}
