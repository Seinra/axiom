# Axiom P4 retry — Retry/RetryWithResult in flow (home decided)

Bounded retries in the existing `flow` package (owner home call). Pairs with
`flow.Race`; `di` retries constructors, no overlap.

## Specs

- S1 (pending-round P4 verbatim 2026-10-23): "`Retry`/`RetryWithResult` (program rank 16, go-utils excutils). Scope: undecided — `flow` vs standalone. Lesson: resilience, pairs with `flow.Race` (di retries constructors already — no overlap). Gate: EXPLICIT owner home/shape call before any ODD."
- S2 (owner home call verbatim): question "¿Dónde vive?" with "Mi lectura: flow es el home natural, junto a Race — misma familia de resiliencia, y di ya reintenta constructores sin pisarse"; selected "Home flow existente"; launch order "Ok, vamos". Scope: `Retry` (error-only) + `RetryWithResult` ([T any], value + error) in `flow`, bounded attempts with backoff, context-aware like `flow.Race`/`flow.Map`.
- S3 (gates, standing): `go test -count=1 ./...` green, `go vet ./...` 0, `gofmt` empty, lint 0; TDD RED→GREEN recorded; verifier Go before delivery (small additive, no high-risk item → inline); native review only if owner orders; delivery needs new issue + approval, push/PR/merge owner-gated.
- S4 (house style, program S5): follow `flow` conventions (context threading, cancellation, errgroup parity where it applies) + godoc with stdlib-gap justification + compiling Example per export; purity (no shared mutable state); edge pins in tests (zero/one/max attempts, immediate success, always-fail, context cancel, backoff bound).

## Explicitly NOT in this round

- No standalone `retry` package (decided: `flow` home).
- No unbounded retries, no jitter policies beyond a fixed backoff (no scope creep without a new gate).

## Tasks

- T0 (S1, S2): ODD + mirror + todo + branch `feat/axiom-p4-retry` from main — route: parent inline — commit: pending (with work-unit)
- T1 (S2-S4): `Retry` + `RetryWithResult` + tests + Examples in `flow` — route: parent inline, TDD RED→GREEN — commit: pending
- T2 (S3): full verification (suite + vet + gofmt + lint, Codegraph blast radius incl. `flow` callers) — route: parent inline — commit: n/a
- T3 (S3): delivery (issue type:feature + status:approved + PR Closes + CI green; merge owner-gated) — route: parent — commit: n/a

## Log

- L1 (request verbatim): "Ok, vamos" (launch P4).
- L2 (home call verbatims): "P4 exige tu call explícito de home/forma antes de cualquier ODD (flow vs standalone). Mi lectura: flow es el home natural, junto a Race — misma familia de resiliencia, y di ya reintenta constructores sin pisarse."; selected "Home flow existente".
- L3 (2026-10-23, T0 done): rama `feat/axiom-p4-retry` desde main; ODD + mirror + todo reconciled antes del primer write; nada implementado.
- L4 (2026-10-23, T1-T2 done): RED observado por comando (`undefined: flow.Retry`) → GREEN 11 tests + 2 Examples; package doc actualizado (sale el out-of-scope de Retry); suite 21/21 -count=1, vet 0, gofmt vacío, blast radius solo tests nuevos; lint local limitado por binario v2.13.2 (mismo error ambiental en orderedmap no tocado), gate real = CI v2.15.0; sin commitear, T3 a orden owner.
