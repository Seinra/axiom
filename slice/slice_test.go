package slice_test

import (
	"fmt"
	"reflect"
	"testing"

	axiomslice "github.com/seinra/axiom/slice"
)

func TestMinFindsMinimum(t *testing.T) {
	lowest, found := axiomslice.Min([]int{3, 1, 2})
	if !found {
		t.Fatal("found = false, want true")
	}
	if lowest != 1 {
		t.Fatalf("Min = %d, want 1", lowest)
	}
}

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

func TestMinEmptyYieldsNotFound(t *testing.T) {
	if _, found := axiomslice.Min([]int{}); found {
		t.Fatal("Min(empty) found = true, want false")
	}
}

func TestMinNilYieldsNotFound(t *testing.T) {
	var numbers []int
	if _, found := axiomslice.Min(numbers); found {
		t.Fatal("Min(nil) found = true, want false")
	}
}

func TestMinSingletonReturnsElement(t *testing.T) {
	lowest, found := axiomslice.Min([]string{"only"})
	if !found || lowest != "only" {
		t.Fatalf("Min(singleton) = %q, %v, want %q, true", lowest, found, "only")
	}
}

func TestMaxFindsMaximum(t *testing.T) {
	highest, found := axiomslice.Max([]int{3, 1, 2})
	if !found {
		t.Fatal("found = false, want true")
	}
	if highest != 3 {
		t.Fatalf("Max = %d, want 3", highest)
	}
}

func TestMaxEmptyYieldsNotFound(t *testing.T) {
	if _, found := axiomslice.Max([]int{}); found {
		t.Fatal("Max(empty) found = true, want false")
	}
}

func TestMaxNilYieldsNotFound(t *testing.T) {
	var numbers []int
	if _, found := axiomslice.Max(numbers); found {
		t.Fatal("Max(nil) found = true, want false")
	}
}

func TestMaxSingletonReturnsElement(t *testing.T) {
	highest, found := axiomslice.Max([]string{"only"})
	if !found || highest != "only" {
		t.Fatalf("Max(singleton) = %q, %v, want %q, true", highest, found, "only")
	}
}

func TestMinBySelectsSmallestKey(t *testing.T) {
	words := []string{"aaa", "b", "cc"}
	shortest, found := axiomslice.MinBy(words, func(candidate string) int { return len(candidate) })
	if !found || shortest != "b" {
		t.Fatalf("MinBy = %q, %v, want %q, true", shortest, found, "b")
	}
}

func TestMinByTieKeepsFirst(t *testing.T) {
	words := []string{"aaa", "bb", "cc"}
	shortest, found := axiomslice.MinBy(words, func(candidate string) int { return len(candidate) })
	if !found || shortest != "bb" {
		t.Fatalf("MinBy tie = %q, %v, want %q, true", shortest, found, "bb")
	}
}

func TestMinByEmptyYieldsNotFound(t *testing.T) {
	if _, found := axiomslice.MinBy([]string{}, func(candidate string) int { return len(candidate) }); found {
		t.Fatal("MinBy(empty) found = true, want false")
	}
}

func TestMinByNilYieldsNotFound(t *testing.T) {
	var words []string
	if _, found := axiomslice.MinBy(words, func(candidate string) int { return len(candidate) }); found {
		t.Fatal("MinBy(nil) found = true, want false")
	}
}

func TestMaxBySelectsLargestKey(t *testing.T) {
	words := []string{"a", "bb", "dddd"}
	longest, found := axiomslice.MaxBy(words, func(candidate string) int { return len(candidate) })
	if !found || longest != "dddd" {
		t.Fatalf("MaxBy = %q, %v, want %q, true", longest, found, "dddd")
	}
}

func TestMaxByTieKeepsFirst(t *testing.T) {
	words := []string{"a", "bb", "cc", "d"}
	longest, found := axiomslice.MaxBy(words, func(candidate string) int { return len(candidate) })
	if !found || longest != "bb" {
		t.Fatalf("MaxBy tie = %q, %v, want %q, true", longest, found, "bb")
	}
}

func TestMaxByEmptyYieldsNotFound(t *testing.T) {
	if _, found := axiomslice.MaxBy([]string{}, func(candidate string) int { return len(candidate) }); found {
		t.Fatal("MaxBy(empty) found = true, want false")
	}
}

func TestMaxByNilYieldsNotFound(t *testing.T) {
	var words []string
	if _, found := axiomslice.MaxBy(words, func(candidate string) int { return len(candidate) }); found {
		t.Fatal("MaxBy(nil) found = true, want false")
	}
}

func TestSumAddsIntegers(t *testing.T) {
	total := axiomslice.Sum([]int{1, 2, 3, 4})
	if total != 10 {
		t.Fatalf("Sum = %d, want 10", total)
	}
}

func TestSumEmptyYieldsZero(t *testing.T) {
	if total := axiomslice.Sum([]int{}); total != 0 {
		t.Fatalf("Sum(empty) = %d, want 0", total)
	}
}

func TestSumNilYieldsZero(t *testing.T) {
	var numbers []int
	if total := axiomslice.Sum(numbers); total != 0 {
		t.Fatalf("Sum(nil) = %d, want 0", total)
	}
}

func TestSumFloats(t *testing.T) {
	total := axiomslice.Sum([]float64{1.5, 2.5})
	if total != 4.0 {
		t.Fatalf("Sum = %v, want 4", total)
	}
}

