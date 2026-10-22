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

// ExampleMapSeqResults shows the ordered collect-all stream: every input
// yields exactly one Result in stream order, so a Fail sits alongside its
// Ok siblings instead of canceling them.
func ExampleMapSeqResults() {
	sequence := slices.Values([]int{1, 2, 3})
	streamed := flow.MapSeqResults(context.Background(), sequence, 2, func(_ context.Context, candidate int) (int, error) {
		if candidate == 2 {
			return 0, fmt.Errorf("bad candidate")
		}
		return candidate * 2, nil
	})
	for outcome := range streamed {
		success, failure := outcome.Unwrap()
		if failure != nil {
			fmt.Println("fail")
			continue
		}
		fmt.Println(success)
	}
	// Output:
	// 2
	// fail
	// 6
}
