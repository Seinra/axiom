package queue_test

import (
	"fmt"

	"github.com/seinra/axiom/queue"
)

// ExampleNew shows the FIFO path: Enqueue adds at the back, Peek reads the
// front without removing, Dequeue removes from the front forward.
func ExampleNew() {
	container := queue.New[string]()
	container.Enqueue("first")
	container.Enqueue("back")
	front, _ := container.Peek()
	fmt.Println(front)
	fmt.Println(container.Len())
	for !container.IsEmpty() {
		next, _ := container.Dequeue()
		fmt.Println(next)
	}
	// Output:
	// first
	// 2
	// first
	// back
}
