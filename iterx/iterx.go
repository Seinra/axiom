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
// Fail-fast exception: ChunkedSeq and WindowsSeq validate size at call time
// and panic on size <= 0 instead of yielding nothing — valid sizes stay
// fully lazy.
//
// Generate, FilterMapSeq, and Unfold take OrNull-shaped callbacks: the bool
// reports presence — never trust the value without it (see the OrNull
// definition in the README guide index).
//
// Which function?
//
// | When you want to ...                               | Reach for ...           |
// |----------------------------------------------------|-------------------------|
// | Transform elements lazily                          | MapSeq                  |
// | Keep matching elements lazily                      | FilterSeq               |
// | Map and flatten in one lazy pass                   | FlatMapSeq / FlattenSeq |
// | Keep / skip the first count                        | Take / Drop             |
// | Keep / skip while a condition holds                | TakeWhile / DropWhile   |
// | Pair positions / pair two sequences                | Enumerate / ZipSeq      |
// | Observe elements without changing flow             | OnEachSeq               |
// | Repeat one value / cycle a slice (bound with Take) | Repeat / Cycle          |
// | Collapse adjacent duplicates                       | DeduplicateSeq          |
// | Unfold evolving state into values                  | Unfold                  |
// | Generate values until a stop                       | Generate                |
// | Map and keep only some results                     | FilterMapSeq            |
// | Fixed chunks, last short (slice.Chunked)           | ChunkedSeq              |
// | Sliding windows (slice.Windows)                    | WindowsSeq              |
// | Slot a separator between elements                  | IntersperseSeq          |
// | Concatenate sequences in order                     | ConcatSeq             |
package iterx

import (
	"fmt"
	"iter"
)

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

// OnEachSeq runs effect once per element during iteration and yields the
// element unchanged. Use it to observe a pipeline while debugging; it adds
// no buffering and propagates early termination upstream.
func OnEachSeq[Element any](sequence iter.Seq[Element], effect func(Element)) iter.Seq[Element] {
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

// DeduplicateSeq drops adjacent duplicates, keeping the first of each run.
// This matches itertools.groupby semantics: only consecutive equals collapse,
// so [1 2 1] keeps all three. For all-seen uniqueness use a set instead.
func DeduplicateSeq[Element comparable](sequence iter.Seq[Element]) iter.Seq[Element] {
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

// Unfold builds a sequence from evolving state: following maps the current
// state to an element plus the next state, stopping when ok is false. The
// following callback is OrNull-shaped: ok reports presence — never trust the
// element or next state without it.
//
//	Example: Unfold(1, func(current int) (int, int, bool) {
//	  if current > 3 { return 0, 0, false }
//	  return current, current+1, true
//	}) yields 1, 2, 3.
//
// Nothing runs until ranged; breaking early stops calling following, and
// re-ranging restarts from initial.
func Unfold[State any, Element any](initial State, following func(State) (element Element, next State, ok bool)) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		current := initial
		for {
			element, upcoming, open := following(current)
			if !open {
				return
			}
			current = upcoming
			if !yield(element) {
				return
			}
		}
	}
}

// Generate builds a sequence by calling produce for each element, stopping
// when ok is false. The produce callback is OrNull-shaped: ok reports
// presence — never trust element without it.
//
//	Example: with candidate starting at 0, Generate(func() (int, bool) {
//	  candidate++
//	  if candidate > 2 { return 0, false }
//	  return candidate, true
//	}) yields 1, 2.
//
// Nothing runs until ranged: Take(Generate(produce), count) calls produce
// exactly count times, and re-ranging restarts the calls.
func Generate[Element any](produce func() (element Element, ok bool)) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		for {
			element, open := produce()
			if !open {
				return
			}
			if !yield(element) {
				return
			}
		}
	}
}

// FilterMapSeq maps each element of sequence and keeps only the mappings
// where keep is true, preserving order. The transform callback is
// OrNull-shaped: keep reports presence — never trust mapped without it.
//
//	Example: FilterMapSeq(values, func(candidate int) (int, bool) {
//	  if candidate%2 != 0 { return 0, false }
//	  return candidate*10, true
//	}) keeps the doubled even inputs.
//
// Transform runs lazily as the caller pulls; breaking early stops pulling
// the source, and re-ranging re-runs from the source.
func FilterMapSeq[Element any, Transformed any](sequence iter.Seq[Element], transform func(Element) (mapped Transformed, keep bool)) iter.Seq[Transformed] {
	return func(yield func(Transformed) bool) {
		for element := range sequence {
			mapped, keep := transform(element)
			if keep && !yield(mapped) {
				return
			}
		}
	}
}

