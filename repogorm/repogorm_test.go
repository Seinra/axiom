// Package repogorm_test verifies the GORM adapter honors the repo contract.
package repogorm_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/seinra/axiom/repo"
	"github.com/seinra/axiom/repogorm"
)

// Lesson is the GORM fixture behind every scenario: an explicit numeric
// primary key plus Catalyst-flavored lesson fields.
type Lesson struct {
	ID    uint `gorm:"primaryKey"`
	Slug  string
	Title string
}

// Store must satisfy repo.Repository at compile time.
var _ repo.Repository[Lesson, uint] = (*repogorm.Store[Lesson, uint])(nil)

// openTestDatabase opens an isolated in-memory SQLite database for one test.
func openTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	// Shared-cache memory DSN so the test's single connection pool sees one database.
	sanitized := strings.ReplaceAll(t.Name(), "/", "_")
	source := fmt.Sprintf("file:%s?mode=memory&cache=shared", sanitized)
	database, openFailure := gorm.Open(sqlite.Open(source), &gorm.Config{})
	if openFailure != nil {
		t.Fatalf("open in-memory database: unexpected failure: %v", openFailure)
	}
	return database
}

// openLessonStore builds a migrated Lesson store seeded with candidates,
// failing fast on setup problems.
func openLessonStore(t *testing.T, candidates ...Lesson) *repogorm.Store[Lesson, uint] {
	t.Helper()
	lessonStore, buildFailure := repogorm.NewStore[Lesson, uint](openTestDatabase(t), []any{Lesson{}}, "id")
	if buildFailure != nil {
		t.Fatalf("NewStore: unexpected failure: %v", buildFailure)
	}
	for _, candidate := range candidates {
		if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
			t.Fatalf("seed Store(%q): unexpected failure: %v", candidate.Slug, storeFailure)
		}
	}
	return lessonStore
}

// lessonIDs extracts primary keys in listed order for exact page assertions.
func lessonIDs(candidates []Lesson) []uint {
	identities := make([]uint, 0, len(candidates))
	for _, candidate := range candidates {
		identities = append(identities, candidate.ID)
	}
	return identities
}

// seedFiveLessons returns five lessons with stable primary keys 1-5.
func seedFiveLessons() []Lesson {
	return []Lesson{
		{ID: 1, Slug: "lesson-a", Title: "Lesson A"},
		{ID: 2, Slug: "lesson-b", Title: "Lesson B"},
		{ID: 3, Slug: "lesson-c", Title: "Lesson C"},
		{ID: 4, Slug: "lesson-d", Title: "Lesson D"},
		{ID: 5, Slug: "lesson-e", Title: "Lesson E"},
	}
}

func TestStoreFetchRoundTrip(t *testing.T) {
	lessonStore := openLessonStore(t)
	candidate := Lesson{ID: 1, Slug: "generics-intro", Title: "Intro to Generics"}
	if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
		t.Fatalf("Store: unexpected failure: %v", storeFailure)
	}
	fetched, fetchFailure := lessonStore.Fetch(1)
	if fetchFailure != nil {
		t.Fatalf("Fetch: unexpected failure: %v", fetchFailure)
	}
	if fetched != candidate {
		t.Fatalf("Fetch = %+v, want %+v", fetched, candidate)
	}
}

func TestFetchMissingReportsNotFound(t *testing.T) {
	lessonStore := openLessonStore(t)
	if _, fetchFailure := lessonStore.Fetch(404); !errors.Is(fetchFailure, repo.ErrRecordNotFound) {
		t.Fatalf("Fetch missing: failure = %v, want ErrRecordNotFound", fetchFailure)
	}
}

func TestRemoveMissingReportsNotFound(t *testing.T) {
	lessonStore := openLessonStore(t)
	if removeFailure := lessonStore.Remove(404); !errors.Is(removeFailure, repo.ErrRecordNotFound) {
		t.Fatalf("Remove missing: failure = %v, want ErrRecordNotFound", removeFailure)
	}
}

