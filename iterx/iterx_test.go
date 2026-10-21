package iterx_test

import (
	"iter"
	"reflect"
	"slices"
	"testing"

	"github.com/seinra/axiom/iterx"
	"github.com/seinra/axiom/slice"
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

func TestRerangeReruns(t *testing.T) {
	pipeline := iterx.MapSeq(naturals(1, 4), func(candidate int) int { return candidate * 2 })
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{2, 4, 6}) {
		t.Fatalf("collect = %v, want [2 4 6]", first)
	}
}

func TestTakeWhileStopsAtFirstFalse(t *testing.T) {
	got := slices.Collect(iterx.TakeWhile(naturals(1, 8), func(candidate int) bool { return candidate < 4 }))
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("TakeWhile = %v, want %v", got, want)
	}
}

func TestTakeWhileEmptyYieldsEmpty(t *testing.T) {
	got := slices.Collect(iterx.TakeWhile(naturals(5, 5), func(candidate int) bool { return true }))
	if len(got) != 0 {
		t.Fatalf("TakeWhile empty = %v, want empty", got)
	}
}

func TestTakeWhileAllKept(t *testing.T) {
	got := slices.Collect(iterx.TakeWhile(naturals(1, 4), func(candidate int) bool { return true }))
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("TakeWhile = %v, want %v", got, want)
	}
}

func TestTakeWhileStopsEarly(t *testing.T) {
	pulled := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.TakeWhile(counting, func(candidate int) bool { return candidate < 4 })
	if pulled != 0 {
		t.Fatalf("pulled before range = %d, want 0 (lazy)", pulled)
	}
	got := slices.Collect(pipeline)
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("TakeWhile = %v, want [1 2 3]", got)
	}
	// Boundary element 4 is pulled to test it, everything after stays untouched.
	if pulled != 4 {
		t.Fatalf("pulled = %d, want 4 (stop right after first false)", pulled)
	}
}

func TestTakeWhileRerangeReruns(t *testing.T) {
	calls := 0
	counting := func(yield func(int) bool) {
		for _, candidate := range []int{1, 2, 3, 1} {
			calls++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.TakeWhile(counting, func(candidate int) bool { return candidate < 3 })
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{1, 2}) {
		t.Fatalf("collect = %v, want [1 2]", first)
	}
	if calls != 6 {
		t.Fatalf("calls = %d, want 6 (3 pulls per range incl boundary)", calls)
	}
}

func TestDropWhileDropsLeadingRunOnly(t *testing.T) {
	got := slices.Collect(iterx.DropWhile(slices.Values([]int{1, 2, 3, 2, 1}), func(candidate int) bool { return candidate < 3 }))
	want := []int{3, 2, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DropWhile = %v, want %v", got, want)
	}
}

func TestDropWhileEmptyYieldsEmpty(t *testing.T) {
	got := slices.Collect(iterx.DropWhile(naturals(5, 5), func(candidate int) bool { return true }))
	if len(got) != 0 {
		t.Fatalf("DropWhile empty = %v, want empty", got)
	}
}

func TestDropWhileAllDroppedYieldsEmpty(t *testing.T) {
	got := slices.Collect(iterx.DropWhile(naturals(1, 4), func(candidate int) bool { return true }))
	if len(got) != 0 {
		t.Fatalf("DropWhile all-dropped = %v, want empty", got)
	}
}

func TestDropWhileNoneDroppedKeepsAll(t *testing.T) {
	got := slices.Collect(iterx.DropWhile(naturals(1, 4), func(candidate int) bool { return false }))
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DropWhile = %v, want %v", got, want)
	}
}

func TestDropWhileLazy(t *testing.T) {
	pulled := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; candidate < 6; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.DropWhile(counting, func(candidate int) bool { return candidate < 3 })
	if pulled != 0 {
		t.Fatalf("pulled before range = %d, want 0 (lazy)", pulled)
	}
	got := slices.Collect(iterx.Take(pipeline, 1))
	if !reflect.DeepEqual(got, []int{3}) {
		t.Fatalf("DropWhile+Take = %v, want [3]", got)
	}
	if pulled != 3 {
		t.Fatalf("pulled = %d, want 3 (leading run plus first kept)", pulled)
	}
}

func TestDropWhileRerangeReruns(t *testing.T) {
	pipeline := iterx.DropWhile(slices.Values([]int{1, 2, 3, 2}), func(candidate int) bool { return candidate < 3 })
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{3, 2}) {
		t.Fatalf("collect = %v, want [3 2]", first)
	}
}