// ChunkedSeq groups sequence into fixed chunks of size, mirroring
// slice.Chunked: the last chunk may be short, and every yielded chunk is a
// fresh copy.
//
// Example: ChunkedSeq(slices.Values([]int{1, 2, 3, 4, 5}), 2) yields
// [1 2], [3 4], [5].
//
// Size <= 0 panics at call time: a non-positive chunk size is a programmer
// error, matching slices.Chunk. Chunks fill lazily as the caller pulls;
// breaking early stops pulling the source, and re-ranging re-runs from
// the source.
//
// Cost: O(n) time and O(size) working space for n elements — every element
// is copied once into a fresh chunk.
func ChunkedSeq[Element any](sequence iter.Seq[Element], size int) iter.Seq[[]Element] {
	if size <= 0 {
		panic(fmt.Sprintf("axiom iterx: ChunkedSeq needs size >= 1, got %d", size))
	}
	return func(yield func([]Element) bool) {
		// No presized backing: a huge size must not allocate upfront. The chunk
		// grows incrementally and never exceeds size elements.
		chunk := make([]Element, 0)
		for element := range sequence {
			chunk = append(chunk, element)
			if len(chunk) < size {
				continue
			}
			snapshot := append([]Element(nil), chunk...)
			chunk = chunk[:0]
			if !yield(snapshot) {
				return
			}
		}
		if len(chunk) == 0 {
			return
		}
		if !yield(append([]Element(nil), chunk...)) {
			return
		}
	}
}

// WindowsSeq yields every consecutive sliding window of exactly size,
// mirroring slice.Windows, and every yielded window is a fresh copy.
//
// Example: WindowsSeq(slices.Values([]int{1, 2, 3}), 2) yields
// [1 2], [2 3].
//
// Size <= 0 panics at call time: a non-positive window size is a programmer
// error, matching slices.Chunk. A size beyond the input length yields an
// empty sequence. Windows fill lazily as the caller pulls; breaking early
// stops pulling the source, and re-ranging re-runs from the source.
//
// Cost: O(n*size) element copies for n elements — every window is a fresh copy.
func WindowsSeq[Element any](sequence iter.Seq[Element], size int) iter.Seq[[]Element] {
	if size <= 0 {
		panic(fmt.Sprintf("axiom iterx: WindowsSeq needs size >= 1, got %d", size))
	}
	return func(yield func([]Element) bool) {
		// No presized backing: a huge size must not allocate upfront. The window
		// grows incrementally and never exceeds size elements.
		window := make([]Element, 0)
		for element := range sequence {
			window = append(window, element)
			if len(window) < size {
				continue
			}
			snapshot := append([]Element(nil), window...)
			window = window[1:]
			if !yield(snapshot) {
				return
			}
		}
	}
}

// IntersperseSeq slots separator between every two elements of sequence,
// mirroring slice.Intersperse.
//
// Example: IntersperseSeq(slices.Values([]int{1, 2, 3}), 0) yields
// 1, 0, 2, 0, 3. Empty input yields nothing; a singleton yields its one
// element with no separator.
//
// Elements flow lazily as the caller pulls; breaking early stops pulling
// the source, and re-ranging re-runs from the source.
func IntersperseSeq[Element any](sequence iter.Seq[Element], separator Element) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		started := false
		for element := range sequence {
			if started {
				if !yield(separator) {
					return
				}
			} else {
				started = true
			}
			if !yield(element) {
				return
			}
		}
	}
}

// ConcatSeq concatenates first with each sequence in rest, in order.
//
// Example: ConcatSeq(slices.Values([]int{1}), slices.Values([]int{2, 3}))
// yields 1, 2, 3.
//
// Inputs range lazily in order: later inputs never start while the caller
// still pulls earlier ones, breaking early stops the current input, and
// re-ranging re-runs from the first input.
func ConcatSeq[Element any](first iter.Seq[Element], rest ...iter.Seq[Element]) iter.Seq[Element] {
	return func(yield func(Element) bool) {
		for element := range first {
			if !yield(element) {
				return
			}
		}
		for _, remaining := range rest {
			for element := range remaining {
				if !yield(element) {
					return
				}
			}
		}
	}
}
