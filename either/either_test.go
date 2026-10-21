package either_test

import (
	"errors"
	"testing"

	"github.com/seinra/axiom/either"
)

func TestLeftHoldsLeftSide(t *testing.T) {
	current := either.Left[string, int]("locked")
	if !current.IsLeft() {
		t.Fatal("IsLeft = false, want true")
	}
	if current.IsRight() {
		t.Fatal("IsRight = true, want false")
	}
}

func TestRightHoldsRightSide(t *testing.T) {
	current := either.Right[string, int](42)
	if !current.IsRight() {
		t.Fatal("IsRight = false, want true")
	}
	if current.IsLeft() {
		t.Fatal("IsLeft = true, want false")
	}
}

func TestZeroValueIsLeftHoldingZero(t *testing.T) {
	testCases := []struct {
		scenario string
		check    func() bool
	}{
		{scenario: "string/int zero reads as Left", check: func() bool {
			var current either.Either[string, int]
			return current.IsLeft() && !current.IsRight()
		}},
		{scenario: "int/string zero reads as Left", check: func() bool {
			var current either.Either[int, string]
			return current.IsLeft() && !current.IsRight()
		}},
		{scenario: "zero Left carries zero payload", check: func() bool {
			var current either.Either[string, int]
			rendered := either.Match(current,
				func(lockReason string) string { return "left:" + lockReason },
				func(granted int) string { return "right" })
			return rendered == "left:"
		}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			if !testCase.check() {
				t.Fatalf("zero-value check failed for %q", testCase.scenario)
			}
		})
	}
}

func TestMatchSplitsBothSides(t *testing.T) {
	testCases := []struct {
		scenario string
		current  either.Either[string, int]
		expected string
	}{
		{scenario: "Left calls onLeft", current: either.Left[string, int]("season pass missing"), expected: "locked: season pass missing"},
		{scenario: "Right calls onRight", current: either.Right[string, int](7), expected: "ready: 7 slots"},
		{scenario: "zero value calls onLeft with empty reason", current: either.Either[string, int]{}, expected: "locked: "},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			rendered := either.Match(testCase.current,
				func(lockReason string) string { return "locked: " + lockReason },
				func(slotCount int) string {
					if slotCount == 7 {
						return "ready: 7 slots"
					}
					return "ready: other"
				})
			if rendered != testCase.expected {
				t.Fatalf("Match = %q, want %q", rendered, testCase.expected)
			}
		})
	}
}

func TestMapRightTransformsRightCarriesLeft(t *testing.T) {
	testCases := []struct {
		scenario      string
		current       either.Either[string, int]
		expectedLabel string
		expectLeft    bool
	}{
		{scenario: "Right maps payload", current: either.Right[string, int](21), expectedLabel: "slots:42", expectLeft: false},
		{scenario: "Left carries untouched", current: either.Left[string, int]("locked"), expectedLabel: "locked", expectLeft: true},
		{scenario: "zero Left carries untouched", current: either.Either[string, int]{}, expectedLabel: "", expectLeft: true},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			mapped := either.MapRight(testCase.current, func(slotCount int) string {
				if slotCount == 21 {
					return "slots:42"
				}
				return "slots:other"
			})
			rendered := either.Match(mapped,
				func(lockReason string) string { return lockReason },
				func(label string) string { return label })
			if rendered != testCase.expectedLabel {
				t.Fatalf("MapRight Match = %q, want %q", rendered, testCase.expectedLabel)
			}
			if mapped.IsLeft() != testCase.expectLeft {
				t.Fatalf("MapRight IsLeft = %v, want %v", mapped.IsLeft(), testCase.expectLeft)
			}
		})
	}
}

func TestMapLeftTransformsLeftCarriesRight(t *testing.T) {
	testCases := []struct {
		scenario      string
		current       either.Either[string, int]
		expectedLabel string
		expectRight   bool
	}{
		{scenario: "Left maps reason", current: either.Left[string, int]("locked"), expectedLabel: "LEFT:locked", expectRight: false},
		{scenario: "Right carries untouched", current: either.Right[string, int](9), expectedLabel: "9", expectRight: true},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			mapped := either.MapLeft(testCase.current, func(lockReason string) string { return "LEFT:" + lockReason })
			rendered := either.Match(mapped,
				func(label string) string { return label },
				func(slotCount int) string {
					if slotCount == 9 {
						return "9"
					}
					return "other"
				})
			if rendered != testCase.expectedLabel {
				t.Fatalf("MapLeft Match = %q, want %q", rendered, testCase.expectedLabel)
			}
			if mapped.IsRight() != testCase.expectRight {
				t.Fatalf("MapLeft IsRight = %v, want %v", mapped.IsRight(), testCase.expectRight)
			}
		})
	}
}