func TestFlatMapSeqExpands(t *testing.T) {
	got := slices.Collect(iterx.FlatMapSeq(slices.Values([]int{1, 2, 3}), func(candidate int) iter.Seq[int] {
		return slices.Values([]int{candidate, candidate * 10})
	}))
	want := []int{1, 10, 2, 20, 3, 30}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FlatMapSeq = %v, want %v", got, want)
	}
}

func TestFlatMapSeqEmptyOuterYieldsEmpty(t *testing.T) {
	got := slices.Collect(iterx.FlatMapSeq(naturals(5, 5), func(candidate int) iter.Seq[int] {
		return slices.Values([]int{candidate})
	}))
	if len(got) != 0 {
		t.Fatalf("FlatMapSeq empty outer = %v, want empty", got)
	}
}

func TestFlatMapSeqEmptyInnerYieldsEmpty(t *testing.T) {
	got := slices.Collect(iterx.FlatMapSeq(slices.Values([]int{1, 2}), func(candidate int) iter.Seq[int] {
		return slices.Values([]int{})
	}))
	if len(got) != 0 {
		t.Fatalf("FlatMapSeq empty inner = %v, want empty", got)
	}
}

func TestFlatMapSeqLazy(t *testing.T) {
	mapped := 0
	pulled := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.FlatMapSeq(counting, func(candidate int) iter.Seq[int] {
		mapped++
		return slices.Values([]int{candidate, candidate * 100})
	})
	if pulled != 0 || mapped != 0 {
		t.Fatalf("before range pulled=%d mapped=%d, want 0 0 (lazy)", pulled, mapped)
	}
	got := slices.Collect(iterx.Take(pipeline, 3))
	if !reflect.DeepEqual(got, []int{1, 100, 2}) {
		t.Fatalf("FlatMapSeq+Take = %v, want [1 100 2]", got)
	}
	if pulled != 2 || mapped != 2 {
		t.Fatalf("pulled=%d mapped=%d, want 2 2 (early termination propagates)", pulled, mapped)
	}
}

func TestFlattenSeqFlattens(t *testing.T) {
	got := slices.Collect(iterx.FlattenSeq(slices.Values([]iter.Seq[int]{
		slices.Values([]int{1, 2}),
		slices.Values([]int{3}),
		slices.Values([]int{}),
		slices.Values([]int{4, 5}),
	})))
	want := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FlattenSeq = %v, want %v", got, want)
	}
}

func TestFlattenSeqEmptyOuterYieldsEmpty(t *testing.T) {
	got := slices.Collect(iterx.FlattenSeq(slices.Values([]iter.Seq[int]{})))
	if len(got) != 0 {
		t.Fatalf("FlattenSeq empty = %v, want empty", got)
	}
}

func TestFlattenSeqLazy(t *testing.T) {
	pulled := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	outer := slices.Values([]iter.Seq[int]{slices.Values([]int{7}), counting})
	pipeline := iterx.FlattenSeq(outer)
	if pulled != 0 {
		t.Fatalf("pulled before range = %d, want 0 (lazy)", pulled)
	}
	got := slices.Collect(iterx.Take(pipeline, 2))
	if !reflect.DeepEqual(got, []int{7, 1}) {
		t.Fatalf("FlattenSeq+Take = %v, want [7 1]", got)
	}
	if pulled != 1 {
		t.Fatalf("pulled = %d, want 1 (early termination propagates)", pulled)
	}
}

func TestZipSeqPairs(t *testing.T) {
	var firstElements []int
	var secondElements []string
	for firstCandidate, secondCandidate := range iterx.ZipSeq(slices.Values([]int{1, 2}), slices.Values([]string{"go", "rust"})) {
		firstElements = append(firstElements, firstCandidate)
		secondElements = append(secondElements, secondCandidate)
	}
	if !reflect.DeepEqual(firstElements, []int{1, 2}) {
		t.Fatalf("ZipSeq first = %v, want [1 2]", firstElements)
	}
	if !reflect.DeepEqual(secondElements, []string{"go", "rust"}) {
		t.Fatalf("ZipSeq second = %v, want [go rust]", secondElements)
	}
}

