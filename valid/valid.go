// Package valid provides small generic validation primitives for learners.
//
// Define a Predicate per condition, wrap each one in a Rule carrying a name
// and a message, then collect Rules in a Validator. Validator.Validate runs
// every Rule and reports all failures; Validator.Valid answers the boolean
// shorthand. The package ships no domain helpers: learners write their own
// Rules, and that restraint is the lesson. Reach for plain ifs and errors
// directly when a single inline check reads clearer than a Rule.
//
// Which function?
//
// | When you want to ...                 | Reach for ...      |
// |---------------------------------------|--------------------|
// | Express one check as a value          | Predicate          |
// | Combine checks                        | And / Or / Not     |
// | Name one check with its message       | Rule (+ Apply)     |
// | Collect rules into a set              | New                |
// | Run every rule, collect all failures  | Validator.Validate |
// | Boolean shorthand for one candidate   | Validator.Valid    |
package valid

import "fmt"

// Predicate reports whether candidate satisfies a single condition.
type Predicate[Value any] func(Value) bool

// And combines predicates into one that holds only when every leg holds.
// An empty conjunction holds vacuously, so And with no predicates always
// reports true. Legs run left to right and stop at the first false, so
// later legs never run after a failure. A nil leg counts as false.
func And[Value any](predicates ...Predicate[Value]) Predicate[Value] {
	return func(candidate Value) bool {
		for _, activePredicate := range predicates {
			if activePredicate == nil {
				return false
			}
			if !activePredicate(candidate) {
				return false
			}
		}
		return true
	}
}

// Or combines predicates into one that holds when any leg holds.
// An empty disjunction fails vacuously, so Or with no predicates always
// reports false. Legs run left to right and stop at the first true, so
// later legs never run after a success. A nil leg counts as false and is skipped.
func Or[Value any](predicates ...Predicate[Value]) Predicate[Value] {
	return func(candidate Value) bool {
		for _, activePredicate := range predicates {
			if activePredicate == nil {
				continue
			}
			if activePredicate(candidate) {
				return true
			}
		}
		return false
	}
}

// Not negates predicate. A nil predicate is a programmer error handled
// as always-false rather than panicking: Not of nil always reports false,
// matching how Validator already accepts empty rule sets gracefully.
func Not[Value any](predicate Predicate[Value]) Predicate[Value] {
	return func(candidate Value) bool {
		if predicate == nil {
			return false
		}
		return !predicate(candidate)
	}
}

// Rule pairs a human-readable Name with a Check and a failure Message.
// Name identifies the rule in errors; Message explains the failure.
type Rule[Value any] struct {
	Name    string
	Check   Predicate[Value]
	Message string
}

// Apply runs the rule Check against candidate.
// Apply returns nil when the Check passes, otherwise a descriptive error
// carrying the rule Name and Message.
func (current Rule[Value]) Apply(candidate Value) error {
	if current.Check == nil {
		return fmt.Errorf("axiom valid: rule %q has no check: %s", current.Name, current.Message)
	}
	if current.Check(candidate) {
		return nil
	}
	return fmt.Errorf("axiom valid: rule %q failed: %s", current.Name, current.Message)
}

// Validator holds an ordered set of Rules applied together.
type Validator[Value any] struct {
	rules []Rule[Value]
}

// New collects rules into a Validator.
// The rules slice is copied so later caller mutations never affect the Validator.
func New[Value any](rules ...Rule[Value]) Validator[Value] {
	copiedRules := make([]Rule[Value], len(rules))
	copy(copiedRules, rules)
	return Validator[Value]{rules: copiedRules}
}

// Validate runs every rule against candidate and collects all failures.
// The result is never nil: an empty slice means candidate is valid.
func (current Validator[Value]) Validate(candidate Value) []error {
	failures := make([]error, 0, len(current.rules))
	for _, activeRule := range current.rules {
		if applyFailure := activeRule.Apply(candidate); applyFailure != nil {
			failures = append(failures, applyFailure)
		}
	}
	return failures
}

// Valid reports whether candidate satisfies every rule.
func (current Validator[Value]) Valid(candidate Value) bool {
	return len(current.Validate(candidate)) == 0
}
