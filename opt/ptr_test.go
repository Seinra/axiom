package opt_test

import (
	"testing"

	"github.com/seinra/axiom/opt"
)

func TestToWrapsFreshPointer(t *testing.T) {
	ptr := opt.To(7)
	if ptr == nil {
		t.Fatal("To(7) = nil, want live pointer")
	}
	if *ptr != 7 {
		t.Fatalf("To(7) points to %d, want 7", *ptr)
	}
}

func TestToPointerIsIndependent(t *testing.T) {
	value := 7
	ptr := opt.To(value)
	*ptr = 99
	if value != 7 {
		t.Fatalf("To aliases input: value = %d, want 7", value)
	}
}

func TestDerefReadsLivePointer(t *testing.T) {
	got, found := opt.Deref(opt.To("gopher"))
	if !found {
		t.Fatal("Deref found = false, want true")
	}
	if got != "gopher" {
		t.Fatalf("Deref = %q, want %q", got, "gopher")
	}
}

func TestDerefNilYieldsZeroFalse(t *testing.T) {
	var ptr *string
	got, found := opt.Deref(ptr)
	if found {
		t.Fatal("Deref(nil) found = true, want false")
	}
	if got != "" {
		t.Fatalf("Deref(nil) = %q, want zero value", got)
	}
}

func TestCoalescePicksFirstLive(t *testing.T) {
	first := opt.To(1)
	second := opt.To(2)
	if got := opt.Coalesce(nil, first, second); got != first {
		t.Fatalf("Coalesce = %p, want first %p", got, first)
	}
}

func TestCoalesceAllNilYieldsNil(t *testing.T) {
	if got := opt.Coalesce[int](nil, nil); got != nil {
		t.Fatalf("Coalesce(nil, nil) = %p, want nil", got)
	}
}

func TestCoalesceEmptyYieldsNil(t *testing.T) {
	if got := opt.Coalesce[int](); got != nil {
		t.Fatalf("Coalesce() = %p, want nil", got)
	}
}

func TestCoalesceAliasesWinner(t *testing.T) {
	winner := opt.To(3)
	got := opt.Coalesce(nil, winner)
	*got = 30
	if *winner != 30 {
		t.Fatalf("Coalesce does not alias winner: winner = %d, want 30", *winner)
	}
}
