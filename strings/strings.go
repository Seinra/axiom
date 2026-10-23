// Package strings provides small rune-aware helpers over string values.
//
// Byte slicing breaks multibyte text: every function here counts runes, never
// bytes. When you need both this package and the standard library in one
// file, alias the import:
//
//	import axiomstrings "github.com/seinra/axiom/strings"
//
// Case converters live in x/text and stay out of this package on purpose.
//
// Which function?
//
// | When you want to ...               | Reach for ... |
// |-----------------------------------|---------------|
// | Split on whitespace into words     | Words         |
// | Split into fixed-rune chunks       | ChunkString   |
// | Cut to a rune budget, no suffix   | Truncate      |
package strings

import (
	"fmt"
	"unicode"
)

// Words splits text on Unicode whitespace into fields.
//
// Blank or empty input yields an empty non-nil slice. Words keep their
// multibyte runes intact: the scan walks runes, so no byte boundary can
// split a character. strings.Fields covers ASCII-centered splitting; Words
// pins the Axiom empty-non-nil contract and documents multibyte behavior.
func Words(text string) []string {
	words := make([]string, 0)
	start := -1
	for index, char := range text {
		if unicode.IsSpace(char) {
			if start >= 0 {
				words = append(words, text[start:index])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = index
		}
	}
	if start >= 0 {
		words = append(words, text[start:])
	}
	return words
}

// ChunkString splits text into chunks of at most size runes.
//
// Size below one panics: a non-positive chunk size is a programmer error,
// matching the size <= 0 panic policy of Chunked and Take. Empty input
// yields an empty non-nil slice; size at or beyond the rune length yields a
// single copy. Byte slicing would tear multibyte runes apart, so chunks are
// cut on rune boundaries. There is no rune-chunk helper in the standard
// library.
func ChunkString(text string, size int) []string {
	if size <= 0 {
		panic(fmt.Sprintf("axiom strings: ChunkString needs size >= 1, got %d", size))
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return make([]string, 0)
	}
	chunked := make([]string, 0, (len(runes)+size-1)/size)
	for start := 0; start < len(runes); start += size {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunked = append(chunked, string(runes[start:end]))
	}
	return chunked
}

// Truncate cuts text to at most maxRunes runes, with no suffix added.
//
// Text within budget returns unchanged. A negative budget panics: it is a
// programmer error, matching the size-below-zero panic policy. Byte cuts
// would tear multibyte runes apart, so the cut lands on a rune boundary.
// Ellipsis-decorating variants stay out on purpose: presentation belongs to
// the caller. The standard library has no rune truncate.
func Truncate(text string, maxRunes int) string {
	if maxRunes < 0 {
		panic(fmt.Sprintf("axiom strings: Truncate needs maxRunes >= 0, got %d", maxRunes))
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes])
}
