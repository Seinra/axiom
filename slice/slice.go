// Package slice provides small typed helpers over []Element, layered on the
// standard library (slices, maps, cmp, iter) rather than replacing it.
//
// Reach for these when lo-style Map/Filter/Reduce/Find/GroupBy reads clearer
// than a hand-rolled loop; reach for the stdlib directly when it already names
// the operation. All helpers are pure: inputs are never mutated.
//
// Which function?
//
// | When you want to ...                      | Reach for ...               |
// |-------------------------------------------|-----------------------------|
// | Transform every element                   | MapSlice / MapIndexed       |
// | Keep only matching elements               | FilterSlice / FilterIndexed |
// | Fold to one value                         | ReduceSlice / Sum           |
// | Find the first match                      | FindSlice                   |
// | Smallest / largest, plain or by key       | Min / Max / MinBy / MaxBy   |
// | Mean of numbers                           | Average                     |
// | Bucket elements by key                    | GroupBySlice / Tally        |
// | Map each element to a slice, then join    | FlatMap / Flatten           |
// | Fixed chunks / sliding windows            | Chunked / Windows           |
// | Pair two slices / split pairs back        | Zip / Unzip (see Pair)      |
// | Drop later duplicates                     | Distinct                    |
// | Split into matching plus rest             | Partition                   |
// | Insert a separator between elements       | Intersperse                 |
// | Sort a copy by key, ties keep input order | SortBy                      |
// | First or last match, fallback when absent  | FindOrElse / FirstOr / LastOr |
// | Check whether all, any, or none match      | Every / Some / None           |
// | Values repeating, or appearing just once   | FindDuplicates / FindUniques  |
// | Positions of every match                   | FindIndexes                   |
// | First or last n elements, plus the rest    | Take / Drop / TakeLast / SkipLast |
// | Running folds element by element          | RunningFold                 |
// | All k-subsets in index order              | Combinations                |
// | All k-arrangements in index order         | Permutations                |
// | Multiply every number                     | Product                     |
// | Most frequent value                       | Mode                        |
// | Project one field per element             | Pluck                       |
// | Append only when absent                   | EnsureUniqueAndAppend       |
// | Every pair across two slices              | CrossJoin (see Pair)        |
// | Shuffled copy from an explicit source     | Shuffle                     |
// | Integer sequences                         | Range / RangeFrom           |
// | Bound a value into a range                | Clamp                       |
//
// The (value, bool) return on FindSlice, Min, Max, MinBy, MaxBy, Mode, and
// Average is this package's OrNull: the bool reports presence and the
// value is the zero value when absent. Empty input or no match is a normal
// outcome here, never an error; for failures with causes, see the result
// package, and for a wrapped maybe, see opt.
package slice

import (
	"cmp"
	"fmt"
	"math"
	"math/rand"
	"slices"
)

// MapSlice applies transform to every element, preserving order.
func MapSlice[Element any, Transformed any](elements []Element, transform func(Element) Transformed) []Transformed {
	mapped := make([]Transformed, len(elements))
	for index, element := range elements {
		mapped[index] = transform(element)
	}
	return mapped
}

// FilterSlice keeps the elements where keep reports true, preserving order.
// The result is never nil: empty input yields an empty slice.
func FilterSlice[Element any](elements []Element, keep func(Element) bool) []Element {
	kept := make([]Element, 0, len(elements))
	for _, element := range elements {
		if keep(element) {
			kept = append(kept, element)
		}
	}
	return kept
}

// ReduceSlice folds elements into a single accumulation starting at initial.
func ReduceSlice[Element any, Accumulation any](elements []Element, initial Accumulation, combine func(Accumulation, Element) Accumulation) Accumulation {
	total := initial
	for _, element := range elements {
		total = combine(total, element)
	}
	return total
}

// FindSlice returns the first element where match reports true.
// found is false when no element matches; matched is then the zero value.
// The (matched, found) pair is the OrNull shape: check found before trusting matched.
func FindSlice[Element any](elements []Element, match func(Element) bool) (matched Element, found bool) {
	for _, element := range elements {
		if match(element) {
			return element, true
		}
	}
	var empty Element
	return empty, false
}

