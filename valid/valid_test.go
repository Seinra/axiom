// Package valid_test verifies the valid package against exam score scenarios.
package valid_test

import (
	"strings"
	"testing"

	"github.com/seinra/axiom/valid"
)

const (
	minExamScore     = 0
	maxExamScore     = 100
	passingExamScore = 60
)

// buildRangeRule returns the Catalyst exam score range rule (0-100).
func buildRangeRule() valid.Rule[int] {
	return valid.Rule[int]{
		Name: "score in range",
		Check: func(candidateScore int) bool {
			return candidateScore >= minExamScore && candidateScore <= maxExamScore
		},
		Message: "exam score must be between 0 and 100",
	}
}

// buildPassingRule returns the Catalyst exam passing rule (score >= 60).
func buildPassingRule() valid.Rule[int] {
	return valid.Rule[int]{
		Name: "score passing",
		Check: func(candidateScore int) bool {
			return candidateScore >= passingExamScore
		},
		Message: "exam score must reach the passing mark",
	}
}

func TestRuleApplyExamScore(t *testing.T) {
	rangeRule := buildRangeRule()
	examCases := []struct {
		scenario       string
		candidateScore int
		wantFailure    bool
	}{
		{scenario: "passing score within range returns nil", candidateScore: 85, wantFailure: false},
		{scenario: "boundary score zero returns nil", candidateScore: 0, wantFailure: false},
		{scenario: "boundary score one hundred returns nil", candidateScore: 100, wantFailure: false},
		{scenario: "failing score above range returns failure", candidateScore: 150, wantFailure: true},
		{scenario: "failing score below range returns failure", candidateScore: -5, wantFailure: true},
	}
	for _, examCase := range examCases {
		t.Run(examCase.scenario, func(t *testing.T) {
			applyFailure := rangeRule.Apply(examCase.candidateScore)
			if !examCase.wantFailure && applyFailure != nil {
				t.Fatalf("Apply(%d) failure = %v, want nil", examCase.candidateScore, applyFailure)
			}
			if examCase.wantFailure && applyFailure == nil {
				t.Fatalf("Apply(%d) failure = nil, want failure carrying rule name", examCase.candidateScore)
			}
			if examCase.wantFailure {
				failureText := applyFailure.Error()
				if !strings.Contains(failureText, rangeRule.Name) {
					t.Fatalf("failure %q does not carry rule name %q", failureText, rangeRule.Name)
				}
				if !strings.Contains(failureText, rangeRule.Message) {
					t.Fatalf("failure %q does not carry rule message %q", failureText, rangeRule.Message)
				}
			}
		})
	}
}

func TestValidatorCollectsMultipleFailures(t *testing.T) {
	scoreValidator := valid.New(buildRangeRule(), buildPassingRule())
	examCases := []struct {
		scenario          string
		candidateScore    int
		wantFailureCount  int
		wantValidatorName []string
	}{
		{scenario: "valid score collects no failures", candidateScore: 85, wantFailureCount: 0},
		{scenario: "passing but out of range collects one failure", candidateScore: 150, wantFailureCount: 1},
		{scenario: "in range but not passing collects one failure", candidateScore: 45, wantFailureCount: 1},
		{scenario: "negative score collects two failures", candidateScore: -5, wantFailureCount: 2},
	}
	for _, examCase := range examCases {
		t.Run(examCase.scenario, func(t *testing.T) {
			validationFailures := scoreValidator.Validate(examCase.candidateScore)
			if validationFailures == nil {
				t.Fatal("Validate returned nil slice, want empty non-nil slice")
			}
			if len(validationFailures) != examCase.wantFailureCount {
				t.Fatalf("Validate(%d) failures = %d, want %d", examCase.candidateScore, len(validationFailures), examCase.wantFailureCount)
			}
			if examCase.wantFailureCount == 2 {
				combinedText := validationFailures[0].Error() + "|" + validationFailures[1].Error()
				if !strings.Contains(combinedText, "score in range") {
					t.Fatalf("combined failures %q missing range rule name", combinedText)
				}
				if !strings.Contains(combinedText, "score passing") {
					t.Fatalf("combined failures %q missing passing rule name", combinedText)
				}
			}
		})
	}
}