func TestSumNamedNumberType(t *testing.T) {
	type customCount int
	total := axiomslice.Sum([]customCount{1, 2, 3})
	if total != 6 {
		t.Fatalf("Sum = %d, want 6", total)
	}
}

func TestFlatMapConcatenates(t *testing.T) {
	got := axiomslice.FlatMap([]int{1, 2, 3}, func(element int) []int { return []int{element, element * 10} })
	want := []int{1, 10, 2, 20, 3, 30}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FlatMap = %v, want %v", got, want)
	}
}

func TestFlatMapNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.FlatMap(numbers, func(element int) []int { return []int{element} })
	if got == nil || len(got) != 0 {
		t.Fatalf("FlatMap(nil) = %v, want empty non-nil", got)
	}
}

func TestFlatMapSkipsNilPieces(t *testing.T) {
	got := axiomslice.FlatMap([]int{1, 2}, func(element int) []int {
		if element%2 == 1 {
			return nil
		}
		return []int{element}
	})
	want := []int{2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FlatMap = %v, want %v", got, want)
	}
}

func TestFlattenConcatenates(t *testing.T) {
	got := axiomslice.Flatten([][]int{{1, 2}, {}, {3}})
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Flatten = %v, want %v", got, want)
	}
}

func TestFlattenNilYieldsEmpty(t *testing.T) {
	var nested [][]int
	got := axiomslice.Flatten(nested)
	if got == nil || len(got) != 0 {
		t.Fatalf("Flatten(nil) = %v, want empty non-nil", got)
	}
}

// TestFlattenLengthTotalOverflowPanics lives in flatten_overflow_test.go
// (//go:build !race): header fabrication trips checkptr.

func TestChunkedSplitsEvenChunks(t *testing.T) {
	got := axiomslice.Chunked([]int{1, 2, 3, 4}, 2)
	want := [][]int{{1, 2}, {3, 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Chunked = %v, want %v", got, want)
	}
}

func TestChunkedLastChunkShort(t *testing.T) {
	got := axiomslice.Chunked([]int{1, 2, 3, 4, 5}, 2)
	want := [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Chunked = %v, want %v", got, want)
	}
}

func TestChunkedInvalidSizePanics(t *testing.T) {
	for _, width := range []int{0, -1} {
		panicked := func() (didPanic any) {
			defer func() { didPanic = recover() }()
			_ = axiomslice.Chunked([]int{1, 2}, width)
			return nil
		}()
		if panicked == nil {
			t.Fatalf("Chunked(size %d) did not panic, want panic", width)
		}
	}
}

func TestChunkedNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.Chunked(numbers, 2)
	if got == nil || len(got) != 0 {
		t.Fatalf("Chunked(nil) = %v, want empty non-nil", got)
	}
}

func TestWindowsSlidesOverElements(t *testing.T) {
	got := axiomslice.Windows([]int{1, 2, 3, 4}, 2)
	want := [][]int{{1, 2}, {2, 3}, {3, 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Windows = %v, want %v", got, want)
	}
}

func TestWindowsFullLengthYieldsSingleton(t *testing.T) {
	got := axiomslice.Windows([]int{1, 2}, 2)
	want := [][]int{{1, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Windows = %v, want %v", got, want)
	}
}

func TestWindowsInvalidSizePanics(t *testing.T) {
	for _, width := range []int{0, -1} {
		panicked := func() (didPanic any) {
			defer func() { didPanic = recover() }()
			_ = axiomslice.Windows([]int{1, 2, 3}, width)
			return nil
		}()
		if panicked == nil {
			t.Fatalf("Windows(size %d) did not panic, want panic", width)
		}
	}
}

func TestWindowsOversizeYieldsEmpty(t *testing.T) {
	got := axiomslice.Windows([]int{1, 2, 3}, 5)
	if got == nil || len(got) != 0 {
		t.Fatalf("Windows(size 5) = %v, want empty non-nil", got)
	}
}

func TestWindowsNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.Windows(numbers, 2)
	if got == nil || len(got) != 0 {
		t.Fatalf("Windows(nil) = %v, want empty non-nil", got)
	}
}

func TestZipStopsAtShorter(t *testing.T) {
	got := axiomslice.Zip([]int{1, 2, 3}, []string{"a", "b"})
	want := []axiomslice.Pair[int, string]{{First: 1, Second: "a"}, {First: 2, Second: "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Zip = %v, want %v", got, want)
	}
}

func TestZipStopsAtShorterSecond(t *testing.T) {
	got := axiomslice.Zip([]int{1}, []string{"a", "b", "c"})
	want := []axiomslice.Pair[int, string]{{First: 1, Second: "a"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Zip = %v, want %v", got, want)
	}
}

func TestChunkedCopiesNotAliased(t *testing.T) {
	numbers := []int{1, 2, 3, 4}
	chunked := axiomslice.Chunked(numbers, 2)
	chunked[0][0] = 99
	chunked[1][1] = 99
	if !reflect.DeepEqual(numbers, []int{1, 2, 3, 4}) {
		t.Fatalf("source mutated to %v, want [1 2 3 4]", numbers)
	}
}

func TestWindowsCopiesNotAliased(t *testing.T) {
	numbers := []int{1, 2, 3}
	windows := axiomslice.Windows(numbers, 2)
	windows[0][0] = 99
	windows[1][1] = 99
	if !reflect.DeepEqual(numbers, []int{1, 2, 3}) {
		t.Fatalf("source mutated to %v, want [1 2 3]", numbers)
	}
}

func TestZipEmptyYieldsEmpty(t *testing.T) {
	got := axiomslice.Zip([]int{}, []string{"a"})
	if got == nil || len(got) != 0 {
		t.Fatalf("Zip(empty) = %v, want empty non-nil", got)
	}
}

func TestZipNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.Zip(numbers, []string{"a"})
	if got == nil || len(got) != 0 {
		t.Fatalf("Zip(nil) = %v, want empty non-nil", got)
	}
}

func TestUnzipSplitsPairs(t *testing.T) {
	pairs := []axiomslice.Pair[int, string]{{First: 1, Second: "a"}, {First: 2, Second: "b"}}
	firsts, seconds := axiomslice.Unzip(pairs)
	if !reflect.DeepEqual(firsts, []int{1, 2}) {
		t.Fatalf("Unzip firsts = %v, want [1 2]", firsts)
	}
	if !reflect.DeepEqual(seconds, []string{"a", "b"}) {
		t.Fatalf("Unzip seconds = %v, want [a b]", seconds)
	}
}

func TestUnzipNilYieldsEmptyPair(t *testing.T) {
	var pairs []axiomslice.Pair[int, string]
	firsts, seconds := axiomslice.Unzip(pairs)
	if firsts == nil || len(firsts) != 0 {
		t.Fatalf("Unzip(nil) firsts = %v, want empty non-nil", firsts)
	}
	if seconds == nil || len(seconds) != 0 {
		t.Fatalf("Unzip(nil) seconds = %v, want empty non-nil", seconds)
	}
}

func TestDistinctKeepsFirstOccurrence(t *testing.T) {
	got := axiomslice.Distinct([]int{2, 1, 2, 3, 1})
	want := []int{2, 1, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Distinct = %v, want %v", got, want)
	}
}

func TestDistinctCollisionCollapses(t *testing.T) {
	got := axiomslice.Distinct([]int{5, 5, 5})
	want := []int{5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Distinct = %v, want %v", got, want)
	}
}

func TestDistinctNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.Distinct(numbers)
	if got == nil || len(got) != 0 {
		t.Fatalf("Distinct(nil) = %v, want empty non-nil", got)
	}
}

