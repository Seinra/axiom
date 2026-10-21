package singleflight_test

import (
	"fmt"

	"github.com/seinra/axiom/singleflight"
)

// ExampleGroup_Call shows one key collapsing into a single execution.
func ExampleGroup_Call() {
	var flight singleflight.Group[string, string]
	executions := 0
	outcome, _, _ := flight.Call("greeting", func() (string, error) {
		executions++
		return "hello", nil
	})
	fmt.Println(outcome)
	fmt.Println(executions)
	// Output:
	// hello
	// 1
}