func TestValidatorValidExamScore(t *testing.T) {
	scoreValidator := valid.New(buildRangeRule(), buildPassingRule())
	examCases := []struct {
		scenario       string
		candidateScore int
		wantValid      bool
	}{
		{scenario: "score in range and passing is valid", candidateScore: 85, wantValid: true},
		{scenario: "score above range is invalid", candidateScore: 150, wantValid: false},
		{scenario: "score below passing is invalid", candidateScore: 45, wantValid: false},
		{scenario: "negative score is invalid", candidateScore: -5, wantValid: false},
	}
	for _, examCase := range examCases {
		t.Run(examCase.scenario, func(t *testing.T) {
			observedValid := scoreValidator.Valid(examCase.candidateScore)
			if observedValid != examCase.wantValid {
				t.Fatalf("Valid(%d) = %v, want %v", examCase.candidateScore, observedValid, examCase.wantValid)
			}
		})
	}
}

func TestEmptyValidatorAcceptsEverything(t *testing.T) {
	emptyValidator := valid.New[int]()
	examCases := []struct {
		scenario       string
		candidateScore int
	}{
		{scenario: "negative score accepted", candidateScore: -999},
		{scenario: "zero accepted", candidateScore: 0},
		{scenario: "mid score accepted", candidateScore: 85},
		{scenario: "score above range accepted", candidateScore: 101},
		{scenario: "large score accepted", candidateScore: 1000},
	}
	for _, examCase := range examCases {
		t.Run(examCase.scenario, func(t *testing.T) {
			validationFailures := emptyValidator.Validate(examCase.candidateScore)
			if validationFailures == nil {
				t.Fatal("Validate returned nil slice, want empty non-nil slice")
			}
			if len(validationFailures) != 0 {
				t.Fatalf("Validate(%d) failures = %d, want 0", examCase.candidateScore, len(validationFailures))
			}
			if !emptyValidator.Valid(examCase.candidateScore) {
				t.Fatalf("Valid(%d) = false, want true", examCase.candidateScore)
			}
		})
	}
}

func TestNilCheckCarriesName(t *testing.T) {
	namelessCheck := valid.Rule[int]{
		Name:    "score checked",
		Message: "exam score failed an absent check",
	}
	applyFailure := namelessCheck.Apply(85)
	if applyFailure == nil {
		t.Fatal("Apply with nil Check failure = nil, want failure carrying rule name")
	}
	failureText := applyFailure.Error()
	if !strings.Contains(failureText, namelessCheck.Name) {
		t.Fatalf("failure %q does not carry rule name %q", failureText, namelessCheck.Name)
	}
	if !strings.Contains(failureText, namelessCheck.Message) {
		t.Fatalf("failure %q does not carry rule message %q", failureText, namelessCheck.Message)
	}
}

// buildRangePredicate returns the Catalyst exam score range predicate (0-100).
func buildRangePredicate() valid.Predicate[int] {
	return func(candidateScore int) bool {
		return candidateScore >= minExamScore && candidateScore <= maxExamScore
	}
}

// buildPassingPredicate returns the Catalyst exam passing predicate (score >= 60).
func buildPassingPredicate() valid.Predicate[int] {
	return func(candidateScore int) bool {
		return candidateScore >= passingExamScore
	}
}

