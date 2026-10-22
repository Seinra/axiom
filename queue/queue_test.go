package queue_test

import (
	"testing"

	"github.com/seinra/axiom/queue"
)

func TestNewIsEmpty(t *testing.T) {
	container := queue.New[int]()
	if container == nil {
		t.Fatal("New = nil, want non-nil queue")
	}
	if got := container.Len(); got != 0 {
		t.Fatalf("Len = %d, want 0", got)
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true")
	}
}

func TestZeroValueIsReady(t *testing.T) {
	var container queue.Queue[string]
	if got := container.Len(); got != 0 {
		t.Fatalf("Len = %d, want 0", got)
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true")
	}
	container.Enqueue("gopher")
	if got := container.Len(); got != 1 {
		t.Fatalf("Len = %d, want 1", got)
	}
	if container.IsEmpty() {
		t.Fatal("IsEmpty = true, want false")
	}
	if got, held := container.Dequeue(); !held || got != "gopher" {
		t.Fatalf("Dequeue = (%q, %t), want (%q, true)", got, held, "gopher")
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true")
	}
}

func TestEnqueueDequeueIsFIFO(t *testing.T) {
	container := queue.New[int]()
	container.Enqueue(1)
	container.Enqueue(2)
	container.Enqueue(3)
	for _, wanted := range []int{1, 2, 3} {
		got, held := container.Dequeue()
		if !held {
			t.Fatalf("Dequeue held = false, want true for %d", wanted)
		}
		if got != wanted {
			t.Fatalf("Dequeue = %d, want %d", got, wanted)
		}
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true after draining")
	}
}

func TestDequeueEmptyYieldsZeroWithoutPanic(t *testing.T) {
	container := queue.New[int]()
	got, held := container.Dequeue()
	if held {
		t.Fatal("Dequeue held = true, want false on empty queue")
	}
	if got != 0 {
		t.Fatalf("Dequeue = %d, want zero value 0", got)
	}
	if got := container.Len(); got != 0 {
		t.Fatalf("Len = %d, want 0 after empty Dequeue", got)
	}
}

func TestPeekEmptyYieldsZeroWithoutPanic(t *testing.T) {
	container := queue.New[int]()
	got, held := container.Peek()
	if held {
		t.Fatal("Peek held = true, want false on empty queue")
	}
	if got != 0 {
		t.Fatalf("Peek = %d, want zero value 0", got)
	}
}

func TestPeekKeepsFront(t *testing.T) {
	container := queue.New[string]()
	container.Enqueue("front")
	container.Enqueue("back")
	for round := 0; round < 2; round++ {
		got, held := container.Peek()
		if !held || got != "front" {
			t.Fatalf("Peek = (%q, %t), want (%q, true)", got, held, "front")
		}
	}
	if got := container.Len(); got != 2 {
		t.Fatalf("Len = %d, want 2 after Peek", got)
	}
	if got, held := container.Dequeue(); !held || got != "front" {
		t.Fatalf("Dequeue = (%q, %t), want (%q, true)", got, held, "front")
	}
	if got, held := container.Peek(); !held || got != "back" {
		t.Fatalf("Peek = (%q, %t), want (%q, true)", got, held, "back")
	}
}

func TestInterleavedKeepsOrder(t *testing.T) {
	container := queue.New[int]()
	for _, candidate := range []int{1, 2, 3, 4, 5} {
		container.Enqueue(candidate)
	}
	for _, wanted := range []int{1, 2, 3} {
		if got, held := container.Dequeue(); !held || got != wanted {
			t.Fatalf("Dequeue = (%d, %t), want (%d, true)", got, held, wanted)
		}
	}
	for _, candidate := range []int{6, 7, 8} {
		container.Enqueue(candidate)
	}
	for _, wanted := range []int{4, 5, 6, 7, 8} {
		if got, held := container.Dequeue(); !held || got != wanted {
			t.Fatalf("Dequeue = (%d, %t), want (%d, true)", got, held, wanted)
		}
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true after draining")
	}
}

func TestCompactionKeepsOrder(t *testing.T) {
	const total = 200
	const drained = 150
	container := queue.New[int]()
	for candidate := 0; candidate < total; candidate++ {
		container.Enqueue(candidate)
	}
	for wanted := 0; wanted < drained; wanted++ {
		if got, held := container.Dequeue(); !held || got != wanted {
			t.Fatalf("Dequeue = (%d, %t), want (%d, true)", got, held, wanted)
		}
	}
	for candidate := total; candidate < total+100; candidate++ {
		container.Enqueue(candidate)
	}
	for wanted := drained; wanted < total+100; wanted++ {
		got, held := container.Dequeue()
		if !held {
			t.Fatalf("Dequeue held = false, want true for %d", wanted)
		}
		if got != wanted {
			t.Fatalf("Dequeue = %d, want %d", got, wanted)
		}
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true after draining")
	}
	if got := container.Len(); got != 0 {
		t.Fatalf("Len = %d, want 0 after draining", got)
	}
}

func TestEnqueueAfterDrainReusesQueue(t *testing.T) {
	container := queue.New[int]()
	container.Enqueue(1)
	if _, held := container.Dequeue(); !held {
		t.Fatal("Dequeue held = false, want true")
	}
	container.Enqueue(2)
	got, held := container.Dequeue()
	if !held || got != 2 {
		t.Fatalf("Dequeue = (%d, %t), want (2, true)", got, held)
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true after draining")
	}
}

func TestQueuesStayIndependent(t *testing.T) {
	first := queue.New[int]()
	second := queue.New[int]()
	first.Enqueue(1)
	first.Enqueue(2)
	if !second.IsEmpty() {
		t.Fatal("second IsEmpty = false, want true; queues share backing")
	}
	if got, held := second.Dequeue(); held || got != 0 {
		t.Fatalf("second Dequeue = (%d, %t), want (0, false)", got, held)
	}
	if got := first.Len(); got != 2 {
		t.Fatalf("first Len = %d, want 2", got)
	}
}

func TestCompactionSlideBoundaryKeepsOrder(t *testing.T) {
	container := queue.New[int]()
	const seeded = 70
	const drained = 64
	for candidate := 0; candidate < seeded; candidate++ {
		container.Enqueue(candidate)
	}
	for wanted := 0; wanted < drained; wanted++ {
		if got, held := container.Dequeue(); !held || got != wanted {
			t.Fatalf("Dequeue = (%d, %t), want (%d, true)", got, held, wanted)
		}
	}
	for candidate := seeded; candidate < seeded+10; candidate++ {
		container.Enqueue(candidate)
	}
	wanted := drained
	for !container.IsEmpty() {
		got, held := container.Dequeue()
		if !held {
			t.Fatalf("Dequeue held = false, want true for %d", wanted)
		}
		if got != wanted {
			t.Fatalf("Dequeue = %d, want %d", got, wanted)
		}
		wanted++
	}
	if wanted != seeded+10 {
		t.Fatalf("drained to %d, want %d", wanted, seeded+10)
	}
	if got := container.Len(); got != 0 {
		t.Fatalf("Len = %d, want 0 after draining", got)
	}
}
