package result_test

import (
	"errors"
	"strings"
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

func TestTrySuccessReturnsValue(t *testing.T) {
	current := result.Try(func() int { return 41 + 1 })
	if !current.IsOk() {
		t.Fatal("Try IsOk = false, want true")
	}
	got, failure := current.Unwrap()
	if failure != nil {
		t.Fatalf("Try Unwrap failure = %v, want nil", failure)
	}
	if got != 42 {
		t.Fatalf("Try value = %d, want 42", got)
	}
}

func TestTryStringPayloadBecomesFailure(t *testing.T) {
	current := result.Try(func() int { panic("try boom") })
	if current.IsOk() {
		t.Fatal("Try panic IsOk = true, want false")
	}
	_, failure := current.Unwrap()
	if failure == nil {
		t.Fatal("Try panic failure = nil, want recovered panic")
	}
	if !strings.Contains(failure.Error(), "axiom result: recovered panic: try boom") {
		t.Fatalf("Try panic failure = %q, want recovered panic with payload", failure.Error())
	}
}

func TestTryErrorPayloadBecomesFailure(t *testing.T) {
	sentinel := errors.New("inner boom")
	current := result.Try(func() int { panic(sentinel) })
	if current.IsOk() {
		t.Fatal("Try error panic IsOk = true, want false")
	}
	_, failure := current.Unwrap()
	if failure == nil {
		t.Fatal("Try error panic failure = nil, want recovered panic")
	}
	if !strings.Contains(failure.Error(), "axiom result: recovered panic: inner boom") {
		t.Fatalf("Try error panic failure = %q, want recovered panic with payload", failure.Error())
	}
}

func TestTryNilAttemptRecovers(t *testing.T) {
	var attempt func() int
	current := result.Try(attempt)
	if current.IsOk() {
		t.Fatal("Try nil IsOk = true, want false")
	}
	_, failure := current.Unwrap()
	if failure == nil {
		t.Fatal("Try nil failure = nil, want recovered panic")
	}
	if !strings.Contains(failure.Error(), "axiom result: recovered panic:") {
		t.Fatalf("Try nil failure = %q, want recovered panic marker", failure.Error())
	}
}

func TestUnwrapOrElseTable(t *testing.T) {
	sentinel := errors.New("fallback cause")
	testCases := []struct {
		name     string
		current  result.Result[int]
		fallback func(error) int
		want     int
	}{
		{
			name:     "success yields held value",
			current:  result.Ok(4),
			fallback: func(failure error) int { return 99 },
			want:     4,
		},
		{
			name:     "failure yields fallback outcome",
			current:  result.Fail[int](sentinel),
			fallback: func(failure error) int { return 9 },
			want:     9,
		},
		{
			name:     "zero yields fallback outcome",
			current:  result.Result[int]{},
			fallback: func(failure error) int { return 7 },
			want:     7,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := result.UnwrapOrElse(testCase.current, testCase.fallback); got != testCase.want {
				t.Fatalf("UnwrapOrElse = %d, want %d", got, testCase.want)
			}
		})
	}
}

func TestUnwrapOrElseForwardsFailure(t *testing.T) {
	sentinel := errors.New("forward me")
	var received error
	got := result.UnwrapOrElse(result.Fail[int](sentinel), func(failure error) int {
		received = failure
		return 9
	})
	if got != 9 {
		t.Fatalf("UnwrapOrElse = %d, want 9", got)
	}
	if !errors.Is(received, sentinel) {
		t.Fatalf("fallback received = %v, want %v", received, sentinel)
	}
}

func TestUnwrapOrElseLazySkipsFallback(t *testing.T) {
	ranFallback := false
	got := result.UnwrapOrElse(result.Ok(5), func(failure error) int {
		ranFallback = true
		return 9
	})
	if got != 5 {
		t.Fatalf("UnwrapOrElse success = %d, want 5", got)
	}
	if ranFallback {
		t.Fatal("UnwrapOrElse ran fallback on success, want lazy skip")
	}
}

