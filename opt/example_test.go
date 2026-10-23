package opt_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/seinra/axiom/opt"
)

// ExampleSome shows the present-value path: hold, transform, read with fallback.
func ExampleSome() {
	present := opt.Some("gopher")
	fmt.Println(present.ValueOr("nobody"))
	fmt.Println(opt.None[string]().ValueOr("nobody"))
	shouted := present.MapValue(strings.ToUpper)
	fmt.Println(shouted.ValueOr("nobody"))
	// Output:
	// gopher
	// nobody
	// GOPHER
}

// ExampleTo shows the fresh-pointer wrap: live, independent, mutable alone.
func ExampleTo() {
	value := 7
	ptr := opt.To(value)
	*ptr = 99
	fmt.Println(*ptr, value)
	// Output:
	// 99 7
}

// ExampleDeref shows the nil-safe read: value plus presence.
func ExampleDeref() {
	got, found := opt.Deref(opt.To("gopher"))
	fmt.Println(got, found)
	var missing *string
	zero, ok := opt.Deref(missing)
	fmt.Println(zero == "", ok)
	// Output:
	// gopher true
	// true false
}

// ExampleCoalesce shows the first live pointer winning, nil when none is.
func ExampleCoalesce() {
	fmt.Println(*opt.Coalesce(nil, opt.To(1), opt.To(2)))
	fmt.Println(opt.Coalesce[int](nil, nil) == nil)
	// Output:
	// 1
	// true
}

// ExampleOption_MarshalJSON shows the null codec: Some renders as its value, None as null.
func ExampleOption_MarshalJSON() {
	payload, _ := json.Marshal(opt.Some("gopher"))
	fmt.Println(string(payload))
	absent, _ := json.Marshal(opt.None[string]())
	fmt.Println(string(absent))
	var restored opt.Option[string]
	_ = json.Unmarshal(payload, &restored)
	fmt.Println(restored.ValueOr("nobody"))
	// Output:
	// "gopher"
	// null
	// gopher
}