func TestPartitionSplitsMatching(t *testing.T) {
	matching, rest := axiomslice.Partition([]int{1, 2, 3, 4}, func(element int) bool { return element%2 == 0 })
	if !reflect.DeepEqual(matching, []int{2, 4}) {
		t.Fatalf("Partition matching = %v, want [2 4]", matching)
	}
	if !reflect.DeepEqual(rest, []int{1, 3}) {
		t.Fatalf("Partition rest = %v, want [1 3]", rest)
	}
}

func TestPartitionNilYieldsEmptyPair(t *testing.T) {
	var numbers []int
	matching, rest := axiomslice.Partition(numbers, func(element int) bool { return true })
	if matching == nil || len(matching) != 0 {
		t.Fatalf("Partition(nil) matching = %v, want empty non-nil", matching)
	}
	if rest == nil || len(rest) != 0 {
		t.Fatalf("Partition(nil) rest = %v, want empty non-nil", rest)
	}
}

func TestPartitionNoneMatchingKeepsRest(t *testing.T) {
	matching, rest := axiomslice.Partition([]int{1, 3}, func(element int) bool { return element%2 == 0 })
	if matching == nil || len(matching) != 0 {
		t.Fatalf("Partition matching = %v, want empty non-nil", matching)
	}
	if !reflect.DeepEqual(rest, []int{1, 3}) {
		t.Fatalf("Partition rest = %v, want [1 3]", rest)
	}
}

func TestMapIndexedUsesPosition(t *testing.T) {
	got := axiomslice.MapIndexed([]int{10, 20, 30}, func(position int, element int) int { return position + element })
	want := []int{10, 21, 32}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MapIndexed = %v, want %v", got, want)
	}
}

func TestMapIndexedNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.MapIndexed(numbers, func(position int, element int) int { return element })
	if got == nil || len(got) != 0 {
		t.Fatalf("MapIndexed(nil) = %v, want empty non-nil", got)
	}
}

func TestFilterIndexedKeepsByPosition(t *testing.T) {
	got := axiomslice.FilterIndexed([]int{10, 20, 30, 40}, func(position int, element int) bool { return position%2 == 0 })
	want := []int{10, 30}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterIndexed = %v, want %v", got, want)
	}
}

func TestFilterIndexedNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.FilterIndexed(numbers, func(position int, element int) bool { return true })
	if got == nil || len(got) != 0 {
		t.Fatalf("FilterIndexed(nil) = %v, want empty non-nil", got)
	}
}

func TestIntersperseInsertsSeparator(t *testing.T) {
	got := axiomslice.Intersperse([]int{1, 2, 3}, 0)
	want := []int{1, 0, 2, 0, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Intersperse = %v, want %v", got, want)
	}
}

func TestIntersperseSingletonCopies(t *testing.T) {
	original := []int{7}
	got := axiomslice.Intersperse(original, 0)
	if !reflect.DeepEqual(got, []int{7}) {
		t.Fatalf("Intersperse(singleton) = %v, want [7]", got)
	}
	got[0] = 99
	if original[0] != 7 {
		t.Fatalf("Intersperse singleton shares backing array: %v", original)
	}
}