func TestZipSeqStopsAtShorterFirst(t *testing.T) {
	count := 0
	for firstCandidate, secondCandidate := range iterx.ZipSeq(slices.Values([]int{1}), slices.Values([]string{"go", "rust", "zig"})) {
		_, _ = firstCandidate, secondCandidate
		count++
	}
	if count != 1 {
		t.Fatalf("ZipSeq count = %d, want 1 (stop at shorter)", count)
	}
}

func TestZipSeqStopsAtShorterSecond(t *testing.T) {
	count := 0
	for firstCandidate, secondCandidate := range iterx.ZipSeq(slices.Values([]int{1, 2, 3}), slices.Values([]string{"go"})) {
		_, _ = firstCandidate, secondCandidate
		count++
	}
	if count != 1 {
		t.Fatalf("ZipSeq count = %d, want 1 (stop at shorter)", count)
	}
}

func TestZipSeqEmptyYieldsEmpty(t *testing.T) {
	count := 0
	for firstCandidate, secondCandidate := range iterx.ZipSeq(naturals(5, 5), slices.Values([]string{"go"})) {
		_, _ = firstCandidate, secondCandidate
		count++
	}
	if count != 0 {
		t.Fatalf("ZipSeq count = %d, want 0", count)
	}
}

func TestZipSeqStopsEarly(t *testing.T) {
	firstPulled := 0
	secondPulled := 0
	firstCounting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			firstPulled++
			if !yield(candidate) {
				return
			}
		}
	}
	secondCounting := func(yield func(string) bool) {
		for _, candidate := range []string{"go", "rust"} {
			secondPulled++
			if !yield(candidate) {
				return
			}
		}
	}
	count := 0
	for firstCandidate, secondCandidate := range iterx.ZipSeq(firstCounting, secondCounting) {
		_, _ = firstCandidate, secondCandidate
		count++
	}
	if count != 2 {
		t.Fatalf("ZipSeq count = %d, want 2 (stop at shorter)", count)
	}
	if firstPulled != 3 {
		t.Fatalf("firstPulled = %d, want 3 (two yields plus one failed pull proving exhaustion check)", firstPulled)
	}
	if secondPulled != 2 {
		t.Fatalf("secondPulled = %d, want 2 (never beyond shorter)", secondPulled)
	}
}

func TestZipSeqTakeStopsEarly(t *testing.T) {
	firstPulled := 0
	secondPulled := 0
	firstCounting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			firstPulled++
			if !yield(candidate) {
				return
			}
		}
	}
	secondCounting := func(yield func(int) bool) {
		for candidate := 10; ; candidate++ {
			secondPulled++
			if !yield(candidate) {
				return
			}
		}
	}
	count := 0
	for firstCandidate, secondCandidate := range iterx.ZipSeq(firstCounting, secondCounting) {
		_, _ = firstCandidate, secondCandidate
		count++
		break
	}
	if count != 1 {
		t.Fatalf("ZipSeq+break count = %d, want 1", count)
	}
	if firstPulled != 1 || secondPulled != 1 {
		t.Fatalf("pulled first=%d second=%d, want 1 1 (early termination propagates)", firstPulled, secondPulled)
	}
}

func TestZipSeqRerangeReruns(t *testing.T) {
	pipeline := iterx.ZipSeq(slices.Values([]int{1, 2}), slices.Values([]string{"go", "rust", "zig"}))
	collect := func() [][2]string {
		var pairs [][2]string
		for firstCandidate, secondCandidate := range pipeline {
			pairs = append(pairs, [2]string{decimalString(firstCandidate), secondCandidate})
		}
		return pairs
	}
	first := collect()
	second := collect()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, [][2]string{{"1", "go"}, {"2", "rust"}}) {
		t.Fatalf("collect = %v, want [[1 go] [2 rust]]", first)
	}
}

func TestOnEachSeqRunsEffect(t *testing.T) {
	var observed []int
	got := slices.Collect(iterx.OnEachSeq(naturals(1, 4), func(candidate int) { observed = append(observed, candidate*10) }))
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("OnEachSeq = %v, want [1 2 3]", got)
	}
	if !reflect.DeepEqual(observed, []int{10, 20, 30}) {
		t.Fatalf("OnEachSeq effects = %v, want [10 20 30]", observed)
	}
}

func TestOnEachSeqEmptyRunsNoEffect(t *testing.T) {
	effects := 0
	got := slices.Collect(iterx.OnEachSeq(naturals(5, 5), func(candidate int) { effects++ }))
	if len(got) != 0 || effects != 0 {
		t.Fatalf("OnEachSeq empty = %v effects=%d, want empty 0", got, effects)
	}
}