// GroupBySlice buckets elements by keyOf, preserving encounter order per key.
func GroupBySlice[Element any, Key comparable](elements []Element, keyOf func(Element) Key) map[Key][]Element {
	grouped := make(map[Key][]Element)
	for _, element := range elements {
		key := keyOf(element)
		grouped[key] = append(grouped[key], element)
	}
	return grouped
}

// Pair holds one positional element from each of two zipped slices.
// It is the shared element type of Zip and Unzip.
type Pair[First any, Second any] struct {
	First  First
	Second Second
}

// number covers the built-in numeric types summable by Sum.
type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Min returns the smallest element.
// Empty or nil input yields the zero value and false, consistent with FindSlice.
// The (lowest, ok) pair is the OrNull shape: check ok before trusting the value.
func Min[Element cmp.Ordered](elements []Element) (Element, bool) {
	if len(elements) == 0 {
		var empty Element
		return empty, false
	}
	lowest := elements[0]
	for _, element := range elements[1:] {
		if element < lowest {
			lowest = element
		}
	}
	return lowest, true
}

// Max returns the largest element.
// Empty or nil input yields the zero value and false, consistent with FindSlice.
// The (highest, ok) pair is the OrNull shape: check ok before trusting the value.
func Max[Element cmp.Ordered](elements []Element) (Element, bool) {
	if len(elements) == 0 {
		var empty Element
		return empty, false
	}
	highest := elements[0]
	for _, element := range elements[1:] {
		if element > highest {
			highest = element
		}
	}
	return highest, true
}

// MinBy returns the element with the smallest keyOf key; ties keep the first minimal element.
// Empty or nil input yields the zero value and false, consistent with FindSlice.
// The (lowest, ok) pair is the OrNull shape: check ok before trusting the value.
func MinBy[Element any, Key cmp.Ordered](elements []Element, keyOf func(Element) Key) (Element, bool) {
	if len(elements) == 0 {
		var empty Element
		return empty, false
	}
	lowest := elements[0]
	lowestKey := keyOf(lowest)
	for _, element := range elements[1:] {
		candidate := keyOf(element)
		if candidate < lowestKey {
			lowest = element
			lowestKey = candidate
		}
	}
	return lowest, true
}

// MaxBy returns the element with the largest keyOf key; ties keep the first maximal element.
// Empty or nil input yields the zero value and false, consistent with FindSlice.
// The (highest, ok) pair is the OrNull shape: check ok before trusting the value.
func MaxBy[Element any, Key cmp.Ordered](elements []Element, keyOf func(Element) Key) (Element, bool) {
	if len(elements) == 0 {
		var empty Element
		return empty, false
	}
	highest := elements[0]
	highestKey := keyOf(highest)
	for _, element := range elements[1:] {
		candidate := keyOf(element)
		if candidate > highestKey {
			highest = element
			highestKey = candidate
		}
	}
	return highest, true
}

// Sum adds every number.
// Empty or nil input yields zero.
func Sum[Number number](numbers []Number) Number {
	var total Number
	for _, element := range numbers {
		total += element
	}
	return total
}

// Average returns the arithmetic mean of numbers.
// Empty or nil input yields 0 and false, consistent with Min and Max.
// The (mean, ok) pair is the OrNull shape: check ok before trusting the mean.
// The sum accumulates in float64, so the result is best-effort precision,
// not exactness, for large magnitudes or non-binary fractions.
func Average[Number number](numbers []Number) (float64, bool) {
	if len(numbers) == 0 {
		return 0, false
	}
	var total float64
	for _, element := range numbers {
		total += float64(element)
	}
	return total / float64(len(numbers)), true
}

// Product multiplies every number, pairing with Sum and Average for
// sum-grade arithmetic over the number constraint.
//
// Empty or nil input yields 1, the multiplicative identity, so Product over
// nothing composes with Reduce-style folds the way Sum yields 0.
//
// Integer overflow panics: a wrapped product is programmer-error data,
// matching the checkedAdd guard behind Flatten. The guard is exact integer
// arithmetic, so named number types are covered without reflection: the 1/2
// probe tells truncating integers from floats, and the wrap probe spots the
// signed minimum whose product with -1 has no representable result (plain
// division would trap there instead of failing the check). Float factors
// skip the check and keep Go semantics: overflow yields ±Inf, never a panic.
//
// slices has no product fold; hand-rolled loops are the only stdlib route.
func Product[Number number](numbers []Number) Number {
	var product Number = 1
	for _, element := range numbers {
		product = checkedMul(product, element)
	}
	return product
}

