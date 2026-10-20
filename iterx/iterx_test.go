package iterx_test

import (
	"iter"
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

func TestTapSeqRunsEffect(t *testing.T) {
	var observed []int
	got := slices.Collect(iterx.TapSeq(naturals(1, 4), func(candidate int) { observed = append(observed, candidate*10) }))
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("TapSeq = %v, want [1 2 3]", got)
	}
	if !reflect.DeepEqual(observed, []int{10, 20, 30}) {
		t.Fatalf("TapSeq effects = %v, want [10 20 30]", observed)
	}
}

func TestTapSeqEmptyRunsNoEffect(t *testing.T) {
	effects := 0
	got := slices.Collect(iterx.TapSeq(naturals(5, 5), func(candidate int) { effects++ }))
	if len(got) != 0 || effects != 0 {
		t.Fatalf("TapSeq empty = %v effects=%d, want empty 0", got, effects)
	}
}

func TestTapSeqStopsEarly(t *testing.T) {
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
	pipeline := iterx.TapSeq(counting, func(candidate int) { effects++ })
	if pulled != 0 || effects != 0 {
		t.Fatalf("before range pulled=%d effects=%d, want 0 0 (lazy)", pulled, effects)
	}
	got := slices.Collect(iterx.Take(pipeline, 2))
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("TapSeq+Take = %v, want [1 2]", got)
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

func TestDedupSeqDropsAdjacentOnly(t *testing.T) {
	got := slices.Collect(iterx.DedupSeq(slices.Values([]int{1, 1, 2, 1, 1, 3, 3})))
	want := []int{1, 2, 1, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DedupSeq = %v, want %v", got, want)
	}
}

func TestDedupSeqEmptyYieldsEmpty(t *testing.T) {
	if got := slices.Collect(iterx.DedupSeq(naturals(5, 5))); len(got) != 0 {
		t.Fatalf("DedupSeq empty = %v, want empty", got)
	}
}

func TestDedupSeqAllSameKeepsOne(t *testing.T) {
	got := slices.Collect(iterx.DedupSeq(slices.Values([]int{9, 9, 9})))
	if !reflect.DeepEqual(got, []int{9}) {
		t.Fatalf("DedupSeq = %v, want [9]", got)
	}
}

func TestDedupSeqNoDuplicatesKeepsAll(t *testing.T) {
	got := slices.Collect(iterx.DedupSeq(naturals(1, 4)))
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("DedupSeq = %v, want [1 2 3]", got)
	}
}

func TestDedupSeqLazy(t *testing.T) {
	pulled := 0
	counting := func(yield func(int) bool) {
		for _, candidate := range []int{5, 5, 5, 6, 7} {
			pulled++
			if !yield(candidate) {
				return
			}
		}
	}
	pipeline := iterx.DedupSeq(counting)
	if pulled != 0 {
		t.Fatalf("pulled before range = %d, want 0 (lazy)", pulled)
	}
	got := slices.Collect(iterx.Take(pipeline, 1))
	if !reflect.DeepEqual(got, []int{5}) {
		t.Fatalf("DedupSeq+Take = %v, want [5]", got)
	}
	if pulled != 1 {
		t.Fatalf("pulled = %d, want 1 (early termination propagates)", pulled)
	}
}

func TestDedupSeqRerangeReruns(t *testing.T) {
	pipeline := iterx.DedupSeq(slices.Values([]int{4, 4, 5}))
	first := slices.Collect(pipeline)
	second := slices.Collect(pipeline)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("re-range = %v, want repeat of %v", second, first)
	}
	if !reflect.DeepEqual(first, []int{4, 5}) {
		t.Fatalf("collect = %v, want [4 5]", first)
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
