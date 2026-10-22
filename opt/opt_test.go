package opt_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/seinra/axiom/opt"
)

func TestSomeIsSome(t *testing.T) {
	current := opt.Some(7)
	if !current.IsSome() {
		t.Fatal("IsSome = false, want true")
	}
	if got := current.ValueOr(0); got != 7 {
		t.Fatalf("ValueOr = %d, want 7", got)
	}
}

func TestNoneIsNone(t *testing.T) {
	current := opt.None[int]()
	if current.IsSome() {
		t.Fatal("IsSome = true, want false")
	}
	if got := current.ValueOr(9); got != 9 {
		t.Fatalf("ValueOr = %d, want 9", got)
	}
}

func TestMapValueTransformsSome(t *testing.T) {
	current := opt.Some(4)
	mapped := current.MapValue(func(success int) string { return "four" })
	if !mapped.IsSome() {
		t.Fatal("IsSome = false, want true")
	}
	if got := mapped.ValueOr(""); got != "four" {
		t.Fatalf("ValueOr = %q, want %q", got, "four")
	}
}

func TestMapValueKeepsNone(t *testing.T) {
	current := opt.None[int]()
	mapped := current.MapValue(func(success int) string { return "unreachable" })
	if mapped.IsSome() {
		t.Fatal("IsSome = true, want false")
	}
}

func TestMethodMapValueChains(t *testing.T) {
	// Exercises the Go 1.27 generic method: MapValue declares its own type params.
	if got := opt.Some(5).MapValue(func(success int) int { return success + 1 }).ValueOr(0); got != 6 {
		t.Fatalf("chained value = %d, want 6", got)
	}
}

func TestZeroValueIsAbsent(t *testing.T) {
	var current opt.Option[int]
	if current.IsSome() {
		t.Fatal("IsSome = true, want false")
	}
	if got := current.ValueOr(3); got != 3 {
		t.Fatalf("ValueOr = %d, want 3", got)
	}
}

func TestOrElseSomeWins(t *testing.T) {
	current := opt.Some(1)
	fallback := opt.Some(2)
	if got := opt.OrElse(current, fallback).ValueOr(0); got != 1 {
		t.Fatalf("OrElse Some/Some = %d, want 1", got)
	}
}

func TestOrElseFallsBackOnNone(t *testing.T) {
	current := opt.None[int]()
	fallback := opt.Some(2)
	if got := opt.OrElse(current, fallback).ValueOr(0); got != 2 {
		t.Fatalf("OrElse None/Some = %d, want 2", got)
	}
}

func TestOrElseNoneNoneStaysNone(t *testing.T) {
	combined := opt.OrElse(opt.None[int](), opt.None[int]())
	if combined.IsSome() {
		t.Fatal("OrElse None/None IsSome = true, want false")
	}
}

func TestAndThenBindsSome(t *testing.T) {
	current := opt.Some(4)
	mapped := opt.AndThen(current, func(held int) opt.Option[string] {
		return opt.Some("bound")
	})
	if got := mapped.ValueOr(""); got != "bound" {
		t.Fatalf("AndThen = %q, want %q", got, "bound")
	}
}

func TestAndThenTransformToNone(t *testing.T) {
	current := opt.Some(4)
	mapped := opt.AndThen(current, func(held int) opt.Option[string] {
		return opt.None[string]()
	})
	if mapped.IsSome() {
		t.Fatal("AndThen to None IsSome = true, want false")
	}
}

func TestAndThenNoneShortCircuits(t *testing.T) {
	ranTransform := false
	current := opt.None[int]()
	mapped := opt.AndThen(current, func(held int) opt.Option[string] {
		ranTransform = true
		return opt.Some("unreachable")
	})
	if mapped.IsSome() {
		t.Fatal("AndThen on None IsSome = true, want false")
	}
	if ranTransform {
		t.Fatal("AndThen ran transform on None, want short-circuit")
	}
}

func TestFilterKeepsMatchingSome(t *testing.T) {
	current := opt.Some(6)
	filtered := opt.Filter(current, func(held int) bool { return held%2 == 0 })
	if got := filtered.ValueOr(0); got != 6 {
		t.Fatalf("Filter keep = %d, want 6", got)
	}
}

