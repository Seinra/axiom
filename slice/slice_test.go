package slice_test

import (
	"reflect"
	"testing"

	axiomslice "github.com/seinra/axiom/slice"
)

func TestMapSliceTransforms(t *testing.T) {
	got := axiomslice.MapSlice([]int{1, 2, 3}, func(success int) int { return success * 2 })
	want := []int{2, 4, 6}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MapSlice = %v, want %v", got, want)
	}
}

func TestMapSliceEmptyStaysEmpty(t *testing.T) {
	got := axiomslice.MapSlice([]int{}, func(success int) int { return success })
	if len(got) != 0 {
		t.Fatalf("MapSlice len = %d, want 0", len(got))
	}
}

func TestFilterSliceKeepsMatching(t *testing.T) {
	got := axiomslice.FilterSlice([]int{1, 2, 3, 4}, func(success int) bool { return success%2 == 0 })
	want := []int{2, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterSlice = %v, want %v", got, want)
	}
}

func TestReduceSliceSums(t *testing.T) {
	got := axiomslice.ReduceSlice([]int{1, 2, 3, 4}, 0, func(total int, success int) int { return total + success })
	if got != 10 {
		t.Fatalf("ReduceSlice = %d, want 10", got)
	}
}

func TestFindSliceFound(t *testing.T) {
	got, found := axiomslice.FindSlice([]string{"go", "rust", "zig"}, func(candidate string) bool { return candidate == "rust" })
	if !found {
		t.Fatal("found = false, want true")
	}
	if got != "rust" {
		t.Fatalf("FindSlice = %q, want %q", got, "rust")
	}
}

func TestFindSliceMissing(t *testing.T) {
	_, found := axiomslice.FindSlice([]string{"go"}, func(candidate string) bool { return candidate == "ada" })
	if found {
		t.Fatal("found = true, want false")
	}
}

func TestGroupBySliceGroups(t *testing.T) {
	got := axiomslice.GroupBySlice([]int{1, 2, 3, 4}, func(success int) string {
		if success%2 == 0 {
			return "even"
		}
		return "odd"
	})
	want := map[string][]int{"even": {2, 4}, "odd": {1, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GroupBySlice = %v, want %v", got, want)
	}
}

func TestNilInputYieldsEmpty(t *testing.T) {
	var numbers []int
	if got := axiomslice.MapSlice(numbers, func(element int) int { return element }); got == nil || len(got) != 0 {
		t.Fatalf("MapSlice(nil) = %v, want empty non-nil", got)
	}
	if got := axiomslice.FilterSlice(numbers, func(element int) bool { return true }); got == nil || len(got) != 0 {
		t.Fatalf("FilterSlice(nil) = %v, want empty non-nil", got)
	}
	if got := axiomslice.ReduceSlice(numbers, 7, func(total int, element int) int { return total + element }); got != 7 {
		t.Fatalf("ReduceSlice(nil) = %d, want 7", got)
	}
	if _, found := axiomslice.FindSlice(numbers, func(element int) bool { return true }); found {
		t.Fatal("FindSlice(nil) found = true, want false")
	}
	if got := axiomslice.GroupBySlice(numbers, func(element int) string { return "all" }); got == nil || len(got) != 0 {
		t.Fatalf("GroupBySlice(nil) = %v, want empty non-nil", got)
	}
}

func TestHelpersNeverMutateInput(t *testing.T) {
	numbers := []int{3, 1, 2}
	_ = axiomslice.MapSlice(numbers, func(element int) int { return element * 2 })
	_ = axiomslice.FilterSlice(numbers, func(element int) bool { return element > 1 })
	_ = axiomslice.ReduceSlice(numbers, 0, func(total int, element int) int { return total + element })
	_, _ = axiomslice.FindSlice(numbers, func(element int) bool { return element == 1 })
	_ = axiomslice.GroupBySlice(numbers, func(element int) string { return "all" })
	want := []int{3, 1, 2}
	if !reflect.DeepEqual(numbers, want) {
		t.Fatalf("input mutated to %v, want %v", numbers, want)
	}
}
