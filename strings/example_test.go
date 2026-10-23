package strings_test

import (
	"fmt"

	axiomstrings "github.com/seinra/axiom/strings"
)

// ExampleWords shows whitespace splitting with multibyte words kept whole.
func ExampleWords() {
	fmt.Println(axiomstrings.Words("  go   rust\tzig\n"))
	fmt.Println(axiomstrings.Words("héllo 世界"))
	fmt.Println(axiomstrings.Words(""))
	// Output:
	// [go rust zig]
	// [héllo 世界]
	// []
}

// ExampleChunkString shows rune-boundary chunks that never tear a character.
func ExampleChunkString() {
	fmt.Println(axiomstrings.ChunkString("héllo", 2))
	fmt.Println(axiomstrings.ChunkString("世界世", 2))
	// Output:
	// [hé ll o]
	// [世界 世]
}

// ExampleTruncate shows the pure rune cut with no suffix added.
func ExampleTruncate() {
	fmt.Println(axiomstrings.Truncate("héllo", 2))
	fmt.Println(axiomstrings.Truncate("go", 9))
	// Output:
	// hé
	// go
}
