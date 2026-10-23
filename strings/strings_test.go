package strings_test

import (
	"reflect"
	"testing"

	"github.com/seinra/axiom/strings"
)

func TestWordsSplitsOnWhitespace(t *testing.T) {
	got := strings.Words("  go   rust\tzig\n")
	want := []string{"go", "rust", "zig"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Words = %q, want %q", got, want)
	}
}

func TestWordsEmptyYieldsEmptyNonNil(t *testing.T) {
	got := strings.Words("")
	if got == nil {
		t.Fatal("Words(\"\") = nil, want empty non-nil slice")
	}
	if len(got) != 0 {
		t.Fatalf("Words(\"\") len = %d, want 0", len(got))
	}
}

func TestWordsBlankYieldsEmptyNonNil(t *testing.T) {
	got := strings.Words("  \t\n ")
	if got == nil || len(got) != 0 {
		t.Fatalf("Words(blank) = %q, want empty non-nil slice", got)
	}
}

func TestWordsKeepsMultibyteWords(t *testing.T) {
	got := strings.Words("héllo 世界")
	want := []string{"héllo", "世界"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Words = %q, want %q", got, want)
	}
}

func TestChunkStringSplitsByRunes(t *testing.T) {
	got := strings.ChunkString("héllo", 2)
	want := []string{"hé", "ll", "o"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ChunkString = %q, want %q", got, want)
	}
}

func TestChunkStringMultibyteNeverSplitsRune(t *testing.T) {
	got := strings.ChunkString("世界世", 2)
	want := []string{"世界", "世"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ChunkString = %q, want %q", got, want)
	}
	for _, chunk := range got {
		if len([]rune(chunk)) > 2 {
			t.Fatalf("chunk %q exceeds 2 runes", chunk)
		}
	}
}

func TestChunkStringEmptyYieldsEmptyNonNil(t *testing.T) {
	got := strings.ChunkString("", 2)
	if got == nil || len(got) != 0 {
		t.Fatalf("ChunkString(\"\") = %q, want empty non-nil slice", got)
	}
}

func TestChunkStringBeyondLengthCopiesAll(t *testing.T) {
	got := strings.ChunkString("go", 9)
	want := []string{"go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ChunkString = %q, want %q", got, want)
	}
}

func TestChunkStringSizeBelowOnePanics(t *testing.T) {
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_ = strings.ChunkString("go", 0)
		return nil
	}()
	if panicked == nil {
		t.Fatal("ChunkString(0) did not panic, want panic")
	}
}

func TestTruncateShortStaysWhole(t *testing.T) {
	if got := strings.Truncate("go", 9); got != "go" {
		t.Fatalf("Truncate = %q, want %q", got, "go")
	}
}

func TestTruncateCutsByRunesNotBytes(t *testing.T) {
	if got := strings.Truncate("héllo", 2); got != "hé" {
		t.Fatalf("Truncate = %q, want %q", got, "hé")
	}
}

func TestTruncateEmptyAndZero(t *testing.T) {
	if got := strings.Truncate("", 3); got != "" {
		t.Fatalf("Truncate(\"\") = %q, want empty", got)
	}
	if got := strings.Truncate("go", 0); got != "" {
		t.Fatalf("Truncate(0) = %q, want empty", got)
	}
}

func TestTruncateNegativePanics(t *testing.T) {
	panicked := func() (didPanic any) {
		defer func() { didPanic = recover() }()
		_ = strings.Truncate("go", -1)
		return nil
	}()
	if panicked == nil {
		t.Fatal("Truncate(-1) did not panic, want panic")
	}
}