func TestIntersperseEmptyYieldsEmpty(t *testing.T) {
	got := axiomslice.Intersperse([]int{}, 0)
	if got == nil || len(got) != 0 {
		t.Fatalf("Intersperse(empty) = %v, want empty non-nil", got)
	}
}

func TestIntersperseNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.Intersperse(numbers, 0)
	if got == nil || len(got) != 0 {
		t.Fatalf("Intersperse(nil) = %v, want empty non-nil", got)
	}
}

func TestTallyCollisionSums(t *testing.T) {
	got := axiomslice.Tally([]string{"a", "b", "a", "c", "b", "a"})
	want := map[string]int{"a": 3, "b": 2, "c": 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Tally = %v, want %v", got, want)
	}
}

func TestTallySingletonTalliesOnce(t *testing.T) {
	got := axiomslice.Tally([]string{"only"})
	want := map[string]int{"only": 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Tally = %v, want %v", got, want)
	}
}

func TestTallyNilYieldsEmptyMap(t *testing.T) {
	var words []string
	got := axiomslice.Tally(words)
	if got == nil || len(got) != 0 {
		t.Fatalf("Tally(nil) = %v, want empty non-nil", got)
	}
}

func TestCollectionHelpersNeverMutateInput(t *testing.T) {
	numbers := []int{3, 1, 2, 1}
	nested := [][]int{{3, 1}, {2, 1}}
	_, _ = axiomslice.Min(numbers)
	_, _ = axiomslice.Max(numbers)
	_, _ = axiomslice.MinBy(numbers, func(element int) int { return element })
	_, _ = axiomslice.MaxBy(numbers, func(element int) int { return element })
	_ = axiomslice.Sum(numbers)
	_ = axiomslice.FlatMap(numbers, func(element int) []int { return []int{element} })
	_ = axiomslice.Flatten(nested)
	_ = axiomslice.Chunked(numbers, 2)
	_ = axiomslice.Windows(numbers, 2)
	_ = axiomslice.Zip(numbers, []string{"a", "b", "c", "d"})
	_, _ = axiomslice.Unzip([]axiomslice.Pair[int, string]{{First: 1, Second: "a"}})
	_ = axiomslice.Distinct(numbers)
	_, _ = axiomslice.Partition(numbers, func(element int) bool { return element > 1 })
	_ = axiomslice.MapIndexed(numbers, func(position int, element int) int { return element })
	_ = axiomslice.FilterIndexed(numbers, func(position int, element int) bool { return true })
	_ = axiomslice.Intersperse(numbers, 0)
	_ = axiomslice.Tally(numbers)
	if !reflect.DeepEqual(numbers, []int{3, 1, 2, 1}) {
		t.Fatalf("input mutated to %v", numbers)
	}
	if !reflect.DeepEqual(nested, [][]int{{3, 1}, {2, 1}}) {
		t.Fatalf("nested mutated to %v", nested)
	}
}

func TestAverageIntsAveragesEvenly(t *testing.T) {
	average, found := axiomslice.Average([]int{1, 2, 3, 4})
	if !found {
		t.Fatal("Average found = false, want true")
	}
	if average != 2.5 {
		t.Fatalf("Average = %v, want 2.5", average)
	}
}

func TestAverageFloatsAveragesFraction(t *testing.T) {
	average, found := axiomslice.Average([]float64{1.5, 2.5})
	if !found {
		t.Fatal("Average found = false, want true")
	}
	if average != 2.0 {
		t.Fatalf("Average = %v, want 2", average)
	}
}

func TestAverageEmptyYieldsNotFound(t *testing.T) {
	average, found := axiomslice.Average([]int{})
	if found {
		t.Fatal("Average(empty) found = true, want false")
	}
	if average != 0 {
		t.Fatalf("Average(empty) = %v, want 0", average)
	}
}

func TestAverageNilYieldsNotFound(t *testing.T) {
	var numbers []int
	average, found := axiomslice.Average(numbers)
	if found {
		t.Fatal("Average(nil) found = true, want false")
	}
	if average != 0 {
		t.Fatalf("Average(nil) = %v, want 0", average)
	}
}

func TestAverageSingletonReturnsElement(t *testing.T) {
	average, found := axiomslice.Average([]int{7})
	if !found || average != 7.0 {
		t.Fatalf("Average(singleton) = %v, %v, want 7, true", average, found)
	}
}

func TestAverageNegativeMixBalances(t *testing.T) {
	average, found := axiomslice.Average([]int{-4, -2, 6})
	if !found || average != 0 {
		t.Fatalf("Average(negative mix) = %v, %v, want 0, true", average, found)
	}
}

func TestAverageNeverMutatesInput(t *testing.T) {
	numbers := []int{3, 1, 2, 1}
	_, _ = axiomslice.Average(numbers)
	if !reflect.DeepEqual(numbers, []int{3, 1, 2, 1}) {
		t.Fatalf("input mutated to %v", numbers)
	}
}

