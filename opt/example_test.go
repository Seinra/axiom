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
