package set_test

import (
	"sync"
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
	_ = first.SymmetricDifference(second)
	if !first.Equal(set.Of(1, 2)) || !second.Equal(set.Of(2, 3)) {
		t.Fatalf("inputs mutated: first = %v, second = %v", first.Members(), second.Members())
	}
}

func TestSymmetricDifferenceKeepsExactlyOneSide(t *testing.T) {
	differed := set.Of(1, 2, 3).SymmetricDifference(set.Of(2, 3, 4))
	want := set.Of(1, 4)
	if !differed.Equal(want) {
		t.Fatalf("SymmetricDifference = %v, want %v", differed.Members(), want.Members())
	}
}

func TestSymmetricDifferenceWithEmpty(t *testing.T) {
	populated := set.Of(1, 2)
	var empty set.Set[int]
	if got := populated.SymmetricDifference(empty); !got.Equal(populated) {
		t.Fatalf("populated ^ empty = %v, want %v", got.Members(), populated.Members())
	}
	if got := empty.SymmetricDifference(populated); !got.Equal(populated) {
		t.Fatalf("empty ^ populated = %v, want %v", got.Members(), populated.Members())
	}
}

func TestSymmetricDifferenceBothEmpty(t *testing.T) {
	var first, second set.Set[int]
	differed := first.SymmetricDifference(second)
	if !differed.IsEmpty() {
		t.Fatalf("empty ^ empty len = %d, want 0", differed.Len())
	}
	if differed.Len() != 0 {
		t.Fatalf("Len = %d, want 0", differed.Len())
	}
}

func TestIsSubsetOf(t *testing.T) {
	if !set.Of(1, 2).IsSubsetOf(set.Of(1, 2, 3)) {
		t.Fatal("IsSubsetOf = false for held members, want true")
	}
	if !set.Of(1, 2).IsSubsetOf(set.Of(1, 2)) {
		t.Fatal("IsSubsetOf = false for equal sets, want true")
	}
	if set.Of(1, 4).IsSubsetOf(set.Of(1, 2, 3)) {
		t.Fatal("IsSubsetOf = true with unheld member, want false")
	}
}

func TestIsSubsetOfEmptyEdges(t *testing.T) {
	var empty set.Set[int]
	if !empty.IsSubsetOf(set.Of(1)) {
		t.Fatal("empty IsSubsetOf populated = false, want true")
	}
	var otherEmpty set.Set[int]
	if !empty.IsSubsetOf(otherEmpty) {
		t.Fatal("empty IsSubsetOf empty = false, want true")
	}
	if set.Of(1).IsSubsetOf(empty) {
		t.Fatal("populated IsSubsetOf empty = true, want false")
	}
}

func TestIsSupersetOf(t *testing.T) {
	if !set.Of(1, 2, 3).IsSupersetOf(set.Of(1, 2)) {
		t.Fatal("IsSupersetOf = false for held members, want true")
	}
	if !set.Of(1, 2).IsSupersetOf(set.Of(1, 2)) {
		t.Fatal("IsSupersetOf = false for equal sets, want true")
	}
	if set.Of(1, 2).IsSupersetOf(set.Of(1, 4)) {
		t.Fatal("IsSupersetOf = true with unheld member, want false")
	}
}

func TestIsSupersetOfEmptyEdges(t *testing.T) {
	var empty set.Set[int]
	if !set.Of(1).IsSupersetOf(empty) {
		t.Fatal("populated IsSupersetOf empty = false, want true")
	}
	var otherEmpty set.Set[int]
	if !empty.IsSupersetOf(otherEmpty) {
		t.Fatal("empty IsSupersetOf empty = false, want true")
	}
	if empty.IsSupersetOf(set.Of(1)) {
		t.Fatal("empty IsSupersetOf populated = true, want false")
	}
}

func TestOverlapsReportsShared(t *testing.T) {
	if !set.Of(1, 2, 3).Overlaps(set.Of(3, 4)) {
		t.Fatal("Overlaps = false for shared member, want true")
	}
	if set.Of(1, 2).Overlaps(set.Of(3, 4)) {
		t.Fatal("Overlaps = true for disjoint sets, want false")
	}
}

func TestOverlapsEmptyEdges(t *testing.T) {
	var empty set.Set[int]
	if empty.Overlaps(set.Of(1)) {
		t.Fatal("empty Overlaps populated = true, want false")
	}
	if set.Of(1).Overlaps(empty) {
		t.Fatal("populated Overlaps empty = true, want false")
	}
	var otherEmpty set.Set[int]
	if empty.Overlaps(otherEmpty) {
		t.Fatal("empty Overlaps empty = true, want false")
	}
}