func TestUnwrapOrElseZeroReportsExplicitFailure(t *testing.T) {
	var current result.Result[int]
	var received error
	got := result.UnwrapOrElse(current, func(failure error) int {
		received = failure
		return 7
	})
	if got != 7 {
		t.Fatalf("UnwrapOrElse zero = %d, want 7", got)
	}
	if received == nil {
		t.Fatal("UnwrapOrElse zero fallback received nil, want explicit error")
	}
}

func TestFlattenResultTable(t *testing.T) {
	sentinelInner := errors.New("inner boom")
	sentinelOuter := errors.New("outer boom")
	testCases := []struct {
		name        string
		current     result.Result[result.Result[int]]
		wantOk      bool
		wantValue   int
		wantFailure error
	}{
		{
			name:      "ok of ok yields inner success",
			current:   result.Ok(result.Ok(3)),
			wantOk:    true,
			wantValue: 3,
		},
		{
			name:        "ok of failure yields inner failure",
			current:     result.Ok(result.Fail[int](sentinelInner)),
			wantOk:      false,
			wantFailure: sentinelInner,
		},
		{
			name:        "failure yields outer failure",
			current:     result.Fail[result.Result[int]](sentinelOuter),
			wantOk:      false,
			wantFailure: sentinelOuter,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			flattened := result.Flatten(testCase.current)
			if flattened.IsOk() != testCase.wantOk {
				t.Fatalf("Flatten IsOk = %v, want %v", flattened.IsOk(), testCase.wantOk)
			}
			got, failure := flattened.Unwrap()
			if testCase.wantOk {
				if failure != nil {
					t.Fatalf("Flatten failure = %v, want nil", failure)
				}
				if got != testCase.wantValue {
					t.Fatalf("Flatten = %d, want %d", got, testCase.wantValue)
				}
				return
			}
			if !errors.Is(failure, testCase.wantFailure) {
				t.Fatalf("Flatten failure = %v, want %v", failure, testCase.wantFailure)
			}
		})
	}
}

func TestFlattenResultZeroOuterReportsError(t *testing.T) {
	var current result.Result[result.Result[int]]
	flattened := result.Flatten(current)
	if flattened.IsOk() {
		t.Fatal("Flatten(zero) IsOk = true, want false")
	}
	if _, failure := flattened.Unwrap(); failure == nil {
		t.Fatal("Flatten(zero) failure = nil, want explicit error")
	}
}

func TestFlattenResultZeroNestingTable(t *testing.T) {
	testCases := []struct {
		name                string
		current             result.Result[result.Result[int]]
		wantMessageFragment string
	}{
		{
			name:                "zero outer resolves to explicit failure",
			current:             result.Result[result.Result[int]]{},
			wantMessageFragment: "nil failure",
		},
		{
			name:                "ok of zero inner resolves to explicit zero failure",
			current:             result.Ok(result.Result[int]{}),
			wantMessageFragment: "zero Result",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			flattened := result.Flatten(testCase.current)
			if flattened.IsOk() {
				t.Fatal("Flatten IsOk = true, want false")
			}
			outcome, failure := flattened.Unwrap()
			if failure == nil {
				t.Fatal("Flatten failure = nil, want explicit error")
			}
			if !strings.Contains(failure.Error(), testCase.wantMessageFragment) {
				t.Fatalf("Flatten failure = %q, want fragment %q", failure.Error(), testCase.wantMessageFragment)
			}
			if outcome != 0 {
				t.Fatalf("Flatten value = %d, want 0", outcome)
			}
		})
	}
}

func pinnedPanics(t *testing.T, label string, call func()) {
	t.Helper()
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		call()
		return nil
	}()
	if panicked == nil {
		t.Fatalf("%s did not panic, want panic (nil func runs on the held success)", label)
	}
}

func pinnedNoPanic(t *testing.T, label string, call func()) {
	t.Helper()
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		call()
		return nil
	}()
	if panicked != nil {
		t.Fatalf("%s panicked with %v, want no panic (nil func never runs on the carried failure)", label, panicked)
	}
}

