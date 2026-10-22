# Axiom findings open round — all 27 parked advisories (post-F3b)

Triage-only slice over the five native-review burns since the last findings close-out. No new behavior; pins land as tests/docs, dismissals carry evidence, stale items close with reason.

## Specs

- S1 (owner orders verbatim 2026-10-22): "Registra el burn, manda delivery con issue status:approved y despues prepara un slice de todos los findings abiertos." Then: "Merge y de ahi el slice". Prior: "Si, manda review nativo".
- S2 (inventory, 27 advisories, all informational, none opened a correction):
  - F3b (8, lineage review-71da8d1c4a63d8cc, resilience+readability+reliability): R2-001W flow.go:565-637, R2-002 flow.go:736-739, R2-003 flow_test.go:1538-1541, R3-001W flow.go:559-578, R3-002 flow_test.go:1427-1433, R3-003 flow_test.go:1683-1725, R4-001W flow.go:585-589, R4-002 flow.go:539-546.
  - F3 (7, lineage review-4cc3f4fe9c6e8ca4): R2-001W flow.go:441-446, R2-002 flow.go:357-366, R2-003 flow.go:494-496, R3-001W flow.go:357, R3-002 flow_test.go:1003-1006, R4-001W flow.go:357, R4-002 flow.go:411-415.
  - F2 (3, lineage review-864498bfb3f6a9f2): R2-001 test:73-74, R3-001 orderedmap.go:192-195, R3-002 orderedmap.go:30-39.
  - F1 (4, lineage review-fb21b297839fa5d5): R2-001 WARNING queue.go:31-36, R2-002 queue.go:39-41, R2-003 queue_test.go:148-154, R3-stack-drain-backing-retention stack.go:57-71.
  - Slice-2 (5, lineage review-e19bfa66f5d06755): R2-1 slice.go:252-266, R2-2 slice.go:314, R2-3 result.go:226-234, R3-001 WARNING result.go:225-243, R3-002 opt.go:184-200.
- S3 (disposition rule): each finding ends dismissed-with-evidence, documented, or pinned-in-tests; no silent drops; no behavior changes beyond test/doc pins; findings never overwrite each other (new observation per batch, cross-reference priors).
- S4 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green; TDD RED→GREEN where a pin adds behavior coverage; verifier Go before delivery; delivery needs new issue + approval, push/PR/merge owner-gated.

## Tasks