func TestOnEachSeqStopsEarly(t *testing.T) {
	effects := 0
	pulled := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.OnEachSeq(counting, func(candidate int) { effects++ })
	if pulled != 0 || effects != 0 {
		t.Fatalf("before range pulled=%d effects=%d, want 0 0 (lazy)", pulled, effects)
	}
	got := slices.Collect(iterx.Take(pipeline, 2))
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("OnEachSeq+Take = %v, want [1 2]", got)
	}
	if effects != 2 || pulled != 2 {
		t.Fatalf("effects=%d pulled=%d, want 2 2 (early termination propagates)", effects, pulled)
	}
}

func TestRepeatTake(t *testing.T) {
	got := slices.Collect(iterx.Take(iterx.Repeat(7), 3))
	if !reflect.DeepEqual(got, []int{7, 7, 7}) {
		t.Fatalf("Repeat+Take = %v, want [7 7 7]", got)
	}
}

func TestRepeatTakeZeroYieldsEmpty(t *testing.T) {
	if got := slices.Collect(iterx.Take(iterx.Repeat(7), 0)); len(got) != 0 {
		t.Fatalf("Repeat+Take(0) = %v, want empty", got)
	}
}

func TestRepeatRerangeReruns(t *testing.T) {
	pipeline := iterx.Take(iterx.Repeat("go"), 2)
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []string{"go", "go"}) {
		t.Fatalf("collect = %v, want [go go]", first)
	}
}

func TestCycleTake(t *testing.T) {
	got := slices.Collect(iterx.Take(iterx.Cycle([]int{10, 20}), 5))
	if !reflect.DeepEqual(got, []int{10, 20, 10, 20, 10}) {
		t.Fatalf("Cycle+Take = %v, want [10 20 10 20 10]", got)
	}
}

func TestCycleEmptyYieldsEmpty(t *testing.T) {
	if got := slices.Collect(iterx.Cycle([]int{})); len(got) != 0 {
		t.Fatalf("Cycle empty = %v, want empty", got)
	}
	if got := slices.Collect(iterx.Take(iterx.Cycle([]int{}), 5)); len(got) != 0 {
		t.Fatalf("Cycle empty+Take = %v, want empty", got)
	}
	if got := slices.Collect(iterx.Cycle[int](nil)); len(got) != 0 {
		t.Fatalf("Cycle nil = %v, want empty", got)
	}
}

func TestCycleRerangeReruns(t *testing.T) {
	pipeline := iterx.Take(iterx.Cycle([]int{1, 2}), 3)
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{1, 2, 1}) {
		t.Fatalf("collect = %v, want [1 2 1]", first)
	}
}

func TestDeduplicateSeqDropsAdjacentOnly(t *testing.T) {
	got := slices.Collect(iterx.DeduplicateSeq(slices.Values([]int{1, 1, 2, 1, 1, 3, 3})))
	want := []int{1, 2, 1, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DeduplicateSeq = %v, want %v", got, want)
	}
}

func TestDeduplicateSeqEmptyYieldsEmpty(t *testing.T) {
	if got := slices.Collect(iterx.DeduplicateSeq(naturals(5, 5))); len(got) != 0 {
		t.Fatalf("DeduplicateSeq empty = %v, want empty", got)
	}
}

func TestDeduplicateSeqAllSameKeepsOne(t *testing.T) {
	got := slices.Collect(iterx.DeduplicateSeq(slices.Values([]int{9, 9, 9})))
	if !reflect.DeepEqual(got, []int{9}) {
		t.Fatalf("DeduplicateSeq = %v, want [9]", got)
	}
}

func TestDeduplicateSeqNoDuplicatesKeepsAll(t *testing.T) {
	got := slices.Collect(iterx.DeduplicateSeq(naturals(1, 4)))
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("DeduplicateSeq = %v, want [1 2 3]", got)
	}
}

