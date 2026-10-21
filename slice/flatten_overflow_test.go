//go:build !race

package slice_test

import (
	"math"
	"strings"
	"testing"
	"unsafe"

	axiomslice "github.com/seinra/axiom/slice"
)

// TestFlattenLengthTotalOverflowPanics pins the length-total guard: inner
// lengths summing past math.MaxInt panic with an axiom message before any
// allocation. Lengths are fabricated via unsafe.Slice (headers only, never
// dereferenced): 1M headers, zero element bytes, no real OOM. result.Flatten
// needs no such guard (single-level unwrap, no length total) and FlatMap
// grows incrementally via append (no precomputed total), so the guard lives
// on slice.Flatten only.
//
// Excluded under -race on purpose: checkptr rejects fabricated headers that
// straddle allocations. The guard logic itself is race-clean (pure integer
// arithmetic); what -race cannot run is the header fabrication, not the
// behavior under test.
func TestFlattenLengthTotalOverflowPanics(t *testing.T) {
	const pieces = 1 << 20
	each := math.MaxInt/pieces + 1
	var base [1]int
	nested := make([][]int, 0, pieces)
	for range pieces {
		nested = append(nested, unsafe.Slice(&base[0], each))
	}
	panicked := func() (recovered any) {
		defer func() { recovered = recover() }()
		_ = axiomslice.Flatten(nested)
		return nil
	}()
	if panicked == nil {
		t.Fatal("Flatten(length total overflow) did not panic, want axiom panic")
	}
	if message, isString := panicked.(string); !isString || !strings.Contains(message, "axiom slice: Flatten") {
		t.Fatalf("Flatten panic = %v, want message carrying axiom slice: Flatten", panicked)
	}
}
