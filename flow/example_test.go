package flow_test

import (
	"context"
	"fmt"
	"slices"

	"github.com/seinra/axiom/flow"
)

// ExampleMap shows the ordered parallel transform.
func ExampleMap() {
	doubled, _ := flow.Map(context.Background(), []int{1, 2, 3}, 2, func(_ context.Context, candidate int) (int, error) {
		return candidate * 2, nil
	})
	fmt.Println(doubled)
	// Output: [2 4 6]
}

// ExampleMapSeq shows the ordered parallel transform over a sequence:
// input streams in, output materializes in stream order.
func ExampleMapSeq() {
	doubled, _ := flow.MapSeq(context.Background(), slices.Values([]int{1, 2, 3}), 2, func(_ context.Context, candidate int) (int, error) {
		return candidate * 2, nil
	})
	fmt.Println(doubled)
	// Output: [2 4 6]
}

// ExampleEachSeq shows the streaming side-effect pass over a sequence: no
// results are collected, so nothing accumulates while the sequence streams.
func ExampleEachSeq() {
	_ = flow.EachSeq(context.Background(), slices.Values([]int{1, 2, 3}), 1, func(_ context.Context, candidate int) error {
		fmt.Println(candidate)
		return nil
	})
	// Output:
	// 1
	// 2
	// 3
}
