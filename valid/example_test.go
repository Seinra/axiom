package valid_test

import (
	"fmt"

	"github.com/seinra/axiom/valid"
)

// ExampleNew shows the Predicate-Rule-Validator exam-score path.
func ExampleNew() {
	passing := valid.Rule[int]{
		Name:    "passing score",
		Check:   func(score int) bool { return score >= 60 },
		Message: "score below passing grade",
	}
	exam := valid.New(passing)
	fmt.Println(exam.Valid(85))
	fmt.Println(exam.Valid(42))
	fmt.Println(exam.Validate(42))
	// Output:
	// true
	// false
	// [axiom valid: rule "passing score" failed: score below passing grade]
}
