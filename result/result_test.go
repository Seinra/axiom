package result_test

import (
	"errors"
	"testing"

	"github.com/seinra/axiom/result"
)

func TestOkStoresSuccess(t *testing.T) {
	current := result.Ok(42)
	if !current.IsOk() {
		t.Fatal("IsOk = false, want true")
	}
	got, failure := current.Unwrap()
	if failure != nil {
		t.Fatalf("Unwrap failure = %v, want nil", failure)
	}
	if got != 42 {
		t.Fatalf("Unwrap value = %d, want 42", got)
	}
}

func TestFailStoresFailure(t *testing.T) {
	sentinel := errors.New("boom")
	current := result.Fail[int](sentinel)
	if current.IsOk() {
		t.Fatal("IsOk = true, want false")
	}
	_, failure := current.Unwrap()
	if !errors.Is(failure, sentinel) {
		t.Fatalf("Unwrap failure = %v, want %v", failure, sentinel)
	}
}

func TestMapTransformsSuccess(t *testing.T) {
	current := result.Ok(21)
	mapped := result.Map(current, func(success int) string {
		if success != 21 {
			t.Fatalf("transform input = %d, want 21", success)
		}
		return "answer"
	})
	got, failure := mapped.Unwrap()
	if failure != nil {
		t.Fatalf("Unwrap failure = %v, want nil", failure)
	}
	if got != "answer" {
		t.Fatalf("Unwrap value = %q, want %q", got, "answer")
	}
}

func TestMapKeepsFailure(t *testing.T) {
	sentinel := errors.New("kept")
	current := result.Fail[int](sentinel)
	mapped := result.Map(current, func(success int) string { return "unreachable" })
	if mapped.IsOk() {
		t.Fatal("IsOk = true, want false")
	}
	_, failure := mapped.Unwrap()
	if !errors.Is(failure, sentinel) {
		t.Fatalf("Unwrap failure = %v, want %v", failure, sentinel)
	}
}

func TestMethodMapValueChains(t *testing.T) {
	// Exercises the Go 1.27 generic method: MapValue declares its own type params.
	got, failure := result.Ok(2).MapValue(func(success int) int { return success * 3 }).Unwrap()
	if failure != nil {
		t.Fatalf("Unwrap failure = %v, want nil", failure)
	}
	if got != 6 {
		t.Fatalf("chained value = %d, want 6", got)
	}
}

func TestMustPanicsOnFailure(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("Must did not panic on failure")
		}
	}()
	_ = result.Must(result.Fail[int](errors.New("must boom")))
}

func TestFailNilYieldsError(t *testing.T) {
	current := result.Fail[int](nil)
	if current.IsOk() {
		t.Fatal("IsOk = true, want false")
	}
	_, failure := current.Unwrap()
	if failure == nil {
		t.Fatal("Unwrap failure = nil, want descriptive error")
	}
}

func TestMustReturnsSuccess(t *testing.T) {
	if got := result.Must(result.Ok("kept")); got != "kept" {
		t.Fatalf("Must = %q, want %q", got, "kept")
	}
}

func TestMapValueCarriesFailure(t *testing.T) {
	sentinel := errors.New("carried")
	mapped := result.Fail[int](sentinel).MapValue(func(success int) string { return "unreachable" })
	if mapped.IsOk() {
		t.Fatal("IsOk = true, want false")
	}
	_, failure := mapped.Unwrap()
	if !errors.Is(failure, sentinel) {
		t.Fatalf("Unwrap failure = %v, want %v", failure, sentinel)
	}
}

func TestZeroValueReportsError(t *testing.T) {
	var current result.Result[int]
	if current.IsOk() {
		t.Fatal("IsOk = true, want false")
	}
	_, failure := current.Unwrap()
	if failure == nil {
		t.Fatal("zero Unwrap failure = nil, want explicit error")
	}
}

func TestMapZeroReportsError(t *testing.T) {
	var current result.Result[int]
	mapped := result.Map(current, func(success int) string { return "unreachable" })
	_, failure := mapped.Unwrap()
	if failure == nil {
		t.Fatal("Map(zero) Unwrap failure = nil, want explicit error")
	}
}
