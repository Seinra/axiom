package flow_test

import (
	"context"
	"fmt"

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
