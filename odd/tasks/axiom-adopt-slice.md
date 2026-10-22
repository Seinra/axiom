# Axiom adopt slice round — OrElse finds + quantifiers + find ext + Take/Drop (issue #41 follow-up)

First thesis-compatible adoption round: 13 slice helpers proven missing by the go-utils/lo survey, all teaching-shaped, all additive. Launches from the ranked program (odd/tasks/axiom-future-slices.md).

## Specs

- S1 (owner order verbatim 2026-10-22): "Ahora empecemos" (lanza la ronda de adopción recomendada: OrElse trio + quantifiers + find ext + Take/Drop).
- S2 (scope, parent 2026-10-22): `FindOrElse`, `FirstOr`, `LastOr` (fallback, no bool — Or-canonical); `Every`, `Some`, `None` (vacuous: Every([])=true, Some([])=false, None([])=true); `None` = !`Some`); `FindDuplicates` (values occurring >1, order of first occurrence), `FindUniques` (values occurring exactly once, input order), `FindIndexes` (all matching indexes ascending); `Take`/`Drop`/`TakeLast`/`SkipLast` (n<0 panics — programmer error, Chunked/Windows size policy; n==0 → Take empty / Drop full copy; n>len → Take full copy / Drop empty; copies, never mutate).
- S3 (house style, program S5): element-only callbacks; Which-table rows + godoc + compiling Examples (FirstOr, Every, FindDuplicates, Take); purity; OrNull shapes where absence is a value (find-index absence = empty slice, not bool); panic policy above; no reflection; stdlib-gap justification in godoc (slices has none of these).
- S4 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green (no source perf change expected); TDD RED→GREEN recorded.

## Tasks

- T1 (S1, S2, S3): Implementar las 13 funciones + tests + Which-table + Examples — route: bounded writer — surfaces: slice/slice.go, slice/slice_test.go, slice/example_test.go — skills: foundry + foundry-testing + cognitive-doc-design — commit: pending
- T2 (S4): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated — commit: pending

## Log

- L1 (request verbatim): "Ahora empecemos"
- L2 (2026-10-22): rama feat/axiom-adopt-slice desde main; ODD + mirror listos; T1 a writer acotado (superficies slice disjuntas); #41 cubre el delivery (Closes #41 + type:test ya aprobados para el pilot; esta ronda amplía el mismo PR o uno nuevo a decisión de entrega).
- L3 (2026-10-22, T1 DONE writer + parent-verified): writer mvizxi5b-6-3c9d entregó las 13 (OrElse trio, quantifiers con vacuous pins + None==!Some, find ext con first-occurrence/ascending, Take-family con panic exacto + zero/full/oversize + copies); Which-table rows + 4 Examples con Output; re-run fresco: suite 16/16 -count=1 ok 0 FAIL, vet 0, gofmt vacío, bench smoke PASS; TDD RED (undefined) → GREEN. T2 verifier + close.
