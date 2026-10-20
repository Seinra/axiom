package opt_test

import (
	"testing"

	"github.com/seinra/axiom/opt"
)

func TestSomeIsSome(t *testing.T) {
	current := opt.Some(7)
	if !current.IsSome() {
		t.Fatal("IsSome = false, want true")
	}
	if got := current.ValueOr(0); got != 7 {
		t.Fatalf("ValueOr = %d, want 7", got)
	}
}

func TestNoneIsNone(t *testing.T) {
	current := opt.None[int]()
	if current.IsSome() {
		t.Fatal("IsSome = true, want false")
	}
	if got := current.ValueOr(9); got != 9 {
		t.Fatalf("ValueOr = %d, want 9", got)
	}
}

func TestMapOptionTransformsSome(t *testing.T) {
	current := opt.Some(4)
	mapped := opt.MapOption(current, func(success int) string { return "four" })
	if !mapped.IsSome() {
		t.Fatal("IsSome = false, want true")
	}
	if got := mapped.ValueOr(""); got != "four" {
		t.Fatalf("ValueOr = %q, want %q", got, "four")
	}
}

func TestMapOptionKeepsNone(t *testing.T) {
	current := opt.None[int]()
	mapped := opt.MapOption(current, func(success int) string { return "unreachable" })
	if mapped.IsSome() {
		t.Fatal("IsSome = true, want false")
	}
}

func TestMethodMapValueChains(t *testing.T) {
	// Exercises the Go 1.27 generic method: MapValue declares its own type params.
	if got := opt.Some(5).MapValue(func(success int) int { return success + 1 }).ValueOr(0); got != 6 {
		t.Fatalf("chained value = %d, want 6", got)
	}
}

func TestZeroValueIsAbsent(t *testing.T) {
	var current opt.Option[int]
	if current.IsSome() {
		t.Fatal("IsSome = true, want false")
	}
	if got := current.ValueOr(3); got != 3 {
		t.Fatalf("ValueOr = %d, want 3", got)
	}
}
