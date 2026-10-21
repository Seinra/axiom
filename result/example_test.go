package result_test

import (
	"errors"
	"fmt"

	"github.com/seinra/axiom/result"
)

// ExampleOk shows the success path: wrap, transform, read explicitly.
func ExampleOk() {
	doubled := result.Map(result.Ok(21), func(value int) int { return value * 2 })
	success, _ := doubled.Unwrap()
	fmt.Println(success)
	failed := result.Fail[int](errors.New("boom"))
	fmt.Println(result.UnwrapOr(failed, -1))
	// Output:
	// 42
	// -1
}