func TestMapNilTransformPanicsOnOkCarriesFailure(t *testing.T) {
	pinnedPanics(t, "Map(Ok, nil)", func() {
		_ = result.Map[int, string](result.Ok(1), nil)
	})
	pinnedPanics(t, "Ok.MapValue(nil)", func() {
		_ = result.Ok(1).MapValue[string](nil)
	})
	sentinel := errors.New("carried")
	var mappedFunc result.Result[string]
	pinnedNoPanic(t, "Map(Fail, nil)", func() {
		mappedFunc = result.Map[int, string](result.Fail[int](sentinel), nil)
	})
	if _, failure := mappedFunc.Unwrap(); !errors.Is(failure, sentinel) {
		t.Fatalf("Map(Fail, nil) failure = %v, want %v", failure, sentinel)
	}
	var mappedMethod result.Result[string]
	pinnedNoPanic(t, "Fail.MapValue(nil)", func() {
		mappedMethod = result.Fail[int](sentinel).MapValue[string](nil)
	})
	if _, failure := mappedMethod.Unwrap(); !errors.Is(failure, sentinel) {
		t.Fatalf("Fail.MapValue(nil) failure = %v, want %v", failure, sentinel)
	}
	var mappedZero result.Result[string]
	pinnedNoPanic(t, "Map(zero, nil)", func() {
		mappedZero = result.Map[int, string](result.Result[int]{}, nil)
	})
	if _, failure := mappedZero.Unwrap(); failure == nil {
		t.Fatal("Map(zero, nil) failure = nil, want explicit error")
	}
}

func TestAndThenNilTransformPanicsOnOkShortCircuitsOnFailure(t *testing.T) {
	pinnedPanics(t, "AndThen(Ok, nil)", func() {
		_ = result.AndThen[int, string](result.Ok(1), nil)
	})
	sentinel := errors.New("kept")
	var chained result.Result[string]
	pinnedNoPanic(t, "AndThen(Fail, nil)", func() {
		chained = result.AndThen[int, string](result.Fail[int](sentinel), nil)
	})
	if _, failure := chained.Unwrap(); !errors.Is(failure, sentinel) {
		t.Fatalf("AndThen(Fail, nil) failure = %v, want %v", failure, sentinel)
	}
	var chainedZero result.Result[string]
	pinnedNoPanic(t, "AndThen(zero, nil)", func() {
		chainedZero = result.AndThen[int, string](result.Result[int]{}, nil)
	})
	if _, failure := chainedZero.Unwrap(); failure == nil {
		t.Fatal("AndThen(zero, nil) failure = nil, want explicit error")
	}
}

func TestOrElseNilHandlerPanicsOnFailureSkipsOnSuccess(t *testing.T) {
	pinnedPanics(t, "OrElse(Fail, nil)", func() {
		_ = result.OrElse(result.Fail[int](errors.New("broken")), nil)
	})
	var recovered result.Result[int]
	pinnedNoPanic(t, "OrElse(Ok, nil)", func() {
		recovered = result.OrElse(result.Ok(5), nil)
	})
	if got, failure := recovered.Unwrap(); failure != nil || got != 5 {
		t.Fatalf("OrElse(Ok, nil) = %d, %v, want 5, nil", got, failure)
	}
}

func TestUnwrapOrElseNilFallbackPanicsOnFailureSkipsOnSuccess(t *testing.T) {
	pinnedPanics(t, "UnwrapOrElse(Fail, nil)", func() {
		_ = result.UnwrapOrElse(result.Fail[int](errors.New("gone")), nil)
	})
	var got int
	pinnedNoPanic(t, "UnwrapOrElse(Ok, nil)", func() {
		got = result.UnwrapOrElse(result.Ok(4), nil)
	})
	if got != 4 {
		t.Fatalf("UnwrapOrElse(Ok, nil) = %d, want 4", got)
	}
}
