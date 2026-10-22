package stack_test

import (
	"fmt"

	"github.com/seinra/axiom/stack"
)

// ExampleNew shows the LIFO path: Push adds on top, Peek reads without
// removing, Pop removes from the top down.
func ExampleNew() {
	container := stack.New[string]()
	container.Push("first")
	container.Push("top")
	top, _ := container.Peek()
	fmt.Println(top)
	fmt.Println(container.Len())
	for !container.IsEmpty() {
		next, _ := container.Pop()
		fmt.Println(next)
	}
	// Output:
	// top
	// 2
	// top
	// first
}
