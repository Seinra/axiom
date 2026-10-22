# Axiom F1 — Stack/Queue (declared-future slice F1)

First declared-future slice from the program (`odd/tasks/axiom-future-slices.md`): LIFO/FIFO over slices, pure constructors + methods, godoc + Which-table + Example. Lesson: data structures 101, "structure your work". stdlib has `container/list` (untyped) and no generic stack/queue.

## Specs

- S1 (owner order verbatim 2026-10-22): "Vamos con los slices declarados" (F1 first per program order adoption → F1 → F2 → F3 → F4; one ODD per slice).
- S2 (scope, parent 2026-10-22): `stack` package: `Stack[Element any]` + `New() *Stack[Element]`; methods `Push(element Element)`, `Pop() (Element, bool)`, `Peek() (Element, bool)`, `Len() int`, `IsEmpty() bool`. `queue` package: `Queue[Element any]` + `New() *Queue[Element]`; methods `Enqueue(element Element)`, `Dequeue() (Element, bool)`, `Peek() (Element, bool)`, `Len() int`, `IsEmpty() bool`.
  - Absence is OrNull `(T,bool)`: empty `Pop`/`Dequeue`/`Peek` yield `(zero, false)` — never an error, never a panic.
  - Zero value ready: `var container stack.Stack[int]` (addressable) works without `New`; methods handle nil backing.
  - Ownership: slice-backed; methods mutate only the receiver's own backing (container semantics, `MemoryRepository.Store` precedent) — caller slices passed anywhere are never retained-aliased... constructors take no slices, so no aliasing surface at all.
  - `Dequeue` amortized O(1) via head offset with documented compaction policy (writer implements, godoc documents, tests pin order — never timing).
  - Do-not-copy documented (`pool.Pool` precedent: slice header shares backing).
  - No other containers, no `container/list` interop, no concurrency (F3 owns that).
- S3 (house style, program S5): package doc with Which-table + godoc carrying stdlib-gap justification + compiling Example per type; full names; OrNull shapes; no reflection; no new dependencies (stdlib only).
- S4 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green; TDD RED→GREEN recorded.

## Tasks

- T1 (S1, S2, S3): Implement stack + queue packages with tests + Which-table rows + Examples — route: bounded writer — surfaces: stack/stack.go, stack/stack_test.go, stack/example_test.go, queue/queue.go, queue/queue_test.go, queue/example_test.go — skills: foundry + foundry-testing + cognitive-doc-design — commit: 7ee6db7
- T2 (S4): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated (new issue + approval) — commit: pending

## Log

- L1 (request verbatim): "Vamos con los slices declarados"
- L2 (2026-10-22): rama feat/axiom-f1-stack-queue desde main; F1 first per program order; F2-F4 staged with their gate notes (F2 LRU observability decision; F3 cancellation+backpressure policy; F4 strongest lesson justification); ODD + mirror listos; T1 a writer acotado; delivery needs new issue + approval.
- L3 (2026-10-22, T1 DONE writer mvj8trjb-3-qy9r + parent-verified): stack + queue packages (OrNull, zero-value ready, receiver-only mutation, documented compaction); parent re-run suite 18/18 -count=1 ok, vet 0, gofmt vacío, lint 0, bench smoke PASS; parent read both impl files fully; verifier mvj8xreb-4-rikp Go (0 BLOCKER, 0 SUGGESTION, 1 WARNING on lint/bench authorization only — both covered by parent evidence); work-unit 7ee6db7 (588+/0, single logical unit over the 400-line heuristic). T2 close + delivery (new issue + approval, native review owner-call).
