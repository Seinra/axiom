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

// ExampleCombinations shows all 2-subsets of three options in index order.
func ExampleCombinations() {
	combos := slice.Combinations([]string{"a", "b", "c"}, 2)
	fmt.Println(combos)
	// Output:
	// [[a b] [a c] [b c]]
}

// ExamplePermutations shows all 2-arrangements of three options in index order.
func ExamplePermutations() {
	orders := slice.Permutations([]string{"a", "b", "c"}, 2)
	fmt.Println(orders)
	// Output:
	// [[a b] [a c] [b a] [b c] [c a] [c b]]
}

// ExampleFirstOr shows the first element, or the fallback when empty.
func ExampleFirstOr() {
	fmt.Println(slice.FirstOr([]string{"go", "rust"}, "none"))
	fmt.Println(slice.FirstOr([]string{}, "none"))
	// Output:
	// go
	// none
}

// ExampleEvery shows the vacuous truth of Every over an all-even slice.
func ExampleEvery() {
	fmt.Println(slice.Every([]int{2, 4, 6}, func(value int) bool { return value%2 == 0 }))
	fmt.Println(slice.Every([]int{2, 3, 6}, func(value int) bool { return value%2 == 0 }))
	// Output:
	// true
	// false
}

// ExampleFindDuplicates shows repeats in order of first occurrence.
func ExampleFindDuplicates() {
	fmt.Println(slice.FindDuplicates([]int{2, 1, 2, 3, 1}))
	// Output:
	// [2 1]
}

// ExampleTake shows the first n elements, copying past the end safely.
func ExampleTake() {
	fmt.Println(slice.Take([]int{1, 2, 3, 4}, 2))
	fmt.Println(slice.Take([]int{1, 2, 3, 4}, 9))
	// Output:
	// [1 2]
	// [1 2 3 4]
}
