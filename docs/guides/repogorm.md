# repogorm — the repo contract over GORM

You serve the `repo.Repository` contract from a GORM database with
deterministic paging. `NewStore` migrates once; `List` always sorts by the
order column you fix at construction.

You reach for this shape when callers need the `repo.Repository` contract over SQL with deterministic paging: you fix the order column once (usually the primary key) because generic code cannot divine schema, and every `List` sorts by it. Models use normal GORM struct tags; the core's no-tags rule stays core-side.

When not to use it: one-off queries stay simpler in `database/sql` or GORM directly — reach for this store when callers program against the `Repository` contract.

**Version:** `github.com/seinra/axiom`, Go 1.27.
**Prereqs:** `import "github.com/seinra/axiom/repogorm"` plus
`gorm.io/gorm` and a driver (tests use `github.com/glebarez/sqlite`).
Pick the order column before constructing — usually the primary key.

## Quick path

1. Open the database and choose the order column (e.g. `"id"`).
2. Build: `repogorm.NewStore[Lesson, uint](database, []any{Lesson{}}, "id")`.
3. Use it as a `repo.Repository`: `Store` / `Fetch` / `Remove` / `List`.

## Which function?

The package doc (`go doc ./repogorm`) is the source; this table mirrors its
"Which function?" table:

| When you want to ... | Reach for ... |
|---|---|
| Build a GORM-backed `Repository` (migrates once) | `NewStore` |
| Persist / fetch / drop one row | `Store.Store` / `Fetch` / `Remove` |
| List one deterministic page by order column | `Store.List` |

## Example

From `repogorm/example_test.go` (`ExampleNewStore`):

```go fragment: excerpt of ExampleNewStore — full runnable at repogorm/example_test.go
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
```

Run it: `go test ./repogorm/ -run ExampleNewStore -v`.

> [!WARNING]
> Prefer a unique order column (the primary key): rows tied on the order column have no defined relative order, so pages over ties may shift between calls.

> [!NOTE]
> The order column is identifier-shaped only (`^[A-Za-z_][A-Za-z0-9_.]*$`, dotted forms like `"lessons.id"` accepted): anything else fails `NewStore` construction instead of reaching SQL.

## Contracts and errors

| Contract | What it means for you |
|---|---|
| Order column is explicit and fixed | Generic code cannot divine schema, so you pass it (`"id"`). Every `List` sorts by it — unlike `MemoryRepository`, pages repeat exactly. Prefer a unique column (the primary key): rows tied on the order column have no defined relative order, so pages over ties may shift between calls. |
| Order column is identifier-shaped | `NewStore` rejects anything but a plain dotted identifier (`^[A-Za-z_][A-Za-z0-9_.]*$`) — `"id; DROP TABLE lessons"` and `"id DESC --"` fail construction instead of reaching SQL. Dotted forms like `"lessons.id"` are accepted. |
| Construction fails eagerly | A `nil` database, an empty order column, or a migration failure yields a descriptive error at `NewStore` — programmer error surfaced once, never control flow. The zero `Store` is not usable. |
| Same page convention as `repo` | 1-based `Number`, `Size` caps, past the end yields an empty non-nil slice, invalid pages wrap `repo.ErrInvalidPage`. Missing rows wrap `repo.ErrRecordNotFound` — match both with `errors.Is`, so callers can swap memory and GORM stores behind one handler. |
| `Store` upserts | Saving a record with an existing identity overwrites it, mirroring `MemoryRepository`. Models use normal GORM struct tags; the core's no-tags rule stays core-side. |

The store is safe for concurrent use through the shared `*gorm.DB` handle.
For one-off queries, reach for `database/sql` or GORM directly instead of
this contract.

## Next step

For the paging convention itself (`Page`/`Offset`/`TotalPages`), see
`docs/guides/repo.md`. To expose a store through the JSON envelope, see
`docs/guides/httpx.md`.