- T1 (S1, S2, S3): Triage 27 in 3 disjoint batches — B1 flow (F3 7 + F3b 8, surfaces flow/*), B2 slice/result/opt (slice-2 5, surfaces slice/* + result/* + opt/*), B3 stack/queue/orderedmap (F1 4 + F2 3, surfaces stack/* + queue/* + orderedmap/*) — route: 3 bounded writers in background, disjoint surfaces — skills: foundry-testing + cognitive-doc-design — commit: 565cb41
- T2 (S4): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated (new issue + approval) — commit: pending
- T3 (S4): Delivery — route: parent; issue type:feature + status:approved + PR Closes #N type:feature + CI green; merge owner-gated — commit: pending

## Log

- L1 (request verbatim): "Registra el burn, manda delivery con issue status:approved y despues prepara un slice de todos los findings abiertos."
- L2 (request verbatim): "Merge y de ahi el slice"
- L3 (prior order verbatim, context): "Si, manda review nativo"
- L4 (2026-10-22, merge a pedido owner): PR #58 merged (628f3da), issue #57 auto-closed; `main` synced locally, tree clean; slice kickoff on `feat/axiom-findings-open-round` from main post-merge.
- L5 (2026-10-22, ubicaciones resueltas pre-ODD): F1 + slice-2 file:line leídos de sus ODD L-rows (`axiom-f1-stack-queue.md:28`, `axiom-slice2-round.md:34`); F2/F3 de Engram (obs 5127/5150); F3b de esta sesión (obs 5165). Nada inventado.
- L6 (2026-10-22, T1 DONE 3 writers + parent-verified): B1 flow 15 (7 documented + 1 pinned + 7 dismissed), B2 5 (3 documented + 2 pinned), B3 7 (3 documented + 2 pinned + 2 dismissed) = 13 + 5 + 9 = 27, 0 behavior change (lib diff comments-only probado vía grep: 14 líneas `//` agregadas / 3 rewordings, 0 código); parent re-run suite 19/19 -count=1 ok, vet 0, gofmt vacío, lint 0 issues; work-unit 565cb41 (136+/3-, 11 files); verifier mvjl488z-4-roik No-Go en trazabilidad (2 BLOCKERs, riesgo behavior nil): sin ledger 27→disposición en el árbol + conteo 10/4/11 ≠ 27 → este ledger appendix lo cierra (13 + 5 + 9 = 27, 5 tests no 4). T2 re-verify + close-out, T3 delivery.
- L7 (2026-10-22, flake preexistente en cuarentena): TestMapSeqPanicCancelsSiblings falla aislado (`-run`, 5/5 + 3/3) pero pasa en paquete completo (suite 19/19 verde, `-count=2 ./flow/` ok); probado en worktree del base 628f3da SIN cambios del slice → falla igual: preexistente, scheduling-dependiente (misma clase que el flake champion-poison obs 4944), no causado por T1 (diff comments-only + tests aditivos). Cuarentena: no se toca el test en este slice; owning ticket = se reporta en el issue de delivery T3.
- L8 (2026-10-22, native review quemado a pedido owner): lineage review-d8d4200e63d0c833 APPROVED + burned (readability+reliability, 229 líneas, group 2/2 limpio via pi_host_relay); 3 advisories informativas parkeadas sin corrección (R2-001 flow/flow.go:420 SUGGESTION, R3-001 WARNING odd/tasks/axiom-findings-open-round.md:31, R3-002 queue/queue_test.go:197-229 SUGGESTION); merge owner-gated (PR #60).

## Disposition ledger (27/27, counts reconciled: 13 documented + 5 pinned + 9 dismissed)

Review-time line numbers drifted (+comment lines); documented rows cite current file:line of the added line. Dismissal evidence = worker proof runs re-runnable by the verifier.

### F3b (lineage review-71da8d1c4a63d8cc, 8)

| ID | Disposition | Artifact / evidence |
| --- | --- | --- |
| R2-001W | documented | flow/flow.go:575 trio-shape one-liner (Emit/TryEmit/Offer) |
| R2-002 | dismissed | `go test -race -count=2 -run 'TestMapSeqResults' ./flow/` ok (B1 handoff) |
| R2-003 | dismissed | same race proof run as R2-002 |
| R3-001W | documented | flow/flow.go:564 mutex one-liner (loadSeqResultsPanic latch) |
| R3-002 | dismissed | same race proof run as R2-002 |
| R3-003 | pinned | TestMapSeqResultsMixedOutcomesKeepInputOrder (flow/flow_test.go) |
| R4-001W | documented | flow/flow.go:593 cancel-drop one-liner (emit cancel branch) |
| R4-002 | documented | flow/flow.go:538 abandoned-slots GC one-liner (worker cancel branch) |

### F3 (lineage review-4cc3f4fe9c6e8ca4, 7)

| ID | Disposition | Artifact / evidence |
| --- | --- | --- |
| R2-001W | dismissed | focused MapSeq/EachSeq proof runs ok (B1 handoff: mid-run-cancel + panic-payload + first-failure runs) |
| R2-002 | documented | flow/flow.go:357 six-param split one-liner (seqFeeder) |
| R2-003 | dismissed | same focused runs as R2-001W |
| R3-001W | documented | flow/flow.go:367 reserve-before-offer one-liner (seqFeeder) |
| R3-002 | dismissed | same focused runs as R2-001W |
| R4-001W | dismissed | `go test -race -count=1 -run 'TestMapSeq\|TestEachSeq' ./flow/` ok (B1 handoff) |
| R4-002 | documented | flow/flow.go:420 only-range-success reword (MapSeq unbounded inputs) |

### F2 (lineage review-864498bfb3f6a9f2, 3)

| ID | Disposition | Artifact / evidence |
| --- | --- | --- |
| R2-001 | dismissed | assertion correct as written (3 Puts, 1 upsert = 2 live); `go test -count=1 ./orderedmap/` ok |
| R3-001 | documented | orderedmap/orderedmap.go:193 LRU-eviction inline comment |
| R3-002 | pinned | TestMapZeroValueOperationsStayPanicFree (orderedmap/orderedmap_test.go) |

### F1 (lineage review-fb21b297839fa5d5, 4)

| ID | Disposition | Artifact / evidence |
| --- | --- | --- |
| R2-001 WARNING | dismissed | covered by package doc (drain-release + slide policy) + TestCompactionKeepsOrder + TestEnqueueAfterDrainReusesQueue; `go test -count=1 ./queue/` ok |
| R2-002 | documented | queue/queue.go:43 compactThreshold 64-entry heuristic one-liner |
| R2-003 | pinned | TestCompactionSlideBoundaryKeepsOrder (queue/queue_test.go) |
| R3-stack-drain-backing-retention | documented | stack/stack.go:52 Pop backing-reuse one-liner |

### Slice-2 (lineage review-e19bfa66f5d06755, 5)

| ID | Disposition | Artifact / evidence |
| --- | --- | --- |
| R2-1 | documented | slice/slice.go:250 checkedMul MinInt-first one-liner |
| R2-2 | documented | slice/slice.go:316-317 Pluck present-path-copy lines |
| R2-3 | documented | result/result.go:222 non-string-error fall-through one-liner |
| R3-001 WARNING | pinned | TestResultUnmarshalJSONNonStringErrorFallsThrough (result/result_test.go) |
| R3-002 | pinned | TestOptionScanInt64IntoIntFails (opt/opt_test.go) |
