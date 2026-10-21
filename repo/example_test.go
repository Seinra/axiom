package repo_test

import (
	"fmt"

	"github.com/seinra/axiom/repo"
)

// ExampleNewMemoryRepository shows the Store/Fetch round trip.
func ExampleNewMemoryRepository() {
	lessonStore := repo.NewMemoryRepository(lessonKeyOf)
	_ = lessonStore.Store(lesson{Slug: "lesson-a", Title: "Lesson A"})
	fetched, _ := lessonStore.Fetch("lesson-a")
	fmt.Println(fetched.Title)
	// Output: Lesson A
}
