package opt_test

import (
	"fmt"
	"strings"

	"github.com/seinra/axiom/opt"
)

// ExampleSome shows the present-value path: hold, transform, read with fallback.
func ExampleSome() {
	present := opt.Some("gopher")
	fmt.Println(present.ValueOr("nobody"))
	fmt.Println(opt.None[string]().ValueOr("nobody"))
	shouted := present.MapValue(strings.ToUpper)
	fmt.Println(shouted.ValueOr("nobody"))
	// Output:
	// gopher
	// nobody
	// GOPHER
}
