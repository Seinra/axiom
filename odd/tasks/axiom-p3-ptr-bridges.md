# Axiom P3 ptr bridges — Deref/Coalesce/To in opt (home decided)

Pointer↔value bridges in the existing `opt` package (owner home call). No new
package: `opt` is nil-safety at the edges, coherent with its OrNull shapes.

## Specs

- S1 (pending-round P3 verbatim 2026-10-23): "pointer↔value bridges: `Deref`/`Coalesce`/`To` (program rank 15). Scope: undecided — `opt` vs new package. Lesson: 'nil-safety at the edges', OrNull-shaped. Gate: EXPLICIT owner home call (opt vs new package) before any ODD."
- S2 (owner home call verbatim): question "¿Dónde viven Deref/Coalesce/To?" with recommendation "opt es el home natural (nil-safety at the edges, sus formas OrNull)"; selected "Home opt existente"; launch order "Ok, vamos con el siguiente". Scope: `To` (T→*T fresh), `Deref` (*T→(T,bool) OrNull), `Coalesce` (...*T→ first non-nil, nil if none).
- S3 (gates, standing): `go test -count=1 ./...` green, `go vet ./...` 0, `gofmt` empty, lint 0; TDD RED→GREEN recorded; verifier Go before delivery (small additive, no high-risk item → inline); native review only if owner orders; delivery needs new issue + approval, push/PR/merge owner-gated.
- S4 (house style, program S5): follow `opt` package conventions + godoc + compiling Example per export; purity (fresh pointer from `To`, copy from `Deref`, documented aliasing of `Coalesce` result); stdlib-gap justification recorded (no scalar ptr helpers in stdlib); edge pins in tests (nil, nil-chain, aliasing: `To` result independent, `Coalesce` result aliases the winning input).

## Explicitly NOT in this round

- No new package (`ptr` or otherwise stays rejected for this scope).
- No `DerefOr`/fallback variants unless the ODD uncovers a gap (no silent scope creep).

## Tasks

- T0 (S1, S2): ODD + mirror + todo + branch `feat/axiom-p3-ptr-bridges` from main — route: parent inline — commit: pending (with work-unit)
- T1 (S2-S4): `To` + `Deref` + `Coalesce` + tests + Examples in `opt` — route: parent inline, TDD RED→GREEN — commit: pending
- T2 (S3): full verification (suite + vet + gofmt + lint, Codegraph blast radius incl. `opt` callers) — route: parent inline — commit: n/a
- T3 (S3): delivery (issue type:feature + status:approved + PR Closes + CI green; merge owner-gated) — route: parent — commit: n/a

## Log

- L1 (request verbatim): "Ok, vamos con el siguiente" (launch P3).
- L2 (home call verbatims): "¿Dónde viven Deref/Coalesce/To?"; recommendation "opt es el home natural (nil-safety at the edges, sus formas OrNull)"; selected "Home opt existente".
- L3 (2026-10-23, T0 done): rama `feat/axiom-p3-ptr-bridges` desde main; ODD + mirror + todo reconciled antes del primer write; nada implementado.
- L4 (2026-10-23, T1-T2 done): RED observado por comando (`undefined: opt.To/Deref`) → GREEN 8 tests + 3 Examples; Which-table en `opt.go`; suite 21/21 -count=1, vet 0, gofmt vacío, blast radius solo tests nuevos; lint local limitado por binario v2.13.2 (mismo error ambiental en orderedmap no tocado), gate real = CI v2.15.0; sin commitear, T3 a orden owner.