func TestDeduplicateSeqLazy(t *testing.T) {
	pulled := 0
	counting := func(yield func(int) bool) {
		for _, candidate := range []int{5, 5, 5, 6, 7} {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.DeduplicateSeq(counting)
	if pulled != 0 {
		t.Fatalf("pulled before range = %d, want 0 (lazy)", pulled)
	}
	got := slices.Collect(iterx.Take(pipeline, 1))
	if !reflect.DeepEqual(got, []int{5}) {
		t.Fatalf("DeduplicateSeq+Take = %v, want [5]", got)
	}
	if pulled != 1 {
		t.Fatalf("pulled = %d, want 1 (early termination propagates)", pulled)
	}
}

func TestDeduplicateSeqRerangeReruns(t *testing.T) {
	pipeline := iterx.DeduplicateSeq(slices.Values([]int{4, 4, 5}))
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{4, 5}) {
		t.Fatalf("collect = %v, want [4 5]", first)
	}
}

func TestUnfoldCountsUp(t *testing.T) {
	pipeline := iterx.Unfold(1, func(current int) (int, int, bool) {
		if current > 3 {
			return 0, 0, false
		}
		return current * 10, current + 1, true
	})
	got := slices.Collect(pipeline)
	want := []int{10, 20, 30}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Unfold = %v, want %v", got, want)
	}
}

func TestUnfoldImmediateStopYieldsEmpty(t *testing.T) {
	pipeline := iterx.Unfold(41, func(current int) (int, int, bool) {
		return 0, current, false
	})
	if got := slices.Collect(pipeline); len(got) != 0 {
		t.Fatalf("Unfold immediate stop = %v, want empty", got)
	}
}

func TestUnfoldLazy(t *testing.T) {
	calls := 0
	pipeline := iterx.Unfold(1, func(current int) (int, int, bool) {
		calls++
		return current, current + 1, true
	})
	if calls != 0 {
		t.Fatalf("calls before range = %d, want 0 (lazy)", calls)
	}
	got := slices.Collect(iterx.Take(pipeline, 3))
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("Unfold+Take = %v, want [1 2 3]", got)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3 (early termination propagates)", calls)
	}
}

func TestUnfoldRerangeReruns(t *testing.T) {
	pipeline := iterx.Unfold(1, func(current int) (int, int, bool) {
		if current > 2 {
			return 0, 0, false
		}
		return current, current + 1, true
	})
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{1, 2}) {
		t.Fatalf("collect = %v, want [1 2]", first)
	}
}

func TestGenerateYieldsUntilFalse(t *testing.T) {
	candidate := 0
	got := slices.Collect(iterx.Generate(func() (int, bool) {
		candidate++
		if candidate > 3 {
			return 0, false
		}
		return candidate * 5, true
	}))
	if !reflect.DeepEqual(got, []int{5, 10, 15}) {
		t.Fatalf("Generate = %v, want [5 10 15]", got)
	}
}

func TestGenerateImmediateStopYieldsEmpty(t *testing.T) {
	got := slices.Collect(iterx.Generate(func() (int, bool) {
		return 0, false
	}))
	if len(got) != 0 {
		t.Fatalf("Generate immediate stop = %v, want empty", got)
	}
}

func TestGenerateTakePullsExactly(t *testing.T) {
	produced := 0
	counting := iterx.Generate(func() (int, bool) {
		produced++
		return produced, true
	})
	if produced != 0 {
		t.Fatalf("produced before range = %d, want 0 (lazy)", produced)
	}
	got := slices.Collect(iterx.Take(counting, 4))
	if !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Fatalf("Generate+Take = %v, want [1 2 3 4]", got)
	}
	if produced != 4 {
		t.Fatalf("produced = %d, want 4 (upstream stops on early break)", produced)
	}
}

func TestGenerateRerangeReruns(t *testing.T) {
	candidate := 0
	pipeline := iterx.Generate(func() (int, bool) {
		candidate++
		if candidate > 2 {
			return 0, false
		}
		return candidate, true
	})
	_ = slices.Collect(pipeline)
	candidate = 0
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(second, []int{1, 2}) {
		t.Fatalf("re-range = %v, want [1 2]", second)
	}
}

func TestFilterMapSeqKeepsOnlyMapped(t *testing.T) {
	got := slices.Collect(iterx.FilterMapSeq(naturals(1, 7), func(candidate int) (int, bool) {
		if candidate%2 != 0 {
			return 0, false
		}
		return candidate * 100, true
	}))
	want := []int{200, 400, 600}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterMapSeq = %v, want %v", got, want)
	}
}

func TestFilterMapSeqNoneKeptYieldsEmpty(t *testing.T) {
	got := slices.Collect(iterx.FilterMapSeq(naturals(1, 4), func(candidate int) (string, bool) {
		return "", false
	}))
	if len(got) != 0 {
		t.Fatalf("FilterMapSeq none kept = %v, want empty", got)
	}
}