func TestAndExamScore(t *testing.T) {
	isInRange := buildRangePredicate()
	isPassing := buildPassingPredicate()
	examCases := []struct {
		scenario       string
		candidateScore int
		combined       valid.Predicate[int]
		wantSatisfied  bool
	}{
		{scenario: "score in range and passing satisfies", candidateScore: 85, combined: valid.And(isInRange, isPassing), wantSatisfied: true},
		{scenario: "score in range but failing violates one leg", candidateScore: 45, combined: valid.And(isInRange, isPassing), wantSatisfied: false},
		{scenario: "score passing but out of range violates one leg", candidateScore: 150, combined: valid.And(isInRange, isPassing), wantSatisfied: false},
		{scenario: "score violating both legs dissatisfies", candidateScore: -5, combined: valid.And(isInRange, isPassing), wantSatisfied: false},
		{scenario: "single passing predicate satisfies alone", candidateScore: 85, combined: valid.And(isPassing), wantSatisfied: true},
		{scenario: "vacuous empty conjunction holds", candidateScore: -999, combined: valid.And[int](), wantSatisfied: true},
	}
	for _, examCase := range examCases {
		t.Run(examCase.scenario, func(t *testing.T) {
			observedSatisfied := examCase.combined(examCase.candidateScore)
			if observedSatisfied != examCase.wantSatisfied {
				t.Fatalf("And(%d) = %v, want %v", examCase.candidateScore, observedSatisfied, examCase.wantSatisfied)
			}
		})
	}
}

func TestAndShortCircuitsOnFirstFalse(t *testing.T) {
	firstCallCount := 0
	secondCallCount := 0
	thirdCallCount := 0
	isAlwaysTrue := func(candidateScore int) bool {
		firstCallCount++
		return candidateScore >= minExamScore
	}
	isNeverTrue := func(candidateScore int) bool {
		secondCallCount++
		return false
	}
	isUnreached := func(candidateScore int) bool {
		thirdCallCount++
		return candidateScore <= maxExamScore
	}
	combined := valid.And(isAlwaysTrue, isNeverTrue, isUnreached)
	if combined(85) {
		t.Fatal("And with a false leg returned true, want false")
	}
	if firstCallCount != 1 {
		t.Fatalf("first predicate calls = %d, want 1", firstCallCount)
	}
	if secondCallCount != 1 {
		t.Fatalf("second predicate calls = %d, want 1", secondCallCount)
	}
	if thirdCallCount != 0 {
		t.Fatalf("third predicate calls = %d, want 0 (short-circuited)", thirdCallCount)
	}
}

func TestOrExamScore(t *testing.T) {
	isPassing := buildPassingPredicate()
	isPerfect := func(candidateScore int) bool { return candidateScore == maxExamScore }
	examCases := []struct {
		scenario       string
		candidateScore int
		combined       valid.Predicate[int]
		wantSatisfied  bool
	}{
		{scenario: "passing score satisfies one leg", candidateScore: 85, combined: valid.Or(isPassing, isPerfect), wantSatisfied: true},
		{scenario: "perfect score satisfies second leg", candidateScore: 100, combined: valid.Or(func(candidateScore int) bool { return false }, isPerfect), wantSatisfied: true},
		{scenario: "failing score dissatisfies every leg", candidateScore: 45, combined: valid.Or(isPassing, isPerfect), wantSatisfied: false},
		{scenario: "single failing predicate dissatisfies alone", candidateScore: 45, combined: valid.Or(isPassing), wantSatisfied: false},
		{scenario: "vacuous empty disjunction fails", candidateScore: 85, combined: valid.Or[int](), wantSatisfied: false},
	}
	for _, examCase := range examCases {
		t.Run(examCase.scenario, func(t *testing.T) {
			observedSatisfied := examCase.combined(examCase.candidateScore)
			if observedSatisfied != examCase.wantSatisfied {
				t.Fatalf("Or(%d) = %v, want %v", examCase.candidateScore, observedSatisfied, examCase.wantSatisfied)
			}
		})
	}
}

func TestOrShortCircuitsOnFirstTrue(t *testing.T) {
	firstCallCount := 0
	secondCallCount := 0
	isPassing := func(candidateScore int) bool {
		firstCallCount++
		return candidateScore >= passingExamScore
	}
	isUnreached := func(candidateScore int) bool {
		secondCallCount++
		return candidateScore <= maxExamScore
	}
	combined := valid.Or(isPassing, isUnreached)
	if !combined(85) {
		t.Fatal("Or with a true first leg returned false, want true")
	}
	if firstCallCount != 1 {
		t.Fatalf("first predicate calls = %d, want 1", firstCallCount)
	}
	if secondCallCount != 0 {
		t.Fatalf("second predicate calls = %d, want 0 (short-circuited)", secondCallCount)
	}
}

