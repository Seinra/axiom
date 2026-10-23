package mathx_test

import (
	"fmt"

	"github.com/seinra/axiom/mathx"
)

// ExampleIsPrime shows the prime test over small and edge inputs.
func ExampleIsPrime() {
	fmt.Println(mathx.IsPrime(7))
	fmt.Println(mathx.IsPrime(9))
	fmt.Println(mathx.IsPrime(1))
	// Output:
	// true
	// false
	// false
}

// ExampleGcd shows Euclid's algorithm, including the zero convention.
func ExampleGcd() {
	fmt.Println(mathx.Gcd(12, 18))
	fmt.Println(mathx.Gcd(0, 5))
	fmt.Println(mathx.Gcd(-12, 18))
	// Output:
	// 6
	// 5
	// 6
}

// ExampleLcm shows the common multiple, zero, and overflow-safe ordering.
func ExampleLcm() {
	fmt.Println(mathx.Lcm(4, 6))
	fmt.Println(mathx.Lcm(0, 5))
	fmt.Println(mathx.Lcm(1<<32, 3<<32))
	// Output:
	// 12
	// 0
	// 12884901888
}
