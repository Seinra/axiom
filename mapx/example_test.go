package mapx_test

import (
	"fmt"

	"github.com/seinra/axiom/mapx"
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
