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
4. Naming per foundry code-style: no single-letter vars (except ≤5-line loop index), no abbreviations (`usr`/`cfg`/`req`/`res`), no generic names (`data`/`item`/`result`); collections are plural nouns; predicates are `is/has/can/should`.
5. Fewest files: one file per boundary, never single-export files (ban 22) — co-locate until the file earns its boundary.
6. Functions ≤50 lines / complexity ≤10 / params ≤4; guard clauses over nesting; sentinel errors + `fmt.Errorf("…: %w", …)` wrapping.
7. Skeleton slice ships only `result` + `opt` + `slice`; everything else is declared future scope, never stubbed speculatively.

## Coverage map (present → future)

| Area | Present (skeleton) | Future (declared, not built) |
| --- | --- | --- |
| Collections | `slice` Map/Filter/Reduce/Find/GroupBy | `mapx`, `set`, OrderedMap/Stack/Queue/LRU |
| FP | `result`, `opt` | Either/Future/Task, Pipe/Compose/Memoize |
| Concurrency | — | fan-in/out, pipeline, worker pool, channels over `iter.Seq` |
| Infra | — | `pool`, `singleflight`, `di`, `repo`, `httpx`, `valid` |
