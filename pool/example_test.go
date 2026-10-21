package pool_test

import (
	"fmt"

	"github.com/seinra/axiom/pool"
)

// ExampleNew shows the typed Get/Put reuse path.
func ExampleNew() {
	buffers := pool.New(func() []byte { return make([]byte, 0, 4) })
	buffer := append(buffers.Get(), "ok"...)
	fmt.Println(string(buffer))
	buffers.Put(buffer[:0])
	fmt.Println(len(append(buffers.Get(), 'x')) > 0)
	// Output:
	// ok
	// true
}
