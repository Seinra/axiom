package slice_test

import (
	"testing"

	axiomslice "github.com/seinra/axiom/slice"
)

// Representative input sizes for the slice hot paths. These are stable
// workload markers, not tuning targets.
const (
	benchMapFilterSize = 1000
	benchFlatOuter     = 10
	benchFlatInner     = 100
	benchChunkCount    = 1000
	benchChunkSize     = 32
	benchComboSize     = 12
	benchComboSelect   = 4
	benchPermSize      = 8
	benchPermSelect    = 3
)

// benchSink collects result lengths so the compiler keeps benchmarked calls.
var benchSink int

func makeBenchRange(elementCount int) []int {
	elements := make([]int, elementCount)
	for position := range elements {
		elements[position] = position
	}
	return elements
}

func BenchmarkMapSlice(b *testing.B) {
	elements := makeBenchRange(benchMapFilterSize)
	double := func(element int) int { return element * 2 }
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		mapped := axiomslice.MapSlice(elements, double)
		benchSink += len(mapped)
	}
}

func BenchmarkFilterSlice(b *testing.B) {
	elements := makeBenchRange(benchMapFilterSize)
	keepEven := func(element int) bool { return element%2 == 0 }
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		kept := axiomslice.FilterSlice(elements, keepEven)
		benchSink += len(kept)
	}
}

func BenchmarkFlatMap(b *testing.B) {
	elements := makeBenchRange(benchFlatOuter)
	replicate := func(element int) []int {
		piece := make([]int, benchFlatInner)
		for position := range piece {
			piece[position] = element
		}
		return piece
	}
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		flattened := axiomslice.FlatMap(elements, replicate)
		benchSink += len(flattened)
	}
}

func BenchmarkFlatten(b *testing.B) {
	nested := make([][]int, benchFlatOuter)
	for outer := range nested {
		inner := make([]int, benchFlatInner)
		for position := range inner {
			inner[position] = outer*benchFlatInner + position
		}
		nested[outer] = inner
	}
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		flattened := axiomslice.Flatten(nested)
		benchSink += len(flattened)
	}
}

func BenchmarkChunked(b *testing.B) {
	elements := makeBenchRange(benchChunkCount)
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		chunked := axiomslice.Chunked(elements, benchChunkSize)
		benchSink += len(chunked)
	}
}

func BenchmarkWindows(b *testing.B) {
	elements := makeBenchRange(benchChunkCount)
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		windows := axiomslice.Windows(elements, benchChunkSize)
		benchSink += len(windows)
	}
}

// BenchmarkCombinations covers C(12,4) = 495 combinations.
func BenchmarkCombinations(b *testing.B) {
	elements := makeBenchRange(benchComboSize)
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		combos := axiomslice.Combinations(elements, benchComboSelect)
		benchSink += len(combos)
	}
}

// BenchmarkPermutations covers P(8,3) = 336 arrangements.
func BenchmarkPermutations(b *testing.B) {
	elements := makeBenchRange(benchPermSize)
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		orders := axiomslice.Permutations(elements, benchPermSelect)
		benchSink += len(orders)
	}
}

// BenchmarkTake pins the copy cost of taking the leading half of n=1000.
func BenchmarkTake(b *testing.B) {
	elements := makeBenchRange(benchMapFilterSize)
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		taken := axiomslice.Take(elements, benchMapFilterSize/2)
		benchSink += len(taken)
	}
}

// BenchmarkDrop pins the copy cost of dropping the leading half of n=1000.
func BenchmarkDrop(b *testing.B) {
	elements := makeBenchRange(benchMapFilterSize)
	b.ResetTimer()
	b.ReportAllocs()
	for iteration := 0; iteration < b.N; iteration++ {
		rest := axiomslice.Drop(elements, benchMapFilterSize/2)
		benchSink += len(rest)
	}
}
