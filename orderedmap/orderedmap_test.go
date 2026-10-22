package orderedmap_test

import (
	"slices"
	"testing"

	"github.com/seinra/axiom/orderedmap"
)

func TestMapNewStartsEmpty(t *testing.T) {
	ordered := orderedmap.New[string, int]()
	if ordered.Len() != 0 {
		t.Fatalf("Len = %d, want 0", ordered.Len())
	}
	if !ordered.IsEmpty() {
		t.Fatal("IsEmpty = false, want true")
	}
	if _, found := ordered.Get("missing"); found {
		t.Fatal("Get on empty map found a key, want absent")
	}
	if keys := ordered.Keys(); len(keys) != 0 {
		t.Fatalf("Keys = %v, want empty", keys)
	}
	if values := ordered.Values(); len(values) != 0 {
		t.Fatalf("Values = %v, want empty", values)
	}
}

func TestMapZeroValueReady(t *testing.T) {
	var ordered orderedmap.Map[string, int]
	if !ordered.IsEmpty() {
		t.Fatal("zero Map IsEmpty = false, want true")
	}
	ordered.Put("alpha", 1)
	value, found := ordered.Get("alpha")
	if !found || value != 1 {
		t.Fatalf("Get after Put on zero Map = (%d, %v), want (1, true)", value, found)
	}
}

func TestMapPutGetRoundTrip(t *testing.T) {
	ordered := orderedmap.New[string, int]()
	ordered.Put("alpha", 1)
	ordered.Put("beta", 2)
	for key, want := range map[string]int{"alpha": 1, "beta": 2} {
		value, found := ordered.Get(key)
		if !found || value != want {
			t.Errorf("Get(%q) = (%d, %v), want (%d, true)", key, value, found, want)
		}
	}
	if ordered.Len() != 2 {
		t.Fatalf("Len = %d, want 2", ordered.Len())
	}
}

func TestMapGetMissingYieldsOrNull(t *testing.T) {
	ordered := orderedmap.New[string, int]()
	ordered.Put("alpha", 1)
	value, found := ordered.Get("missing")
	if found {
		t.Fatal("Get on missing key returned found = true, want false")
	}
	if value != 0 {
		t.Fatalf("Get on missing key returned %d, want zero value", value)
	}
}

func TestMapUpsertKeepsFirstInsertionPosition(t *testing.T) {
	ordered := orderedmap.New[string, int]()
	ordered.Put("alpha", 1)
	ordered.Put("beta", 2)
	ordered.Put("alpha", 10)
	if ordered.Len() != 3-1 {
		t.Fatalf("Len after upsert = %d, want 2", ordered.Len())
	}
	value, found := ordered.Get("alpha")
	if !found || value != 10 {
		t.Fatalf("Get(alpha) after upsert = (%d, %v), want (10, true)", value, found)
	}
	if keys := ordered.Keys(); !slices.Equal(keys, []string{"alpha", "beta"}) {
		t.Fatalf("Keys after upsert = %v, want [alpha beta]", keys)
	}
	if values := ordered.Values(); !slices.Equal(values, []int{10, 2}) {
		t.Fatalf("Values after upsert = %v, want [10 2]", values)
	}
}

func TestMapDeleteMissingKeyIsNoOp(t *testing.T) {
	ordered := orderedmap.New[string, int]()
	ordered.Put("alpha", 1)
	ordered.Delete("missing")
	if ordered.Len() != 1 {
		t.Fatalf("Len after deleting missing key = %d, want 1", ordered.Len())
	}
	if keys := ordered.Keys(); !slices.Equal(keys, []string{"alpha"}) {
		t.Fatalf("Keys after deleting missing key = %v, want [alpha]", keys)
	}
}

func TestMapDeleteRemovesEntry(t *testing.T) {
	ordered := orderedmap.New[string, int]()
	ordered.Put("alpha", 1)
	ordered.Put("beta", 2)
	ordered.Put("gamma", 3)
	ordered.Delete("beta")
	if _, found := ordered.Get("beta"); found {
		t.Fatal("Get on deleted key returned found = true, want false")
	}
	if keys := ordered.Keys(); !slices.Equal(keys, []string{"alpha", "gamma"}) {
		t.Fatalf("Keys after Delete = %v, want [alpha gamma]", keys)
	}
	if values := ordered.Values(); !slices.Equal(values, []int{1, 3}) {
		t.Fatalf("Values after Delete = %v, want [1 3]", values)
	}
	if ordered.Len() != 2 {
		t.Fatalf("Len after Delete = %d, want 2", ordered.Len())
	}
}

func TestMapRePutAfterDeleteTakesBack(t *testing.T) {
	ordered := orderedmap.New[string, int]()
	ordered.Put("alpha", 1)
	ordered.Put("beta", 2)
	ordered.Delete("alpha")
	ordered.Put("alpha", 10)
	if keys := ordered.Keys(); !slices.Equal(keys, []string{"beta", "alpha"}) {
		t.Fatalf("Keys after re-Put = %v, want [beta alpha]", keys)
	}
	value, found := ordered.Get("alpha")
	if !found || value != 10 {
		t.Fatalf("Get(alpha) after re-Put = (%d, %v), want (10, true)", value, found)
	}
}

