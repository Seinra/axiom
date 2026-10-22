package mapx_test

import (
	"fmt"

	"github.com/seinra/axiom/mapx"
	"github.com/seinra/axiom/slice"
)

// ExampleKeyBy shows the index-and-read path: key a slice, read with fallback.
func ExampleKeyBy() {
	lessons := []lesson{{slug: "a", level: 1}, {slug: "b", level: 2}}
	bySlug := mapx.KeyBy(lessons, func(entry lesson) string { return entry.slug })
	fmt.Println(mapx.ValueOr(bySlug, "a", lesson{}).level)
	fmt.Println(mapx.ValueOr(bySlug, "missing", lesson{level: -1}).level)
	titles := mapx.MapValues(bySlug, func(entry lesson) int { return entry.level * 10 })
	fmt.Println(mapx.ValueOr(titles, "b", -1))
	// Output:
	// 1
	// -1
	// 20
}

// ExampleFromEntries shows the pairs bridge: Zip feeds FromEntries, and values select back.
func ExampleFromEntries() {
	entries := mapx.FromEntries(slice.Zip([]string{"a", "b"}, []int{1, 2}))
	fmt.Println(entries)
	fmt.Println(mapx.PickByValues(entries, 2))
	// Output:
	// map[a:1 b:2]
	// map[b:2]
}
