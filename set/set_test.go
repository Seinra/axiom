package set_test

import (
	"testing"

	"github.com/seinra/axiom/set"
)

func TestOfDedupsMembers(t *testing.T) {
	present := set.Of("go", "rust", "go")
	if got := present.Len(); got != 2 {
		t.Fatalf("Len = %d, want 2", got)
	}
	for _, member := range []string{"go", "rust"} {
		if !present.Has(member) {
			t.Fatalf("Has(%q) = false, want true", member)
		}
	}
	if present.Has("zig") {
		t.Fatal("Has(zig) = true, want false")
	}
}

func TestZeroValueIsUsable(t *testing.T) {
	var present set.Set[string]
	if !present.IsEmpty() {
		t.Fatal("IsEmpty = false, want true")
	}
	present.Add("go")
	if !present.Has("go") {
		t.Fatal("Has(go) = false, want true")
	}
	if got := present.Len(); got != 1 {
		t.Fatalf("Len = %d, want 1", got)
	}
}

func TestAddIsIdempotent(t *testing.T) {
	present := set.New[int]()
	present.Add(7)
	present.Add(7)
	if got := present.Len(); got != 1 {
		t.Fatalf("Len = %d, want 1", got)
	}
}

func TestRemoveDeletesOnlyTarget(t *testing.T) {
	present := set.Of("a", "b")
	present.Remove("missing")
	if got := present.Len(); got != 2 {
		t.Fatalf("Len after missing remove = %d, want 2", got)
	}
	present.Remove("a")
	if present.Has("a") {
		t.Fatal("Has(a) = true, want false")
	}
	if !present.Has("b") {
		t.Fatal("Has(b) = false, want true")
	}
}

func TestClearEmpties(t *testing.T) {
	present := set.Of(1, 2, 3)
	present.Clear()
	if !present.IsEmpty() {
		t.Fatal("IsEmpty = false, want true")
	}
	if got := present.Len(); got != 0 {
		t.Fatalf("Len = %d, want 0", got)
	}
}

func TestEqualComparesMembers(t *testing.T) {
	first := set.Of(1, 2, 3)
	second := set.Of(3, 2, 1)
	if !first.Equal(second) {
		t.Fatal("Equal = false for same members, want true")
	}
	third := set.Of(1, 2)
	if first.Equal(third) {
		t.Fatal("Equal = true for different members, want false")
	}
	var emptyA, emptyB set.Set[int]
	if !emptyA.Equal(emptyB) {
		t.Fatal("Equal = false for two empty sets, want true")
	}
}

func TestUnionCombines(t *testing.T) {
	combined := set.Of(1, 2).Union(set.Of(2, 3))
	if got := combined.Len(); got != 3 {
		t.Fatalf("Len = %d, want 3", got)
	}
	for _, member := range []int{1, 2, 3} {
		if !combined.Has(member) {
			t.Fatalf("Has(%d) = false, want true", member)
		}
	}
}

func TestIntersectionKeepsShared(t *testing.T) {
	shared := set.Of(1, 2, 3).Intersection(set.Of(2, 3, 4))
	if got := shared.Len(); got != 2 {
		t.Fatalf("Len = %d, want 2", got)
	}
	for _, member := range []int{2, 3} {
		if !shared.Has(member) {
			t.Fatalf("Has(%d) = false, want true", member)
		}
	}
	if shared.Has(1) || shared.Has(4) {
		t.Fatal("Intersection holds unshared member")
	}
}

func TestDifferenceRemovesOther(t *testing.T) {
	remaining := set.Of(1, 2, 3).Difference(set.Of(2, 4))
	if got := remaining.Len(); got != 2 {
		t.Fatalf("Len = %d, want 2", got)
	}
	for _, member := range []int{1, 3} {
		if !remaining.Has(member) {
			t.Fatalf("Has(%d) = false, want true", member)
		}
	}
	if remaining.Has(2) {
		t.Fatal("Has(2) = true, want false")
	}
}

func TestMembersSnapshotHoldsAll(t *testing.T) {
	present := set.Of("x", "y")
	members := present.Members()
	if len(members) != 2 {
		t.Fatalf("len(Members) = %d, want 2", len(members))
	}
	for _, member := range members {
		if !present.Has(member) {
			t.Fatalf("snapshot holds %q missing from set", member)
		}
	}
}

func TestMapValueTransformsAndDedups(t *testing.T) {
	// Exercises the Go 1.27 generic method: MapValue declares its own type params.
	parities := set.Of(1, 2, 3, 4).MapValue(func(candidate int) int { return candidate % 2 })
	if got := parities.Len(); got != 2 {
		t.Fatalf("Len = %d, want 2", got)
	}
	for _, member := range []int{0, 1} {
		if !parities.Has(member) {
			t.Fatalf("Has(%d) = false, want true", member)
		}
	}
}

func TestCombinationsNeverMutateInputs(t *testing.T) {
	first := set.Of(1, 2)
	second := set.Of(2, 3)
	_ = first.Union(second)
	_ = first.Intersection(second)
	_ = first.Difference(second)
	if !first.Equal(set.Of(1, 2)) || !second.Equal(set.Of(2, 3)) {
		t.Fatalf("inputs mutated: first = %v, second = %v", first.Members(), second.Members())
	}
}
