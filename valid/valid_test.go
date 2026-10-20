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
