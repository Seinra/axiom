# VISION — Axiom Product Vision

## Vision

The Echo of Go generics: one small, dependable library that makes the hard parts of generics easy — for humans and for AI — so apps like Catalyst teach and ship typed code instead of `map[string]any`.

## Invariants (non-negotiable)

1. Stdlib-first: `slices`, `maps`, `cmp`, `iter` before hand-rolled; every new symbol justifies what stdlib lacked.
2. Typed over `any`: public API names its constraints; `any` lives only at codec edges behind typed wrappers.
3. Names a stranger can grep: no single-letter vars, no abbreviations, no generic names; godoc on every export with an example.
4. Tree green after every iteration; one logical feature per iteration.
5. AI-readable: package docs answer "which function?" without reading bodies; stable names across releases.
6. Private now, public at maturity: consumed via `replace` while developing; published open-source (MIT/BSD) once it passes the library bar — owner decision 2026-10-20. The DLC is its onboarding ramp: students learn on the same library they keep.
7. Teachable API: every package doubles as course content for the Catalyst generics DLC — names, godoc, and examples must work as lesson material, not just as reference.

## Non-goals

- Not a mega-framework: small composable packages, never "install this whole world".
- No Echo/Gin/GORM imports in core; adapters (`httpx`, `repo`) own those edges alone.
- No `any`-soup helpers, no code generation, no reflection-first APIs.

## MVP line

`skeleton`: `result` + `opt` + `slice` green with docs → then one ODD feature per future package (`mapx`, `set`, `iterx`, concurrency, `pool`, `singleflight`, `di`, `repo`, `httpx`, `valid`) until the S5/S12 coverage map is real.

## Currency rule

Every close-out declares per-artifact disposition (UPDATED vs verified-unchanged with reason); a close-out without it is invalid. On drift, flag it in the open log (STATUS.md debt ledger) — never silently rewrite ratified intent.

## Provenance

Owner request 2026-10-20 (biblioteca de genéricos, futuro framework, consumida por Catalyst) + owner decisions (module `github.com/seinra/axiom`; skeleton + full loop) + scout evidence (Catalyst `go 1.27`, no own generics) + Go 1.27 generic-methods fetch.
