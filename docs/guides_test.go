package docs

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/seinra/axiom/orderedmap"
	"github.com/seinra/axiom/queue"
	"github.com/seinra/axiom/stack"
	"github.com/seinra/axiom/task"
)

// TestFamilyGuidesStackQueue executes the docs/guides/stack.md and
// docs/guides/queue.md examples and asserts the exact outputs documented
// there: LIFO drains top-down, FIFO drains front-forward, Peek never
// removes, empty is (zero, false) rather than a panic.
func TestFamilyGuidesStackQueue(t *testing.T) {
	t.Parallel()

	container := stack.New[string]()
	container.Push("first")
	container.Push("top")
	pileTop, _ := container.Peek()
	drainedPile := []string{pileTop, fmt.Sprint(container.Len())}
	for !container.IsEmpty() {
		next, _ := container.Pop()
		drainedPile = append(drainedPile, next)
	}

	line := queue.New[string]()
	line.Enqueue("first")
	line.Enqueue("back")
	front, _ := line.Peek()
	servedLine := []string{front, fmt.Sprint(line.Len())}
	for !line.IsEmpty() {
		next, _ := line.Dequeue()
		servedLine = append(servedLine, next)
	}

	got := []string{
		fmt.Sprint(drainedPile),
		fmt.Sprint(servedLine),
	}
	want := []string{
		"[top 2 top first]",
		"[first 2 first back]",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("family guide output[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestFamilyGuidesOrderedMap executes the docs/guides/orderedmap.md example
// and asserts the exact outputs documented there: first-insertion replay,
// upsert-keeps-seat, and LRU recency eviction.
func TestFamilyGuidesOrderedMap(t *testing.T) {
	t.Parallel()

	ordered := orderedmap.New[string, int]()
	ordered.Put("beta", 2)
	ordered.Put("alpha", 1)
	ordered.Put("beta", 20)

	recent := orderedmap.NewLRU[string, int](2)
	recent.Put("alpha", 1)
	recent.Put("beta", 2)
	_, _ = recent.Get("alpha")
	recent.Put("gamma", 3)
	_, betaHeld := recent.Get("beta")
	alphaValue, _ := recent.Get("alpha")

	got := []string{
		fmt.Sprint(ordered.Keys()),
		fmt.Sprint(ordered.Values()),
		fmt.Sprint(betaHeld),
		fmt.Sprint(alphaValue),
		fmt.Sprint(recent.Len()),
	}
	want := []string{
		"[beta alpha]",
		"[20 1]",
		"false",
		"1",
		"2",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("family guide output[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestFamilyGuidesTask executes the docs/guides/task.md example and asserts
// the exact outputs documented there: eager Spawn joined by Await, and one
// shared failure reading identically on every waiter.
func TestFamilyGuidesTask(t *testing.T) {
	t.Parallel()

	future := task.Spawn(context.Background(), func(_ context.Context) (int, error) {
		return 40 + 2, nil
	})
	success, failure := future.Await().Unwrap()

	compute := func(_ context.Context) (int, error) {
		return 0, errors.New("boom")
	}
	shared := task.Spawn(context.Background(), compute)
	_, first := shared.Await().Unwrap()
	_, second := shared.Await().Unwrap()

	got := []string{
		fmt.Sprintf("%d %v", success, failure),
		fmt.Sprintf("%v %v", first, second),
	}
	want := []string{
		"42 <nil>",
		"boom boom",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("family guide output[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