func TestIsDisjointTable(t *testing.T) {
	testCases := []struct {
		name   string
		first  set.Set[int]
		second set.Set[int]
		want   bool
	}{
		{
			name:   "disjoint holdings",
			first:  set.Of(1, 2),
			second: set.Of(3, 4),
			want:   true,
		},
		{
			name:   "shared member",
			first:  set.Of(1, 2, 3),
			second: set.Of(3, 4),
			want:   false,
		},
		{
			name:   "identical non-empty",
			first:  set.Of(1, 2),
			second: set.Of(1, 2),
			want:   false,
		},
		{
			name:   "subset shares members",
			first:  set.Of(1),
			second: set.Of(1, 2),
			want:   false,
		},
		{
			name:   "empty against populated",
			first:  set.Set[int]{},
			second: set.Of(1),
			want:   true,
		},
		{
			name:   "populated against empty",
			first:  set.Of(1),
			second: set.Set[int]{},
			want:   true,
		},
		{
			name:   "both empty",
			first:  set.Set[int]{},
			second: set.New[int](),
			want:   true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.first.IsDisjoint(testCase.second); got != testCase.want {
				t.Fatalf("IsDisjoint = %v, want %v", got, testCase.want)
			}
			if got := testCase.first.IsDisjoint(testCase.second); got != !testCase.first.Overlaps(testCase.second) {
				t.Fatalf("IsDisjoint = %v, want complement of Overlaps %v", got, testCase.first.Overlaps(testCase.second))
			}
		})
	}
}

func TestEmptySnapshotIsNonNil(t *testing.T) {
	var zero set.Set[string]
	if members := zero.Members(); members == nil || len(members) != 0 {
		t.Fatalf("zero Members() = %v, want empty non-nil", members)
	}
	if members := set.New[int]().Members(); members == nil || len(members) != 0 {
		t.Fatalf("New Members() = %v, want empty non-nil", members)
	}
	if members := set.Of[int]().Members(); members == nil || len(members) != 0 {
		t.Fatalf("Of() Members() = %v, want empty non-nil", members)
	}
	cleared := set.Of(1, 2)
	cleared.Clear()
	if members := cleared.Members(); members == nil || len(members) != 0 {
		t.Fatalf("cleared Members() = %v, want empty non-nil", members)
	}
}

func TestConcurrentHammerStaysRaceClean(t *testing.T) {
	const members = 64
	const workers = 8
	base := set.New[int]()
	for candidate := 0; candidate < members; candidate++ {
		base.Add(candidate)
	}
	snapshot := base.Members()
	if len(snapshot) != members {
		t.Fatalf("len(base Members) = %d, want %d", len(snapshot), members)
	}
	failures := make(chan string, workers)
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func(worker int) {
			defer group.Done()
			for candidate := 0; candidate < members; candidate++ {
				if !base.Has(candidate) {
					failures <- "worker missing member"
					return
				}
			}
			if base.Len() != members {
				failures <- "worker observes mutated base"
				return
			}
			if base.IsEmpty() {
				failures <- "worker sees empty base"
				return
			}
			if got := len(base.Members()); got != members {
				failures <- "worker snapshot shrank"
				return
			}
			if !base.Equal(set.Of(snapshot...)) {
				failures <- "worker Equal mismatch"
				return
			}
			local := set.New[int]()
			for candidate := 0; candidate < members; candidate++ {
				local.Add(worker*members + candidate)
			}
			if local.Len() != members {
				failures <- "worker local lost members"
				return
			}
			if !local.Has(worker*members) || local.Has(-1) {
				failures <- "worker local membership wrong"
				return
			}
			local.Remove(worker * members)
			if local.Has(worker*members) || local.Len() != members-1 {
				failures <- "worker local remove failed"
				return
			}
			if !local.IsDisjoint(base) && worker > 0 {
				failures <- "worker local overlaps base, want disjoint offsets"
				return
			}
		}(worker)
	}
	group.Wait()
	close(failures)
	for failure := range failures {
		t.Fatal(failure)
	}
	if base.Len() != members {
		t.Fatalf("base Len = %d, want %d (hammer must not mutate the shared set)", base.Len(), members)
	}
}
