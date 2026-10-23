package mathx_test

import (
	"testing"

	"github.com/seinra/axiom/mathx"
)

func TestIsPrimeSmallPrimes(t *testing.T) {
	for _, prime := range []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31} {
		if !mathx.IsPrime(prime) {
			t.Fatalf("IsPrime(%d) = false, want true", prime)
		}
	}
}

func TestIsPrimeComposites(t *testing.T) {
	for _, composite := range []int{4, 6, 8, 9, 10, 12, 15, 21, 25, 27, 49, 100} {
		if mathx.IsPrime(composite) {
			t.Fatalf("IsPrime(%d) = true, want false", composite)
		}
	}
}

func TestIsPrimeBelowTwo(t *testing.T) {
	for _, n := range []int{-17, -2, -1, 0, 1} {
		if mathx.IsPrime(n) {
			t.Fatalf("IsPrime(%d) = true, want false", n)
		}
	}
}

func TestIsPrimeLargePrime(t *testing.T) {
	if !mathx.IsPrime(104729) {
		t.Fatal("IsPrime(104729) = false, want true")
	}
}

func TestIsPrimeLargeComposite(t *testing.T) {
	if mathx.IsPrime(104730) {
		t.Fatal("IsPrime(104730) = true, want false")
	}
}

func TestGcdBasic(t *testing.T) {
	if got := mathx.Gcd(12, 18); got != 6 {
		t.Fatalf("Gcd(12, 18) = %d, want 6", got)
	}
}

func TestGcdCoprime(t *testing.T) {
	if got := mathx.Gcd(7, 13); got != 1 {
		t.Fatalf("Gcd(7, 13) = %d, want 1", got)
	}
}

func TestGcdZero(t *testing.T) {
	if got := mathx.Gcd(0, 5); got != 5 {
		t.Fatalf("Gcd(0, 5) = %d, want 5", got)
	}
	if got := mathx.Gcd(0, 0); got != 0 {
		t.Fatalf("Gcd(0, 0) = %d, want 0", got)
	}
}

func TestGcdNegativesYieldNonNegative(t *testing.T) {
	if got := mathx.Gcd(-12, 18); got != 6 {
		t.Fatalf("Gcd(-12, 18) = %d, want 6", got)
	}
	if got := mathx.Gcd(-12, -18); got != 6 {
		t.Fatalf("Gcd(-12, -18) = %d, want 6", got)
	}
}

func TestLcmBasic(t *testing.T) {
	if got := mathx.Lcm(4, 6); got != 12 {
		t.Fatalf("Lcm(4, 6) = %d, want 12", got)
	}
}

func TestLcmZero(t *testing.T) {
	if got := mathx.Lcm(0, 5); got != 0 {
		t.Fatalf("Lcm(0, 5) = %d, want 0", got)
	}
}

func TestLcmNegativesYieldNonNegative(t *testing.T) {
	if got := mathx.Lcm(-4, 6); got != 12 {
		t.Fatalf("Lcm(-4, 6) = %d, want 12", got)
	}
}

func TestLcmDivideBeforeMultiplyAvoidsOverflow(t *testing.T) {
	first, second := 1<<32, 3<<32
	if naive := first * second; naive == 3<<32 {
		t.Fatalf("test premise broken: %d * %d did not overflow", first, second)
	}
	if got := mathx.Lcm(first, second); got != 3<<32 {
		t.Fatalf("Lcm(2^32, 3*2^32) = %d, want %d", got, 3<<32)
	}
}
