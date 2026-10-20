// Package slice provides small typed helpers over []Element, layered on the
// standard library (slices, maps, cmp, iter) rather than replacing it.
//
// Reach for these when lo-style Map/Filter/Reduce/Find/GroupBy reads clearer
// than a hand-rolled loop; reach for the stdlib directly when it already names
// the operation. All helpers are pure: inputs are never mutated.
package slice

import "cmp"

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

// FlatMap maps each element to a slice and concatenates the pieces in order.
// Nil or empty input yields an empty non-nil slice; nil pieces contribute nothing.
func FlatMap[Element any, Transformed any](elements []Element, transform func(Element) []Transformed) []Transformed {
	flattened := make([]Transformed, 0, len(elements))
	for _, element := range elements {
		flattened = append(flattened, transform(element)...)
	}
	return flattened
}

// Flatten concatenates nested slices in order.
// Nil or empty input yields an empty non-nil slice.
func Flatten[Element any](nested [][]Element) []Element {
	total := 0
	for _, inner := range nested {
		total += len(inner)
	}
	flattened := make([]Element, 0, total)
	for _, inner := range nested {
		flattened = append(flattened, inner...)
	}
	return flattened
}

// Chunked splits elements into consecutive chunks of at most size, copying each chunk.
// Size <= 0, or nil/empty input, yields an empty non-nil slice; the last chunk may be short.
func Chunked[Element any](elements []Element, size int) [][]Element {
	if size <= 0 || len(elements) == 0 {
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
// Size <= 0, size beyond the input length, or nil/empty input yields an empty non-nil slice.
func Windows[Element any](elements []Element, size int) [][]Element {
	if size <= 0 || size > len(elements) {
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

// Counts tallies occurrences by element.
// Nil or empty input yields an empty non-nil map.
func Counts[Element comparable](elements []Element) map[Element]int {
	counts := make(map[Element]int, len(elements))
	for _, element := range elements {
		counts[element]++
	}
	return counts
}
