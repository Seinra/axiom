# Axiom findings slice — all open parked advisories (no new scope)

Close every open finding with evidence or a code pin. Fresh items get fixed; stale items get verified-as-covered or parked-with-reason. Nothing new is built.

## Specs

- S1 (owner order verbatim 2026-10-22): "Arma un slice de los findings"
- S2 (owner scope verbatim 2026-10-22): "De todos los findings abiertos eh"
- S3 (fresh, native review review-b82ed88c618bdae1, all SUGGESTION informational): F1 R2-doc-example-vacuous (example_test.go:48 — ExampleEvery never shows the empty→true case it documents); F2 R2-quantifier-godoc-dup (slice.go:580-611 — three near-identical quantifier paragraphs); F3 R2-takefamily-panic-doc-dup (slice.go:677-761 — four near-identical panic paragraphs); F4 R3-nil-input-take-family (slice.go:692-704, WARNING — nil-input behavior unpinned); F5 verifier SUG-2 (no Take-family allocs bench).
- S4 (stale, re-triaged 2026-10-22): R3-flatten-zero-inner-coverage (result.go:136-138) COVERED by later work (explicit zero-outer Fail + TestFlattenResultZeroOuterReportsError + TestFlattenResultZeroNestingTable); R2-001 iterx (ChunkedSeq snapshot/detach) COVERED (explanatory comments + TestChunkedSeqFreshCopy + lazy-backing docs); either R2-001 WARNING pre-rename UNMAPPABLE (no verbatim; current MapRight/MapLeft document transform-skipping both sides) — stays parked with reason, same class as #6.
- S5 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green; house style S5 (program doc) holds.

## Tasks

- T1 (S1-S3): Fix F1-F4 + add BenchmarkTake allocs pin — route: bounded writer — surfaces: slice/slice_test.go, slice/example_test.go, slice/bench_test.go (NO source changes expected; STOP + report if a fix exposes an implementation bug) — skills: foundry + foundry-testing + cognitive-doc-design — commit: pending
- T2 (S4, S5): Record stale dispositions + verifier + close-out — route: parent; push/PR owner-gated (new issue + approval, #43 closed) — commit: pending

## Log

- L1 (request verbatim): "Arma un slice de los findings"
- L2 (scope verbatim): "De todos los findings abiertos eh"
- L3 (2026-10-22, inventory DONE parent-inline): fresh 4 + SUG-2 go to writer (slice-scoped); R3-flatten-zero-inner verified covered (2 zero tests); iterx R2-001 verified covered (comments + FreshCopy pin); either R2-001 stays parked unmappable post-rename (verbatim unknown, current code documents skip behavior); #6 stays discarded (owner order). Rama feat/axiom-findings-slice desde main.
- L4 (2026-10-22, T1 DONE writer + parent-verified): writer mvj0ri88-8-zxs0 entregó F1 (ExampleEvery empty→true), F4 (TestTakeFamilyNilYieldsEmptyNonNil 16 subtests incl. MaxInt), F5 (BenchmarkTake/Drop allocs); F2/F3 DISMISSED con evidencia (godoc se lee por símbolo en pkg.go.dev: cada helper necesita su contrato completo; las gap-lines son house style S3; lo idéntico real era solo la policy compartida, ya referenciada). Parent re-run fresco: suite 16/16 ok 0 FAIL, vet 0, gofmt vacío, bench smoke PASS. T2 close + delivery.
- L5 (2026-10-22, review disposition findings-fix): candidato sha256:6eda24f0… (slice cerrado, solo tests + prosa) explícitamente sin nuevo review nativo por orden owner — "no hace falta el review"; veredicto previo + re-verificación parent como gate.
