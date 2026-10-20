package repogorm_test

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/seinra/axiom/repo"
	"github.com/seinra/axiom/repogorm"
)

// ExampleNewStore shows the migrate-store-fetch round trip.
func ExampleNewStore() {
	database, _ := gorm.Open(sqlite.Open("file:example?mode=memory&cache=shared"), &gorm.Config{})
	lessonStore, _ := repogorm.NewStore[Lesson, uint](database, []any{Lesson{}}, "id")
	_ = lessonStore.Store(Lesson{ID: 1, Slug: "lesson-a", Title: "Lesson A"})
	fetched, _ := lessonStore.Fetch(uint(1))
	fmt.Println(fetched.Slug)
	listed, _ := lessonStore.List(repo.Page{Number: 1, Size: 10})
	fmt.Println(len(listed))
	// Output:
	// lesson-a
	// 1
}
