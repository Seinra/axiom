# Axiom slice perf pilot — benchmarks + fuzz (no new projects)

In-repo pilot: benchmarks for hot paths + native fuzz for edge policies, all inside existing `slice` tests with stdlib tooling. No new repos, deps, or infra. If it pays off, roll out to other packages.

## Specs

- S1 (owner order 2026-10-22 verbatim): "Manda, inyecta bien las skills al delegar"
- S2 (scope, parent 2026-10-22): benchmarks `MapSlice`, `FilterSlice`, `FlatMap`/`Flatten`, `Chunked`, `Windows`, `Combinations`, `Permutations` at representative sizes (small + linear-scale); native fuzz `Combinations`, `Permutations` (+ `Chunked`/`Windows` size policy) with seed corpus from existing table cases.
- S3 (properties, parent 2026-10-22): count == C(n,k)/P(n,k); index-order preserved; no input mutation; no panic on valid k; invalid k<0 stays with unit tests (covered, not fuzzed).
- S4 (gates, standing): `go test -count=1 ./...` green (seed corpus runs as unit tests), `go vet`, `gofmt` empty, lint 0; bench evidence via `go test -bench=. -benchtime=100x -run=^$ ./slice/` smoke + full bench numbers recorded; long fuzz sessions stay local, CI runs seeds only.
- S5 (no-goals): no API changes, no docs changes, no CI workflow changes, no new dependencies.

## Tasks

- T1 (S2, S4): Benchmarks hot paths — route: bounded writer — surfaces: slice/bench_test.go (new) — skills: foundry + foundry-testing — commit: pending
- T2 (S2, S3, S4): Fuzz targets + seeds — route: bounded writer (same unit, same surfaces + slice/fuzz_test.go) — skills: foundry + foundry-testing — commit: pending
- T3 (S4, S5): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated — commit: pending

## Log

- L1 (request verbatim): "Manda, inyecta bien las skills al delegar"
- L2 (2026-10-22): rama feat/axiom-slice-perf desde main; ODD + mirror listos; T1+T2 a un writer acotado (una unidad, superficies disjuntas slice/*_test.go nuevos); backstop de contexto vigente.
- L3 (2026-10-22, T1+T2 DONE writer + parent-verified): writer mviwu6pg-3-p19n entregó slice/bench_test.go (8 benchmarks, tamaños S2) + slice/fuzz_test.go (3 targets + seeds, oráculos integer-math propios, panic policy no-fuzzeada); 0 API/docs/CI tocados. Parent re-run fresco: suite 16/16 -count=1 ok, vet 0, gofmt vacío, bench smoke 100x PASS (Map 2461ns … Windows 52038ns/970a … Comb 14592ns/505a … Perm 6519ns/345a, i7-13620H); fuzz 15s x3 PASS writer (~6M execs c/u, 0 crashes). T3 verifier + close.
- L4 (2026-10-22, T3 verifier Go mvix0h75-4-20wm): 0 BLOCKER/WARNING + 2 SUGGESTIONs rollout-scope sin acción (curvas multi-size, segundo tipo fuzz); re-evidencia propia: 8/8 bench, 3/3 fuzz, seeds PASS (171 líneas), vet 0, gofmt vacío, 0 cambios API/docs/CI/deps (+423/-0). Slice 423 líneas: work-unit único + este close-out. Push/PR owner-gated.
- L5 (2026-10-22, native review quemado a pedido owner): lineage review-4e8ce2c0ac31e2b9 APPROVED + burned (readability+reliability, 425 líneas); 4 advisories informativas parkeadas (R3-chunk-boundary-loose + R3-combo-budget-dead + R3-no-distinctness SUGGESTIONs, R3-fuzz-oracle-overflow WARNING, todas en slice/fuzz_test.go); sin correction. Push/PR owner-gated.
- L6 (2026-10-22, findings F1-F4 cerrados): writer mvixhxht-5-kbfq (last-chunk remainder pins, overflow-safe expectedChunks + TestChunkedWindowsHugeSize MaxInt, dead combo-budget skip fuera con nota C(10,5)=252, distinctness pins con duplicados permitidos); parent sacó const huérfana fuzzMaxCombinations (lint); re-run fresco: suite 16/16 -count=1 ok, 0 FAIL, vet 0, gofmt vacío, bench smoke PASS, MaxInt unit PASS. Issue #41 creado para el delivery.
