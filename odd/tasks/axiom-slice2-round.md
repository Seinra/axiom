# Axiom slice-2 round — Product/Mode, Pluck, EnsureUniqueAndAppend, CrossJoin, Shuffle, Range/Clamp (slice) + FromEntries/ToEntries, PickByValues (mapx) + JSON/SQL codecs (opt, result)

Second thesis-compatible adoption round from the future program (`odd/tasks/axiom-future-slices.md`, ranks 8-13 + 3-4). All additive, all teaching-shaped. Launches after the program PR #48 merged.

## Specs

- S1 (owner orders verbatim 2026-10-22): "Vamos con 3 y 4 en el orden que recomiendes" (launches T3 program + T4 slice-2 round; order 3-then-4 recommended and executed: issue #47 + PR #48 merged `babfaa9`) + "Mergea y arranca con T4".
- S2 (scope): `slice`: `Product`, `Mode`, `Pluck`, `EnsureUniqueAndAppend`, `CrossJoin`, `Shuffle`, `Range`/`RangeFrom`, `Clamp`. `mapx`: `FromEntries`/`ToEntries`, `PickByValues`. `opt` + `result`: JSON marshal/unmarshal + SQL Scan/Value codecs.
  - `Product`: overflow-safe sum-grade arithmetic (pairs with `Average`/`Sum`); empty yields the multiplicative identity `1`; arithmetic overflow panics (programmer-error data, `checkedAdd` precedent).
  - `Mode`: most frequent value, OrNull `(T,bool)` shape (empty yields zero + false); ties keep first occurrence (`MaxBy` tie precedent).
  - `Pluck`: field projection without reflection — `Pluck[Element,Field](elements []Element, field func(Element) Field) []Field`; element-only callback; nil yields empty non-nil.
  - `EnsureUniqueAndAppend`: set-building on slices — appends candidate iff absent (`comparable`); always returns a fresh slice, input never mutated.
  - `CrossJoin`: cartesian pairs — `CrossJoin[First,Second](first []First, second []Second) []Pair[First,Second]`; outer-first-major order (for each first, for each second); empty/nil either side yields empty non-nil.
  - `Shuffle`: Fisher-Yates over a fresh copy; purity-compatible — explicit `*rand.Rand` source param (nil panics, programmer error), never the global rand, input never mutated.
  - `Range`/`RangeFrom`: `Range(count int) []int` yields `0..count-1` (negative panics); `RangeFrom(start, count int) []int` yields `start..start+count-1`.
  - `Clamp`: `Clamp[Number cmp.Ordered](value, low, high Number) Number`; `low > high` panics (programmer error).
  - `FromEntries`/`ToEntries` (`mapx`): `FromEntries[Key comparable,Value any](pairs []slice.Pair[Key,Value]) map[Key]Value` (later wins on duplicate keys, documented; pairs with `Zip`); `ToEntries(entries map[Key]Value) []slice.Pair[Key,Value]` (order unspecified — map iteration; callers assert membership, never order). `mapx` may import `slice` for `Pair` (acyclic; precedent `either` imports `result`).
  - `PickByValues` (`mapx`): `PickByValues[Key comparable,Value comparable](entries map[Key]Value, values ...Value) map[Key]Value`; nil yields empty non-nil; never mutates.
  - Codecs (`opt`, `result`): JSON `Some(v)` renders as `v`, `None` as `null` (unmarshal `null` yields `None`; anything else unmarshalable yields an error, never a panic); `Ok(v)` renders as `v`, `Fail(err)` as `{"error":msg}` (unmarshal restores `Fail`); SQL `Value`/`Scan` with `NULL` rendering as `None` and `Value` on `Fail` yielding `(nil, heldErr)`; `Scan` accepts sources directly assignable to `T` (type assertion, no reflection).
- S3 (house style, program S5): element-only callbacks (indexed variants explicit, never `(v,i,s)`); Which-table row + godoc + compiling Example per behaviour group (representative: Product, Pluck, CrossJoin, Shuffle, Range, FromEntries, one JSON codec); purity (frozen decision 8); OrNull `(T,bool)` shapes where absence is a value; panic policy above; no reflection; stdlib-gap justification recorded in godoc; edge + never-mutate pins in tests.
- S4 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green (no source perf change expected); TDD RED→GREEN recorded.

## Tasks

- T1 (S1, S2, S3): Implement the slice helpers + mapx bridges + opt/result codecs + tests + Which-table rows + Examples — route: bounded writer — surfaces: slice/slice.go, slice/slice_test.go, slice/example_test.go, mapx/mapx.go, mapx/mapx_test.go, mapx/example_test.go, opt/opt.go, opt/opt_test.go, opt/example_test.go, result/result.go, result/result_test.go, result/example_test.go — skills: foundry + foundry-testing + cognitive-doc-design — commit: 1768a5c
- T2 (S4): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated (new issue + approval) — commit: pending

## Log

- L1 (request verbatim): "Vamos con 3 y 4 en el orden que recomiendes"
- L2 (decisions verbatim): Range home "Vamos en slice entonces."; codecs scope "JSON + SQL, ambos pkgs" (selected option); launch "Mergea y arranca con T4".
- L3 (2026-10-22): rama feat/axiom-slice2-round desde main post-#48 (`babfaa9`); ODD + mirror listos; T1 a writer acotado; delivery needs new issue + approval.
- L4 (2026-10-22, T1 DONE writer mvj5dxgw-1-ezm4 + parent-verified): 8 slice helpers + 3 mapx bridges + opt/result JSON+SQL codecs, Which-table rows + godoc + Examples; parent re-run suite 16/16 -count=1 ok, vet 0, gofmt vacío, lint 0, bench smoke PASS; verifier mvj63ng7-2-norz Go (0 BLOCKER, 1 WARNING + 1 SUGGESTION test-only, both applied inline + focused PASS + suite re-green); work-unit 1768a5c (1594+/2-, single logical unit over the 400-line heuristic). T2 close + delivery (new issue + approval, native review owner-call).
- L5 (2026-10-22, native review quemado a pedido owner): lineage review-e19bfa66f5d06755 APPROVED + burned (readability+reliability, 1632 líneas); 5 advisories informativas parkeadas sin corrección (R2-1 slice.go:252-266, R2-2 slice.go:314, R2-3 result.go:226-234 SUGGESTIONs; R3-001 result.go:225-243 WARNING; R3-002 opt.go:184-200 SUGGESTION); push/PR owner-gated (nuevo issue + aprobación).
