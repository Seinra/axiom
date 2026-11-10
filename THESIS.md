# THESIS — Axiom Product Thesis

> Status: **PUBLIC STANDING** — this is the contract that gets cited. Frozen; it changes only by your explicit thesis-change.

## One-line identity

Axiom: one dependable, composable, educative generics library that makes the hard parts of generics easy — for humans and for AI — so apps like Catalyst teach and ship typed code instead of `map[string]any`.

## Invariants (non-negotiable)

1. Stdlib-first: `slices`, `maps`, `cmp`, `iter` before hand-rolled; every new symbol justifies what stdlib lacked.
2. Typed over `any`: public API names its constraints; `any` lives only at codec edges behind typed wrappers.
3. Names a stranger can grep: no single-letter vars, no abbreviations, no generic names; godoc on every export with an example.
4. Tree green after every iteration; one logical feature per iteration.
5. AI-readable: package docs answer "which function?" without reading bodies; stable names across releases.
6. Public MIT: go gettable since v0.1.0 (`go get github.com/seinra/axiom@v0.2.0`); local checkout + `replace` only while developing. The DLC is its onboarding ramp: students learn on the same library they keep.
7. Teachable API: every package doubles as course content for the Catalyst generics DLC — names, godoc, and examples must work as lesson material, not just as reference.
8. Standalone teachability (owner decision 2026-10-21): the split is inside Axiom — reference (godoc + Which-tables) vs. guide (decision ladders + compiling Examples), all in-repo and standalone. Catalyst DLC consumes that material for curriculum (prerequisites, exercises), never hosts missing content. Axiom is self-sufficient to learn; Catalyst adds classroom.

## Non-goals

- Not a mega-framework: small composable packages, never "install this whole world".
- No Echo/Gin/GORM imports in core; adapters (`httpx`, `repo`) own those edges alone.
- No `any`-soup helpers, no code generation, no reflection-first APIs.

## MVP — shipped

No skeleton anymore: 21 pkgs on main, `v0.2.0` tagged, go gettable, tests green — `result`, `opt`, `either`, `slice`, `set`, `mapx`, `iterx`, `valid`, `singleflight`, `pool`, `di`, `repo` (+GORM adapter `repogorm`), `httpx` (Echo v5), `flow`, `stack`, `queue`, `orderedmap`, `task`, `mathx`, `strings`.

## Rejected (short list)

Rejected under the stdlib-gap + teachability gate unless new evidence reopens them:

- SQL-style Join: adapter concern for a concrete store, not teachable generics.
- singleflight stats/observability: no DLC lesson needs it; YAGNI.
- DoChan/sharding/batching variants: YAGNI.
- Stdlib-owned duplicates, never reimplemented: sort/reverse/concat (`slices`), lazy memoize (`sync.OnceValue`), atomic wrappers (`sync/atomic`).
- Tuples T2–T9, `Lazy`, gods-style containers: fail the teachability-per-cost gate (Go lacks variadic generics); `Zip`/`Pair` covers the common case.
- Reflection-first or code-generated APIs; mutable in-place variants; domain helpers (dates, URLs, query strings).

## Provenance

Owner request 2026-10-20 (biblioteca de genéricos, futuro framework, consumida por Catalyst) + owner decisions (module `github.com/seinra/axiom`; skeleton + full loop) + scout evidence (Catalyst `go 1.27`, no own generics) + Go 1.27 generic-methods fetch.

## What this document is not

No currency, no open decisions, no history. Evolution (drift, open decisions, currency per close-out, DLC roadmap) lives in the private loop mother (`VISION.md`, Conduit-tracked), not here.
