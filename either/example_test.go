package either_test

import (
	"fmt"

	"github.com/seinra/axiom/either"
)

// ExampleRight shows the two-sided path: hold one side, render both with Match.
func ExampleRight() {
	render := func(reason string) string { return "locked: " + reason }
	describe := func(packTitle string) string { return "ready: " + packTitle }
	granted := either.Right[string, string]("Frozen Depths")
	fmt.Println(either.Match(granted, render, describe))
	locked := either.Left[string, string]("season pass missing")
	fmt.Println(either.Match(locked, render, describe))
	// Output:
	// ready: Frozen Depths
	// locked: season pass missing
}