// checkedMul returns running*factor, panicking when the product overflows
// the element range. See Product for the guard policy.
// The MinInt-times--1 check runs first because plain division would trap there instead of failing the check.
func checkedMul[Number number](running, factor Number) Number {
	if running == 0 || factor == 0 {
		return running * factor
	}
	if running-1 > running && factor+1 == 0 {
		panic(fmt.Sprintf("axiom slice: Product overflows the element range: %v * %v", running, factor))
	}
	var one, two Number = 1, 2
	if one/two == 0 {
		if product := running * factor; product/factor != running {
			panic(fmt.Sprintf("axiom slice: Product overflows the element range: %v * %v", running, factor))
		}
	}
	return running * factor
}

// Mode returns the most frequent value.
// Empty or nil input yields the zero value and false, consistent with Min
// and Max. Ties keep the first occurrence in input order, matching the
// MaxBy tie precedent.
// The (frequent, ok) pair is the OrNull shape: check ok before trusting the value.
//
// slices has no most-frequent helper; Tally plus a loop is the only stdlib route.
func Mode[Element comparable](elements []Element) (frequent Element, ok bool) {
	if len(elements) == 0 {
		var empty Element
		return empty, false
	}
	counts := Tally(elements)
	frequent = elements[0]
	frequentCount := counts[frequent]
	for _, element := range elements[1:] {
		if counts[element] > frequentCount {
			frequent = element
			frequentCount = counts[element]
		}
	}
	return frequent, true
}

// Pluck projects one field per element, preserving order.
// The field callback sees the element only: indexed variants are explicit
// elsewhere in this package (MapIndexed, FilterIndexed), never (value,
// index, slice) triples.
//
// Nil or empty input yields an empty non-nil slice.
//
// Go generics have no field projection: struct reflection is the only
// language route, and this package never reflects (see the VISION non-goal
// carried by the future program). lo calls the same shape Pluck.
func Pluck[Element any, Field any](elements []Element, field func(Element) Field) []Field {
	plucked := make([]Field, len(elements))
	for index, element := range elements {
		plucked[index] = field(element)
	}
	return plucked
}

// EnsureUniqueAndAppend appends candidate only when no element equals it,
// building a set on top of a slice. The result always holds distinct
// values only when the input already does: duplicates already present stay
// untouched, and only the candidate is deduplicated.
//
// The result is always a fresh slice: present candidates return a copy of
// the input, absent candidates return the input plus the candidate. The
// input is never mutated and neither path aliases it. The present-path copy
// is deliberate: callers may mutate the result without affecting the input.
//
// Nil input plus a candidate yields a one-element slice.
//
// slices has Contains but no append-if-absent: Contains plus append need a
// manual length check, which this folds into one call.
func EnsureUniqueAndAppend[Element comparable](elements []Element, candidate Element) []Element {
	for _, element := range elements {
		if element == candidate {
			present := make([]Element, len(elements))
			copy(present, elements)
			return present
		}
	}
	grown := make([]Element, len(elements)+1)
	copy(grown, elements)
	grown[len(elements)] = candidate
	return grown
}

// CrossJoin returns every pair across first and second in outer-first-major
// order: for each first element, every second element in order. It pairs
// with Zip (positional pairing) and Combinations (k-subsets of one slice):
// the exam use-case cousin of Combinations in the future program.
//
// Either side nil or empty yields an empty non-nil slice.
//
// The stdlib has no cartesian helper; nested loops are the only route,
// and hand-rolled nests rarely document their order. The Pair element type
// is shared with Zip and Unzip.
func CrossJoin[First any, Second any](first []First, second []Second) []Pair[First, Second] {
	joined := make([]Pair[First, Second], 0, len(first)*len(second))
	for _, outer := range first {
		for _, inner := range second {
			joined = append(joined, Pair[First, Second]{First: outer, Second: inner})
		}
	}
	return joined
}

