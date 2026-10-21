package iterx_test

import (
	"fmt"
	"slices"

	"github.com/seinra/axiom/iterx"
)

// ExampleMapSeq shows the lazy pipeline: transform, filter, take, collect.
func ExampleMapSeq() {
	doubled := iterx.MapSeq(slices.Values([]int{1, 2, 3}), func(value int) int { return value * 2 })
	fmt.Println(slices.Collect(doubled))
	evens := iterx.FilterSeq(slices.Values([]int{1, 2, 3, 4}), func(value int) bool { return value%2 == 0 })
	fmt.Println(slices.Collect(evens))
	firstTwo := iterx.Take(slices.Values([]int{1, 2, 3, 4}), 2)
	fmt.Println(slices.Collect(firstTwo))
	// Output:
	// [2 4 6]
	// [2 4]
	// [1 2]
}