func TestRemoveDeletesStoredLesson(t *testing.T) {
	lessonStore := openLessonStore(t, Lesson{ID: 7, Slug: "retired-lesson", Title: "Retired"})
	if removeFailure := lessonStore.Remove(7); removeFailure != nil {
		t.Fatalf("Remove: unexpected failure: %v", removeFailure)
	}
	if _, fetchFailure := lessonStore.Fetch(7); !errors.Is(fetchFailure, repo.ErrRecordNotFound) {
		t.Fatalf("Fetch after Remove: failure = %v, want ErrRecordNotFound", fetchFailure)
	}
}

func TestStoreOverwritesSameIdentity(t *testing.T) {
	lessonStore := openLessonStore(t, Lesson{ID: 3, Slug: "upsert-me", Title: "First"})
	if storeFailure := lessonStore.Store(Lesson{ID: 3, Slug: "upsert-me", Title: "Second"}); storeFailure != nil {
		t.Fatalf("second Store: unexpected failure: %v", storeFailure)
	}
	fetched, fetchFailure := lessonStore.Fetch(3)
	if fetchFailure != nil {
		t.Fatalf("Fetch: unexpected failure: %v", fetchFailure)
	}
	if fetched.Title != "Second" {
		t.Fatalf("Fetch Title = %q, want %q (Store upserts)", fetched.Title, "Second")
	}
	listed, listFailure := lessonStore.List(repo.Page{Number: 1, Size: 10})
	if listFailure != nil {
		t.Fatalf("List: unexpected failure: %v", listFailure)
	}
	if len(listed) != 1 {
		t.Fatalf("List len = %d, want 1 (upsert keeps one record per identity)", len(listed))
	}
}

func TestListPagesAreDeterministic(t *testing.T) {
	lessonStore := openLessonStore(t, seedFiveLessons()...)
	wantPages := [][]uint{{1, 2}, {3, 4}, {5}}
	// List orders by the explicit order column, so every call repeats the
	// same slices: assert exact slices twice, not just membership.
	for round := range 2 {
		for pageIndex, wantIdentities := range wantPages {
			listed, listFailure := lessonStore.List(repo.Page{Number: pageIndex + 1, Size: 2})
			if listFailure != nil {
				t.Fatalf("round %d List page %d: unexpected failure: %v", round, pageIndex+1, listFailure)
			}
			if gotIdentities := lessonIDs(listed); !slices.Equal(gotIdentities, wantIdentities) {
				t.Fatalf("round %d page %d IDs = %v, want %v", round, pageIndex+1, gotIdentities, wantIdentities)
			}
		}
	}
}

func TestListPastEndYieldsEmptyNonNil(t *testing.T) {
	lessonStore := openLessonStore(t, seedFiveLessons()...)
	trailing, trailingFailure := lessonStore.List(repo.Page{Number: 9, Size: 2})
	if trailingFailure != nil {
		t.Fatalf("List past end: unexpected failure: %v", trailingFailure)
	}
	if len(trailing) != 0 {
		t.Fatalf("List past end len = %d, want 0", len(trailing))
	}
	if trailing == nil {
		t.Fatal("List past end = nil slice, want empty non-nil slice")
	}
}

func TestListRejectsInvalidPage(t *testing.T) {
	lessonStore := openLessonStore(t, seedFiveLessons()...)
	invalidPages := []struct {
		scenario string
		page     repo.Page
	}{
		{scenario: "zero number", page: repo.Page{Number: 0, Size: 2}},
		{scenario: "negative number", page: repo.Page{Number: -1, Size: 2}},
		{scenario: "zero size", page: repo.Page{Number: 1, Size: 0}},
		{scenario: "negative size", page: repo.Page{Number: 1, Size: -1}},
	}
	for _, pageCase := range invalidPages {
		t.Run(pageCase.scenario, func(t *testing.T) {
			if _, listFailure := lessonStore.List(pageCase.page); !errors.Is(listFailure, repo.ErrInvalidPage) {
				t.Fatalf("List(%+v): failure = %v, want ErrInvalidPage", pageCase.page, listFailure)
			}
		})
	}
}