func TestSortByOrdersByKey(t *testing.T) {
	testCases := []struct {
		scenario string
		input    []int
		expected []int
	}{
		{scenario: "unsorted orders ascending", input: []int{3, 1, 2}, expected: []int{1, 2, 3}},
		{scenario: "already sorted stays ordered", input: []int{1, 2, 3}, expected: []int{1, 2, 3}},
		{scenario: "reversed orders ascending", input: []int{5, 4, 3, 2, 1}, expected: []int{1, 2, 3, 4, 5}},
		{scenario: "duplicates sort in order", input: []int{2, 1, 2, 1}, expected: []int{1, 1, 2, 2}},
		{scenario: "singleton stays singleton", input: []int{7}, expected: []int{7}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			sorted := axiomslice.SortBy(testCase.input, func(element int) int { return element })
			if !reflect.DeepEqual(sorted, testCase.expected) {
				t.Fatalf("SortBy(%v) = %v, want %v", testCase.input, sorted, testCase.expected)
			}
		})
	}
}

func TestSortByKeyOrdersWordsByLength(t *testing.T) {
	words := []string{"bbbb", "a", "ccc"}
	sorted := axiomslice.SortBy(words, func(candidate string) int { return len(candidate) })
	expected := []string{"a", "ccc", "bbbb"}
	if !reflect.DeepEqual(sorted, expected) {
		t.Fatalf("SortBy lengths = %v, want %v", sorted, expected)
	}
}

func TestSortByTieKeepsFirst(t *testing.T) {
	type player struct {
		label string
		score int
	}
	players := []player{{label: "first", score: 1}, {label: "second", score: 1}, {label: "third", score: 2}}
	sorted := axiomslice.SortBy(players, func(candidate player) int { return candidate.score })
	if len(sorted) != 3 || sorted[0].label != "first" || sorted[1].label != "second" || sorted[2].label != "third" {
		t.Fatalf("SortBy tie order = %v, want first/second/third stable", sorted)
	}
}

func TestSortByEmptyYieldsEmpty(t *testing.T) {
	sorted := axiomslice.SortBy([]int{}, func(element int) int { return element })
	if sorted == nil || len(sorted) != 0 {
		t.Fatalf("SortBy(empty) = %v, want empty non-nil", sorted)
	}
}

func TestSortByNilYieldsEmpty(t *testing.T) {
	var numbers []int
	sorted := axiomslice.SortBy(numbers, func(element int) int { return element })
	if sorted == nil || len(sorted) != 0 {
		t.Fatalf("SortBy(nil) = %v, want empty non-nil", sorted)
	}
}

func TestSortByNeverMutatesInput(t *testing.T) {
	numbers := []int{3, 1, 2}
	sorted := axiomslice.SortBy(numbers, func(element int) int { return element })
	sorted[0] = 99
	if !reflect.DeepEqual(numbers, []int{3, 1, 2}) {
		t.Fatalf("input mutated to %v, want [3 1 2]", numbers)
	}
}

func TestSortByStabilityAcrossManyTies(t *testing.T) {
	type player struct {
		label string
		score int
	}
	players := []player{
		{label: "k0a", score: 0}, {label: "k1a", score: 1}, {label: "k2a", score: 2},
		{label: "k0b", score: 0}, {label: "k1b", score: 1}, {label: "k2b", score: 2},
		{label: "k0c", score: 0}, {label: "k1c", score: 1}, {label: "k2c", score: 2},
		{label: "k0d", score: 0}, {label: "k1d", score: 1}, {label: "k2d", score: 2},
	}
	sorted := axiomslice.SortBy(players, func(candidate player) int { return candidate.score })
	labels := make([]string, 0, len(sorted))
	for _, winner := range sorted {
		labels = append(labels, winner.label)
	}
	expected := []string{"k0a", "k0b", "k0c", "k0d", "k1a", "k1b", "k1c", "k1d", "k2a", "k2b", "k2c", "k2d"}
	if !reflect.DeepEqual(labels, expected) {
		t.Fatalf("SortBy stable labels = %v, want %v", labels, expected)
	}
}

func TestRunningFoldRunningTotals(t *testing.T) {
	testCases := []struct {
		scenario string
		input    []int
		initial  int
		expected []int
	}{
		{scenario: "prefix sums from zero", input: []int{1, 2, 3, 4}, initial: 0, expected: []int{1, 3, 6, 10}},
		{scenario: "nonzero initial offsets totals", input: []int{1, 2}, initial: 10, expected: []int{11, 13}},
		{scenario: "singleton holds combined initial", input: []int{5}, initial: 2, expected: []int{7}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			actual := axiomslice.RunningFold(testCase.input, testCase.initial, func(total int, element int) int { return total + element })
			if !reflect.DeepEqual(actual, testCase.expected) {
				t.Fatalf("RunningFold(%v) = %v, want %v", testCase.input, actual, testCase.expected)
			}
		})
	}
}

func TestRunningFoldConcatenatesStrings(t *testing.T) {
	actual := axiomslice.RunningFold([]string{"a", "b", "c"}, "", func(total string, element string) string { return total + element })
	expected := []string{"a", "ab", "abc"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("RunningFold strings = %v, want %v", actual, expected)
	}
}

func TestRunningFoldLengthMatchesInput(t *testing.T) {
	input := []int{4, 3, 2, 1}
	actual := axiomslice.RunningFold(input, 0, func(total int, element int) int { return total + element })
	if len(actual) != len(input) {
		t.Fatalf("RunningFold len = %d, want %d", len(actual), len(input))
	}
}

func TestRunningFoldEmptyYieldsEmpty(t *testing.T) {
	actual := axiomslice.RunningFold([]int{}, 0, func(total int, element int) int { return total + element })
	if actual == nil || len(actual) != 0 {
		t.Fatalf("RunningFold(empty) = %v, want empty non-nil", actual)
	}
}

