# repo — records by identity, pages by convention

You persist, fetch, and drop records addressed by an identity, and list
them one page at a time. The core is pure generics over the standard
library: identity extraction stays a caller-supplied `keyOf` function —
no struct tags, no reflection.

**Version:** `github.com/seinra/axiom`, Go 1.27.
**Prereqs:** `import "github.com/seinra/axiom/repo"`. Define your record
type plus a `keyOf` extractor; construct with `NewMemoryRepository`.

## Quick path

1. Define the record and its `keyOf`: `func lessonKeyOf(c lesson) string { return c.Slug }`.
2. Build: `lessonStore := repo.NewMemoryRepository(lessonKeyOf)`.
3. `Store` / `Fetch` / `Remove` single records; `List(repo.Page{Number: 1, Size: 10})` for pages.

## Which function?

The package doc (`go doc ./repo`) is the source; this table mirrors its
"Which function?" table:

| When you want to ... | Reach for ... |
|---|---|
| Persist / fetch / drop one record | `Repository.Store` / `Fetch` / `Remove` |
| List one page | `Repository.List` |
| In-memory `Repository` for tests | `NewMemoryRepository` (`MemoryRepository`) |
| Select a page; count skipped records | `Page` / `Page.Offset` |
| Count pages; check for a next page | `TotalPages` / `HasNextPage` |
| Match missing record / bad page | `ErrRecordNotFound` / `ErrInvalidPage` |

## Example

From `repo/example_test.go` (`ExampleNewMemoryRepository`):

```go fragment: excerpt of ExampleNewMemoryRepository — full runnable at repo/example_test.go
lessonStore := repo.NewMemoryRepository(lessonKeyOf)
_ = lessonStore.Store(lesson{Slug: "lesson-a", Title: "Lesson A"})
fetched, _ := lessonStore.Fetch("lesson-a")
fmt.Println(fetched.Title)
// Output: Lesson A
```

Run it: `go test ./repo/ -run ExampleNewMemoryRepository -v`.

## Contracts and errors

| Contract | What it means for you |
|---|---|
| Pages are 1-based | The first page is `Number: 1`, never `0`. `List` rejects `Number < 1` or `Size < 1` with `ErrInvalidPage` (match with `errors.Is`). `Page.Offset()` reports `(Number-1)*Size` skipped records. |
| `Size` is your policy, not a clamp | `List` rejects `Size < 1` but never clamps an oversized `Size` — cap it at the handler edge before listing. |
| Past the end is empty, not an error | Listing past the last record yields an empty non-nil slice with a `nil` failure. Overflowed offsets land past the end too, never in a negative-index panic. |
| `MemoryRepository.List` order is unspecified | Map iteration randomizes, so assert membership, never order. Deterministic ordering is the adapter's job — see `docs/guides/repogorm.md`. |
| `Store` upserts | Storing a record with an existing identity overwrites it; one record per identity. Missing `Fetch`/`Remove` report `ErrRecordNotFound` wrapped with the identity. |
| `TotalPages` divides upward | A remainder needs one more page; non-positive totals or sizes yield `0`. `HasNextPage` reports `false` for invalid pages and empty stores. |

The zero `MemoryRepository` is not usable — construct with
`NewMemoryRepository` (a `nil` `keyOf` panics). Do not copy a repository
after use: the mutex stays with its map.

## Next step

For the same contract over GORM with deterministic paging, see
`docs/guides/repogorm.md`. To hold the repository behind a named
singleton, see recipe (c) in `docs/cookbook.md`.
