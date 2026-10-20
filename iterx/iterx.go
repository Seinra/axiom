// Package iterx provides small lazy adapters over iter.Seq, complementing the
// eager helpers in the slice package.
//
// The stdlib iter package defines the sequence protocol on purpose without
// transforms; slices.Collect and maps.Collect consume sequences back into
// containers. Each adapter returns a new sequence: nothing runs until the
// caller ranges or collects, and early termination (break, Take) propagates
// upstream — see TestTakeStopsEarly. Re-ranging a sequence re-runs the whole
// pipeline from the source.
//
// Which function?
//
// | When you want to ...                          | Reach for ...           |
// |------------------------------------------------|-------------------------|
// | Transform elements lazily                      | MapSeq                  |
// | Keep matching elements lazily                  | FilterSeq               |
// | Map and flatten in one lazy pass               | FlatMapSeq / FlattenSeq |
// | Keep / skip the first count                    | Take / Drop             |
// | Keep / skip while a condition holds            | TakeWhile / DropWhile   |
// | Pair positions / pair two sequences            | Enumerate / ZipSeq      |
// | Observe elements without changing flow         | TapSeq                  |
// | Repeat one value / cycle a slice (bound with Take) | Repeat / Cycle       |
// | Collapse adjacent duplicates                   | DedupSeq                |
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

// TakeWhile yields the leading elements of sequence while keep reports true.
// It stops at the first false without yielding it. The boundary element IS
// consumed from the source: range pulls each element before keep runs, so a
// counter observes one pull for the failing element. Everything after the
// boundary stays untouched, and breaking early propagates upstream.
func TakeWhile[Element any](sequence iter.Seq[Element], keep func(Element) bool) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		for element := range sequence {
			if !keep(element) {
				return
			}
			if !yield(element) {
				return
			}
		}
	}
}

// DropWhile skips the leading elements of sequence while drop reports true,
// then yields the remainder unchanged. Once drop reports false the leading
// run is over: later elements yield even when drop would accept them.
func DropWhile[Element any](sequence iter.Seq[Element], drop func(Element) bool) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		dropping := true
		for element := range sequence {
			if dropping {
				if drop(element) {
					continue
				}
				dropping = false
			}
			if !yield(element) {
				return
			}
		}
	}
}

// FlatMapSeq maps each element of sequence to an inner sequence and yields
// every inner element in order. Inner sequences range lazily: transform runs
// only as the caller pulls, and stopping early never advances the source
// past what was yielded.
func FlatMapSeq[Element, Transformed any](sequence iter.Seq[Element], transform func(Element) iter.Seq[Transformed]) iter.Seq[Transformed] {
	return func(yield func(Transformed) bool) {
		for element := range sequence {
			for mapped := range transform(element) {
				if !yield(mapped) {
					return
				}
			}
		}
	}
}

// FlattenSeq yields every element of every inner sequence in order.
// Inner sequences range lazily as the caller pulls.
func FlattenSeq[Element any](sequences iter.Seq[iter.Seq[Element]]) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		for inner := range sequences {
			for element := range inner {
				if !yield(element) {
					return
				}
			}
		}
	}
}

// ZipSeq pairs elements of first and second positionally, stopping at the
// shorter input. Both inputs run through iter.Pull with deferred stop calls,
// so no goroutine leaks when either side exhausts or the caller breaks early.
func ZipSeq[First, Second any](first iter.Seq[First], second iter.Seq[Second]) iter.Seq2[First, Second] {
	return func(yield func(First, Second) bool) {
		nextFirst, stopFirst := iter.Pull(first)
		defer stopFirst()
		nextSecond, stopSecond := iter.Pull(second)
		defer stopSecond()
		for {
			firstElement, firstOpen := nextFirst()
			if !firstOpen {
				return
			}
			secondElement, secondOpen := nextSecond()
			if !secondOpen {
				return
			}
			if !yield(firstElement, secondElement) {
				return
			}
		}
	}
}

// TapSeq runs effect once per element during iteration and yields the
// element unchanged. Use it to observe a pipeline while debugging; it adds
// no buffering and propagates early termination upstream.
func TapSeq[Element any](sequence iter.Seq[Element], effect func(Element)) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		for element := range sequence {
			effect(element)
			if !yield(element) {
				return
			}
		}
	}
}

// Repeat yields element forever. Realize it only through a bounded consumer
// such as Take; ranging it directly never terminates.
func Repeat[Element any](element Element) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		for {
			if !yield(element) {
				return
			}
		}
	}
}

// Cycle repeats elements forever in order. An empty or nil slice yields an
// empty sequence instead of looping forever. Realize infinite output only
// through a bounded consumer such as Take.
func Cycle[Element any](elements []Element) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		if len(elements) == 0 {
			return
		}
		for {
			for _, element := range elements {
				if !yield(element) {
					return
				}
			}
		}
	}
}

// DedupSeq drops adjacent duplicates, keeping the first of each run.
// This matches itertools.groupby semantics: only consecutive equals collapse,
// so [1 2 1] keeps all three. For all-seen uniqueness use a set instead.
func DedupSeq[Element comparable](sequence iter.Seq[Element]) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		started := false
		var previous Element
		for current := range sequence {
			if started && current == previous {
				continue
			}
			started = true
			previous = current
			if !yield(current) {
				return
			}
		}
	}
}
