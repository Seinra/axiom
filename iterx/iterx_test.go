package iterx_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/seinra/axiom/iterx"
)

func naturals(start int, stop int) func(func(int) bool) {
	return func(yield func(int) bool) {
		for candidate := start; candidate < stop; candidate++ {
			if !yield(candidate) {
				return
			}
		}
	}
}

func TestMapSeqTransforms(t *testing.T) {
	got := slices.Collect(iterx.MapSeq(naturals(1, 4), func(candidate int) int { return candidate * 10 }))
	want := []int{10, 20, 30}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MapSeq = %v, want %v", got, want)
	}
}

func TestFilterSeqKeepsMatching(t *testing.T) {
	got := slices.Collect(iterx.FilterSeq(naturals(1, 7), func(candidate int) bool { return candidate%2 == 0 }))
	want := []int{2, 4, 6}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterSeq = %v, want %v", got, want)
	}
}

func TestTakeLimits(t *testing.T) {
	got := slices.Collect(iterx.Take(naturals(1, 100), 3))
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Take = %v, want %v", got, want)
	}
}

func TestTakeBeyondLengthCollectsAll(t *testing.T) {
	got := slices.Collect(iterx.Take(naturals(1, 3), 10))
	want := []int{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Take = %v, want %v", got, want)
	}
}

func TestTakeNonPositiveYieldsEmpty(t *testing.T) {
	for _, count := range []int{0, -5} {
		if got := slices.Collect(iterx.Take(naturals(1, 4), count)); len(got) != 0 {
			t.Fatalf("Take(%d) = %v, want empty", count, got)
		}
	}
}

func TestTakeStopsEarly(t *testing.T) {
	// Laziness proof: the pipeline must not pull beyond what Take keeps.
	pulled := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	got := slices.Collect(iterx.Take(counting, 2))
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("Take = %v, want [1 2]", got)
	}
	if pulled != 2 {
		t.Fatalf("pulled = %d, want 2 (short-circuit must propagate)", pulled)
	}
}

func TestDropSkips(t *testing.T) {
	got := slices.Collect(iterx.Drop(naturals(1, 6), 3))
	want := []int{4, 5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Drop = %v, want %v", got, want)
	}
}

func TestDropBeyondLengthYieldsEmpty(t *testing.T) {
	if got := slices.Collect(iterx.Drop(naturals(1, 3), 10)); len(got) != 0 {
		t.Fatalf("Drop = %v, want empty", got)
	}
}

func TestDropNonPositiveKeepsAll(t *testing.T) {
	for _, count := range []int{0, -5} {
		got := slices.Collect(iterx.Drop(naturals(1, 4), count))
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Drop(%d) = %v, want %v", count, got, want)
		}
	}
}

func TestEnumerateIndexes(t *testing.T) {
	indexed := make(map[int]string)
	for position, name := range iterx.Enumerate(slices.Values([]string{"go", "rust"})) {
		indexed[position] = name
	}
	want := map[int]string{0: "go", 1: "rust"}
	if !reflect.DeepEqual(indexed, want) {
		t.Fatalf("Enumerate = %v, want %v", indexed, want)
	}
}

func TestPipelineComposes(t *testing.T) {
	// Filter → Map → Take chains left to right, evaluated lazily.
	got := slices.Collect(iterx.Take(
		iterx.MapSeq(
			iterx.FilterSeq(naturals(1, 100), func(candidate int) bool { return candidate%2 == 1 }),
			func(candidate int) int { return candidate * candidate },
		),
		3,
	))
	want := []int{1, 9, 25}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pipeline = %v, want %v", got, want)
	}
}