func TestFilterDropsNonMatchingSome(t *testing.T) {
	current := opt.Some(7)
	filtered := opt.Filter(current, func(held int) bool { return held%2 == 0 })
	if filtered.IsSome() {
		t.Fatal("Filter drop IsSome = true, want false")
	}
}

func TestFilterNoneStaysNone(t *testing.T) {
	ranKeep := false
	current := opt.None[int]()
	filtered := opt.Filter(current, func(held int) bool {
		ranKeep = true
		return true
	})
	if filtered.IsSome() {
		t.Fatal("Filter on None IsSome = true, want false")
	}
	if ranKeep {
		t.Fatal("Filter ran keep on None, want short-circuit")
	}
}

func TestValueOrElseSomeSkipsFallback(t *testing.T) {
	ranFallback := false
	current := opt.Some(5)
	got := opt.ValueOrElse(current, func() int {
		ranFallback = true
		return 9
	})
	if got != 5 {
		t.Fatalf("ValueOrElse Some = %d, want 5", got)
	}
	if ranFallback {
		t.Fatal("ValueOrElse ran fallback on Some, want lazy skip")
	}
}

func TestValueOrElseNoneRunsFallback(t *testing.T) {
	current := opt.None[int]()
	got := opt.ValueOrElse(current, func() int { return 9 })
	if got != 9 {
		t.Fatalf("ValueOrElse None = %d, want 9", got)
	}
}

func TestToResultSomeYieldsOk(t *testing.T) {
	converted := opt.ToResult(opt.Some(3), errors.New("unused"))
	got, failure := converted.Unwrap()
	if failure != nil {
		t.Fatalf("ToResult Some failure = %v, want nil", failure)
	}
	if got != 3 {
		t.Fatalf("ToResult Some = %d, want 3", got)
	}
}

func TestToResultNoneYieldsFailure(t *testing.T) {
	sentinel := errors.New("missing")
	converted := opt.ToResult(opt.None[int](), sentinel)
	if converted.IsOk() {
		t.Fatal("ToResult None IsOk = true, want false")
	}
	if _, failure := converted.Unwrap(); !errors.Is(failure, sentinel) {
		t.Fatalf("ToResult None failure = %v, want %v", failure, sentinel)
	}
}

func TestToResultNilFailureGuards(t *testing.T) {
	converted := opt.ToResult(opt.None[int](), nil)
	if converted.IsOk() {
		t.Fatal("ToResult nil IsOk = true, want false")
	}
	if _, failure := converted.Unwrap(); failure == nil {
		t.Fatal("ToResult nil failure = nil, want guard error")
	}
}

func TestFlattenTable(t *testing.T) {
	testCases := []struct {
		name      string
		current   opt.Option[opt.Option[int]]
		wantSome  bool
		wantValue int
	}{
		{
			name:     "none yields none",
			current:  opt.None[opt.Option[int]](),
			wantSome: false,
		},
		{
			name:     "some none yields none",
			current:  opt.Some(opt.None[int]()),
			wantSome: false,
		},
		{
			name:      "some some yields inner",
			current:   opt.Some(opt.Some(8)),
			wantSome:  true,
			wantValue: 8,
		},
		{
			name:     "zero yields none",
			current:  opt.Option[opt.Option[int]]{},
			wantSome: false,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			flattened := opt.Flatten(testCase.current)
			if flattened.IsSome() != testCase.wantSome {
				t.Fatalf("Flatten IsSome = %v, want %v", flattened.IsSome(), testCase.wantSome)
			}
			if testCase.wantSome {
				if got := flattened.ValueOr(0); got != testCase.wantValue {
					t.Fatalf("Flatten = %d, want %d", got, testCase.wantValue)
				}
			}
		})
	}
}

