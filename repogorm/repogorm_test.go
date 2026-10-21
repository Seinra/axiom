// Package repogorm_test verifies the GORM adapter honors the repo contract.
package repogorm_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
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

// openClosableLessonStore builds a migrated Lesson store and keeps its
// database handle so the test can close it and pin the failure branches.
func openClosableLessonStore(t *testing.T, candidates ...Lesson) (*repogorm.Store[Lesson, uint], *gorm.DB) {
	t.Helper()
	database := openTestDatabase(t)
	lessonStore, buildFailure := repogorm.NewStore[Lesson, uint](database, []any{Lesson{}}, "id")
	if buildFailure != nil {
		t.Fatalf("NewStore: unexpected failure: %v", buildFailure)
	}
	for _, candidate := range candidates {
		if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
			t.Fatalf("seed Store(%q): unexpected failure: %v", candidate.Slug, storeFailure)
		}
	}
	return lessonStore, database
}

// closeTestDatabase closes the pooled handle so later operations pin the
// wrapped-driver-error branches instead of the happy path.
func closeTestDatabase(t *testing.T, database *gorm.DB) {
	t.Helper()
	sqlHandle, handleFailure := database.DB()
	if handleFailure != nil {
		t.Fatalf("database handle: unexpected failure: %v", handleFailure)
	}
	if closeFailure := sqlHandle.Close(); closeFailure != nil {
		t.Fatalf("close database: unexpected failure: %v", closeFailure)
	}
}

func TestStoreOnClosedDatabaseReportsError(t *testing.T) {
	lessonStore, database := openClosableLessonStore(t)
	closeTestDatabase(t, database)
	if storeFailure := lessonStore.Store(Lesson{ID: 1, Slug: "closed", Title: "Closed"}); storeFailure == nil {
		t.Fatal("Store on closed database: expected failure, got nil")
	}
}

func TestFetchOnClosedDatabaseIsNotNotFound(t *testing.T) {
	lessonStore, database := openClosableLessonStore(t, Lesson{ID: 1, Slug: "kept", Title: "Kept"})
	closeTestDatabase(t, database)
	_, fetchFailure := lessonStore.Fetch(1)
	if fetchFailure == nil {
		t.Fatal("Fetch on closed database: expected failure, got nil")
	}
	if errors.Is(fetchFailure, repo.ErrRecordNotFound) {
		t.Fatalf("Fetch on closed database: failure = %v, want driver failure, not ErrRecordNotFound", fetchFailure)
	}
}

func TestRemoveOnClosedDatabaseReportsError(t *testing.T) {
	lessonStore, database := openClosableLessonStore(t, Lesson{ID: 1, Slug: "kept", Title: "Kept"})
	closeTestDatabase(t, database)
	if removeFailure := lessonStore.Remove(1); removeFailure == nil {
		t.Fatal("Remove on closed database: expected failure, got nil")
	}
}

func TestListOnClosedDatabaseReportsError(t *testing.T) {
	lessonStore, database := openClosableLessonStore(t, seedFiveLessons()...)
	closeTestDatabase(t, database)
	if _, listFailure := lessonStore.List(repo.Page{Number: 1, Size: 2}); listFailure == nil {
		t.Fatal("List on closed database: expected failure, got nil")
	}
}

// TestListWithTiedOrderColumnReturnsFullSet pins the documented tie
// contract: rows tied on the order column have no defined relative order,
// so the test sorts before comparing and never asserts a tie sequence.
func TestListWithTiedOrderColumnReturnsFullSet(t *testing.T) {
	lessonStore, buildFailure := repogorm.NewStore[Lesson, uint](openTestDatabase(t), []any{Lesson{}}, "title")
	if buildFailure != nil {
		t.Fatalf("NewStore(title column): unexpected failure: %v", buildFailure)
	}
	tied := []Lesson{
		{ID: 1, Slug: "slug-a", Title: "same-title"},
		{ID: 2, Slug: "slug-b", Title: "same-title"},
		{ID: 3, Slug: "slug-c", Title: "same-title"},
		{ID: 4, Slug: "slug-d", Title: "same-title"},
	}
	for _, candidate := range tied {
		if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
			t.Fatalf("seed Store(%q): unexpected failure: %v", candidate.Slug, storeFailure)
		}
	}
	wantIdentities := []uint{1, 2, 3, 4}
	// One full page is a single query: every tied row appears exactly once
	// regardless of tie order, so the sorted set is stable across calls.
	for round := range 2 {
		listed, listFailure := lessonStore.List(repo.Page{Number: 1, Size: 10})
		if listFailure != nil {
			t.Fatalf("round %d List: unexpected failure: %v", round, listFailure)
		}
		gotIdentities := lessonIDs(listed)
		slices.Sort(gotIdentities)
		if !slices.Equal(gotIdentities, wantIdentities) {
			t.Fatalf("round %d tied IDs = %v, want %v (set, tie order undefined)", round, gotIdentities, wantIdentities)
		}
	}
}

func TestConcurrentStoreFetchListHammer(t *testing.T) {
	database := openTestDatabase(t)
	sqlHandle, handleFailure := database.DB()
	if handleFailure != nil {
		t.Fatalf("database handle: unexpected failure: %v", handleFailure)
	}
	// Single pooled connection: goroutines still race through Store
	// concurrently (-race watches), but shared-cache SQLite never trips
	// table locks between pooled connections.
	sqlHandle.SetMaxOpenConns(1)
	lessonStore, buildFailure := repogorm.NewStore[Lesson, uint](database, []any{Lesson{}}, "id")
	if buildFailure != nil {
		t.Fatalf("NewStore: unexpected failure: %v", buildFailure)
	}
	for _, candidate := range seedFiveLessons() {
		if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
			t.Fatalf("seed Store(%q): unexpected failure: %v", candidate.Slug, storeFailure)
		}
	}
	const hammerWorkers = 16
	failures := make([]error, hammerWorkers)
	var waitGroup sync.WaitGroup
	startSignal := make(chan struct{})
	for workerIdx := range hammerWorkers {
		waitGroup.Add(1)
		go func(workerSlot int) {
			defer waitGroup.Done()
			<-startSignal
			identity := uint(100 + workerSlot)
			candidate := Lesson{ID: identity, Slug: fmt.Sprintf("hammer-%d", workerSlot), Title: "Hammer"}
			if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
				failures[workerSlot] = storeFailure
				return
			}
			fetched, fetchFailure := lessonStore.Fetch(identity)
			if fetchFailure != nil {
				failures[workerSlot] = fetchFailure
				return
			}
			if fetched != candidate {
				failures[workerSlot] = fmt.Errorf("Fetch = %+v, want %+v", fetched, candidate)
				return
			}
			if _, listFailure := lessonStore.List(repo.Page{Number: 1, Size: 50}); listFailure != nil {
				failures[workerSlot] = listFailure
			}
		}(workerIdx)
	}
	close(startSignal)
	waitGroup.Wait()
	for workerIdx, workerFailure := range failures {
		if workerFailure != nil {
			t.Fatalf("worker %d failure = %v, want nil", workerIdx, workerFailure)
		}
	}
	listed, listFailure := lessonStore.List(repo.Page{Number: 1, Size: 50})
	if listFailure != nil {
		t.Fatalf("final List: unexpected failure: %v", listFailure)
	}
	if len(listed) != 5+hammerWorkers {
		t.Fatalf("final List len = %d, want %d", len(listed), 5+hammerWorkers)
	}
}
