package slice_test

import (
	"math"
	"slices"
	"testing"

	axiomslice "github.com/seinra/axiom/slice"
)

// Local fuzz budgets keep short sessions fast: inputs beyond these bounds
// are skipped, never failed.
const (
	fuzzMaxElements     = 10
	fuzzMaxPermutations = 5000
	fuzzMaxChunkInputs  = 256
)

// binomialCoefficient returns C(elementCount, selectCount) with plain integer
// math. Callers only invoke it when 0 <= selectCount <= elementCount.
func binomialCoefficient(elementCount, selectCount int) int {
	choose := selectCount
	if elementCount-choose < choose {
		choose = elementCount - choose
	}
	total := 1
	for position := 1; position <= choose; position++ {
		total = total * (elementCount - choose + position) / position
	}
	return total
}

// permutationCount returns P(elementCount, selectCount) with plain integer
// math. Callers only invoke it when 0 <= selectCount <= elementCount.
func permutationCount(elementCount, selectCount int) int {
	total := 1
	for position := 0; position < selectCount; position++ {
		total *= elementCount - position
	}
	return total
}

// isSubsequence reports whether candidate appears in order inside haystack,
// so combinations keep the input order even with duplicate values.
func isSubsequence(candidate, haystack []byte) bool {
	offset := 0
	for _, wanted := range candidate {
		matched := false
		for offset < len(haystack) {
			current := haystack[offset]
			offset++
			if current == wanted {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// multisetContained reports whether arrangement draws only elements present
// in source without exceeding their multiplicity.
func multisetContained(arrangement, source []byte) bool {
	var remaining [256]int
	for _, element := range source {
		remaining[element]++
	}
	for _, element := range arrangement {
		if remaining[element] == 0 {
			return false
		}
		remaining[element]--
	}
	return true
}

// allDistinct reports whether every element in values appears exactly once,
// so oracles only pin output uniqueness when the input has no duplicates.
func allDistinct(values []byte) bool {
	seen := make(map[byte]struct{}, len(values))
	for _, element := range values {
		if _, dup := seen[element]; dup {
			return false
		}
		seen[element] = struct{}{}
	}
	return true
}

// requireDistinctResults pins pairwise-distinct outputs whenever the input
// elements are all distinct. Equal values from duplicate inputs remain
// allowed and correct, so mixed inputs skip this check.
func requireDistinctResults(t *testing.T, snapshot []byte, actual [][]byte, label string) {
	t.Helper()
	if !allDistinct(snapshot) {
		return
	}
	seen := make(map[string]struct{}, len(actual))
	for _, arrangement := range actual {
		key := string(arrangement)
		if _, dup := seen[key]; dup {
			t.Fatalf("%s duplicate %v for distinct input %v", label, arrangement, snapshot)
		}
		seen[key] = struct{}{}
	}
}

func requireUnmutated(t *testing.T, snapshot, elements []byte, label string) {
	t.Helper()
	if !slices.Equal(snapshot, elements) {
		t.Fatalf("%s mutated input to %v, want %v", label, elements, snapshot)
	}
}

func checkCombinationSet(t *testing.T, snapshot []byte, actualCombos [][]byte, selectCount int) {
	t.Helper()
	if actualCombos == nil {
		t.Fatal("Combinations returned nil, want non-nil")
	}
	expectedCount := 0
	if selectCount <= len(snapshot) {
		expectedCount = binomialCoefficient(len(snapshot), selectCount)
	}
	if len(actualCombos) != expectedCount {
		t.Fatalf("Combinations len = %d, want %d (C(%d,%d))", len(actualCombos), expectedCount, len(snapshot), selectCount)
	}
	if expectedCount == 0 {
		return
	}
	if !slices.Equal(actualCombos[0], snapshot[:selectCount]) {
		t.Fatalf("Combinations first = %v, want leading %v", actualCombos[0], snapshot[:selectCount])
	}
	for _, combo := range actualCombos {
		if len(combo) != selectCount {
			t.Fatalf("combination len = %d, want %d", len(combo), selectCount)
		}
		if !isSubsequence(combo, snapshot) {
			t.Fatalf("combination %v breaks input order of %v", combo, snapshot)
		}
	}
	requireDistinctResults(t, snapshot, actualCombos, "Combinations")
}

func checkPermutationSet(t *testing.T, snapshot []byte, actualOrders [][]byte, selectCount int) {
	t.Helper()
	if actualOrders == nil {
		t.Fatal("Permutations returned nil, want non-nil")
	}
	expectedCount := 0
	if selectCount <= len(snapshot) {
		expectedCount = permutationCount(len(snapshot), selectCount)
	}
	if len(actualOrders) != expectedCount {
		t.Fatalf("Permutations len = %d, want %d (P(%d,%d))", len(actualOrders), expectedCount, len(snapshot), selectCount)
	}
	if expectedCount == 0 {
		return
	}
	if !slices.Equal(actualOrders[0], snapshot[:selectCount]) {
		t.Fatalf("Permutations first = %v, want leading %v", actualOrders[0], snapshot[:selectCount])
	}
	for _, order := range actualOrders {
		if len(order) != selectCount {
			t.Fatalf("arrangement len = %d, want %d", len(order), selectCount)
		}
		if !multisetContained(order, snapshot) {
			t.Fatalf("arrangement %v draws outside input %v", order, snapshot)
		}
	}
	requireDistinctResults(t, snapshot, actualOrders, "Permutations")
}

func checkChunkCoverage(t *testing.T, snapshot []byte, actualChunks [][]byte, chunkSize int) {
	t.Helper()
	if actualChunks == nil {
		t.Fatal("Chunked returned nil, want non-nil")
	}
	joined := make([]byte, 0, len(snapshot))
	for index, chunk := range actualChunks {
		if len(chunk) == 0 || len(chunk) > chunkSize {
			t.Fatalf("chunk len = %d, want 1..%d", len(chunk), chunkSize)
		}
		if index < len(actualChunks)-1 && len(chunk) != chunkSize {
			t.Fatalf("chunk %d len = %d, want %d (non-final chunks are full)", index, len(chunk), chunkSize)
		}
		joined = append(joined, chunk...)
	}
	if !slices.Equal(joined, snapshot) {
		t.Fatalf("chunk concat = %v, want %v", joined, snapshot)
	}
	if len(actualChunks) > 0 {
		remainder := len(snapshot) % chunkSize
		wantLast := remainder
		if remainder == 0 {
			wantLast = chunkSize
		}
		if last := len(actualChunks[len(actualChunks)-1]); last != wantLast {
			t.Fatalf("last chunk len = %d, want %d", last, wantLast)
		}
	}
	// Overflow-safe chunk count: len(snapshot) <= 256 always, but chunkSize
	// is a fuzzed int that can reach math.MaxInt, so len+size-1 would wrap.
	var expectedChunks int
	if chunkSize >= len(snapshot) {
		if len(snapshot) > 0 {
			expectedChunks = 1
		}
	} else {
		expectedChunks = (len(snapshot) + chunkSize - 1) / chunkSize
	}
	if len(actualChunks) != expectedChunks {
		t.Fatalf("chunk count = %d, want %d", len(actualChunks), expectedChunks)
	}
}

func checkWindowCoverage(t *testing.T, snapshot []byte, actualWindows [][]byte, windowSize int) {
	t.Helper()
	if actualWindows == nil {
		t.Fatal("Windows returned nil, want non-nil")
	}
	expectedCount := 0
	if windowSize <= len(snapshot) {
		expectedCount = len(snapshot) - windowSize + 1
	}
	if len(actualWindows) != expectedCount {
		t.Fatalf("window count = %d, want %d", len(actualWindows), expectedCount)
	}
	for start, window := range actualWindows {
		if len(window) != windowSize {
			t.Fatalf("window len = %d, want %d", len(window), windowSize)
		}
		if !slices.Equal(window, snapshot[start:start+windowSize]) {
			t.Fatalf("window %d = %v, want %v", start, window, snapshot[start:start+windowSize])
		}
	}
}

// FuzzCombinations pins count, index order, and input immutability over the
// table-case seed corpus. Negative selectCount stays with unit tests: it
// panics by design, so the fuzzer skips it instead of reporting a crash.
func FuzzCombinations(f *testing.F) {
	f.Add([]byte(nil), 2)
	f.Add([]byte{}, 1)
	f.Add([]byte{1, 2, 3}, 0)
	f.Add([]byte{1, 2, 3}, 1)
	f.Add([]byte{1, 2, 3}, 2)
	f.Add([]byte{1, 2, 3}, 3)
	f.Add([]byte{1, 2, 3, 4}, 2)
	f.Add([]byte{2, 1, 2}, 2)
	f.Add([]byte{1, 2, 3}, 5)
	f.Fuzz(func(t *testing.T, elements []byte, selectCount int) {
		if selectCount < 0 {
			t.Skip("negative selectCount panics by design (unit tests pin it)")
		}
		if len(elements) > fuzzMaxElements {
			t.Skip("oversize input would explode the combination count")
		}
		// No combination-count budget guard: len(elements) <= 10 caps the
		// enumeration at C(10,5) = 252, already below any useful skip budget.
		snapshot := slices.Clone(elements)
		actualCombos := axiomslice.Combinations(elements, selectCount)
		checkCombinationSet(t, snapshot, actualCombos, selectCount)
		requireUnmutated(t, snapshot, elements, "Combinations")
	})
}

// FuzzPermutations pins count, input-order start, multiset membership, and
// input immutability over the table-case seed corpus. Negative selectCount
// stays with unit tests: it panics by design, so the fuzzer skips it.
func FuzzPermutations(f *testing.F) {
	f.Add([]byte(nil), 2)
	f.Add([]byte{}, 1)
	f.Add([]byte{1, 2, 3}, 0)
	f.Add([]byte{1, 2, 3}, 1)
	f.Add([]byte{1, 2, 3}, 2)
	f.Add([]byte{1, 2, 3}, 3)
	f.Add([]byte{2, 1, 2}, 2)
	f.Add([]byte{1, 2, 3}, 5)
	f.Fuzz(func(t *testing.T, elements []byte, selectCount int) {
		if selectCount < 0 {
			t.Skip("negative selectCount panics by design (unit tests pin it)")
		}
		if len(elements) > fuzzMaxElements {
			t.Skip("oversize input would explode the arrangement count")
		}
		if selectCount <= len(elements) && permutationCount(len(elements), selectCount) > fuzzMaxPermutations {
			t.Skip("arrangement count beyond the local fuzz budget")
		}
		snapshot := slices.Clone(elements)
		actualOrders := axiomslice.Permutations(elements, selectCount)
		checkPermutationSet(t, snapshot, actualOrders, selectCount)
		requireUnmutated(t, snapshot, elements, "Permutations")
	})
}

// FuzzChunkedWindows pins non-nil output plus chunk coverage (concat of chunks
// equals the input) and window placement over valid sizes. Non-positive sizes
// panic by design and stay with unit tests: the fuzzer only feeds valid sizes.
func FuzzChunkedWindows(f *testing.F) {
	f.Add([]byte(nil), 2)
	f.Add([]byte{}, 2)
	f.Add([]byte{1, 2, 3, 4}, 2)
	f.Add([]byte{1, 2, 3, 4, 5}, 2)
	f.Add([]byte{1, 2}, 2)
	f.Add([]byte{1, 2, 3}, 5)
	f.Add([]byte{1, 2, 3, 4}, 1)
	f.Fuzz(func(t *testing.T, elements []byte, windowSize int) {
		if windowSize <= 0 {
			t.Skip("non-positive size panics by design (unit tests pin it)")
		}
		if len(elements) > fuzzMaxChunkInputs {
			t.Skip("oversize input beyond the local fuzz budget")
		}
		snapshot := slices.Clone(elements)
		actualChunks := axiomslice.Chunked(elements, windowSize)
		actualWindows := axiomslice.Windows(elements, windowSize)
		checkChunkCoverage(t, snapshot, actualChunks, windowSize)
		checkWindowCoverage(t, snapshot, actualWindows, windowSize)
		requireUnmutated(t, snapshot, elements, "Chunked/Windows")
	})
}

// TestChunkedWindowsHugeSize pins the overflow-safe chunk oracle with an
// adversarial size fuzz seeds may never hit: chunkSize = math.MaxInt would
// wrap len+size-1, so Chunked must yield one full chunk and Windows none.
func TestChunkedWindowsHugeSize(t *testing.T) {
	elements := []byte{1, 2, 3, 4}
	snapshot := slices.Clone(elements)
	checkChunkCoverage(t, snapshot, axiomslice.Chunked(elements, math.MaxInt), math.MaxInt)
	checkWindowCoverage(t, snapshot, axiomslice.Windows(elements, math.MaxInt), math.MaxInt)
	requireUnmutated(t, snapshot, elements, "Chunked/Windows")

	empty := []byte{}
	emptySnapshot := slices.Clone(empty)
	checkChunkCoverage(t, emptySnapshot, axiomslice.Chunked(empty, math.MaxInt), math.MaxInt)
	checkWindowCoverage(t, emptySnapshot, axiomslice.Windows(empty, math.MaxInt), math.MaxInt)
	requireUnmutated(t, emptySnapshot, empty, "Chunked/Windows")
}