// Shuffle returns a Fisher-Yates shuffled copy of elements, driven by an
// explicit random source.
//
// The source is explicit so shuffles stay deterministic under test: seed it
// with rand.New(rand.NewSource(seed)) and the order replays. A nil source
// panics: silently falling back to the global rand would hide shared state
// behind a pure signature. The input is never mutated; nil or empty input
// yields an empty non-nil slice.
//
// math/rand.Shuffle mutates in place behind an untyped swap closure; this
// wrapper is the purity-compatible shape (frozen decision 8): fresh copy in,
// fresh copy out, source in hand.
func Shuffle[Element any](elements []Element, source *rand.Rand) []Element {
	if source == nil {
		panic("axiom slice: Shuffle needs a non-nil *rand.Rand source, got nil")
	}
	shuffled := make([]Element, len(elements))
	copy(shuffled, elements)
	source.Shuffle(len(shuffled), func(first, second int) {
		shuffled[first], shuffled[second] = shuffled[second], shuffled[first]
	})
	return shuffled
}

// Range returns the integers 0..count-1 in order: generate, don't allocate
// by hand. Negative counts panic: a negative length is a programmer error,
// matching the size <= 0 panic policy of Chunked and Windows.
//
// Zero yields an empty non-nil slice.
//
// The stdlib has no int range: a counting loop is the only route, and the
// loop rarely documents its bounds. lo calls the same shape Range.
func Range(count int) []int {
	if count < 0 {
		panic(fmt.Sprintf("axiom slice: Range needs count >= 0, got %d", count))
	}
	counted := make([]int, count)
	for index := range counted {
		counted[index] = index
	}
	return counted
}

// RangeFrom returns the integers start..start+count-1 in order, the offset
// cousin of Range. Negative counts panic like Range; a negative start is
// fine and counts up through zero.
//
// Zero counts yield an empty non-nil slice.
//
// The stdlib has no offset range either; RangeFrom(0, count) equals Range(count).
func RangeFrom(start, count int) []int {
	if count < 0 {
		panic(fmt.Sprintf("axiom slice: RangeFrom needs count >= 0, got %d", count))
	}
	ranged := make([]int, count)
	for index := range ranged {
		ranged[index] = start + index
	}
	return ranged
}