func TestMapRightSkipsTransformOnCarriedSide(t *testing.T) {
	ranTransform := false
	current := either.Left[string, int]("locked")
	mapped := either.MapRight(current, func(slotCount int) string {
		ranTransform = true
		return "unreachable"
	})
	if !mapped.IsLeft() {
		t.Fatal("MapRight on Left IsLeft = false, want true")
	}
	if ranTransform {
		t.Fatal("MapRight ran transform on Left, want skip")
	}
}

func TestMapLeftSkipsTransformOnCarriedSide(t *testing.T) {
	ranTransform := false
	current := either.Right[string, int](5)
	mapped := either.MapLeft(current, func(lockReason string) string {
		ranTransform = true
		return "unreachable"
	})
	if !mapped.IsRight() {
		t.Fatal("MapLeft on Right IsRight = false, want true")
	}
	if ranTransform {
		t.Fatal("MapLeft ran transform on Right, want skip")
	}
}

func TestSwapExchangesSides(t *testing.T) {
	swapped := either.Left[string, int]("locked").Swap()
	if !swapped.IsRight() {
		t.Fatal("Swap(Left) IsRight = false, want true")
	}
	rendered := either.Match(swapped,
		func(packTitle int) string { return "unexpected left" },
		func(lockReason string) string { return lockReason })
	if rendered != "locked" {
		t.Fatalf("Swap(Left) payload = %q, want %q", rendered, "locked")
	}

	back := either.Right[string, int](3).Swap()
	if !back.IsLeft() {
		t.Fatal("Swap(Right) IsLeft = false, want true")
	}
}

func TestSwapRoundTrip(t *testing.T) {
	testCases := []struct {
		scenario string
		check    func() bool
	}{
		{scenario: "Left survives double swap", check: func() bool {
			original := either.Left[string, int]("locked")
			roundTripped := original.Swap().Swap()
			if !roundTripped.IsLeft() {
				return false
			}
			return either.Match(roundTripped,
				func(lockReason string) string { return lockReason },
				func(slotCount int) string { return "unexpected right" }) == "locked"
		}},
		{scenario: "Right survives double swap", check: func() bool {
			original := either.Right[string, int](11)
			roundTripped := original.Swap().Swap()
			if !roundTripped.IsRight() {
				return false
			}
			return either.Match(roundTripped,
				func(lockReason string) string { return "unexpected left" },
				func(slotCount int) string {
					if slotCount == 11 {
						return "kept"
					}
					return "changed"
				}) == "kept"
		}},
		{scenario: "zero Left survives double swap", check: func() bool {
			var original either.Either[string, int]
			roundTripped := original.Swap().Swap()
			return roundTripped.IsLeft()
		}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			if !testCase.check() {
				t.Fatalf("round trip failed for %q", testCase.scenario)
			}
		})
	}
}

func TestToResultBothPaths(t *testing.T) {
	sentinel := errors.New("season pass missing")
	testCases := []struct {
		scenario      string
		current       either.Either[string, int]
		expectOk      bool
		expectedValue int
	}{
		{scenario: "Right becomes Ok", current: either.Right[string, int](42), expectOk: true, expectedValue: 42},
		{scenario: "Left becomes Fail", current: either.Left[string, int]("season pass missing"), expectOk: false},
		{scenario: "zero Left becomes Fail", current: either.Either[string, int]{}, expectOk: false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			converted := either.ToResult(testCase.current, func(lockReason string) error {
				if lockReason == "season pass missing" {
					return sentinel
				}
				return errors.New("locked: " + lockReason)
			})
			if converted.IsOk() != testCase.expectOk {
				t.Fatalf("ToResult IsOk = %v, want %v", converted.IsOk(), testCase.expectOk)
			}
			outcome, failure := converted.Unwrap()
			if testCase.expectOk {
				if failure != nil {
					t.Fatalf("ToResult Ok failure = %v, want nil", failure)
				}
				if outcome != testCase.expectedValue {
					t.Fatalf("ToResult Ok value = %d, want %d", outcome, testCase.expectedValue)
				}
				return
			}
			if failure == nil {
				t.Fatal("ToResult Fail failure = nil, want error")
			}
		})
	}
}

func TestToResultLeftCarriesSentinel(t *testing.T) {
	sentinel := errors.New("locked")
	converted := either.ToResult(either.Left[string, int]("any reason"), func(lockReason string) error {
		return sentinel
	})
	if converted.IsOk() {
		t.Fatal("ToResult(Left) IsOk = true, want false")
	}
	_, failure := converted.Unwrap()
	if !errors.Is(failure, sentinel) {
		t.Fatalf("ToResult failure = %v, want %v", failure, sentinel)
	}
}

func TestToResultRightIgnoresFailureOf(t *testing.T) {
	current := either.Right[string, int](8)
	converted := either.ToResult(current, func(lockReason string) error {
		t.Fatal("failureOf ran on Right, want skip")
		return errors.New("unreachable")
	})
	got, failure := converted.Unwrap()
	if failure != nil {
		t.Fatalf("ToResult(Right) failure = %v, want nil", failure)
	}
	if got != 8 {
		t.Fatalf("ToResult(Right) value = %d, want 8", got)
	}
}
