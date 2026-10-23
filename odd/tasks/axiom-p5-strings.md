# Axiom P5 strings — Words/ChunkString/Truncate in new strings package

Rune-aware string splitting in a NEW `strings` package (owner launch order;
creation call covered by "Manda!" on this pre-agreed scope). Case converters
stay out: they live in `x/text`.

## Specs

- S1 (pending-round P5 verbatim 2026-10-23): "strings package: `Words`/`ChunkString`/`Truncate` (program rank 17, lo + go-utils stringutils). Scope: new strings package (case converters live in x/text — out). Lesson: 'runes vs bytes'. Gate: EXPLICIT owner package-creation call before any ODD."
- S2 (launch + creation call verbatim): "Manda!" proceeds on the pre-agreed scope above (new `strings` package; case converters out). Architect notes recorded: package name `strings` shadows stdlib on import — callers alias when both are needed (documented in package doc); `Truncate` is a pure rune cut with no ellipsis (presentation stays out); `ChunkString` counts runes, never bytes.
- S3 (gates, standing): `go test -count=1 ./...` green, `go vet ./...` 0, `gofmt` empty, lint 0; TDD RED→GREEN recorded; verifier Go before delivery (small additive, no high-risk item → inline); native review only if owner orders; delivery needs new issue + approval, push/PR/merge owner-gated.
- S4 (house style, program S5): follow house conventions (Which-table row + godoc + compiling Example per export, purity, OrNull/empty-non-nil shapes, panic policy for programmer errors, no reflection, stdlib-gap justification recorded, edge pins in tests incl. multibyte runes).

## Explicitly NOT in this round

- Case converters (upper/lower/title live in `x/text` — out, per scope).
- Ellipsis-decorating truncate variants and byte-based splitters (out; the lesson is runes).

## Tasks

- T0 (S1, S2): ODD + mirror + todo + branch `feat/axiom-p5-strings` from main — route: parent inline — commit: pending (with work-unit)
- T1 (S2-S4): `Words` + `ChunkString` + `Truncate` + tests + Examples in `strings` — route: parent inline, TDD RED→GREEN — commit: pending
- T2 (S3): full verification (suite + vet + gofmt + lint, Codegraph blast radius) — route: parent inline — commit: n/a
- T3 (S3): delivery (issue type:feature + status:approved + PR Closes + CI green; merge owner-gated) — route: parent — commit: n/a

## Log

- L1 (request verbatim): "Manda!" (launch P5 on the pre-agreed scope: new `strings` package, case converters out).
- L2 (2026-10-23, T0 done): rama `feat/axiom-p5-strings` desde main; ODD + mirror + todo reconciled antes del primer write; nada implementado.
- L3 (2026-10-23, T1-T2 done): RED observado por comando (`no non-test Go files`) → GREEN 13 tests + 3 Examples; suite 22/22 -count=1, vet 0, gofmt vacío, blast radius solo tests nuevos; lint local limitado por binario v2.13.2 (mismo error ambiental en orderedmap no tocado), gate real = CI v2.15.0; sin commitear, T3 a orden owner.
