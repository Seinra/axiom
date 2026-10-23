package docs

import (
	"fmt"
	"testing"

	"github.com/seinra/axiom/opt"
	"github.com/seinra/axiom/result"
	axiomslice "github.com/seinra/axiom/slice"
)

// TestFoundationsMechanism executes the docs/guides/foundations.md example
// and asserts the exact outputs documented there: inference from value
// arguments, explicit type arguments where inference has nothing to read,
// a constraint-governed call, a generic method with its own type params,
// and each package's zero-value rule.
func TestFoundationsMechanism(t *testing.T) {
	t.Parallel()

	// Inference from value arguments: no type list written.
	doubled := axiomslice.MapSlice([]int{1, 2, 3}, func(n int) int { return n * 2 })

	// Explicit where inference has nothing to read: zero-arg constructor.
	missing := opt.None[string]().ValueOr("nobody")

	// Constraint-governed call: Min works because int is cmp.Ordered.
	smallest, found := axiomslice.Min([]int{3, 1, 2})

	// Generic method with its own type params, chained left to right.
	chainedValue, _ := result.Ok(21).MapValue(func(value int) int { return value * 2 }).Unwrap()

	// Zero values follow each package's rule, never a guess.
	var emptyOpt opt.Option[int]
	var zeroResult result.Result[int]
	_, zeroFailure := zeroResult.Unwrap()

	got := []string{
		fmt.Sprint(doubled),
		missing,
		fmt.Sprintf("%d %v", smallest, found),
		fmt.Sprint(chainedValue),
		fmt.Sprint(emptyOpt.ValueOr(0)),
		fmt.Sprint(zeroFailure != nil),
	}
	want := []string{
		"[2 4 6]",
		"nobody",
		"1 true",
		"42",
		"0",
		"true",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("foundations output[%d] = %q, want %q (full: %q)", i, got[i], want[i], got)
		}
	}
}
