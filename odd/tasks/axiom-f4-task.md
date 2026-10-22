# Axiom F4 — deferred cancelable Task returning Result

Fourth declared slice (`odd/tasks/axiom-future-slices.md` program row): deferred async computation with cancel — "async done right". Owner-locked direction: Task diferido cancelable → Result. Size: large (new package + new abstraction; resume test fails → ODD-tracked).

## Specs

- S1 (owner orders verbatim 2026-10-22): "Larga F4". Direction locked by owner answer: "Task diferido cancelable → Result" — "Unidad diferida que devuelve Result: pánico→Fail, cancelación con join, sin leaks. Lección 'async done right', compone con flow y Try." (rejected: Future-handle with Await; parking F4 unbuilt).
- S2 (scope, program row + decision): new `task` package. Deferred async computation with cancel that returns `result.Result`; transform panic becomes `Fail` (Try precedent — deliberate contrast with F3b, where transform panic re-panics in the puller); cancellation joins all goroutines, none leak (eventual-goroutine-baseline test with deadline, F3b precedent); composes with `flow` (bounded fan-out over tasks) and `Try` (sync panic boundary). Exact API surface (type/function names, start semantics, await shape) is pinned in T1 design lock — no invented exports before sign-off.
- S3 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green; TDD RED→GREEN recorded.
- S4 (house style, program S5): Which-table row + godoc + compiling Example per export; element-only callbacks; purity (frozen decision 8); OrNull n/a (`Result` is the absence-and-failure channel); panic→Fail per S1 (not F3b re-panic); no reflection; stdlib-gap justification recorded (goroutines + channels + errgroup exist — what `task` adds is the cancel+panic+leak contract in one teachable unit); edge + never-mutate pins in tests.

## Tasks

- T1 (S1, S2, S4): Design lock — verifiable requirements (MUST/SHOULD/MAY + Given/When/Then) + exact API shape + lesson note + Which-table template; owner sign-off — route: parent + foundry-requirements + cognitive-doc-design — commit: ODD-only (this document, branch feat/axiom-f4-task)
- T2 (S1-S4): Implement Task with tests + Which-table + Example — route: bounded writer — surfaces: task/task.go, task/task_test.go, task/example_test.go — skills: foundry + foundry-testing + cognitive-doc-design — commit: f3a27f6
- T3 (S3): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated (new issue + approval) — commit: pending
- T4 (S3): Delivery — route: parent; issue type:feature + status:approved + PR Closes #N type:feature + CI green; merge owner-gated — commit: pending

## Log

- L1 (request verbatim): "Larga F4"
- L2 (direction lock verbatim, owner answer 2026-10-22): "Task diferido cancelable → Result" ("Unidad diferida que devuelve Result: pánico→Fail, cancelación con join, sin leaks. Lección 'async done right', compone con flow y Try."). Program gate it satisfies: "weakest stdlib-first story; needs the strongest lesson justification" — justification is the cancel+panic+leak contract (async done right) as DLC lesson material, stdlib-first core (context + sync + result only).
- L3 (2026-10-22): classified large (new package + new abstraction, resume test fails); rama feat/axiom-f4-task desde main (post-#60); ODD + mirror + todo reconciled before first write; exact API surface explicitly NOT invented here — pinned in T1 with owner sign-off.
- L4 (2026-10-22, T1 DONE + owner sign-off sin flips): eager + idempotente + pánico→Fail (política exacta Try) + nil-compute/zero-Task panic; API: `task.Task[Output]` + `Spawn(active, compute)` + `Await() Result`; requirements R1-R6 + edges + out-of-scope + lesson note abajo en ## Design; T2 a writer acotado (task/task.go, task/task_test.go, task/example_test.go).
- L5 (2026-10-22, recomendación Catalyst vía owner): hueco F4 = "el capítulo async pendiente"; recomienda opción 1 (handle de lifecycle, anti-mónada); sin F4 el DLC termina sin el capítulo "ninguna goroutine sin dueño", el único que enseña disciplina y no solo herramientas. Confirma el lock de L2 sin cambiarlo: `Task` es handle de lifecycle (Spawn/Await/join) y `Then`/`Map` ya están fuera de alcance (anti-mónada). Lección titulada en ## Design.
- L6 (2026-10-22, T2 DONE + verifier Go + close-out): work-unit f3a27f6 (436+/0, task/* 3 files, suite 20/20 -count=1 ok, vet 0, gofmt vacío, lint 0, bench smoke PASS slice+flow 1x); verifier mvk5tqwy-6-p78w Go all S# (0 BLOCKER, 1 bench WARNING closed by parent smoke line, 1 nil-ctx-test SUGGESTION parked with reason: deterministic gap, zero behavior risk, no perturbar la unidad verificada); slice boundary: 485 authored lines over ~400 heuristic = single logical unit (new package), no more code commits — delivery strategy applies. T4 delivery (new issue + approval + PR + CI, merge owner-gated).
- L7 (2026-10-22, native review quemado a pedido owner): lineage review-bbfc21782dbf0d65 APPROVED + burned (resilience+readability+reliability, 488 líneas, group 3/3 limpio via pi_host_relay); 8 advisories informativas parkeadas sin corrección (R2-001/002/003, R3-001 WARNING + R3-002/003, R4-001 WARNING + R4-002); merge owner-gated (PR pendiente de #61 approved).

## Design (T1 sign-off, 2026-10-22)

API exacta:

```go
package task

type Task[Output any] struct { /* opaco: done chan, once, outcome */ }

func Spawn[Output any](active context.Context, compute func(context.Context) (Output, error)) Task[Output]
func (future Task[Output]) Await() result.Result[Output]
```

Contrato:

- R1 eager: `Spawn` lanza `compute` en goroutine antes de retornar. Check: Given spawneado, When 50ms sin `Await`, Then compute corrió.
- R2 idempotente: todo `Await` (N goroutines, antes/después) devuelve el `Result` igual (error con `errors.Is`).
- R3 pánico→Fail política `Try` exacta (`axiom result: recovered panic: %w|%v`). Check: payloads error y no-error.
- R4 cancel+join sin leaks: ctx cancelado → `Await` da `Fail(ctx.Err())`; test eventual-goroutine-baseline con deadline. Ctx pre-cancelado: `compute` nunca corre.
- R5 stdlib-only: `context` + `sync` + `result`.
- R6 Which-table: background single result → `task.Spawn`/`Await`; sync panic boundary → `Try`; parallel fan-out → `flow.Map`; lazy ordered stream → `flow.MapSeqResults`; dedup concurrente → `singleflight`.
- Edges: `compute == nil` → panic en `Spawn`; `Await` sobre `Task` cero → panic.
- Out: combinadores `Then`/`Map`, timeouts (ctx deadlines), `TaskGroup` (errgroup/`flow`), buffers tuneables, métricas.
- Lección DLC: "ninguna goroutine sin dueño" — async done right: todo cómputo spawneado se joinea, todo pánico se vuelve dato, toda cancelación se observa, todo resultado se comparte (vs goroutine+channel un receptor/join manual/pánico fatal; errgroup scope grupal; singleflight solo dedup). Capítulo de disciplina, no de herramientas.
