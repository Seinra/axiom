# Axiom F3 — fan-in/out pipeline over iter.Seq, B1 shape (declared-future slice F3)

Third declared-future slice from the program (`odd/tasks/axiom-future-slices.md`): bounded fan-out + ordered fan-in over `iter.Seq`, errgroup parity with `flow`. Lifts flow's YAGNI ban on pipelines surgically, inside `flow`. Lesson: concurrency you can reason about — prompt failure over unbounded input. H shape (lazy `iter.Seq[result.Result]` out) staged as F3b, not this slice.

## Specs

- S1 (owner orders verbatim 2026-10-22): "Vamos con los slices declarados" + "B1 ahora, H después" (B1: streaming-in via feeder, eager ordered out, fail-fast; H staged as F3b follow-up, not killed).
- S2 (scope, parent 2026-10-22): `flow` package, two functions, no new types, no builder.
  - `MapSeq[Input, Output](active context.Context, sequence iter.Seq[Input], workerCount int, transform func(context.Context, Input) (Output, error)) ([]Output, error)`: drains `sequence` through a feeder goroutine into unbuffered positions; workers bounded by `workerCount` (`≤0` falls back to `runtime.NumCPU`, flow convention); output materialized in sequence order; in-flight work bounded by the runner count.
  - Fail-fast errgroup parity with `Map`: first transform failure cancels siblings AND the feeder, wins as the returned failure; unstarted inputs never run after the first failure; in-flight transforms run to completion.
  - Feeder selects on `ctx.Done` on every send (a canceled run never wedges the feeder); every spawned goroutine joins via `WaitGroup` on all paths (flow precedent).
  - Nil sequence yields empty non-nil output without spawning (ranging over a nil `iter.Seq` would panic: guard it); empty sequence yields the same result shape (spawn is unavoidable: emptiness is only knowable by pulling); already-canceled context returns `(nil, active.Err())` and starts nothing.
  - Panic policy inherited from `Map`: first payload latches, siblings cancel, caller re-panics; panic wins over failure.
  - Transform receives the cancellable context and must honor it (flow contract).
  - `EachSeq[Input](active context.Context, sequence iter.Seq[Input], workerCount int, consume func(context.Context, Input) error) error`: fully streaming O(1) memory side-effect pass with the same fail-fast contract; implemented via `MapSeq` (`ForEach`-via-`Map` precedent).
  - Godoc states the shape honestly: input streams, output materializes (anti-laziness-trap disclaimer); stdlib-gap justification (no parallel ordered fan-in over `iter.Seq`); lesson (prompt failure + cancellation over unbounded inputs).
  - Out: lazy-out (H/F3b), fluent builder (anti-idiomatic precedent stands), `FilterSeq`, tunable buffers, unbounded workers.
- S3 (house style, program S5): Which-table rows for `MapSeq`/`EachSeq` + godoc + one compiling Example per function; full names; OrNull n/a (failures are errors, errgroup parity); panic policy above; no reflection; stdlib (`context`, `sync`) only.
- S4 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green; TDD RED→GREEN recorded.

## Tasks

- T1 (S1, S2, S3): Implement MapSeq + EachSeq with tests + Which-table rows + Examples — route: bounded writer — surfaces: flow/flow.go, flow/flow_test.go, flow/example_test.go — skills: foundry + foundry-testing + cognitive-doc-design — commit: 89706ff
- T2 (S4): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated (new issue + approval) — commit: pending

## Log

- L1 (request verbatim): "Vamos con los slices declarados"
- L2 (shape decisions verbatim): "B1 ahora, H después" (after "Estas 100% seguro de que no hay algo mejor que lo que propones en B1?" — challenger H steelmanned: lazy `iter.Seq[result.Result]` out is the fuller F3 letter but 3-5x machinery with reorder-buffer + abandon-join + policy forks; sequenced as F3b once B1's mental model lands, coexistence safe via distinct names).
- L3 (2026-10-22): rama feat/axiom-f3-pipeline desde main (post-#54); ODD + mirror listos; T1 a writer acotado; delivery needs new issue + approval.
- L4 (2026-10-22, T1 DONE writer mvjg2ruy-7-zqma + parent-verified): MapSeq + EachSeq B1 (feeder+unbuffered, fail-fast parity, inherited panic policy); parent re-run suite 19/19 -count=1 ok, vet 0, gofmt vacío, lint 0, -race -count=2 flow ok, bench smoke PASS; parent found + fixed mid-run-cancel silent holes → (nil, ctx.Err()) + test (deliberate parity deviation from Map, Map untouched); verifier mvjggkzv-8-8q5h Go (0 BLOCKER, 1 empty-shape SUGGESTION applied, 1 generator-panic WARNING parked: feeder-crash on generator panic needs ordering decision, loud-never-silent so no urgency); work-unit 89706ff (636+/1-, single logical unit over the 400-line heuristic). T2 close + delivery (new issue + approval, native review owner-call).
