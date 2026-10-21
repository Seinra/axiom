package slice_test

import (
	"fmt"

	"github.com/seinra/axiom/slice"
)

// ExampleMapSlice shows the transform-filter-fold pipeline over a slice.
func ExampleMapSlice() {
	doubled := slice.MapSlice([]int{1, 2, 3}, func(value int) int { return value * 2 })
	fmt.Println(doubled)
	halved := slice.FilterSlice(doubled, func(value int) bool { return value%4 == 0 })
	fmt.Println(halved)
	total := slice.ReduceSlice(doubled, 0, func(accumulation, value int) int { return accumulation + value })
	fmt.Println(total)
	// Output:
	// [2 4 6]
	// [4]
	// 12
}