func TestMapKeysValuesFollowInsertionOrder(t *testing.T) {
	ordered := orderedmap.New[int, string]()
	keys := []int{3, 1, 2}
	for _, key := range keys {
		ordered.Put(key, string(rune('a'+key)))
	}
	if got := ordered.Keys(); !slices.Equal(got, keys) {
		t.Fatalf("Keys = %v, want %v", got, keys)
	}
	want := []string{"d", "b", "c"}
	if got := ordered.Values(); !slices.Equal(got, want) {
		t.Fatalf("Values = %v, want %v", got, want)
	}
}

func TestMapIsEmptyTransitions(t *testing.T) {
	ordered := orderedmap.New[string, int]()
	ordered.Put("alpha", 1)
	if ordered.IsEmpty() {
		t.Fatal("IsEmpty = true after Put, want false")
	}
	ordered.Delete("alpha")
	if !ordered.IsEmpty() {
		t.Fatal("IsEmpty = false after deleting the only key, want true")
	}
}

func TestNewLRURejectsNonPositiveCapacity(t *testing.T) {
	for _, capacity := range []int{0, -1, -100} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("NewLRU(%d) did not panic, want panic", capacity)
				}
			}()
			orderedmap.NewLRU[string, int](capacity)
		}()
	}
}

func TestLRUPutGetRoundTrip(t *testing.T) {
	recent := orderedmap.NewLRU[string, int](2)
	recent.Put("alpha", 1)
	value, found := recent.Get("alpha")
	if !found || value != 1 {
		t.Fatalf("Get(alpha) = (%d, %v), want (1, true)", value, found)
	}
	if recent.Len() != 1 {
		t.Fatalf("Len = %d, want 1", recent.Len())
	}
}

func TestLRUGetMissingYieldsOrNull(t *testing.T) {
	recent := orderedmap.NewLRU[string, int](2)
	recent.Put("alpha", 1)
	value, found := recent.Get("missing")
	if found {
		t.Fatal("Get on missing key returned found = true, want false")
	}
	if value != 0 {
		t.Fatalf("Get on missing key returned %d, want zero value", value)
	}
	if recent.Len() != 1 {
		t.Fatalf("Len after missing Get = %d, want 1", recent.Len())
	}
}

func TestLRUEvictsLeastRecentlyUsedBeyondCapacity(t *testing.T) {
	recent := orderedmap.NewLRU[string, int](2)
	recent.Put("alpha", 1)
	recent.Put("beta", 2)
	recent.Put("gamma", 3)
	if _, found := recent.Get("alpha"); found {
		t.Fatal("Get(alpha) after eviction returned found = true, want false")
	}
	for key, want := range map[string]int{"beta": 2, "gamma": 3} {
		value, found := recent.Get(key)
		if !found || value != want {
			t.Errorf("Get(%q) = (%d, %v), want (%d, true)", key, value, found, want)
		}
	}
	if recent.Len() != 2 {
		t.Fatalf("Len beyond capacity = %d, want 2", recent.Len())
	}
}

func TestLRUGetRefreshesRecency(t *testing.T) {
	recent := orderedmap.NewLRU[string, int](2)
	recent.Put("alpha", 1)
	recent.Put("beta", 2)
	if _, found := recent.Get("alpha"); !found {
		t.Fatal("Get(alpha) before eviction returned found = false, want true")
	}
	recent.Put("gamma", 3)
	if _, found := recent.Get("beta"); found {
		t.Fatal("Get(beta) after eviction returned found = true, want false")
	}
	if _, found := recent.Get("alpha"); !found {
		t.Fatal("Get(alpha) after eviction returned found = false, want true")
	}
}

func TestLRUPutRefreshesRecency(t *testing.T) {
	recent := orderedmap.NewLRU[string, int](2)
	recent.Put("alpha", 1)
	recent.Put("beta", 2)
	recent.Put("alpha", 10)
	value, found := recent.Get("alpha")
	if !found || value != 10 {
		t.Fatalf("Get(alpha) after update = (%d, %v), want (10, true)", value, found)
	}
	recent.Put("gamma", 3)
	if _, found := recent.Get("beta"); found {
		t.Fatal("Get(beta) after eviction returned found = true, want false")
	}
	if recent.Len() != 2 {
		t.Fatalf("Len = %d, want 2", recent.Len())
	}
}

func TestLRUCapacityOneKeepsOnlyNewest(t *testing.T) {
	recent := orderedmap.NewLRU[string, int](1)
	recent.Put("alpha", 1)
	recent.Put("beta", 2)
	if _, found := recent.Get("alpha"); found {
		t.Fatal("Get(alpha) after eviction returned found = true, want false")
	}
	value, found := recent.Get("beta")
	if !found || value != 2 {
		t.Fatalf("Get(beta) = (%d, %v), want (2, true)", value, found)
	}
}

func TestMapKeysValuesReturnFreshCopies(t *testing.T) {
	entries := orderedmap.New[string, int]()
	entries.Put("alpha", 1)
	entries.Put("beta", 2)
	keys := entries.Keys()
	keys[0] = "hacked"
	values := entries.Values()
	values[0] = 99
	if got := entries.Keys(); len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("Keys after mutation = %v, want [alpha beta]", got)
	}
	if got, found := entries.Get("alpha"); !found || got != 1 {
		t.Fatalf("Get(alpha) after Values mutation = (%d, %v), want (1, true)", got, found)
	}
}