// Clamp bounds value into the low..high range: below yields low, above
// yields high, inside passes through. Equal bounds always yield that bound.
//
// Low above high panics: an inverted range is a programmer error, matching
// the negative-count panic policy of Range.
//
// The stdlib has no clamp: min/max builtins need a two-call nest
// (min(max(value, low), high)) that hides the intent and never checks the
// range. lo calls the same shape Clamp.
func Clamp[Number cmp.Ordered](value, low, high Number) Number {
	if low > high {
		panic(fmt.Sprintf("axiom slice: Clamp needs low <= high, got %v > %v", low, high))
	}
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

// FlatMap maps each element to a slice and concatenates the pieces in order.
// Nil or empty input yields an empty non-nil slice; nil pieces contribute nothing.
func FlatMap[Element any, Transformed any](elements []Element, transform func(Element) []Transformed) []Transformed {
	flattened := make([]Transformed, 0, len(elements))
	for _, element := range elements {
		flattened = append(flattened, transform(element)...)
	}
	return flattened
}

// checkedAdd returns total+add, panicking on math.MaxInt overflow: inner
// lengths summing past MaxInt are a programmer error, matching the size <= 0
// panic policy of Chunked and Windows.
func checkedAdd(total, add int) int {
	if add > math.MaxInt-total {
		panic(fmt.Sprintf("axiom slice: Flatten needs total length <= %d, got overflow", math.MaxInt))
	}
	return total + add
}

// Flatten concatenates nested slices in order.
// Nil or empty input yields an empty non-nil slice.
//
// Same elements as slices.Concat(nested...) for non-empty input (Concat
// yields nil on empty; Flatten yields empty non-nil); kept as the
// discoverable one-arg shape with preallocated backing.
//
// Length-total overflow panics: inner lengths summing past math.MaxInt are a
// programmer error, matching the size <= 0 panic policy of Chunked and
// Windows. result.Flatten needs no such guard (single-level unwrap, no
// length total); FlatMap grows incrementally via append (no precomputed total).
func Flatten[Element any](nested [][]Element) []Element {
	total := 0
	for _, inner := range nested {
		total = checkedAdd(total, len(inner))
	}
	flattened := make([]Element, 0, total)
	for _, inner := range nested {
		flattened = append(flattened, inner...)
	}
	return flattened
}

// Chunked splits elements into consecutive chunks of at most size, copying each chunk.
// Size <= 0 panics: a non-positive chunk size is a programmer error, matching
// slices.Chunk. Nil/empty input yields an empty non-nil slice; the last chunk
// may be short.
//
// Cost: O(n) time and O(n) space for n elements — every element is copied once.
func Chunked[Element any](elements []Element, size int) [][]Element {
	if size <= 0 {
		panic(fmt.Sprintf("axiom slice: Chunked needs size >= 1, got %d", size))
	}
	if len(elements) == 0 {
		return make([][]Element, 0)
	}
	chunked := make([][]Element, 0, (len(elements)+size-1)/size)
	for start := 0; start < len(elements); start += size {
		end := start + size
		if end > len(elements) {
			end = len(elements)
		}
		chunk := make([]Element, end-start)
		copy(chunk, elements[start:end])
		chunked = append(chunked, chunk)
	}
	return chunked
}

// Windows returns every consecutive sliding window of exactly size, copying each window.
// Size <= 0 panics: a non-positive window size is a programmer error, matching
// slices.Chunk. Size beyond the input length, or nil/empty input, yields an
// empty non-nil slice.
//
// Cost: O(n*size) time and space for n elements — every window copies size elements.
func Windows[Element any](elements []Element, size int) [][]Element {
	if size <= 0 {
		panic(fmt.Sprintf("axiom slice: Windows needs size >= 1, got %d", size))
	}
	if size > len(elements) {
		return make([][]Element, 0)
	}
	windows := make([][]Element, 0, len(elements)-size+1)
	for start := 0; start+size <= len(elements); start++ {
		window := make([]Element, size)
		copy(window, elements[start:start+size])
		windows = append(windows, window)
	}
	return windows
}

// Zip pairs elements positionally, stopping at the shorter input.
// Either side nil/empty yields an empty non-nil slice.
func Zip[First any, Second any](first []First, second []Second) []Pair[First, Second] {
	shortest := len(first)
	if len(second) < shortest {
		shortest = len(second)
	}
	zipped := make([]Pair[First, Second], 0, shortest)
	for index := 0; index < shortest; index++ {
		zipped = append(zipped, Pair[First, Second]{First: first[index], Second: second[index]})
	}
	return zipped
}

// Unzip splits pairs back into two slices preserving order.
// Nil or empty input yields two empty non-nil slices.
func Unzip[First any, Second any](pairs []Pair[First, Second]) ([]First, []Second) {
	firsts := make([]First, 0, len(pairs))
	seconds := make([]Second, 0, len(pairs))
	for _, pair := range pairs {
		firsts = append(firsts, pair.First)
		seconds = append(seconds, pair.Second)
	}
	return firsts, seconds
}

// Distinct drops later duplicates, keeping first-occurrence order.
// Nil or empty input yields an empty non-nil slice.
func Distinct[Element comparable](elements []Element) []Element {
	seen := make(map[Element]struct{}, len(elements))
	unique := make([]Element, 0, len(elements))
	for _, element := range elements {
		if _, duplicate := seen[element]; duplicate {
			continue
		}
		seen[element] = struct{}{}
		unique = append(unique, element)
	}
	return unique
}

// Partition splits elements into matching (keep true) and rest, both preserving order.
// Both results are non-nil; nil/empty input yields two empty slices.
func Partition[Element any](elements []Element, keep func(Element) bool) (matching []Element, rest []Element) {
	matching = make([]Element, 0, len(elements))
	rest = make([]Element, 0, len(elements))
	for _, element := range elements {
		if keep(element) {
			matching = append(matching, element)
		} else {
			rest = append(rest, element)
		}
	}
	return matching, rest
}

// MapIndexed applies transform to every element with its position, preserving order.
// Nil input yields an empty non-nil slice.
func MapIndexed[Element any, Transformed any](elements []Element, transform func(int, Element) Transformed) []Transformed {
	mapped := make([]Transformed, len(elements))
	for index, element := range elements {
		mapped[index] = transform(index, element)
	}
	return mapped
}

// FilterIndexed keeps elements where keep(position, element) reports true, preserving order.
// The result is never nil: nil/empty input yields an empty slice.
func FilterIndexed[Element any](elements []Element, keep func(int, Element) bool) []Element {
	kept := make([]Element, 0, len(elements))
	for index, element := range elements {
		if keep(index, element) {
			kept = append(kept, element)
		}
	}
	return kept
}

// Intersperse places separator between every two elements.
// Empty input yields an empty non-nil slice; a singleton yields a one-element copy.
func Intersperse[Element any](elements []Element, separator Element) []Element {
	if len(elements) == 0 {
		return make([]Element, 0)
	}
	spaced := make([]Element, 0, 2*len(elements)-1)
	for index, element := range elements {
		if index > 0 {
			spaced = append(spaced, separator)
		}
		spaced = append(spaced, element)
	}
	return spaced
}

// Tally counts occurrences of each element.
// Nil or empty input yields an empty non-nil map.
func Tally[Element comparable](elements []Element) map[Element]int {
	counts := make(map[Element]int, len(elements))
	for _, element := range elements {
		counts[element]++
	}
	return counts
}

// SortBy returns a stable sorted copy of elements ordered by the keyOf key.
// Ties keep input order via slices.SortedStableFunc. The input is never mutated.
//
// For example, ordering words by length keeps equal lengths in encounter order:
//
//	sorted := SortBy(words, func(candidate string) int { return len(candidate) })
//
// Nil or empty input yields an empty non-nil slice.
func SortBy[Element any, Key cmp.Ordered](elements []Element, keyOf func(Element) Key) []Element {
	if len(elements) == 0 {
		return make([]Element, 0)
	}
	return slices.SortedStableFunc(slices.Values(elements), func(first Element, second Element) int {
		return cmp.Compare(keyOf(first), keyOf(second))
	})
}

// RunningFold returns the running accumulations of combining elements from initial.
// The result length equals the input length: the first slot holds initial
// combined with the first element, and each later slot folds the next element in.
// The input is never mutated.
//
// For example, prefix sums read directly:
//
//	totals := RunningFold(numbers, 0, func(total int, element int) int { return total + element })
//
// Nil or empty input yields an empty non-nil slice.
func RunningFold[Element any, Accumulation any](elements []Element, initial Accumulation, combine func(Accumulation, Element) Accumulation) []Accumulation {
	if len(elements) == 0 {
		return make([]Accumulation, 0)
	}
	running := make([]Accumulation, 0, len(elements))
	current := initial
	for _, element := range elements {
		current = combine(current, element)
		running = append(running, current)
	}
	return running
}

// Combinations returns all k-subsets of elements preserving input order.
//
// Each combination keeps the relative order of the input and the result lists
// combinations in lexicographic order by index: the first combination holds
// the first k elements and later combinations advance the rightmost index possible.
//
// k < 0 panics: a negative selection size is a programmer error, matching the
// size <= 0 panic policy of Chunked and Windows. k == 0 yields exactly one
// empty combination. k beyond the input length, or nil/empty input with k > 0,
// yields an empty non-nil slice. Every inner slice is a fresh copy; the input
// is never mutated.
//
// Cost: O(C(n,k)*k) time and space for n elements — every combination copies k elements.
func Combinations[Element any](elements []Element, k int) [][]Element {
	if k < 0 {
		panic(fmt.Sprintf("axiom slice: Combinations needs k >= 0, got %d", k))
	}
	if k == 0 {
		return [][]Element{make([]Element, 0)}
	}
	if k > len(elements) {
		return make([][]Element, 0)
	}
	combos := make([][]Element, 0)
	current := make([]Element, 0, k)
	var build func(start int)
	build = func(start int) {
		if len(current) == k {
			combo := make([]Element, k)
			copy(combo, current)
			combos = append(combos, combo)
			return
		}
		for position := start; position < len(elements); position++ {
			current = append(current, elements[position])
			build(position + 1)
			current = current[:len(current)-1]
		}
	}
	build(0)
	return combos
}

// Permutations returns all k-arrangements of elements in lexicographic order by index.
//
// Each arrangement lists k distinct input positions and the result starts with
// the input order: the first arrangement holds the first k elements. Later
// arrangements advance in lexicographic order over index tuples.
//
// k < 0 panics: a negative selection size is a programmer error, matching the
// size <= 0 panic policy of Chunked and Windows. k == 0 yields exactly one
// empty arrangement. k beyond the input length, or nil/empty input with k > 0,
// yields an empty non-nil slice. Every inner slice is a fresh copy; the input
// is never mutated.
//
// Cost: O(P(n,k)*k) time and space for n elements — every arrangement copies k elements.
func Permutations[Element any](elements []Element, k int) [][]Element {
	if k < 0 {
		panic(fmt.Sprintf("axiom slice: Permutations needs k >= 0, got %d", k))
	}
	if k == 0 {
		return [][]Element{make([]Element, 0)}
	}
	if k > len(elements) {
		return make([][]Element, 0)
	}
	orders := make([][]Element, 0)
	used := make([]bool, len(elements))
	current := make([]Element, 0, k)
	var build func()
	build = func() {
		if len(current) == k {
			ordered := make([]Element, k)
			copy(ordered, current)
			orders = append(orders, ordered)
			return
		}
		for position := range elements {
			if used[position] {
				continue
			}
			used[position] = true
			current = append(current, elements[position])
			build()
			current = current[:len(current)-1]
			used[position] = false
		}
	}
	build()
	return orders
}

// FindOrElse returns the first element where match reports true, or fallback
// when no element matches.
//
// Empty or nil input yields fallback without calling match; there is no bool
// to check. This is the Or-canonical shape: absence is a value, not a flag.
// slices has no fallback find; FindSlice reports presence as a bool instead.
func FindOrElse[Element any](elements []Element, match func(Element) bool, fallback Element) Element {
	for _, element := range elements {
		if match(element) {
			return element
		}
	}
	return fallback
}

// FirstOr returns the first element, or fallback when the slice is empty.
//
// Empty or nil input yields fallback. This is the Or-canonical shape:
// absence is a value, not a bool, for callers that already have a default.
// slices has no first-with-fallback; indexing elements[0] panics on empty.
func FirstOr[Element any](elements []Element, fallback Element) Element {
	if len(elements) == 0 {
		return fallback
	}
	return elements[0]
}

// LastOr returns the last element, or fallback when the slice is empty.
//
// Empty or nil input yields fallback. This is the Or-canonical shape:
// absence is a value, not a bool, for callers that already have a default.
// slices has no last-with-fallback; indexing elements[len(elements)-1] panics on empty.
func LastOr[Element any](elements []Element, fallback Element) Element {
	if len(elements) == 0 {
		return fallback
	}
	return elements[len(elements)-1]
}

// Every reports whether test holds for every element.
//
// Empty or nil input yields true: vacuously, nothing violates the test.
// This mirrors the universal-quantifier convention, not an emptiness check.
// slices has no quantifiers; hand-rolled loops are the only stdlib route.
func Every[Element any](elements []Element, test func(Element) bool) bool {
	for _, element := range elements {
		if !test(element) {
			return false
		}
	}
	return true
}

// Some reports whether test holds for at least one element.
//
// Empty or nil input yields false: no element can satisfy the test.
// slices has no quantifiers; hand-rolled loops are the only stdlib route.
func Some[Element any](elements []Element, test func(Element) bool) bool {
	for _, element := range elements {
		if test(element) {
			return true
		}
	}
	return false
}

// None reports whether test holds for no element.
//
// Empty or nil input yields true: no element violates the absence.
// None is the exact complement of Some over the same input and test.
// slices has no quantifiers; hand-rolled loops are the only stdlib route.
func None[Element any](elements []Element, test func(Element) bool) bool {
	for _, element := range elements {
		if test(element) {
			return false
		}
	}
	return true
}

// FindDuplicates returns the values occurring more than once, in order of
// first occurrence, each reported exactly once.
//
// Values appearing exactly once are skipped. Nil or empty input, or input
// with no repeats, yields an empty non-nil slice.
// slices has no duplicate finder; maps-plus-loop is the only stdlib route.
func FindDuplicates[Element comparable](elements []Element) []Element {
	counts := make(map[Element]int, len(elements))
	for _, element := range elements {
		counts[element]++
	}
	duplicates := make([]Element, 0, len(elements))
	emitted := make(map[Element]struct{}, len(elements))
	for _, element := range elements {
		if _, done := emitted[element]; done {
			continue
		}
		if counts[element] > 1 {
			emitted[element] = struct{}{}
			duplicates = append(duplicates, element)
		}
	}
	return duplicates
}

// FindUniques returns the values occurring exactly once, in input order.
//
// Values appearing more than once are skipped. Nil or empty input, or input
// where every value repeats, yields an empty non-nil slice.
// slices has no uniques finder; maps-plus-loop is the only stdlib route.
func FindUniques[Element comparable](elements []Element) []Element {
	counts := make(map[Element]int, len(elements))
	for _, element := range elements {
		counts[element]++
	}
	uniques := make([]Element, 0, len(elements))
	for _, element := range elements {
		if counts[element] == 1 {
			uniques = append(uniques, element)
		}
	}
	return uniques
}

// FindIndexes returns the positions where match reports true, ascending.
//
// Absence is an empty non-nil slice, not a bool or -1: the OrNull shape for
// find-index absence. Nil or empty input, or no match, yields an empty
// non-nil slice.
// slices has no matching-indexes helper; IndexFunc reports only the first.
func FindIndexes[Element any](elements []Element, match func(Element) bool) []int {
	indexes := make([]int, 0, len(elements))
	for position, element := range elements {
		if match(element) {
			indexes = append(indexes, position)
		}
	}
	return indexes
}

// Take returns the first count elements as a fresh copy.
//
// Count below zero panics: a negative count is a programmer error, matching
// the size <= 0 panic policy of Chunked and Windows. Count zero yields an
// empty non-nil slice; count at or beyond the input length yields a full
// copy. The input is never mutated and the result shares no backing array.
// slices has no take-first-n; slicing elements[:count] aliases and panics past the end.
func Take[Element any](elements []Element, count int) []Element {
	if count < 0 {
		panic(fmt.Sprintf("axiom slice: Take needs count >= 0, got %d", count))
	}
	if count >= len(elements) {
		taken := make([]Element, len(elements))
		copy(taken, elements)
		return taken
	}
	taken := make([]Element, count)
	copy(taken, elements[:count])
	return taken
}

// Drop returns the elements after the first count as a fresh copy.
//
// Count below zero panics: a negative count is a programmer error, matching
// the size <= 0 panic policy of Chunked and Windows. Count zero yields a
// full copy; count at or beyond the input length yields an empty non-nil
// slice. The input is never mutated and the result shares no backing array.
// slices has no drop-first-n; slicing elements[count:] aliases and panics past the end.
func Drop[Element any](elements []Element, count int) []Element {
	if count < 0 {
		panic(fmt.Sprintf("axiom slice: Drop needs count >= 0, got %d", count))
	}
	if count >= len(elements) {
		return make([]Element, 0)
	}
	rest := make([]Element, len(elements)-count)
	copy(rest, elements[count:])
	return rest
}

// TakeLast returns the last count elements as a fresh copy.
//
// Count below zero panics: a negative count is a programmer error, matching
// the size <= 0 panic policy of Chunked and Windows. Count zero yields an
// empty non-nil slice; count at or beyond the input length yields a full
// copy. The input is never mutated and the result shares no backing array.
// slices has no take-last-n; slicing the tail aliases and panics past the end.
func TakeLast[Element any](elements []Element, count int) []Element {
	if count < 0 {
		panic(fmt.Sprintf("axiom slice: TakeLast needs count >= 0, got %d", count))
	}
	if count >= len(elements) {
		taken := make([]Element, len(elements))
		copy(taken, elements)
		return taken
	}
	taken := make([]Element, count)
	copy(taken, elements[len(elements)-count:])
	return taken
}

// SkipLast returns the elements before the last count as a fresh copy.
//
// Count below zero panics: a negative count is a programmer error, matching
// the size <= 0 panic policy of Chunked and Windows. Count zero yields a
// full copy; count at or beyond the input length yields an empty non-nil
// slice. The input is never mutated and the result shares no backing array.
// slices has no skip-last-n; slicing the head aliases and panics past the end.
func SkipLast[Element any](elements []Element, count int) []Element {
	if count < 0 {
		panic(fmt.Sprintf("axiom slice: SkipLast needs count >= 0, got %d", count))
	}
	if count >= len(elements) {
		return make([]Element, 0)
	}
	kept := make([]Element, len(elements)-count)
	copy(kept, elements[:len(elements)-count])
	return kept
}
