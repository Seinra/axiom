package result_test

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/seinra/axiom/result"
)

// ExampleOk shows the success path: wrap, transform, read explicitly.
func ExampleOk() {
	doubled := result.Ok(21).MapValue(func(value int) int { return value * 2 })
	success, _ := doubled.Unwrap()
	fmt.Println(success)
	failed := result.Fail[int](errors.New("boom"))
	fmt.Println(result.UnwrapOr(failed, -1))
	// Output:
	// 42
	// -1
}

// ExampleResult_MarshalJSON shows the outcome codec: Ok renders as its value, Fail as an envelope.
func ExampleResult_MarshalJSON() {
	payload, _ := json.Marshal(result.Ok(42))
	fmt.Println(string(payload))
	failed, _ := json.Marshal(result.Fail[int](errors.New("boom")))
	fmt.Println(string(failed))
	// Output:
	// 42
	// {"error":"boom"}
}
