# ARCHITECTURE

Component layout, data flow, and frozen decisions for Axiom. Read before designing any package.

## Layout (library, not service)

```text
axiom/
├── AXIOM_LOOP.md
├── VISION.md
├── loop-config.yaml
├── .loop-modules/
├── odd/tasks/
├── result/        # Result[T] + Must/Try helpers (skeleton: YES)
├── opt/           # Option[T] (skeleton: YES)
├── slice/         # Slice helpers over stdlib slices (skeleton: YES)
├── mapx/          # (future) Map helpers over stdlib maps
├── set/           # (future) Set[T comparable]
├── iterx/         # (future) iter.Seq/Seq2 adapters
├── pool/          # (future) generic sync.Pool
├── singleflight/  # (future) generic dedup + batching
├── di/            # (future) Provide/Invoke, scopes
├── repo/          # (future) Repository[T,ID] + Page[T]
├── httpx/         # (future) Echo Binder[T] / Response[T]
└── valid/         # (future) Validator[T] + predicates
```

Library rule (adapted from foundry package-organization): every package is public API — that raises the bar, it never lowers it. `internal/` only for test helpers shared across packages that must NOT become API.

## Data flow

```text
Catalyst (consumer)
   │  import github.com/seinra/axiom/<pkg>
   ▼
axiom packages (pure Go, stdlib + cmp/slices/maps/iter only)
   │  no Echo/GORM imports in core (httpx/repo adapt at the edge)
   ▼
stdlib (slices, maps, cmp, iter, sync, sync/atomic)
```

Core packages MUST NOT import Echo, Gin, or GORM. Adapter packages (`httpx`, `repo`) own those deps alone.

## Frozen decisions

1. Go 1.27 minimum; generic methods allowed on concrete types; NO generic interface methods (language limit, GO-GENERIC-METHODS).
2. Stdlib-first: `slices`/`maps`/`cmp`/`iter` before hand-rolled (Simplicity Ladder rung 3).
3. No `any` in domain-facing API (craft-floor ban 17); `any` only at JSON/codec edges with typed wrappers.
4. Naming per foundry code-style: no single-letter vars (except ≤5-line loop index), no abbreviations (`usr`/`cfg`/`req`/`res`), no generic names (`data`/`item`/`result`); collections are plural nouns; predicates are `is/has/can/should`. Constructors: bare `New()` is the default constructor for its type — zero-arg or with required dependencies (`pool.New(maker)`, `valid.New(rules...)`); stateful/adapting construction gets a named constructor (`NewStore`, `NewMemoryRepository`). Single-letter identifiers verified absent (ban-13/-14 hold).
5. Fewest files: one file per boundary, never single-export files (ban 22) — co-locate until the file earns its boundary.
6. Functions ≤50 lines / complexity ≤10 / params ≤4; guard clauses over nesting; sentinel errors + `fmt.Errorf("…: %w", …)` wrapping.
7. Convenience outweighs minimalism (owner decision 2026-10-20): a stdlib-adjacent helper that saves lines or teaches better stays even if it wraps stdlib thinly; YAGNI still kills speculative generality, never proven convenience.
7. Skeleton slice ships only `result` + `opt` + `slice`; everything else is declared future scope, never stubbed speculatively.
8. Purity (owner decision 2026-10-22, deduced from package docs + never-mutate pins, then ratified): all helpers are pure — inputs are never mutated, outputs are fresh values. Mutable in-place variants are rejected: callers rely on inputs staying intact, and Go already offers mutation via plain loops and the stdlib when the caller opts in.

## Coverage map (present → future)

| Area | Present (skeleton) | Future (declared, not built) |
| --- | --- | --- |
| Collections | `slice` Map/Filter/Reduce/Find/GroupBy | `mapx`, `set`, OrderedMap/Stack/Queue/LRU |
| FP | `result`, `opt` | Either/Future/Task, Pipe/Compose/Memoize |
| Concurrency | — | fan-in/out, pipeline, worker pool, channels over `iter.Seq` |
| Infra | — | `pool`, `singleflight`, `di`, `repo`, `httpx`, `valid` |

## Coverage gaps (cross-language survey 2026-10-20)

Compared against Rust itertools, Kotlin stdlib, C# LINQ, Java Streams, Swift
Algorithms, and C++ ranges. P0 = unanimous across libs, small. P1 =
valuable, medium. P2 = stdlib-gap + teachable candidates (owner decision
2026-10-20: gate-stdlib-ensenable, Engram axiom/p2/gate-stdlib-ensenable).
P2 enters when the stdlib lacks it AND it is teachable as DLC lesson material
(godoc + example), even if Catalyst has no production use yet. Axiom precedes
the DLC, so curriculum need gates P2, not production usage.

| Area | P0 (next slices) | P1 | P2 (parked) |
| --- | --- | --- | --- |
| `slice` | Min/Max/Sum(+By), FlatMap/Flatten, Chunked/Windows, Zip/Unzip, Distinct, Partition, MapIndexed/FilterIndexed, Intersperse, Tally | Average, SortBy convenience | — |
| `iterx` | TakeWhile/DropWhile, FlatMap/FlattenSeq, ZipSeq, OnEachSeq, Repeat/Cycle, DeduplicateSeq | — | — |
| `set`/`mapx` | SymmetricDifference/IsSubsetOf/Overlaps, Omit/Invert/ValueOr, GroupByMapped | — | — |
| `opt`/`result` | OrElse/AndThen/Filter/ValueOrElse/ToResult, MapFailure/UnwrapOr | Try (panic→Result) | `either` package |
| `valid`/`repo`/`httpx`/`di` | And/Or/Not predicates | TotalPages/HasNext, ValidationWriteError envelope, di test Replace | — |
| combos | — | — | combinations/permutations, Join, Race, singleflight stats |

## Non-goals (rejected under the stdlib-gap + teachable gate, 2026-10-20)

- SQL-style Join: adapter concern for a concrete store, not teachable generics.
- singleflight stats/observability: no DLC lesson needs it; YAGNI.
- DoChan/sharding/batching variants: YAGNI (see singleflight package doc).
- Stdlib-owned duplicates, never reimplemented: sort/reverse/concat
  (`slices`), lazy memoize (`sync.OnceValue`), atomic wrappers (`sync/atomic`).
