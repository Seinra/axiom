// Package slice provides small typed helpers over []Element, layered on the
// standard library (slices, maps, cmp, iter) rather than replacing it.
//
// Reach for these when lo-style Map/Filter/Reduce/Find/GroupBy reads clearer
// than a hand-rolled loop; reach for the stdlib directly when it already names
// the operation. All helpers are pure: inputs are never mutated.
package slice

// MapSlice returns one transformed success per element, preserving order.
func MapSlice[Element any, Transformed any](elements []Element, transform func(Element) Transformed) []Transformed {
	mapped := make([]Transformed, len(elements))
	for index, success := range elements {
		mapped[index] = transform(success)
	}
	return mapped
}

// FilterSlice keeps the elements where keep reports true, preserving order.
// The result is never nil: empty input yields an empty slice.
func FilterSlice[Element any](elements []Element, keep func(Element) bool) []Element {
	kept := make([]Element, 0, len(elements))
	for _, success := range elements {
		if keep(success) {
			kept = append(kept, success)
		}
	}
	return kept
}

// ReduceSlice folds elements into a single accumulation starting at initial.
func ReduceSlice[Element any, Accumulation any](elements []Element, initial Accumulation, combine func(Accumulation, Element) Accumulation) Accumulation {
	total := initial
	for _, success := range elements {
		total = combine(total, success)
	}
	return total
}

// FindSlice returns the first element where match reports true.
// found is false when no element matches; matched is then the zero value.
func FindSlice[Element any](elements []Element, match func(Element) bool) (matched Element, found bool) {
	for _, success := range elements {
		if match(success) {
			return success, true
		}
	}
	var empty Element
	return empty, false
}

// GroupBySlice buckets elements by keyOf, preserving encounter order per key.
func GroupBySlice[Element any, Key comparable](elements []Element, keyOf func(Element) Key) map[Key][]Element {
	grouped := make(map[Key][]Element)
	for _, success := range elements {
		key := keyOf(success)
		grouped[key] = append(grouped[key], success)
	}
	return grouped
}