func TestRunningFoldNilYieldsEmpty(t *testing.T) {
	var numbers []int
	actual := axiomslice.RunningFold(numbers, 0, func(total int, element int) int { return total + element })
	if actual == nil || len(actual) != 0 {
		t.Fatalf("RunningFold(nil) = %v, want empty non-nil", actual)
	}
}

func TestRunningFoldNeverMutatesInput(t *testing.T) {
	numbers := []int{1, 2, 3}
	actual := axiomslice.RunningFold(numbers, 0, func(total int, element int) int { return total + element })
	actual[0] = 99
	if !reflect.DeepEqual(numbers, []int{1, 2, 3}) {
		t.Fatalf("input mutated to %v, want [1 2 3]", numbers)
	}
}

func pinnedPanics(t *testing.T, label string, call func()) {
	t.Helper()
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		call()
		return nil
	}()
	if panicked == nil {
		t.Fatalf("%s did not panic, want panic (nil func runs on a non-empty input)", label)
	}
}

func TestNilFuncPanicsOnNonEmptyInput(t *testing.T) {
	testCases := []struct {
		name string
		call func()
	}{
		{name: "MapSlice", call: func() { _ = axiomslice.MapSlice[int, int]([]int{1, 2}, nil) }},
		{name: "FilterSlice", call: func() { _ = axiomslice.FilterSlice([]int{1, 2}, nil) }},
		{name: "ReduceSlice", call: func() { _ = axiomslice.ReduceSlice([]int{1, 2}, 0, nil) }},
		{name: "FindSlice", call: func() { _, _ = axiomslice.FindSlice([]int{1, 2}, nil) }},
		{name: "GroupBySlice", call: func() { _ = axiomslice.GroupBySlice[int, string]([]int{1, 2}, nil) }},
		{name: "MinBy", call: func() { _, _ = axiomslice.MinBy[string, int]([]string{"b", "a"}, nil) }},
		{name: "MaxBy", call: func() { _, _ = axiomslice.MaxBy[string, int]([]string{"b", "a"}, nil) }},
		{name: "FlatMap", call: func() { _ = axiomslice.FlatMap[int, int]([]int{1, 2}, nil) }},
		{name: "Partition", call: func() { _, _ = axiomslice.Partition([]int{1, 2}, nil) }},
		{name: "MapIndexed", call: func() { _ = axiomslice.MapIndexed[int, int]([]int{1, 2}, nil) }},
		{name: "FilterIndexed", call: func() { _ = axiomslice.FilterIndexed([]int{1, 2}, nil) }},
		{name: "SortBy", call: func() { _ = axiomslice.SortBy[int, int]([]int{2, 1}, nil) }},
		{name: "RunningFold", call: func() { _ = axiomslice.RunningFold([]int{1, 2}, 0, nil) }},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			pinnedPanics(t, testCase.name+"(non-empty, nil)", testCase.call)
		})
	}
}

func TestNilFuncSkipsOnEmptyInput(t *testing.T) {
	var numbers []int
	if got := axiomslice.MapSlice[int, int](numbers, nil); got == nil || len(got) != 0 {
		t.Fatalf("MapSlice(nil, nil) = %v, want empty non-nil", got)
	}
	if got := axiomslice.FilterSlice(numbers, nil); got == nil || len(got) != 0 {
		t.Fatalf("FilterSlice(nil, nil) = %v, want empty non-nil", got)
	}
	if got := axiomslice.ReduceSlice(numbers, 7, nil); got != 7 {
		t.Fatalf("ReduceSlice(nil, 7, nil) = %d, want 7", got)
	}
	if _, found := axiomslice.FindSlice(numbers, nil); found {
		t.Fatal("FindSlice(nil, nil) found = true, want false")
	}
	if got := axiomslice.GroupBySlice[int, string](numbers, nil); got == nil || len(got) != 0 {
		t.Fatalf("GroupBySlice(nil, nil) = %v, want empty non-nil", got)
	}
	var words []string
	if _, found := axiomslice.MinBy[string, int](words, nil); found {
		t.Fatal("MinBy(nil, nil) found = true, want false")
	}
	if _, found := axiomslice.MaxBy[string, int](words, nil); found {
		t.Fatal("MaxBy(nil, nil) found = true, want false")
	}
	if got := axiomslice.FlatMap[int, int](numbers, nil); got == nil || len(got) != 0 {
		t.Fatalf("FlatMap(nil, nil) = %v, want empty non-nil", got)
	}
	if matching, rest := axiomslice.Partition(numbers, nil); matching == nil || rest == nil || len(matching) != 0 || len(rest) != 0 {
		t.Fatalf("Partition(nil, nil) = %v, %v, want two empty non-nil slices", matching, rest)
	}
	if got := axiomslice.MapIndexed[int, int](numbers, nil); got == nil || len(got) != 0 {
		t.Fatalf("MapIndexed(nil, nil) = %v, want empty non-nil", got)
	}
	if got := axiomslice.FilterIndexed(numbers, nil); got == nil || len(got) != 0 {
		t.Fatalf("FilterIndexed(nil, nil) = %v, want empty non-nil", got)
	}
	if got := axiomslice.SortBy[int, int](numbers, nil); got == nil || len(got) != 0 {
		t.Fatalf("SortBy(nil, nil) = %v, want empty non-nil", got)
	}
	if got := axiomslice.RunningFold(numbers, 0, nil); got == nil || len(got) != 0 {
		t.Fatalf("RunningFold(nil, 0, nil) = %v, want empty non-nil", got)
	}
}

