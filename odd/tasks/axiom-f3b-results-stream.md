# Axiom F3b — lazy collect-all Result stream (MapSeqResults)

F3 follow-up from the staged H shape (`odd/tasks/axiom-f3-pipeline.md` L2): lazy ordered fallible parallel pipeline over `iter.Seq`, collect-all semantics. Approved with the Aegis consumer's requirements as FLOOR, never ceiling (owner: minimums bind from below, craft decides from above). Lesson: validate-all streaming — every outcome accounted, never silent, never truncated.

## Specs

- S1 (owner orders verbatim 2026-10-22): "Vamos con F3b" (with "Al menos debe cumplir con eso, esto es un 'al menos' eh." and "eso son los minimos, si resolvemos 'ponerle mas' porque corresponde, se hace").
- S2 (scope, parent 2026-10-22, Aegis minimums as floor): `flow` package, one function, no new types, no builder.
  - `MapSeqResults[Input, Output any](active context.Context, sequence iter.Seq[Input], workerCount int, transform func(context.Context, Input) (Output, error)) iter.Seq[result.Result[Output]]` — name never bare `MapSeq` (that symbol is the fail-fast materialized one).
  - Lazy: nothing runs until pulled (call-counter pinned); bounded fan-out (`workerCount ≤0` falls back to `runtime.NumCPU`, flow convention); ordered fan-in through a bounded reorder buffer with documented fixed capacity (not tunable); in-flight + buffered stay bounded while consumed streaming.
  - Collect-all: a `Fail` never cancels siblings; every consumed input yields exactly one `Result` (`Ok` or `Fail`), nth output ↔ nth input, pinned.
  - Cancellation: `active.Done()` cancels everything and joins all goroutines; downstream early break (yield returns false) cancels workers + feeder and joins them — no leaked goroutines, pinned by an eventual-goroutine-baseline test with deadline.
  - Pre-canceled context yields an empty stream (first pull gets nothing); nil sequence yields an empty stream without spawning.
  - Panic policy: a transform panic latches, cancels siblings, and re-panics in the pulling goroutine (never escapes a worker, never silent); panic wins over `Fail`. A panic inside the user's generator propagates on the puller's stack (loud, same as ranging it directly).
  - `flow` imports `result` for this symbol (verified acyclic: result is stdlib-only).
  - Godoc states the shape honestly: lazy in AND out; `slices.Collect` on it is O(n) as the caller's documented decision; stdlib-gap justification (no lazy fallible parallel streams — `iter.Seq` is infallible by construction); the validate-all lesson.
  - Out: fail-fast flag/mode on this symbol (fail-fast already exists: `Map`, `MapSeq`, `EachSeq`; abort via range-break, which already cancels), `Filter`, tunable buffers, unbounded workers.
- S3 (house style, program S5): Which-table row + godoc + compiling Example (collect-all incl. a `Fail` element); full names; OrNull n/a (`Result` is the absence-and-failure channel); panic policy above; no reflection; `context` + `sync` + `result` only.
- S4 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green; TDD RED→GREEN recorded.

## Tasks

- T1 (S1, S2, S3): Implement MapSeqResults with tests + Which-table row + Example — route: bounded writer — surfaces: flow/flow.go, flow/flow_test.go, flow/example_test.go — skills: foundry + foundry-testing + cognitive-doc-design — commit: 4d93080
- T2 (S4): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated (new issue + approval) — commit: pending

## Log

- L1 (request verbatim): "Vamos con F3b"
- L2 (requirements verbatim, Aegis agent via owner, FLOOR): collect-all (i) closed the fork; name MapSeqResults/MapAllSeq never bare MapSeq; Fail never cancels siblings; ctx-Done + break cancel + join; tests pin order + exactly-one-Result + lazy + break-without-leaks; godoc O(n)-Collect note; no builders/Filter/tunable-buffers, context+sync only. Engram aegis/5142 verified real. Floor-not-ceiling per owner reaffirmation.
- L3 (2026-10-22): rama feat/axiom-f3b-results-stream desde main (post-#56); ODD + mirror listos; T1 a writer acotado; delivery needs new issue + approval.
- L4 (2026-10-22, T1 DONE writer mvjhhlcc-9-wohj + parent-verified): MapSeqResults collect-all lazy (drive-on-caller-stack, bounded window 64, buffered reply slots); TDD RED (undefined flow.MapSeqResults, compile FAIL) → GREEN (13 tests + Example + suite) per writer handoff; parent re-run suite 19/19 -count=1 ok, vet 0, gofmt vacío, -race -count=2 flow ok, smoke PASS, full impl read; parent found gocyclo-11 BLOCKER (writer skipped lint) → extracted seqResultsPump, re-green (full lint 0, suite, vet, race); verifier mvjhvbxj-a-6nqn No-Go on same BLOCKER + TDD-record WARNING → both closed here; work-unit 4d93080 (792+/11-, single logical unit over the 400-line heuristic). T2 close + delivery (new issue + approval, native review owner-call).