func TestFilterMapSeqEmptyYieldsEmpty(t *testing.T) {
	got := slices.Collect(iterx.FilterMapSeq(naturals(5, 5), func(candidate int) (int, bool) {
		return candidate, true
	}))
	if len(got) != 0 {
		t.Fatalf("FilterMapSeq empty = %v, want empty", got)
	}
}

func TestFilterMapSeqLazy(t *testing.T) {
	pulled := 0
	mapped := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.FilterMapSeq(counting, func(candidate int) (int, bool) {
		mapped++
		if candidate%2 != 0 {
			return 0, false
		}
		return candidate * 10, true
	})
	if pulled != 0 || mapped != 0 {
		t.Fatalf("before range pulled=%d mapped=%d, want 0 0 (lazy)", pulled, mapped)
	}
	got := slices.Collect(iterx.Take(pipeline, 2))
	if !reflect.DeepEqual(got, []int{20, 40}) {
		t.Fatalf("FilterMapSeq+Take = %v, want [20 40]", got)
	}
	if pulled != 4 || mapped != 4 {
		t.Fatalf("pulled=%d mapped=%d, want 4 4 (early termination propagates)", pulled, mapped)
	}
}

func TestFilterMapSeqRerangeReruns(t *testing.T) {
	pipeline := iterx.FilterMapSeq(slices.Values([]int{1, 2, 3}), func(candidate int) (int, bool) {
		return candidate * 2, candidate%2 == 1
	})
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{2, 6}) {
		t.Fatalf("collect = %v, want [2 6]", first)
	}
}

func TestChunkedSeqSplits(t *testing.T) {
	got := slices.Collect(iterx.ChunkedSeq(slices.Values([]int{1, 2, 3, 4, 5}), 2))
	want := [][]int{{1, 2}, {3, 4}, {5}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ChunkedSeq = %v, want %v", got, want)
	}
}

func TestChunkedSeqInvalidSizePanics(t *testing.T) {
	for _, chunkSize := range []int{0, -3} {
		panicked := func() (didPanic any) {
			defer func() { didPanic = recover() }()
			_ = iterx.ChunkedSeq(naturals(1, 4), chunkSize)
			return nil
		}()
		if panicked == nil {
			t.Fatalf("ChunkedSeq(%d) did not panic, want panic", chunkSize)
		}
	}
}

func TestChunkedSeqEmptyYieldsEmpty(t *testing.T) {
	if got := slices.Collect(iterx.ChunkedSeq(naturals(5, 5), 2)); len(got) != 0 {
		t.Fatalf("ChunkedSeq empty = %v, want empty", got)
	}
}

func TestChunkedSeqMirrorsSliceChunked(t *testing.T) {
	inputs := [][]int{{}, {1}, {1, 2, 3, 4, 5}, {1, 2, 3, 4}}
	for _, input := range inputs {
		for _, chunkSize := range []int{1, 2, 3, 5} {
			got := slices.Collect(iterx.ChunkedSeq(slices.Values(input), chunkSize))
			want := slice.Chunked(input, chunkSize)
			if len(got) == 0 && len(want) == 0 {
				continue
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("ChunkedSeq(%v, %d) = %v, want %v", input, chunkSize, got, want)
			}
		}
	}
}

func TestChunkedSeqFreshCopy(t *testing.T) {
	got := slices.Collect(iterx.ChunkedSeq(slices.Values([]int{1, 2, 3}), 2))
	if len(got) != 2 {
		t.Fatalf("ChunkedSeq count = %d, want 2", len(got))
	}
	got[0][0] = 999
	second := slices.Collect(iterx.ChunkedSeq(slices.Values([]int{1, 2, 3}), 2))
	if !reflect.DeepEqual(second, [][]int{{1, 2}, {3}}) {
		t.Fatalf("ChunkedSeq re-collect = %v, want [[1 2] [3]] (each chunk fresh)", second)
	}
	if got[0][0] != 999 || second[0][0] != 1 {
		t.Fatalf("mutation leaked: got=%v second=%v", got, second)
	}
}

func TestChunkedSeqStopsEarly(t *testing.T) {
	pulled := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.ChunkedSeq(counting, 3)
	if pulled != 0 {
		t.Fatalf("pulled before range = %d, want 0 (lazy)", pulled)
	}
	got := slices.Collect(iterx.Take(pipeline, 2))
	if !reflect.DeepEqual(got, [][]int{{1, 2, 3}, {4, 5, 6}}) {
		t.Fatalf("ChunkedSeq+Take = %v, want [[1 2 3] [4 5 6]]", got)
	}
	if pulled != 6 {
		t.Fatalf("pulled = %d, want 6 (upstream stops on early break)", pulled)
	}
}

