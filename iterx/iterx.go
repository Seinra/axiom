// Package iterx provides small lazy adapters over iter.Seq, complementing the
// eager helpers in the slice package.
//
// The stdlib iter package defines the sequence protocol on purpose without
// transforms; slices.Collect and maps.Collect consume sequences back into
// containers. Each adapter returns a new sequence: nothing runs until the
// caller ranges or collects, and early termination (break, Take) propagates
// upstream — see TestTakeStopsEarly. Re-ranging a sequence re-runs the whole
// pipeline from the source.
package iterx

import "iter"

// MapSeq transforms every element of sequence, preserving order.
func MapSeq[Element any, Transformed any](sequence iter.Seq[Element], transform func(Element) Transformed) iter.Seq[Transformed] {
	return func(yield func(Transformed) bool) {
		for element := range sequence {
			if !yield(transform(element)) {
				return
			}
		}
	}
}

// FilterSeq keeps the elements of sequence where keep reports true,
// preserving order.
func FilterSeq[Element any](sequence iter.Seq[Element], keep func(Element) bool) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		for element := range sequence {
			if keep(element) && !yield(element) {
				return
			}
		}
	}
}

// Take keeps at most the first count elements of sequence.
// A non-positive count yields an empty sequence.
// Pulls exactly what it yields: stopping early never advances the source
// past the kept elements (see TestTakeStopsEarly). That precision needs
// iter.Pull — a range loop would request one element too many before noticing
// the limit, because the pull happens as part of the iteration step.
func Take[Element any](sequence iter.Seq[Element], count int) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		if count <= 0 {
			return
		}
		next, stop := iter.Pull(sequence)
		defer stop()
		for range count {
			element, ok := next()
			if !ok {
				return
			}
			if !yield(element) {
				return
			}
		}
	}
}

// Drop skips the first count elements of sequence, yielding the rest.
// A non-positive count yields the whole sequence.
func Drop[Element any](sequence iter.Seq[Element], count int) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		skipped := 0
		for element := range sequence {
			if skipped < count {
				skipped++
				continue
			}
			if !yield(element) {
				return
			}
		}
	}
}

// Enumerate pairs every element of sequence with its zero-based position.
func Enumerate[Element any](sequence iter.Seq[Element]) iter.Seq2[int, Element] {
	return func(yield func(int, Element) bool) {
		position := 0
		for element := range sequence {
			if !yield(position, element) {
				return
			}
			position++
		}
	}
}
