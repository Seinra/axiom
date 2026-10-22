# Axiom P2 combinations/permutations — slice slice (issue #36)

Bounded slice-only track: Combinations + Permutations over slices, deterministic order, godoc + Which-table + compiling Example. First additive test of the v0 bar (odd/tasks/axiom-v0-bar.md).

## Specs

- S1 (issue #36 verbatim, title `[feat]: P2 combinations/permutations (stdlib-gap + teachable)`): "P2 slice candidate per gate-stdlib-ensenable: stdlib (slices/maps/iter) has no combinations/permutations; teachable as DLC lesson (exam distractor generation = candidate use case from audit). Scope: Combinations/Permutations over slices with deterministic order, godoc + Which-table + compiling Example. No implementation until approved."
- S2 (issue #36 Consumer fit verbatim): "DLC lesson material (combinatorics for exam generation); standalone-teachable per VISION invariant 8."
- S3 (owner order 2026-10-22 verbatim): "Ok, manda #36"
- S4 (gate-stdlib-ensenable, decided 2026-10-20): P2 enters iff stdlib lacks it AND it teaches as DLC lesson material (godoc + Example), even with no Catalyst production use yet.
- S5 (v0 bar constraint, odd/tasks/axiom-v0-bar.md): additive-only freeze; new exports MUST carry godoc + Which-table entry + compiling Example and pass `go test -count=1 ./...`, `go vet`, `gofmt`, lint 0.

## Tasks

- T1 (S1, S4): Explorar superficie slice + gap stdlib — route: parent inline read-only (Codegraph/zvec-grep ya activos) — commit: n/a
- T2 (S1, S2, S5): Implementar Combinations/Permutations + tests TDD — route: bounded writer — surfaces: slice/* — skills: foundry + foundry-testing + cognitive-doc-design — commit: pending
- T3 (S5): Verifier + Which-table/Example + close-out — route: parent + independent verifier; push/PR owner-gated — commit: pending

## Log

- L1 (request verbatim): "Ok, manda #36"
- L2 (2026-10-22): rama feat/axiom-p2-combos creada desde main (base fcf8271); ODD preparado sin implementación; T2 a writer acotado por context backstop.
- L3 (2026-10-22, T1 DONE parent-inline): stdlib gap probado (`go doc slices | grep -c -i -E "combin|permut"` → 0; lista completa sin combinatoria); superficie: shape-ops sin sufijo Slice (Chunked/Windows/Distinct/Partition/Intersperse) → nombres `Combinations`/`Permutations` con k (`Combinations(elements, k)`, `Permutations(elements, k)`); política bordes: k<0 panic (rima con Chunked/Windows size<=0), k==0 → un combo vacío, k>len → vacío non-nil, copias frescas, orden lexicográfico por índice, puros sin mutar; Which-table en slice.go doc + Examples compilables.
- L4 (2026-10-22, T2 DONE writer + parent-verified): writer mvitmt3t-1-ctn9 entregó 362 inserciones / 0 borrados en slice/slice.go + slice/slice_test.go + slice/example_test.go (S1+S2+S4+S5; S3 orden sin código); parent spot-check: DFS start-index (Combinations) y used[] (Permutations), mensajes panic exactos, k==0/k>len/nil/copias/mutate conformes, Which-table slice.go:27-28, Examples con Output; re-run fresco: `go test ./slice/ -count=1` ok + `go test -count=1 ./...` árbol completo verde, vet 0, gofmt vacío; TDD RED (undefined) → GREEN registrado por writer. T3 verifier independiente + close.
- L5 (2026-10-22, T3 verifier Go): gentle-ai-verify mvittlsz-2-k37c, 0 BLOCKER/WARNING + 2 SUGGESTIONs docs-only aplicadas aquí (dedup L3, conteo preciso); re-evidencia propia: export count 26→28 (+2 exactos), solo hunks `+`, Examples PASS, `go test ./... -count=1` verde, vet 0, gofmt vacío, lint slice/... 0 issues. Close-out: push/PR owner-gated.
- L6 (2026-10-22, review disposition): candidato sha256:c3ac9442… (slice ×3 + ODD + STATUS, código aditivo) explícitamente sin review nativo por orden owner — "Sin review entonces"; verifier independiente Go como gate.
