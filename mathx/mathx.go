// Package mathx provides small scalar integer helpers with no stdlib home.
//
// Reach for these when the operation is pure number theory over an integer
// rather than a slice traversal (for those, see the slice package).
// math and slices offer no prime test, so IsPrime lives here.
package mathx

// IsPrime reports whether n is a prime number.
//
// Values below 2 (including negatives, zero, and one) yield false: there is
// no divisor story to tell there. Two yields true; every other even yields
// false without looping. Trial division then checks odd divisors up to the
// square root, written as d <= n/d so the bound itself cannot overflow.
// math has primality only via big.Int.ProbablyPrime; slices has no scalar
// number theory at all.
func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for divisor := 3; divisor <= n/divisor; divisor += 2 {
		if n%divisor == 0 {
			return false
		}
	}
	return true
}

// absInt returns the non-negative magnitude of value.
func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// Gcd returns the greatest common divisor of a and b, always non-negative.
//
// Gcd(0, 0) yields 0 by convention; Gcd(0, x) yields |x|. Euclid's algorithm
// drives the remainders strictly down, so it always terminates: the oldest
// nontrivial algorithm still in daily use. math/big offers GCD only for
// big integers, and there is no scalar version in the standard library.
func Gcd(a, b int) int {
	a, b = absInt(a), absInt(b)
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// Lcm returns the least common multiple of a and b, always non-negative.
//
// Either input zero yields zero. The Gcd factors out first and the division
// runs before the multiplication, so pairs whose naive product overflows
// still fit when the true multiple does — the same divide-before-multiply
// discipline as the overflow-safe arithmetic elsewhere in this module family.
// There is no scalar version in the standard library.
func Lcm(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	divisor := Gcd(a, b)
	return absInt(a / divisor * b)
}
