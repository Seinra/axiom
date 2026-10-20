package mapx_test

import (
	"reflect"
	"testing"

	"github.com/seinra/axiom/mapx"
)

type lesson struct {
	slug  string
	level int
}

func TestKeyByIndexesElements(t *testing.T) {
	indexed := mapx.KeyBy(
		[]lesson{{slug: "go-basics", level: 1}, {slug: "go-maps", level: 2}},
		func(candidate lesson) string { return candidate.slug },
	)
	want := map[string]lesson{"go-basics": {slug: "go-basics", level: 1}, "go-maps": {slug: "go-maps", level: 2}}
	if !reflect.DeepEqual(indexed, want) {
		t.Fatalf("KeyBy = %v, want %v", indexed, want)
	}
}

func TestKeyByLastWinsCollision(t *testing.T) {
	indexed := mapx.KeyBy(
		[]lesson{{slug: "dup", level: 1}, {slug: "dup", level: 2}},
		func(candidate lesson) string { return candidate.slug },
	)
	if got := indexed["dup"].level; got != 2 {
		t.Fatalf("collision level = %d, want 2 (last wins)", got)
	}
}

func TestFilterKeepsMatching(t *testing.T) {
	kept := mapx.Filter(
		map[string]int{"a": 1, "b": 2, "c": 3},
		func(key string, score int) bool { return score%2 == 1 },
	)
	want := map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(kept, want) {
		t.Fatalf("Filter = %v, want %v", kept, want)
	}
}

func TestFilterNilYieldsEmpty(t *testing.T) {
	var entries map[string]int
	kept := mapx.Filter(entries, func(key string, score int) bool { return true })
	if kept == nil {
		t.Fatal("Filter(nil) = nil, want empty non-nil map")
	}
	if len(kept) != 0 {
		t.Fatalf("len = %d, want 0", len(kept))
	}
}

func TestMapValuesTransforms(t *testing.T) {
	graded := mapx.MapValues(
		map[string]int{"ana": 85, "bea": 45},
		func(score int) string {
			if score >= 60 {
				return "approved"
			}
			return "retry"
		},
	)
	want := map[string]string{"ana": "approved", "bea": "retry"}
	if !reflect.DeepEqual(graded, want) {
		t.Fatalf("MapValues = %v, want %v", graded, want)
	}
}

func TestMapValuesNilYieldsEmpty(t *testing.T) {
	var entries map[string]int
	mapped := mapx.MapValues(entries, func(score int) string { return "unreachable" })
	if mapped == nil {
		t.Fatal("MapValues(nil) = nil, want empty non-nil map")
	}
}

func TestMergeResolvesCollisions(t *testing.T) {
	// Tally-style counts: collisions sum instead of overwriting.
	combined := mapx.Merge(
		map[string]int{"a": 1, "b": 2},
		map[string]int{"b": 3, "c": 4},
		func(key string, first int, second int) int { return first + second },
	)
	want := map[string]int{"a": 1, "b": 5, "c": 4}
	if !reflect.DeepEqual(combined, want) {
		t.Fatalf("Merge = %v, want %v", combined, want)
	}
}

func TestMergeNeverMutatesInputs(t *testing.T) {
	first := map[string]int{"a": 1}
	second := map[string]int{"a": 2}
	_ = mapx.Merge(first, second, func(key string, oldScore int, newScore int) int { return newScore })
	if first["a"] != 1 || second["a"] != 2 {
		t.Fatalf("inputs mutated: first = %v, second = %v", first, second)
	}
}

func TestPickSelectsKeys(t *testing.T) {
	picked := mapx.Pick(map[string]int{"a": 1, "b": 2, "c": 3}, "c", "a", "missing")
	want := map[string]int{"a": 1, "c": 3}
	if !reflect.DeepEqual(picked, want) {
		t.Fatalf("Pick = %v, want %v", picked, want)
	}
}

func TestPickNilYieldsEmpty(t *testing.T) {
	var entries map[string]int
	picked := mapx.Pick(entries, "a")
	if picked == nil {
		t.Fatal("Pick(nil) = nil, want empty non-nil map")
	}
}

func TestKeyByNilYieldsEmpty(t *testing.T) {
	var courses []lesson
	indexed := mapx.KeyBy(courses, func(candidate lesson) string { return candidate.slug })
	if indexed == nil {
		t.Fatal("KeyBy(nil) = nil, want empty non-nil map")
	}
	if len(indexed) != 0 {
		t.Fatalf("len = %d, want 0", len(indexed))
	}
}