func TestCombinationsTwoOfThreeOrdersLexicographic(t *testing.T) {
	got := axiomslice.Combinations([]int{1, 2, 3}, 2)
	want := [][]int{{1, 2}, {1, 3}, {2, 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Combinations = %v, want %v", got, want)
	}
}

func TestCombinationsTableDriven(t *testing.T) {
	testCases := []struct {
		scenario    string
		input       []int
		selectCount int
		expected    [][]int
	}{
		{scenario: "one of three yields singletons in order", input: []int{1, 2, 3}, selectCount: 1, expected: [][]int{{1}, {2}, {3}}},
		{scenario: "three of three yields singleton input order", input: []int{1, 2, 3}, selectCount: 3, expected: [][]int{{1, 2, 3}}},
		{scenario: "two of four yields six in index order", input: []int{1, 2, 3, 4}, selectCount: 2, expected: [][]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			got := axiomslice.Combinations(testCase.input, testCase.selectCount)
			if !reflect.DeepEqual(got, testCase.expected) {
				t.Fatalf("Combinations(%v, %d) = %v, want %v", testCase.input, testCase.selectCount, got, testCase.expected)
			}
		})
	}
}

func TestCombinationsStringsPreserveInputOrder(t *testing.T) {
	got := axiomslice.Combinations([]string{"a", "b", "c"}, 2)
	want := [][]string{{"a", "b"}, {"a", "c"}, {"b", "c"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Combinations strings = %v, want %v", got, want)
	}
}

func TestCombinationsZeroYieldsOneEmpty(t *testing.T) {
	got := axiomslice.Combinations([]int{1, 2, 3}, 0)
	if got == nil || len(got) != 1 {
		t.Fatalf("Combinations(k 0) = %v, want one empty combination", got)
	}
	if got[0] == nil || len(got[0]) != 0 {
		t.Fatalf("Combinations(k 0)[0] = %v, want empty non-nil", got[0])
	}
	var numbers []int
	gotNil := axiomslice.Combinations(numbers, 0)
	if gotNil == nil || len(gotNil) != 1 || gotNil[0] == nil || len(gotNil[0]) != 0 {
		t.Fatalf("Combinations(nil, 0) = %v, want one empty non-nil combination", gotNil)
	}
}

func TestCombinationsOversizeYieldsEmpty(t *testing.T) {
	got := axiomslice.Combinations([]int{1, 2, 3}, 5)
	if got == nil || len(got) != 0 {
		t.Fatalf("Combinations(k 5) = %v, want empty non-nil", got)
	}
}

func TestCombinationsNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.Combinations(numbers, 2)
	if got == nil || len(got) != 0 {
		t.Fatalf("Combinations(nil) = %v, want empty non-nil", got)
	}
}

func TestCombinationsEmptyYieldsEmpty(t *testing.T) {
	got := axiomslice.Combinations([]int{}, 1)
	if got == nil || len(got) != 0 {
		t.Fatalf("Combinations(empty) = %v, want empty non-nil", got)
	}
}

func TestCombinationsNegativePanics(t *testing.T) {
	for _, selectCount := range []int{-1, -5} {
		panicked := func() (didPanic any) {
			defer func() { didPanic = recover() }()
			_ = axiomslice.Combinations([]int{1, 2}, selectCount)
			return nil
		}()
		if panicked == nil {
			t.Fatalf("Combinations(k %d) did not panic, want panic", selectCount)
		}
		want := fmt.Sprintf("axiom slice: Combinations needs k >= 0, got %d", selectCount)
		if fmt.Sprintf("%v", panicked) != want {
			t.Fatalf("Combinations panic = %v, want %q", panicked, want)
		}
	}
}

func TestCombinationsCopiesNotAliased(t *testing.T) {
	numbers := []int{1, 2, 3}
	combos := axiomslice.Combinations(numbers, 2)
	if len(combos) == 0 {
		t.Fatal("Combinations returned empty, want non-empty seed for alias pin")
	}
	combos[0][0] = 99
	if !reflect.DeepEqual(numbers, []int{1, 2, 3}) {
		t.Fatalf("source mutated to %v, want [1 2 3]", numbers)
	}
	if !reflect.DeepEqual(combos[1], []int{1, 3}) {
		t.Fatalf("sibling combo mutated to %v, want [1 3]", combos[1])
	}
}

func TestCombinationsInputMutationStaysDetached(t *testing.T) {
	numbers := []int{1, 2, 3}
	combos := axiomslice.Combinations(numbers, 2)
	if len(combos) == 0 {
		t.Fatal("Combinations returned empty, want non-empty seed for detach pin")
	}
	numbers[0] = 99
	if reflect.DeepEqual(combos[0], []int{99, 2}) {
		t.Fatalf("combo aliases input, got %v after input mutation", combos[0])
	}
}

func TestCombinationsNeverMutatesInput(t *testing.T) {
	numbers := []int{1, 2, 3}
	_ = axiomslice.Combinations(numbers, 2)
	if !reflect.DeepEqual(numbers, []int{1, 2, 3}) {
		t.Fatalf("input mutated to %v, want [1 2 3]", numbers)
	}
}

