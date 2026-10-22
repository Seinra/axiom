package orderedmap_test

import (
	"fmt"

	"github.com/seinra/axiom/orderedmap"
)

// ExampleNew shows the ordered-map lesson: unlike the builtin map, Keys
// replays first-insertion order, and an upsert keeps the key's original
// seat while replacing its value.
func ExampleNew() {
	ordered := orderedmap.New[string, int]()
	ordered.Put("beta", 2)
	ordered.Put("alpha", 1)
	ordered.Put("beta", 20)
	fmt.Println(ordered.Keys())
	fmt.Println(ordered.Values())
	ordered.Delete("beta")
	ordered.Put("beta", 30)
	fmt.Println(ordered.Keys())
	// Output:
	// [beta alpha]
	// [20 1]
	// [alpha beta]
}

// ExampleNewLRU shows the caching lesson: the cache holds at most two
// entries, a Get hit refreshes recency, and the next Put evicts the entry
// left untouched the longest.
func ExampleNewLRU() {
	recent := orderedmap.NewLRU[string, int](2)
	recent.Put("alpha", 1)
	recent.Put("beta", 2)
	_, _ = recent.Get("alpha")
	recent.Put("gamma", 3)
	_, betaHeld := recent.Get("beta")
	alphaValue, _ := recent.Get("alpha")
	fmt.Println(betaHeld)
	fmt.Println(alphaValue)
	fmt.Println(recent.Len())
	// Output:
	// false
	// 1
	// 2
}