func TestMergeNilNilYieldsEmpty(t *testing.T) {
	var first, second map[string]int
	combined := mapx.Merge(first, second, func(key string, oldScore int, newScore int) int { return oldScore + newScore })
	if combined == nil {
		t.Fatal("Merge(nil, nil) = nil, want empty non-nil map")
	}
}

func TestOmitExcludesKeys(t *testing.T) {
	omitted := mapx.Omit(map[string]int{"a": 1, "b": 2, "c": 3}, "a", "c")
	want := map[string]int{"b": 2}
	if !reflect.DeepEqual(omitted, want) {
		t.Fatalf("Omit = %v, want %v", omitted, want)
	}
}

func TestOmitMissingKeysIgnored(t *testing.T) {
	omitted := mapx.Omit(map[string]int{"a": 1, "b": 2}, "missing", "a")
	want := map[string]int{"b": 2}
	if !reflect.DeepEqual(omitted, want) {
		t.Fatalf("Omit = %v, want %v", omitted, want)
	}
}

func TestOmitNilYieldsEmpty(t *testing.T) {
	var entries map[string]int
	omitted := mapx.Omit(entries, "a")
	if omitted == nil {
		t.Fatal("Omit(nil) = nil, want empty non-nil map")
	}
	if len(omitted) != 0 {
		t.Fatalf("len = %d, want 0", len(omitted))
	}
}

func TestOmitNeverMutatesInputs(t *testing.T) {
	entries := map[string]int{"a": 1, "b": 2}
	_ = mapx.Omit(entries, "a")
	want := map[string]int{"a": 1, "b": 2}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("inputs mutated: entries = %v, want %v", entries, want)
	}
}

func TestInvertGroupsCollidingValues(t *testing.T) {
	grouped := mapx.Invert(map[string]string{"a": "x", "b": "y", "c": "x"})
	if len(grouped) != 2 {
		t.Fatalf("len(Invert) = %d, want 2", len(grouped))
	}
	for _, key := range grouped["x"] {
		if key != "a" && key != "c" {
			t.Fatalf("Invert[x] holds %q, want a or c", key)
		}
	}
	if len(grouped["x"]) != 2 {
		t.Fatalf("len(Invert[x]) = %d, want 2", len(grouped["x"]))
	}
	if !reflect.DeepEqual(grouped["y"], []string{"b"}) {
		t.Fatalf("Invert[y] = %v, want [b]", grouped["y"])
	}
}

func TestInvertNilYieldsEmpty(t *testing.T) {
	var entries map[string]string
	grouped := mapx.Invert(entries)
	if grouped == nil {
		t.Fatal("Invert(nil) = nil, want empty non-nil map")
	}
	if len(grouped) != 0 {
		t.Fatalf("len = %d, want 0", len(grouped))
	}
}

func TestValueOrReturnsHeldOrFallback(t *testing.T) {
	entries := map[string]int{"a": 1}
	if got := mapx.ValueOr(entries, "a", 99); got != 1 {
		t.Fatalf("ValueOr(held) = %d, want 1", got)
	}
	if got := mapx.ValueOr(entries, "missing", 99); got != 99 {
		t.Fatalf("ValueOr(missing) = %d, want 99", got)
	}
}

func TestValueOrNilYieldsFallback(t *testing.T) {
	var entries map[string]int
	if got := mapx.ValueOr(entries, "a", 7); got != 7 {
		t.Fatalf("ValueOr(nil, a) = %d, want 7", got)
	}
}

func TestGroupByMappedGroupsValues(t *testing.T) {
	courses := []lesson{{slug: "go-basics", level: 1}, {slug: "go-maps", level: 1}, {slug: "rust-adv", level: 2}}
	grouped := mapx.GroupByMapped(
		courses,
		func(candidate lesson) int { return candidate.level },
		func(candidate lesson) string { return candidate.slug },
	)
	if len(grouped[1]) != 2 || len(grouped[2]) != 1 {
		t.Fatalf("GroupByMapped = %v, want 2 in level 1 and 1 in level 2", grouped)
	}
}

func TestGroupByMappedNilYieldsEmpty(t *testing.T) {
	var courses []lesson
	grouped := mapx.GroupByMapped(
		courses,
		func(candidate lesson) int { return candidate.level },
		func(candidate lesson) string { return candidate.slug },
	)
	if grouped == nil {
		t.Fatal("GroupByMapped(nil) = nil, want empty non-nil map")
	}
	if len(grouped) != 0 {
		t.Fatalf("len = %d, want 0", len(grouped))
	}
}