func TestChunkedSeqRerangeReruns(t *testing.T) {
	pipeline := iterx.ChunkedSeq(slices.Values([]int{1, 2, 3}), 2)
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
}

func TestWindowsSeqSlides(t *testing.T) {
	got := slices.Collect(iterx.WindowsSeq(slices.Values([]int{1, 2, 3, 4}), 2))
	want := [][]int{{1, 2}, {2, 3}, {3, 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("WindowsSeq = %v, want %v", got, want)
	}
}

func TestWindowsSeqInvalidSizePanics(t *testing.T) {
	for _, windowSize := range []int{0, -2} {
		panicked := func() (didPanic any) {
			defer func() { didPanic = recover() }()
			_ = iterx.WindowsSeq(naturals(1, 4), windowSize)
			return nil
		}()
		if panicked == nil {
			t.Fatalf("WindowsSeq(%d) did not panic, want panic", windowSize)
		}
	}
}

func TestWindowsSeqOversizeYieldsEmpty(t *testing.T) {
	if got := slices.Collect(iterx.WindowsSeq(slices.Values([]int{1, 2}), 5)); len(got) != 0 {
		t.Fatalf("WindowsSeq oversize = %v, want empty", got)
	}
}

func TestWindowsSeqEmptyYieldsEmpty(t *testing.T) {
	if got := slices.Collect(iterx.WindowsSeq(naturals(5, 5), 2)); len(got) != 0 {
		t.Fatalf("WindowsSeq empty = %v, want empty", got)
	}
}

func TestWindowsSeqMirrorsSliceWindows(t *testing.T) {
	inputs := [][]int{{}, {1}, {1, 2, 3, 4, 5}}
	for _, input := range inputs {
		for _, windowSize := range []int{1, 2, 3, 5, 6} {
			got := slices.Collect(iterx.WindowsSeq(slices.Values(input), windowSize))
			want := slice.Windows(input, windowSize)
			if len(got) == 0 && len(want) == 0 {
				continue
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("WindowsSeq(%v, %d) = %v, want %v", input, windowSize, got, want)
			}
		}
	}
}

func TestWindowsSeqFreshCopy(t *testing.T) {
	got := slices.Collect(iterx.WindowsSeq(slices.Values([]int{1, 2, 3}), 2))
	if len(got) != 2 {
		t.Fatalf("WindowsSeq count = %d, want 2", len(got))
	}
	got[0][0] = 999
	if got[1][0] == 999 {
		t.Fatalf("WindowsSeq windows alias: %v (each window fresh)", got)
	}
	second := slices.Collect(iterx.WindowsSeq(slices.Values([]int{1, 2, 3}), 2))
	if !reflect.DeepEqual(second, [][]int{{1, 2}, {2, 3}}) {
		t.Fatalf("WindowsSeq re-collect = %v, want [[1 2] [2 3]]", second)
	}
}

func TestWindowsSeqStopsEarly(t *testing.T) {
	pulled := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; ; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.WindowsSeq(counting, 2)
	if pulled != 0 {
		t.Fatalf("pulled before range = %d, want 0 (lazy)", pulled)
	}
	got := slices.Collect(iterx.Take(pipeline, 2))
	if !reflect.DeepEqual(got, [][]int{{1, 2}, {2, 3}}) {
		t.Fatalf("WindowsSeq+Take = %v, want [[1 2] [2 3]]", got)
	}
	if pulled != 3 {
		t.Fatalf("pulled = %d, want 3 (upstream stops on early break)", pulled)
	}
}

func TestWindowsSeqRerangeReruns(t *testing.T) {
	pipeline := iterx.WindowsSeq(slices.Values([]int{1, 2, 3}), 2)
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
}

func TestIntersperseSeqSlotsSeparator(t *testing.T) {
	got := slices.Collect(iterx.IntersperseSeq(slices.Values([]int{1, 2, 3}), 0))
	want := []int{1, 0, 2, 0, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("IntersperseSeq = %v, want %v", got, want)
	}
}

func TestIntersperseSeqEmptyYieldsEmpty(t *testing.T) {
	if got := slices.Collect(iterx.IntersperseSeq(naturals(5, 5), 0)); len(got) != 0 {
		t.Fatalf("IntersperseSeq empty = %v, want empty", got)
	}
}

func TestIntersperseSeqSingletonKeepsOne(t *testing.T) {
	got := slices.Collect(iterx.IntersperseSeq(slices.Values([]int{7}), 0))
	if !reflect.DeepEqual(got, []int{7}) {
		t.Fatalf("IntersperseSeq singleton = %v, want [7]", got)
	}
}

func TestIntersperseSeqMirrorsSliceIntersperse(t *testing.T) {
	inputs := [][]int{{}, {1}, {1, 2, 3, 4}}
	for _, input := range inputs {
		got := slices.Collect(iterx.IntersperseSeq(slices.Values(input), 9))
		want := slice.Intersperse(input, 9)
		if len(got) == 0 && len(want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("IntersperseSeq(%v) = %v, want %v", input, got, want)
		}
	}
}

func TestIntersperseSeqStopsEarly(t *testing.T) {
	pulled := 0
	counting := func(yield func(int) bool) {
		for candidate := 1; candidate < 100; candidate++ {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.IntersperseSeq(counting, 0)
	if pulled != 0 {
		t.Fatalf("pulled before range = %d, want 0 (lazy)", pulled)
	}
	got := slices.Collect(iterx.Take(pipeline, 3))
	if !reflect.DeepEqual(got, []int{1, 0, 2}) {
		t.Fatalf("IntersperseSeq+Take = %v, want [1 0 2]", got)
	}
	if pulled != 2 {
		t.Fatalf("pulled = %d, want 2 (upstream stops on early break)", pulled)
	}
}

func TestIntersperseSeqRerangeReruns(t *testing.T) {
	pipeline := iterx.IntersperseSeq(slices.Values([]int{1, 2}), 0)
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{1, 0, 2}) {
		t.Fatalf("collect = %v, want [1 0 2]", first)
	}
}

func TestConcatSeqConcatenates(t *testing.T) {
	got := slices.Collect(iterx.ConcatSeq(naturals(1, 3), naturals(10, 12), naturals(20, 21)))
	want := []int{1, 2, 10, 11, 20}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ConcatSeq = %v, want %v", got, want)
	}
}

func TestConcatSeqSingleYieldsOnlyFirst(t *testing.T) {
	got := slices.Collect(iterx.ConcatSeq(naturals(1, 4)))
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("ConcatSeq single = %v, want [1 2 3]", got)
	}
}

