# Axiom findings R2 — 11 open advisories (F4 8 + findings-review 3)

Second-round triage over the two native-review burns landed after the 27-item open round. Zero behavior change planned; pins land as tests/docs, dismissals carry evidence. Finding IDs are lineage-scoped (review-bbfc21782dbf0d65 vs review-d8d4200e63d0c833) — never confuse with the 27-ledger IDs of odd/tasks/axiom-findings-open-round.md.

## Specs

- S1 (owner orders verbatim 2026-10-23): "Cuantos findings quedan abiertos?" Then: "Verifica correctamente que sean validos (o sea que siga existiendo el dramita a corregir) y prepara un slice para atacarlos."
- S2 (inventory, 11 advisories, parent-verified against current tree 2026-10-23; task/* + flow.go + queue_test.go untouched since their burns, so review-time lines hold except noted drift):
  - F4 (lineage review-bbfc21782dbf0d65, 8, all VALID — site exists verbatim):
    - R2-001 task/task.go:80-82 (settled/done/future constructor triple) — readability, wording/naming improvable.
    - R2-002 task/task.go:88-99 (`go func` body, Try-inside-Flatten nesting) — readability.
    - R2-003 task/task.go:119 (`<-future.done`) — readability, thin single-line note.
    - R3-001W task/task.go:114-123 (whole Await body) — reliability WARNING: cancel-wins semantic (computed Ok discarded when ctx is done) genuinely debatable, needs disposition.
    - R3-002 task/task_test.go:41-52 (baseline tail + eager-test head) — reliability.
    - R3-003 task/task.go:103-113 (Await godoc + signature) — reliability, claims checkable.
    - R4-001W task/task_test.go:35-44 (`waitForGoroutineBaseline`) — resilience WARNING: Gosched-poll with deadline is timing-flavored.
    - R4-002 task/task_test.go:37-40 (poll loop core) — resilience, same helper.
  - Findings-review (lineage review-d8d4200e63d0c833, 3):
    - R2-001 flow/flow.go:420 (Only-range-success godoc sentence, added in prior triage) — VALID minor: tangled wording, reword candidate.
    - R3-001 odd/tasks/axiom-findings-open-round.md:31 WARNING — CLOSED by parent count 2026-10-23 (obs axiom/findings-r2/r3-001-closed): ledger holds 27 rows exactly once, 13+5+9 reconciles, tables well-formed; verification note added to the ledger intro. No receipt text needed.
    - R3-002 queue/queue_test.go:197-229 (`TestCompactionSlideBoundaryKeepsOrder`) — VALID: hardcoded `seeded=70/drained=64` mirrors unexported `compactThreshold=64`; a threshold change would silently de-fang the boundary pin (same coupling the T2 verifier already flagged).
- S3 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green; TDD RED→GREEN where a pin adds coverage; verifier Go before delivery; delivery needs new issue + approval, push/PR/merge owner-gated.
- S4 (house style, program S5): Which-table row + godoc + compiling Example per export touched; element-only callbacks; purity; no reflection; stdlib-gap justification recorded; edge + never-mutate pins in tests.

## Tasks

- T1 (S1, S2): Triage 11 in 2 disjoint batches — B1 task/* (F4 8: task/task.go, task/task_test.go, task/example_test.go), B2 (findings-review 3: flow/flow.go godoc-only, queue/queue_test.go, odd/tasks/axiom-findings-open-round.md docs-only for the R3-001 row) — route: 2 bounded writers in background, disjoint surfaces — skills: foundry-testing + cognitive-doc-design — commit: pending
- T2 (S3): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated (new issue + approval) — commit: 18f8418 work-unit + close-out docs (this branch)
- T3 (S3): Delivery — route: parent; issue type:feature + status:approved + PR Closes #N + CI green; merge owner-gated — commit: pending

## Log

- L1 (request verbatim): "Cuantos findings quedan abiertos?"
- L2 (request verbatim): "Verifica correctamente que sean validos (o sea que siga existiendo el dramita a corregir) y prepara un slice para atacarlos."
- L3 (2026-10-23, validity verdicts, parent-read sites): 10 VALID (site exists verbatim in current tree, concern plausibly live) + 1 site-drifted (R3-001 ODD:31 — burn-record commit shifted the ledger; resolve from receipt at kickoff). task/* + flow.go + queue_test.go unchanged since their burns (only ODD/STATUS docs commits after), so all other review-time lines hold. Verdicts above are site-existence + plausibility, NOT the dispositions — those belong to T1 with the receipt texts.
- L4 (2026-10-23): rama feat/axiom-findings-r2 desde main (post-#62, árbol limpio); ODD + mirror + todo reconciled before first write; implementation needs owner launch order.
- L5 (2026-10-23, T2 DONE + verifier Go + close-out): work-unit 18f8418 (62+/2-, task/* + flow.go + queue_test.go, suite 20/20 -count=1 ok, vet 0, gofmt vacío, lint 0 issues); verifier mvk968sl-9-c4na Go all S# (0 BLOCKERs; WARNING closed by committing note+spec; loose-assertion SUGGESTION noted no-action, assertion sound); new tests spot-checked (exact oracles, no sleeps/tautologies); slice 64 lines — no delivery strategy pressure. T3 delivery.