func TestNotExamScore(t *testing.T) {
	isPassing := buildPassingPredicate()
	examCases := []struct {
		scenario       string
		candidateScore int
		negated        valid.Predicate[int]
		wantSatisfied  bool
	}{
		{scenario: "failing score negates to true", candidateScore: 45, negated: valid.Not(isPassing), wantSatisfied: true},
		{scenario: "passing score negates to false", candidateScore: 85, negated: valid.Not(isPassing), wantSatisfied: false},
		{scenario: "nil predicate negates to always false", candidateScore: 45, negated: valid.Not[int](nil), wantSatisfied: false},
		{scenario: "nil predicate stays false on passing score", candidateScore: 85, negated: valid.Not[int](nil), wantSatisfied: false},
	}
	for _, examCase := range examCases {
		t.Run(examCase.scenario, func(t *testing.T) {
			observedSatisfied := examCase.negated(examCase.candidateScore)
			if observedSatisfied != examCase.wantSatisfied {
				t.Fatalf("Not(%d) = %v, want %v", examCase.candidateScore, observedSatisfied, examCase.wantSatisfied)
			}
		})
	}
}

func TestComposedAndThroughValidator(t *testing.T) {
	scoreValidator := valid.New(valid.Rule[int]{
		Name:    "score in range and passing",
		Check:   valid.And(buildRangePredicate(), buildPassingPredicate()),
		Message: "exam score must be between 0 and 100 and reach the passing mark",
	})
	examCases := []struct {
		scenario       string
		candidateScore int
		wantValid      bool
	}{
		{scenario: "score in range and passing is valid", candidateScore: 85, wantValid: true},
		{scenario: "score in range but failing is invalid", candidateScore: 45, wantValid: false},
		{scenario: "score passing but out of range is invalid", candidateScore: 150, wantValid: false},
		{scenario: "negative score is invalid", candidateScore: -5, wantValid: false},
	}
	for _, examCase := range examCases {
		t.Run(examCase.scenario, func(t *testing.T) {
			observedValid := scoreValidator.Valid(examCase.candidateScore)
			if observedValid != examCase.wantValid {
				t.Fatalf("Valid(%d) = %v, want %v", examCase.candidateScore, observedValid, examCase.wantValid)
			}
		})
	}
}

func TestAndNilLegCountsAsFalse(t *testing.T) {
	isPassing := buildPassingPredicate()
	if valid.And[int](nil)(85) {
		t.Fatal("And(nil)(85) = true, want false (nil leg counts as false)")
	}
	if valid.And(isPassing, nil)(85) {
		t.Fatal("And(passing, nil)(85) = true, want false (nil leg counts as false)")
	}
	if valid.And[int](nil, isPassing)(85) {
		t.Fatal("And(nil, passing)(85) = true, want false (nil leg counts as false)")
	}
	secondCalls := 0
	reached := func(candidateScore int) bool {
		secondCalls++
		return true
	}
	if valid.And[int](nil, reached)(85) {
		t.Fatal("And(nil, reached)(85) = true, want false")
	}
	if secondCalls != 0 {
		t.Fatalf("second predicate calls = %d, want 0 (nil leg returns before later legs)", secondCalls)
	}
}