func TestConcatSeqEmptySidesYieldRemainder(t *testing.T) {
	got := slices.Collect(iterx.ConcatSeq(naturals(5, 5), naturals(1, 3), naturals(5, 5)))
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("ConcatSeq empties = %v, want [1 2]", got)
	}
	if got := slices.Collect(iterx.ConcatSeq(naturals(5, 5))); len(got) != 0 {
		t.Fatalf("ConcatSeq all empty = %v, want empty", got)
	}
}

func TestConcatSeqStopsEarly(t *testing.T) {
	firstPulled := 0
	secondPulled := 0
	firstCounting := func(yield func(int) bool) {
		for candidate := 1; candidate < 100; candidate++ {
			firstPulled++
			if !yield(candidate) {
				return
			}
		}
	}
	secondCounting := func(yield func(int) bool) {
		for candidate := 100; candidate < 200; candidate++ {
			secondPulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.ConcatSeq(firstCounting, secondCounting)
	if firstPulled != 0 || secondPulled != 0 {
		t.Fatalf("before range first=%d second=%d, want 0 0 (lazy)", firstPulled, secondPulled)
	}
	got := slices.Collect(iterx.Take(pipeline, 2))
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("ConcatSeq+Take = %v, want [1 2]", got)
	}
	if firstPulled != 2 || secondPulled != 0 {
		t.Fatalf("pulled first=%d second=%d, want 2 0 (never touches later input)", firstPulled, secondPulled)
	}
}

func TestConcatSeqRerangeReruns(t *testing.T) {
	pipeline := iterx.ConcatSeq(slices.Values([]int{1}), slices.Values([]int{2, 3}))
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{1, 2, 3}) {
		t.Fatalf("collect = %v, want [1 2 3]", first)
	}
}

func decimalString(candidate int) string {
	if candidate == 0 {
		return "0"
	}
	negative := candidate < 0
	if negative {
		candidate = -candidate
	}
	var digits []byte
	for candidate > 0 {
		digits = append([]byte{byte('0' + candidate%10)}, digits...)
		candidate /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
