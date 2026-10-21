package docs

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/seinra/axiom/di"
	"github.com/seinra/axiom/flow"
	"github.com/seinra/axiom/opt"
	"github.com/seinra/axiom/repo"
	"github.com/seinra/axiom/result"
	axiomslice "github.com/seinra/axiom/slice"
)

// TestAbsenceLadder executes cookbook recipe (a): OrNull → opt → result,
// and asserts the exact outputs documented in docs/cookbook.md.
func TestAbsenceLadder(t *testing.T) {
	t.Parallel()

	lessons := []string{"generics-intro", "slices-deep", "maps-fast"}
	// Rung 1: (T, bool) OrNull at the caller's own boundary.
	slug, found := axiomslice.FindSlice(lessons, func(s string) bool { return s == "slices-deep" })
	missing, missingFound := axiomslice.FindSlice(lessons, func(s string) bool { return s == "nope" })
	// Rung 2: opt.Option to chain or transform the absence.
	maybe := opt.Some(slug)
	// Rung 3: result.Result when absence becomes a failure with a cause.
	outcome := opt.ToResult(opt.None[string](), errors.New("lesson not found"))

	got := []string{
		fmt.Sprintf("%s %v", slug, found),
		fmt.Sprintf("%v %v", missing == "", missingFound),
		maybe.ValueOr("nobody"),
		opt.None[string]().ValueOr("nobody"),
		result.UnwrapOr(outcome, "fallback-lesson"),
	}
	want := []string{
		"slices-deep true",
		"true false",
		"slices-deep",
		"nobody",
		"fallback-lesson",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("absence ladder outputs = %q, want %q", got, want)
	}
}

// TestRaceFirstSuccess executes cookbook recipe (b): Race over mirrors.
// Same shape (switch + select on ctx.Done); the work is immediate so the
// test stays deterministic with no timers.
func TestRaceFirstSuccess(t *testing.T) {
	t.Parallel()

	mirrors := []string{"slow-mirror", "fast-mirror", "dead-mirror"}
	winner, err := flow.Race(context.Background(), mirrors, 3, func(ctx context.Context, mirror string) (string, error) {
		switch mirror {
		case "fast-mirror":
			return "lesson from " + mirror, nil
		case "slow-mirror":
			<-ctx.Done()
			return "", ctx.Err()
		default:
			return "", errors.New("mirror unreachable: " + mirror)
		}
	})
	if err != nil {
		t.Fatalf("Race failure = %v, want nil", err)
	}
	got := fmt.Sprintf("%s %v", winner, err == nil)
	if want := "lesson from fast-mirror true"; got != want {
		t.Fatalf("Race output = %q, want %q", got, want)
	}
}

type cookbookLesson struct {
	Slug  string
	Title string
}

func cookbookLessonKeyOf(candidate cookbookLesson) string { return candidate.Slug }

// TestWiredStore executes cookbook recipe (c): di + repo MemoryRepository,
// and asserts the documented outputs (title plus singleton identity).
func TestWiredStore(t *testing.T) {
	t.Parallel()

	injector := di.New()
	if err := di.Provide(injector, "lessons", func(*di.Injector) (*repo.MemoryRepository[cookbookLesson, string], error) {
		return repo.NewMemoryRepository(cookbookLessonKeyOf), nil
	}); err != nil {
		t.Fatalf("Provide failure = %v, want nil", err)
	}
	lessons, err := di.Resolve[*repo.MemoryRepository[cookbookLesson, string]](injector, "lessons")
	if err != nil {
		t.Fatalf("Resolve failure = %v, want nil", err)
	}
	if err := lessons.Store(cookbookLesson{Slug: "generics-intro", Title: "Intro to Generics"}); err != nil {
		t.Fatalf("Store failure = %v, want nil", err)
	}
	fetched, err := lessons.Fetch("generics-intro")
	if err != nil {
		t.Fatalf("Fetch failure = %v, want nil", err)
	}
	if fetched.Title != "Intro to Generics" {
		t.Fatalf("Fetch title = %q, want %q", fetched.Title, "Intro to Generics")
	}
	again, err := di.Resolve[*repo.MemoryRepository[cookbookLesson, string]](injector, "lessons")
	if err != nil {
		t.Fatalf("second Resolve failure = %v, want nil", err)
	}
	if again != lessons {
		t.Fatal("second Resolve returned a different store, want the same singleton")
	}
}
