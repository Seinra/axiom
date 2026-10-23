# Axiom mathx IsPrime round — new package, IsPrime sure, Gcd/Lcm gated

New `mathx` package for scalar integer helpers. `IsPrime` ships; `Gcd`/`Lcm` only if they re-pass the program gate in T2.

## Specs

- S1 (pending-round P2 verbatim 2026-10-23): "`mathx`? package: `IsPrime` (+`Gcd`/`Lcm` weak) (program rank 14). Scope: undecided — new `mathx` package vs existing home. Lesson: primes sieve with `Range`+`Generate`. Gate: EXPLICIT owner package-creation call before any ODD (program rule for new packages)."
- S2 (owner package call verbatim): selected "Paquete mathx nuevo" to "¿Dónde vive?"; scope agreed verbatim: "P2 arranca con `IsPrime` seguro y revalidamos `Gcd`/`Lcm` en el ODD" + "Me cierra". Architect verdict recorded: new package correct (no existing home fits scalar integer arithmetic; `slice` is helpers over `[]Element`); `Gcd`/`Lcm` weak (trivial loops, thin lesson) → T2 verdict.
- S3 (gates, standing): `go test -count=1 ./...` green, `go vet ./...` 0, `gofmt` empty, lint 0; TDD RED→GREEN recorded; verifier Go before delivery (small additive, no high-risk item → inline verify); native review only if owner orders; delivery needs new issue + approval, push/PR/merge owner-gated.
- S4 (house style, program S5): scalar funcs (element-only callbacks n/a); package doc + godoc + compiling Example per export; purity (deterministic, nothing to mutate); stdlib-gap justification recorded (`math`/`slices` have no prime test); edge pins in tests (negatives, 0, 1, 2, small/large primes and composites).

## Explicitly NOT in this round

- `Gcd`/`Lcm` ship only on a T2 gate pass; otherwise parked with reason (no silent inclusion).
- No other `mathx` surface (scalar Abs/Max/Min and friends are stdlib-owned or unscoped).

## Tasks

- T0 (S1, S2): ODD + mirror + todo + branch `feat/axiom-mathx-isprime` from main — route: parent inline — commit: pending (with work-unit)
- T1 (S2-S4): `IsPrime` + edge tests + Example + package doc — route: parent inline, TDD RED→GREEN — commit: pending
- T2 (S2-S4): `Gcd`/`Lcm` gate verdict (ship or park with reason) — route: parent inline — commit: pending (same unit if shipped)
- T3 (S3): full verification (suite + vet + gofmt + lint, Codegraph blast radius) — route: parent inline — commit: n/a
- T4 (S3): delivery (issue type:feature + status:approved + PR Closes + CI green; merge owner-gated) — route: parent — commit: n/a

## Log

- L1 (request verbatim): "Me cierra" (scope: "P2 arranca con `IsPrime` seguro y revalidamos `Gcd`/`Lcm` en el ODD", package call: "Paquete mathx nuevo").
- L2 (prior verbatims): "Pero arma el slice para el futuro asi los atacamos por orden de prioridad"; "Te parece correcto darle paquete nuevo en lugar de usar uno existente?".
- L3 (2026-10-23, T0 done): rama `feat/axiom-mathx-isprime` desde main; ODD + mirror + todo reconciled antes del primer write; nada implementado.
- L4 (2026-10-23, T1-T3 done): RED observado dos veces (`no non-test Go files`, `undefined: Gcd/Lcm`) → GREEN 13 tests + 3 Examples; suite 21/21 -count=1, vet 0, gofmt vacío, blast radius solo tests nuevos; lint local limitado por binario v2.13.2 (error ambiental en orderedmap no tocado), gate real = CI v2.15.0; T2 veredicto SHIP con lecciones Euclid + overflow-safe Lcm; sin commitear, T4 a orden owner.
