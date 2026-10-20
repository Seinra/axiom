package repo_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/seinra/axiom/repo"
)

// lesson is a Catalyst-flavored course lesson used as the Record in every scenario.
type lesson struct {
	Slug  string
	Title string
}

// lessonKeyOf extracts the slug identity; it mirrors how callers supply keyOf.
func lessonKeyOf(candidate lesson) string {
	return candidate.Slug
}

// seedLessons stores every candidate in lessonStore, failing fast on setup errors.
func seedLessons(t *testing.T, lessonStore *repo.MemoryRepository[lesson, string], candidates []lesson) {
	t.Helper()
	for _, candidate := range candidates {
		if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
			t.Fatalf("seed Store(%q): unexpected failure: %v", candidate.Slug, storeFailure)
		}
	}
}

func TestMemoryRepositorySatisfiesRepository(t *testing.T) {
	var _ repo.Repository[lesson, string] = repo.NewMemoryRepository(lessonKeyOf)
}

func TestStoreFetchRoundTrip(t *testing.T) {
	lessonStore := repo.NewMemoryRepository(lessonKeyOf)
	candidate := lesson{Slug: "generics-intro", Title: "Intro to Generics"}
	if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
		t.Fatalf("Store: unexpected failure: %v", storeFailure)
	}
	fetched, fetchFailure := lessonStore.Fetch("generics-intro")
	if fetchFailure != nil {
		t.Fatalf("Fetch: unexpected failure: %v", fetchFailure)
	}
	if fetched != candidate {
		t.Fatalf("Fetch = %+v, want %+v", fetched, candidate)
	}
}

func TestFetchMissingReportsNotFound(t *testing.T) {
	lessonStore := repo.NewMemoryRepository(lessonKeyOf)
	if _, fetchFailure := lessonStore.Fetch("no-such-lesson"); !errors.Is(fetchFailure, repo.ErrRecordNotFound) {
		t.Fatalf("Fetch missing: failure = %v, want ErrRecordNotFound", fetchFailure)
	}
}

func TestRemoveMissingReportsNotFound(t *testing.T) {
	lessonStore := repo.NewMemoryRepository(lessonKeyOf)
	if removeFailure := lessonStore.Remove("no-such-lesson"); !errors.Is(removeFailure, repo.ErrRecordNotFound) {
		t.Fatalf("Remove missing: failure = %v, want ErrRecordNotFound", removeFailure)
	}
}

func TestRemoveDeletesStoredLesson(t *testing.T) {
	lessonStore := repo.NewMemoryRepository(lessonKeyOf)
	seedLessons(t, lessonStore, []lesson{{Slug: "retired-lesson", Title: "Retired"}})
	if removeFailure := lessonStore.Remove("retired-lesson"); removeFailure != nil {
		t.Fatalf("Remove: unexpected failure: %v", removeFailure)
	}
	if _, fetchFailure := lessonStore.Fetch("retired-lesson"); !errors.Is(fetchFailure, repo.ErrRecordNotFound) {
		t.Fatalf("Fetch after Remove: failure = %v, want ErrRecordNotFound", fetchFailure)
	}
	listed, listFailure := lessonStore.List(repo.Page{Number: 1, Size: 10})
	if listFailure != nil {
		t.Fatalf("List after Remove: unexpected failure: %v", listFailure)
	}
	if len(listed) != 0 {
		t.Fatalf("List after Remove len = %d, want 0", len(listed))
	}
}

func TestStoreOverwritesSameIdentity(t *testing.T) {
	lessonStore := repo.NewMemoryRepository(lessonKeyOf)
	seedLessons(t, lessonStore, []lesson{{Slug: "upsert-me", Title: "First"}})
	seedLessons(t, lessonStore, []lesson{{Slug: "upsert-me", Title: "Second"}})
	fetched, fetchFailure := lessonStore.Fetch("upsert-me")
	if fetchFailure != nil {
		t.Fatalf("Fetch: unexpected failure: %v", fetchFailure)
	}
	if fetched.Title != "Second" {
		t.Fatalf("Fetch Title = %q, want %q (Store overwrites)", fetched.Title, "Second")
	}
	listed, listFailure := lessonStore.List(repo.Page{Number: 1, Size: 10})
	if listFailure != nil {
		t.Fatalf("List: unexpected failure: %v", listFailure)
	}
	if len(listed) != 1 {
		t.Fatalf("List len = %d, want 1 (upsert keeps one record per identity)", len(listed))
	}
}

func TestPageOffsetComputesRow(t *testing.T) {
	offsetCases := []struct {
		scenario string
		page     repo.Page
		want     int
	}{
		{scenario: "first page starts at zero", page: repo.Page{Number: 1, Size: 10}, want: 0},
		{scenario: "second page skips one size", page: repo.Page{Number: 2, Size: 10}, want: 10},
		{scenario: "third page of two", page: repo.Page{Number: 3, Size: 2}, want: 4},
	}
	for _, offsetCase := range offsetCases {
		if got := offsetCase.page.Offset(); got != offsetCase.want {
			t.Fatalf("%s: Offset() = %d, want %d", offsetCase.scenario, got, offsetCase.want)
		}
	}
}