func TestNewStoreRejectsNilDatabase(t *testing.T) {
	if _, buildFailure := repogorm.NewStore[Lesson, uint](nil, []any{Lesson{}}, "id"); buildFailure == nil {
		t.Fatal("NewStore(nil database): expected failure, got nil")
	}
}

func TestNewStoreRejectsEmptyOrderColumn(t *testing.T) {
	if _, buildFailure := repogorm.NewStore[Lesson, uint](openTestDatabase(t), []any{Lesson{}}, ""); buildFailure == nil {
		t.Fatal("NewStore(empty order column): expected failure, got nil")
	}
}

func TestNewStoreSurfacesMigrationFailure(t *testing.T) {
	// A closed database makes the one AutoMigrate call fail and surface.
	database := openTestDatabase(t)
	sqlHandle, handleFailure := database.DB()
	if handleFailure != nil {
		t.Fatalf("database handle: unexpected failure: %v", handleFailure)
	}
	if closeFailure := sqlHandle.Close(); closeFailure != nil {
		t.Fatalf("close database: unexpected failure: %v", closeFailure)
	}
	if _, buildFailure := repogorm.NewStore[Lesson, uint](database, []any{Lesson{}}, "id"); buildFailure == nil {
		t.Fatal("NewStore(closed database): expected failure, got nil")
	}
}

func TestNewStoreRejectsShapedOrderColumn(t *testing.T) {
	shaped := []string{"id; DROP TABLE lessons", "id DESC --", "lessons.id) OR (1=1", ".id", "9lives"}
	for _, column := range shaped {
		if _, buildFailure := repogorm.NewStore[Lesson, uint](openTestDatabase(t), []any{Lesson{}}, column); buildFailure == nil {
			t.Fatalf("NewStore(%q): expected failure, got nil", column)
		}
	}
}

func TestNewStoreAcceptsDottedOrderColumn(t *testing.T) {
	lessonStore, buildFailure := repogorm.NewStore[Lesson, uint](openTestDatabase(t), []any{Lesson{}}, "lessons.id")
	if buildFailure != nil {
		t.Fatalf("NewStore(dotted column): unexpected failure: %v", buildFailure)
	}
	if lessonStore == nil {
		t.Fatal("NewStore(dotted column) = nil store, want non-nil")
	}
}

// TestListFollowsConfiguredOrderColumn pins the paging contract: List sorts
// by the order column fixed at construction, so a unique column (here slug)
// yields deterministic pages independent of insertion order.
func TestListFollowsConfiguredOrderColumn(t *testing.T) {
	lessonStore, buildFailure := repogorm.NewStore[Lesson, uint](openTestDatabase(t), []any{Lesson{}}, "slug")
	if buildFailure != nil {
		t.Fatalf("NewStore(slug column): unexpected failure: %v", buildFailure)
	}
	seeded := []Lesson{
		{ID: 1, Slug: "lesson-c", Title: "Lesson C"},
		{ID: 2, Slug: "lesson-a", Title: "Lesson A"},
		{ID: 3, Slug: "lesson-b", Title: "Lesson B"},
	}
	for _, candidate := range seeded {
		if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
			t.Fatalf("seed Store(%q): unexpected failure: %v", candidate.Slug, storeFailure)
		}
	}
	listed, listFailure := lessonStore.List(repo.Page{Number: 1, Size: 10})
	if listFailure != nil {
		t.Fatalf("List: unexpected failure: %v", listFailure)
	}
	wantSlugs := []string{"lesson-a", "lesson-b", "lesson-c"}
	gotSlugs := make([]string, 0, len(listed))
	for _, candidate := range listed {
		gotSlugs = append(gotSlugs, candidate.Slug)
	}
	if !slices.Equal(gotSlugs, wantSlugs) {
		t.Fatalf("List slug order = %v, want %v", gotSlugs, wantSlugs)
	}
}
