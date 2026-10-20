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

func TestMapFailureTransformsFailure(t *testing.T) {
	sentinel := errors.New("original")
	current := result.Fail[int](sentinel)
	mapped := result.MapFailure(current, func(failure error) error {
		return errors.New("wrapped: " + failure.Error())
	})
	if mapped.IsOk() {
		t.Fatal("MapFailure IsOk = true, want false")
	}
	_, failure := mapped.Unwrap()
	if failure == nil || failure.Error() != "wrapped: original" {
		t.Fatalf("MapFailure failure = %v, want wrapped original", failure)
	}
}

func TestMapFailureSuccessPassesThrough(t *testing.T) {
	ranTransform := false
	current := result.Ok(11)
	mapped := result.MapFailure(current, func(failure error) error {
		ranTransform = true
		return errors.New("unreachable")
	})
	got, failure := mapped.Unwrap()
	if failure != nil {
		t.Fatalf("MapFailure success failure = %v, want nil", failure)
	}
	if got != 11 {
		t.Fatalf("MapFailure success = %d, want 11", got)
	}
	if ranTransform {
		t.Fatal("MapFailure ran transform on success, want pass-through")
	}
}

func TestMapFailureNilTransformGuards(t *testing.T) {
	current := result.Fail[int](errors.New("original"))
	mapped := result.MapFailure(current, func(failure error) error { return nil })
	if mapped.IsOk() {
		t.Fatal("MapFailure nil IsOk = true, want false")
	}
	if _, failure := mapped.Unwrap(); failure == nil {
		t.Fatal("MapFailure nil failure = nil, want guard error")
	}
}

func TestAndThenBindsSuccess(t *testing.T) {
	current := result.Ok(7)
	mapped := result.AndThen(current, func(held int) result.Result[string] {
		return result.Ok("bound")
	})
	got, failure := mapped.Unwrap()
	if failure != nil {
		t.Fatalf("AndThen failure = %v, want nil", failure)
	}
	if got != "bound" {
		t.Fatalf("AndThen = %q, want %q", got, "bound")
	}
}

func TestAndThenTransformFailurePropagates(t *testing.T) {
	sentinel := errors.New("bind failed")
	current := result.Ok(7)
	mapped := result.AndThen(current, func(held int) result.Result[string] {
		return result.Fail[string](sentinel)
	})
	if mapped.IsOk() {
		t.Fatal("AndThen to failure IsOk = true, want false")
	}
	if _, failure := mapped.Unwrap(); !errors.Is(failure, sentinel) {
		t.Fatalf("AndThen failure = %v, want %v", failure, sentinel)
	}
}

func TestAndThenFailureShortCircuits(t *testing.T) {
	ranTransform := false
	sentinel := errors.New("kept")
	current := result.Fail[int](sentinel)
	mapped := result.AndThen(current, func(held int) result.Result[string] {
		ranTransform = true
		return result.Ok("unreachable")
	})
	if mapped.IsOk() {
		t.Fatal("AndThen on failure IsOk = true, want false")
	}
	if _, failure := mapped.Unwrap(); !errors.Is(failure, sentinel) {
		t.Fatalf("AndThen failure = %v, want %v", failure, sentinel)
	}
	if ranTransform {
		t.Fatal("AndThen ran transform on failure, want short-circuit")
	}
}

func TestOrElseSuccessSkipsRecover(t *testing.T) {
	ranRecover := false
	current := result.Ok(5)
	got, failure := result.OrElse(current, func(failure error) result.Result[int] {
		ranRecover = true
		return result.Ok(9)
	}).Unwrap()
	if failure != nil {
		t.Fatalf("OrElse success failure = %v, want nil", failure)
	}
	if got != 5 {
		t.Fatalf("OrElse success = %d, want 5", got)
	}
	if ranRecover {
		t.Fatal("OrElse ran recover on success, want skip")
	}
}

func TestOrElseFailureRecoversToOk(t *testing.T) {
	current := result.Fail[int](errors.New("fell over"))
	got, failure := result.OrElse(current, func(failure error) result.Result[int] {
		return result.Ok(9)
	}).Unwrap()
	if failure != nil {
		t.Fatalf("OrElse recover failure = %v, want nil", failure)
	}
	if got != 9 {
		t.Fatalf("OrElse recover = %d, want 9", got)
	}
}

func TestOrElseFailureRecoverStaysFailure(t *testing.T) {
	sentinel := errors.New("still broken")
	current := result.Fail[int](errors.New("fell over"))
	recovered := result.OrElse(current, func(failure error) result.Result[int] {
		return result.Fail[int](sentinel)
	})
	if recovered.IsOk() {
		t.Fatal("OrElse recover IsOk = true, want false")
	}
	if _, failure := recovered.Unwrap(); !errors.Is(failure, sentinel) {
		t.Fatalf("OrElse failure = %v, want %v", failure, sentinel)
	}
}

func TestUnwrapOrSuccess(t *testing.T) {
	if got := result.UnwrapOr(result.Ok(4), 9); got != 4 {
		t.Fatalf("UnwrapOr success = %d, want 4", got)
	}
}

func TestUnwrapOrFailureYieldsFallback(t *testing.T) {
	if got := result.UnwrapOr(result.Fail[int](errors.New("gone")), 9); got != 9 {
		t.Fatalf("UnwrapOr failure = %d, want 9", got)
	}
}
