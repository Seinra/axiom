package stack_test

import (
	"testing"

	"github.com/seinra/axiom/stack"
)

func TestNewIsEmpty(t *testing.T) {
	container := stack.New[int]()
	if container == nil {
		t.Fatal("New = nil, want non-nil stack")
	}
	if got := container.Len(); got != 0 {
		t.Fatalf("Len = %d, want 0", got)
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true")
	}
}

func TestZeroValueIsReady(t *testing.T) {
	var container stack.Stack[string]
	if got := container.Len(); got != 0 {
		t.Fatalf("Len = %d, want 0", got)
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true")
	}
	container.Push("gopher")
	if got := container.Len(); got != 1 {
		t.Fatalf("Len = %d, want 1", got)
	}
	if container.IsEmpty() {
		t.Fatal("IsEmpty = true, want false")
	}
	if got, held := container.Pop(); !held || got != "gopher" {
		t.Fatalf("Pop = (%q, %t), want (%q, true)", got, held, "gopher")
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true")
	}
}

func TestPushPopIsLIFO(t *testing.T) {
	container := stack.New[int]()
	container.Push(1)
	container.Push(2)
	container.Push(3)
	for _, wanted := range []int{3, 2, 1} {
		got, held := container.Pop()
		if !held {
			t.Fatalf("Pop held = false, want true for %d", wanted)
		}
		if got != wanted {
			t.Fatalf("Pop = %d, want %d", got, wanted)
		}
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true after draining")
	}
}

func TestPopEmptyYieldsZeroWithoutPanic(t *testing.T) {
	container := stack.New[int]()
	got, held := container.Pop()
	if held {
		t.Fatal("Pop held = true, want false on empty stack")
	}
	if got != 0 {
		t.Fatalf("Pop = %d, want zero value 0", got)
	}
	if got := container.Len(); got != 0 {
		t.Fatalf("Len = %d, want 0 after empty Pop", got)
	}
}

func TestPeekEmptyYieldsZeroWithoutPanic(t *testing.T) {
	container := stack.New[int]()
	got, held := container.Peek()
	if held {
		t.Fatal("Peek held = true, want false on empty stack")
	}
	if got != 0 {
		t.Fatalf("Peek = %d, want zero value 0", got)
	}
}

func TestPeekKeepsTop(t *testing.T) {
	container := stack.New[string]()
	container.Push("first")
	container.Push("top")
	for round := 0; round < 2; round++ {
		got, held := container.Peek()
		if !held || got != "top" {
			t.Fatalf("Peek = (%q, %t), want (%q, true)", got, held, "top")
		}
	}
	if got := container.Len(); got != 2 {
		t.Fatalf("Len = %d, want 2 after Peek", got)
	}
	if got, held := container.Pop(); !held || got != "top" {
		t.Fatalf("Pop = (%q, %t), want (%q, true)", got, held, "top")
	}
	if got, held := container.Peek(); !held || got != "first" {
		t.Fatalf("Peek = (%q, %t), want (%q, true)", got, held, "first")
	}
}

func TestPushAfterDrainReusesStack(t *testing.T) {
	container := stack.New[int]()
	container.Push(1)
	if _, held := container.Pop(); !held {
		t.Fatal("Pop held = false, want true")
	}
	container.Push(2)
	got, held := container.Pop()
	if !held || got != 2 {
		t.Fatalf("Pop = (%d, %t), want (2, true)", got, held)
	}
	if !container.IsEmpty() {
		t.Fatal("IsEmpty = false, want true after draining")
	}
}

func TestStacksStayIndependent(t *testing.T) {
	first := stack.New[int]()
	second := stack.New[int]()
	first.Push(1)
	first.Push(2)
	if !second.IsEmpty() {
		t.Fatal("second IsEmpty = false, want true; stacks share backing")
	}
	if got, held := second.Pop(); held || got != 0 {
		t.Fatalf("second Pop = (%d, %t), want (0, false)", got, held)
	}
	if got := first.Len(); got != 2 {
		t.Fatalf("first Len = %d, want 2", got)
	}
}