func TestFlattenZeroNestingTable(t *testing.T) {
	const fallbackValue = 99
	testCases := []struct {
		name    string
		current opt.Option[opt.Option[int]]
	}{
		{
			name:    "zero outer yields none",
			current: opt.Option[opt.Option[int]]{},
		},
		{
			name:    "some of zero inner yields none",
			current: opt.Some(opt.Option[int]{}),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			flattened := opt.Flatten(testCase.current)
			if flattened.IsSome() {
				t.Fatal("Flatten IsSome = true, want false")
			}
			if outcome := flattened.ValueOr(fallbackValue); outcome != fallbackValue {
				t.Fatalf("Flatten ValueOr = %d, want fallback %d", outcome, fallbackValue)
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
		t.Fatalf("%s panicked with %v, want no panic (nil func never runs on the short-circuit path)", label, panicked)
	}
}

func TestMapValueNilTransformPanicsOnSomeSkipsOnNone(t *testing.T) {
	pinnedPanics(t, "Some.MapValue(nil)", func() {
		_ = opt.Some(1).MapValue[string](nil)
	})
	var mappedMethod opt.Option[string]
	pinnedNoPanic(t, "None.MapValue(nil)", func() {
		mappedMethod = opt.None[int]().MapValue[string](nil)
	})
	if mappedMethod.IsSome() {
		t.Fatal("None.MapValue(nil) IsSome = true, want false")
	}
}

func TestAndThenNilTransformPanicsOnSomeSkipsOnNone(t *testing.T) {
	pinnedPanics(t, "AndThen(Some, nil)", func() {
		_ = opt.AndThen[int, string](opt.Some(1), nil)
	})
	var chained opt.Option[string]
	pinnedNoPanic(t, "AndThen(None, nil)", func() {
		chained = opt.AndThen[int, string](opt.None[int](), nil)
	})
	if chained.IsSome() {
		t.Fatal("AndThen(None, nil) IsSome = true, want false")
	}
}

func TestFilterNilKeepPanicsOnSomeSkipsOnNone(t *testing.T) {
	pinnedPanics(t, "Filter(Some, nil)", func() {
		_ = opt.Filter(opt.Some(1), nil)
	})
	var filtered opt.Option[int]
	pinnedNoPanic(t, "Filter(None, nil)", func() {
		filtered = opt.Filter(opt.None[int](), nil)
	})
	if filtered.IsSome() {
		t.Fatal("Filter(None, nil) IsSome = true, want false")
	}
}

func TestValueOrElseNilFallbackPanicsOnNoneSkipsOnSome(t *testing.T) {
	pinnedPanics(t, "ValueOrElse(None, nil)", func() {
		_ = opt.ValueOrElse(opt.None[int](), nil)
	})
	var got int
	pinnedNoPanic(t, "ValueOrElse(Some, nil)", func() {
		got = opt.ValueOrElse(opt.Some(5), nil)
	})
	if got != 5 {
		t.Fatalf("ValueOrElse(Some, nil) = %d, want 5", got)
	}
}

func TestOptionMarshalJSONSomeRendersValue(t *testing.T) {
	payload, failure := json.Marshal(opt.Some(4))
	if failure != nil {
		t.Fatalf("Marshal(Some) failure = %v, want nil", failure)
	}
	if string(payload) != "4" {
		t.Fatalf("Marshal(Some(4)) = %s, want 4", payload)
	}
}

func TestOptionMarshalJSONStringRendersQuoted(t *testing.T) {
	payload, failure := json.Marshal(opt.Some("gopher"))
	if failure != nil {
		t.Fatalf("Marshal(Some) failure = %v, want nil", failure)
	}
	if string(payload) != `"gopher"` {
		t.Fatalf("Marshal(Some(gopher)) = %s, want quoted gopher", payload)
	}
}

func TestOptionMarshalJSONNoneRendersNull(t *testing.T) {
	payload, failure := json.Marshal(opt.None[int]())
	if failure != nil {
		t.Fatalf("Marshal(None) failure = %v, want nil", failure)
	}
	if string(payload) != "null" {
		t.Fatalf("Marshal(None) = %s, want null", payload)
	}
}

func TestOptionUnmarshalJSONValueYieldsSome(t *testing.T) {
	var current opt.Option[int]
	if failure := json.Unmarshal([]byte("4"), &current); failure != nil {
		t.Fatalf("Unmarshal(4) failure = %v, want nil", failure)
	}
	if got := current.ValueOr(0); got != 4 {
		t.Fatalf("Unmarshal(4) = %d, want 4", got)
	}
}

func TestOptionUnmarshalJSONNullYieldsNone(t *testing.T) {
	for _, payload := range []string{"null", " null "} {
		current := opt.Some(7)
		if failure := json.Unmarshal([]byte(payload), &current); failure != nil {
			t.Fatalf("Unmarshal(%q) failure = %v, want nil", payload, failure)
		}
		if current.IsSome() {
			t.Fatalf("Unmarshal(%q) IsSome = true, want false", payload)
		}
	}
}

func TestOptionUnmarshalJSONInvalidYieldsError(t *testing.T) {
	current := opt.Some(7)
	if failure := json.Unmarshal([]byte("{bad}"), &current); failure == nil {
		t.Fatal("Unmarshal({bad}) failure = nil, want error")
	}
	if got := current.ValueOr(0); got != 7 {
		t.Fatalf("Unmarshal({bad}) clobbered receiver to %d, want untouched 7", got)
	}
}

func TestOptionUnmarshalJSONTrailingGarbageYieldsError(t *testing.T) {
	var current opt.Option[int]
	if failure := json.Unmarshal([]byte("4 x"), &current); failure == nil {
		t.Fatal("Unmarshal(4 x) failure = nil, want error for trailing bytes")
	}
}

func TestOptionJSONRoundTrip(t *testing.T) {
	payload, failure := json.Marshal(opt.Some("gopher"))
	if failure != nil {
		t.Fatalf("Marshal failure = %v, want nil", failure)
	}
	var restored opt.Option[string]
	if failure := json.Unmarshal(payload, &restored); failure != nil {
		t.Fatalf("Unmarshal failure = %v, want nil", failure)
	}
	if got := restored.ValueOr("nobody"); got != "gopher" {
		t.Fatalf("round trip = %q, want %q", got, "gopher")
	}
	nullPayload, failure := json.Marshal(opt.None[string]())
	if failure != nil {
		t.Fatalf("Marshal(None) failure = %v, want nil", failure)
	}
	var restoredNone opt.Option[string]
	if failure := json.Unmarshal(nullPayload, &restoredNone); failure != nil {
		t.Fatalf("Unmarshal(null) failure = %v, want nil", failure)
	}
	if restoredNone.IsSome() {
		t.Fatal("round trip None IsSome = true, want false")
	}
}

func TestOptionValueSomeReturnsHeld(t *testing.T) {
	stored, failure := opt.Some(int64(7)).Value()
	if failure != nil {
		t.Fatalf("Value(Some) failure = %v, want nil", failure)
	}
	if stored != int64(7) {
		t.Fatalf("Value(Some(7)) = %v, want 7", stored)
	}
}

func TestOptionValueNoneYieldsNull(t *testing.T) {
	stored, failure := opt.None[int64]().Value()
	if failure != nil {
		t.Fatalf("Value(None) failure = %v, want nil", failure)
	}
	if stored != nil {
		t.Fatalf("Value(None) = %v, want nil (NULL)", stored)
	}
}

func TestOptionScanNullYieldsNone(t *testing.T) {
	current := opt.Some("seed")
	if failure := current.Scan(nil); failure != nil {
		t.Fatalf("Scan(NULL) failure = %v, want nil", failure)
	}
	if current.IsSome() {
		t.Fatal("Scan(NULL) IsSome = true, want false")
	}
}

func TestOptionScanAssignableYieldsSome(t *testing.T) {
	var current opt.Option[string]
	if failure := current.Scan("gopher"); failure != nil {
		t.Fatalf("Scan(gopher) failure = %v, want nil", failure)
	}
	if got := current.ValueOr("nobody"); got != "gopher" {
		t.Fatalf("Scan(gopher) = %q, want %q", got, "gopher")
	}
}

func TestOptionScanInt64IntoIntFails(t *testing.T) {
	current := opt.Some(3)
	if failure := current.Scan(int64(7)); failure == nil {
		t.Fatal("Scan(int64 into Option[int]) failure = nil, want error (drivers yield int64, never int)")
	}
	if got := current.ValueOr(0); got != 3 {
		t.Fatalf("Scan(int64 into Option[int]) clobbered receiver to %d, want untouched 3", got)
	}
}

func TestOptionScanMismatchYieldsError(t *testing.T) {
	current := opt.Some("kept")
	if failure := current.Scan(int64(1)); failure == nil {
		t.Fatal("Scan(int64 into Option[string]) failure = nil, want error (assertion only, no conversion)")
	}
	if got := current.ValueOr("nobody"); got != "kept" {
		t.Fatalf("Scan(mismatch) clobbered receiver to %q, want untouched kept", got)
	}
}
