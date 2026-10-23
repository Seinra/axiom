package task_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/seinra/axiom/task"
)

// ExampleSpawn shows the eager background computation: Spawn starts compute
// at once and Await joins it for the shared outcome.
func ExampleSpawn() {
	future := task.Spawn(context.Background(), func(_ context.Context) (int, error) {
		return 40 + 2, nil
	})
	success, failure := future.Await().Unwrap()
	fmt.Println(success, failure)
	// Output: 42 <nil>
}

// ExampleTask_Await shows the shared failure path: every Await reports the
// same Result, so a compute failure reads identically on each waiter.
func ExampleTask_Await() {
	compute := func(_ context.Context) (int, error) {
		return 0, errors.New("boom")
	}
	future := task.Spawn(context.Background(), compute)
	_, first := future.Await().Unwrap()
	_, second := future.Await().Unwrap()
	fmt.Println(first, second)
	// Output: boom boom
}
