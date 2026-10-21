package di_test

import (
	"fmt"

	"github.com/seinra/axiom/di"
)

// ExampleResolve shows the Provide/Resolve lazy-singleton path.
func ExampleResolve() {
	injector := di.New()
	_ = di.Provide(injector, "greeting", func(*di.Injector) (string, error) { return "hello", nil })
	greeting, _ := di.Resolve[string](injector, "greeting")
	fmt.Println(greeting)
	// Output: hello
}
