package set_test

import (
	"fmt"

	"github.com/seinra/axiom/set"
)

// ExampleOf shows the distinct-membership path: seed, query, combine.
func ExampleOf() {
	roster := set.Of("ada", "grace")
	fmt.Println(roster.Has("ada"))
	fmt.Println(roster.Len())
	combined := roster.Union(set.Of("grace", "hedy"))
	fmt.Println(combined.Len())
	fmt.Println(combined.Has("hedy"))
	// Output:
	// true
	// 2
	// 3
	// true
}