func TestPermutationsTwoOfThreeOrdersLexicographic(t *testing.T) {
	got := axiomslice.Permutations([]int{1, 2, 3}, 2)
	want := [][]int{{1, 2}, {1, 3}, {2, 1}, {2, 3}, {3, 1}, {3, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Permutations = %v, want %v", got, want)
	}
}

func TestPermutationsTableDriven(t *testing.T) {
	testCases := []struct {
		scenario    string
		input       []int
		selectCount int
		expected    [][]int
	}{
		{scenario: "one of three yields singletons in input order", input: []int{1, 2, 3}, selectCount: 1, expected: [][]int{{1}, {2}, {3}}},
		{scenario: "three of three yields six in index order", input: []int{1, 2, 3}, selectCount: 3, expected: [][]int{{1, 2, 3}, {1, 3, 2}, {2, 1, 3}, {2, 3, 1}, {3, 1, 2}, {3, 2, 1}}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.scenario, func(t *testing.T) {
			got := axiomslice.Permutations(testCase.input, testCase.selectCount)
			if !reflect.DeepEqual(got, testCase.expected) {
				t.Fatalf("Permutations(%v, %d) = %v, want %v", testCase.input, testCase.selectCount, got, testCase.expected)
			}
		})
	}
}

func TestPermutationsStringsStartWithInputOrder(t *testing.T) {
	got := axiomslice.Permutations([]string{"a", "b", "c"}, 2)
	want := [][]string{{"a", "b"}, {"a", "c"}, {"b", "a"}, {"b", "c"}, {"c", "a"}, {"c", "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Permutations strings = %v, want %v", got, want)
	}
}

func TestPermutationsZeroYieldsOneEmpty(t *testing.T) {
	got := axiomslice.Permutations([]int{1, 2, 3}, 0)
	if got == nil || len(got) != 1 {
		t.Fatalf("Permutations(k 0) = %v, want one empty arrangement", got)
	}
	if got[0] == nil || len(got[0]) != 0 {
		t.Fatalf("Permutations(k 0)[0] = %v, want empty non-nil", got[0])
	}
	var numbers []int
	gotNil := axiomslice.Permutations(numbers, 0)
	if gotNil == nil || len(gotNil) != 1 || gotNil[0] == nil || len(gotNil[0]) != 0 {
		t.Fatalf("Permutations(nil, 0) = %v, want one empty non-nil arrangement", gotNil)
	}
}

func TestPermutationsOversizeYieldsEmpty(t *testing.T) {
	got := axiomslice.Permutations([]int{1, 2, 3}, 5)
	if got == nil || len(got) != 0 {
		t.Fatalf("Permutations(k 5) = %v, want empty non-nil", got)
	}
}

func TestPermutationsNilYieldsEmpty(t *testing.T) {
	var numbers []int
	got := axiomslice.Permutations(numbers, 2)
	if got == nil || len(got) != 0 {
		t.Fatalf("Permutations(nil) = %v, want empty non-nil", got)
	}
}

func TestPermutationsEmptyYieldsEmpty(t *testing.T) {
	got := axiomslice.Permutations([]int{}, 1)
	if got == nil || len(got) != 0 {
		t.Fatalf("Permutations(empty) = %v, want empty non-nil", got)
	}
}

func TestPermutationsNegativePanics(t *testing.T) {
	for _, selectCount := range []int{-1, -5} {
		panicked := func() (didPanic any) {
			defer func() { didPanic = recover() }()
			_ = axiomslice.Permutations([]int{1, 2}, selectCount)
			return nil
		}()
		if panicked == nil {
			t.Fatalf("Permutations(k %d) did not panic, want panic", selectCount)
		}
		want := fmt.Sprintf("axiom slice: Permutations needs k >= 0, got %d", selectCount)
		if fmt.Sprintf("%v", panicked) != want {
			t.Fatalf("Permutations panic = %v, want %q", panicked, want)
		}
	}
}

func TestPermutationsCopiesNotAliased(t *testing.T) {
	numbers := []int{1, 2, 3}
	orders := axiomslice.Permutations(numbers, 2)
	if len(orders) == 0 {
		t.Fatal("Permutations returned empty, want non-empty seed for alias pin")
	}
	orders[0][0] = 99
	if !reflect.DeepEqual(numbers, []int{1, 2, 3}) {
		t.Fatalf("source mutated to %v, want [1 2 3]", numbers)
	}
	if !reflect.DeepEqual(orders[1], []int{1, 3}) {
		t.Fatalf("sibling order mutated to %v, want [1 3]", orders[1])
	}
}

func TestPermutationsInputMutationStaysDetached(t *testing.T) {
	numbers := []int{1, 2, 3}
	orders := axiomslice.Permutations(numbers, 2)
	if len(orders) == 0 {
		t.Fatal("Permutations returned empty, want non-empty seed for detach pin")
	}
	numbers[0] = 99
	if reflect.DeepEqual(orders[0], []int{99, 2}) {
		t.Fatalf("order aliases input, got %v after input mutation", orders[0])
	}
}

func TestPermutationsNeverMutatesInput(t *testing.T) {
	numbers := []int{1, 2, 3}
	_ = axiomslice.Permutations(numbers, 2)
	if !reflect.DeepEqual(numbers, []int{1, 2, 3}) {
		t.Fatalf("input mutated to %v, want [1 2 3]", numbers)
	}
}