func TestListPagesCoverSeededLessons(t *testing.T) {
	lessonStore := repo.NewMemoryRepository(lessonKeyOf)
	seeds := []lesson{
		{Slug: "lesson-a", Title: "Lesson A"},
		{Slug: "lesson-b", Title: "Lesson B"},
		{Slug: "lesson-c", Title: "Lesson C"},
		{Slug: "lesson-d", Title: "Lesson D"},
		{Slug: "lesson-e", Title: "Lesson E"},
	}
	seedLessons(t, lessonStore, seeds)
	seedSlugs := make(map[string]struct{}, len(seeds))
	for _, candidate := range seeds {
		seedSlugs[candidate.Slug] = struct{}{}
	}
	pageCases := []struct {
		pageNumber int
		wantCount  int
	}{
		{pageNumber: 1, wantCount: 2},
		{pageNumber: 2, wantCount: 2},
		{pageNumber: 3, wantCount: 1},
	}
	// Pages are sliced from a freshly collected snapshot per call, and map
	// iteration randomizes: page lengths are deterministic, but membership
	// across separate calls is not. Assert lengths plus seed membership here;
	// the union over every seed is asserted on one stable snapshot below.
	for _, pageCase := range pageCases {
		listed, listFailure := lessonStore.List(repo.Page{Number: pageCase.pageNumber, Size: 2})
		if listFailure != nil {
			t.Fatalf("List page %d: unexpected failure: %v", pageCase.pageNumber, listFailure)
		}
		if len(listed) != pageCase.wantCount {
			t.Fatalf("List page %d len = %d, want %d", pageCase.pageNumber, len(listed), pageCase.wantCount)
		}
		for _, candidate := range listed {
			if _, present := seedSlugs[candidate.Slug]; !present {
				t.Fatalf("List page %d: slug %q was never seeded", pageCase.pageNumber, candidate.Slug)
			}
		}
	}
	complete, completeFailure := lessonStore.List(repo.Page{Number: 1, Size: len(seeds)})
	if completeFailure != nil {
		t.Fatalf("List full snapshot: unexpected failure: %v", completeFailure)
	}
	if len(complete) != len(seeds) {
		t.Fatalf("List full snapshot len = %d, want %d", len(complete), len(seeds))
	}
	seenSlugs := make(map[string]struct{}, len(seeds))
	for _, candidate := range complete {
		seenSlugs[candidate.Slug] = struct{}{}
	}
	for _, candidate := range seeds {
		if _, present := seenSlugs[candidate.Slug]; !present {
			t.Fatalf("seed %q missing from full snapshot", candidate.Slug)
		}
	}
	trailing, trailingFailure := lessonStore.List(repo.Page{Number: 4, Size: 2})
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
	lessonStore := repo.NewMemoryRepository(lessonKeyOf)
	seedLessons(t, lessonStore, []lesson{{Slug: "only-lesson", Title: "Only"}})
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

func TestConcurrentStoreFetchHammer(t *testing.T) {
	lessonStore := repo.NewMemoryRepository(lessonKeyOf)
	seedLessons(t, lessonStore, []lesson{{Slug: "shared-lesson", Title: "Shared seed"}})
	const workerCount = 32
	const roundCount = 100
	failureCh := make(chan error, workerCount)
	var workerGroup sync.WaitGroup
	for worker := range workerCount {
		workerGroup.Add(1)
		go func(workerIndex int) {
			defer workerGroup.Done()
			slug := fmt.Sprintf("worker-%02d", workerIndex)
			for range roundCount {
				candidate := lesson{Slug: slug, Title: fmt.Sprintf("title-%d", workerIndex)}
				if storeFailure := lessonStore.Store(candidate); storeFailure != nil {
					failureCh <- storeFailure
					return
				}
				shared := lesson{Slug: "shared-lesson", Title: fmt.Sprintf("shared-by-%d", workerIndex)}
				if storeFailure := lessonStore.Store(shared); storeFailure != nil {
					failureCh <- storeFailure
					return
				}
				if _, fetchFailure := lessonStore.Fetch(slug); fetchFailure != nil {
					failureCh <- fetchFailure
					return
				}
				if _, listFailure := lessonStore.List(repo.Page{Number: 1, Size: 4}); listFailure != nil {
					failureCh <- listFailure
					return
				}
			}
		}(worker)
	}
	workerGroup.Wait()
	close(failureCh)
	for failure := range failureCh {
		t.Fatalf("concurrent worker failure: %v", failure)
	}
	listed, listFailure := lessonStore.List(repo.Page{Number: 1, Size: workerCount + 1})
	if listFailure != nil {
		t.Fatalf("final List: unexpected failure: %v", listFailure)
	}
	if len(listed) != workerCount+1 {
		t.Fatalf("final List len = %d, want %d (one record per worker plus shared)", len(listed), workerCount+1)
	}
}