func TestOrNilLegIsSkipped(t *testing.T) {
	isPassing := buildPassingPredicate()
	if valid.Or[int](nil)(85) {
		t.Fatal("Or(nil)(85) = true, want false (all-nil disjunction fails)")
	}
	if !valid.Or(nil, isPassing)(85) {
		t.Fatal("Or(nil, passing)(85) = false, want true (nil leg skipped)")
	}
	if valid.Or(nil, isPassing)(45) {
		t.Fatal("Or(nil, passing)(45) = true, want false (no leg holds)")
	}
	isFailing := func(candidateScore int) bool { return false }
	if valid.Or(isFailing, nil)(45) {
		t.Fatal("Or(failing, nil)(45) = true, want false (nil leg counts as false)")
	}
	reachedCalls := 0
	reached := func(candidateScore int) bool {
		reachedCalls++
		return true
	}
	if !valid.Or[int](nil, reached)(85) {
		t.Fatal("Or(nil, reached)(85) = false, want true (nil leg skipped, later leg runs)")
	}
	if reachedCalls != 1 {
		t.Fatalf("reached calls = %d, want 1 (nil leg skipped, not terminal)", reachedCalls)
	}
	unreachedCalls := 0
	unreached := func(candidateScore int) bool {
		unreachedCalls++
		return true
	}
	if !valid.Or(isPassing, unreached)(85) {
		t.Fatal("Or(passing, unreached)(85) = false, want true")
	}
	if unreachedCalls != 0 {
		t.Fatalf("unreached calls = %d, want 0 (short-circuited after true)", unreachedCalls)
	}
}

func TestNewCopiesRulesSliceIsolation(t *testing.T) {
	rangeRule := buildRangeRule()
	rules := []valid.Rule[int]{rangeRule}
	scoreValidator := valid.New(rules...)
	// Mutate the caller slice after New: overwrite the slot and grow the
	// slice. Neither may leak into the Validator.
	rules[0] = buildPassingRule()
	grown := append(rules, buildPassingRule())
	if len(grown) != 2 {
		t.Fatalf("grown rules = %d, want 2 (append happened, Validator must still hold the original)", len(grown))
	}
	// 150 violates the original range rule only, so the isolated Validator
	// reports exactly one failure carrying the original name and message.
	validationFailures := scoreValidator.Validate(150)
	if len(validationFailures) != 1 {
		t.Fatalf("Validate(150) failures = %d, want 1 (original range rule retained)", len(validationFailures))
	}
	if failureText := validationFailures[0].Error(); !strings.Contains(failureText, rangeRule.Name) {
		t.Fatalf("failure %q does not carry original rule name %q", failureText, rangeRule.Name)
	}
	if failureText := validationFailures[0].Error(); !strings.Contains(failureText, rangeRule.Message) {
		t.Fatalf("failure %q does not carry original rule message %q", failureText, rangeRule.Message)
	}
}

func TestValidateFailureOrderFollowsRuleOrder(t *testing.T) {
	rangeRule := buildRangeRule()
	passingRule := buildPassingRule()
	forwardValidator := valid.New(rangeRule, passingRule)
	forwardFailures := forwardValidator.Validate(-5)
	if len(forwardFailures) != 2 {
		t.Fatalf("forward Validate(-5) failures = %d, want 2", len(forwardFailures))
	}
	if failureText := forwardFailures[0].Error(); !strings.Contains(failureText, rangeRule.Name) {
		t.Fatalf("forward failures[0] = %q, want range rule first", failureText)
	}
	if failureText := forwardFailures[1].Error(); !strings.Contains(failureText, passingRule.Name) {
		t.Fatalf("forward failures[1] = %q, want passing rule second", failureText)
	}
	reversedValidator := valid.New(passingRule, rangeRule)
	reversedFailures := reversedValidator.Validate(-5)
	if len(reversedFailures) != 2 {
		t.Fatalf("reversed Validate(-5) failures = %d, want 2", len(reversedFailures))
	}
	if failureText := reversedFailures[0].Error(); !strings.Contains(failureText, passingRule.Name) {
		t.Fatalf("reversed failures[0] = %q, want passing rule first", failureText)
	}
	if failureText := reversedFailures[1].Error(); !strings.Contains(failureText, rangeRule.Name) {
		t.Fatalf("reversed failures[1] = %q, want range rule second", failureText)
	}
}
