# Axiom docs T2 — mathx/strings family guides + index rows

Closes freshness findings F3+F4: two family guides in house shape with CI-executed tests, plus README guide-index and quickstart family rows. No library code.

## Specs

- S1 (audit evidence verbatim 2026-10-23): F3 — "no `docs/guides/mathx.md`, no `docs/guides/strings.md` (6 exports godoc-only)"; F4 — "README guide-index + `docs/quickstart.md` family tables with no `mathx`/`strings` rows (discoverability dead-end)".
- S2 (launch order verbatim): "Arranca" (T2 of the parked slices; S4 house shape per slice).
- S3 (gates, standing): prose-only → build gate + review, no RED invented; verifier with review pack (foundry-review, never foundry-craft) before delivery; native review never runs on docs-only; delivery needs new issue + approval, push/PR/merge owner-gated.
- S4 (house style, program S5 + trio): guide = concept → steps → Which-table (mirrors godoc) → runnable Example (adapted from the package `Example*`, same behavior) → presence/lesson section → next step; CI-executed test per guide in `docs/guides_test.go` style (new test funcs, same pins); second person active present; new rows appended after `task` (age order).

## Explicitly NOT in this round

- No Which-table backfills for slice/opt/flow (that's T3, separate slice).
- No rewrites of existing guides beyond appending rows where the template requires (index table only).

## Tasks

- T0 (S1, S2): ODD + mirror + todo + branch `feat/axiom-docs-t2` from main — route: parent inline — commit: pending
- T1 (S2-S4): `docs/guides/mathx.md` + executed test — route: parent inline — commit: pending
- T2 (S2-S4): `docs/guides/strings.md` + executed test — route: parent inline — commit: pending
- T3 (S2-S4): README index rows + quickstart family rows for both — route: parent inline — commit: pending
- T4 (S3): verification (docs suite + vet + gofmt + read-backs) + memory — route: parent inline — commit: n/a

## Log

- L1 (request verbatim): "Arranca" (T2 of parked slices).
- L2 (2026-10-23, T0 done): rama `feat/axiom-docs-t2` desde main; ODD + mirror reconciled antes del primer write; shape copiado de `stack.md` + `guides_test.go` precedent.
- L3 (2026-10-23, T1-T3 done): guías mathx/strings + 2 tests ejecutados PASS + index/quickstart rows; docs suite green, vet 0, gofmt vacío; sin commitear, delivery a orden owner.
