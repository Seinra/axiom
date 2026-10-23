# Axiom docs T3 — Which-table rows for the 8 landed exports

Mirrors godoc into the three family guides: slice (Head/Tail/Initial), opt (To/Deref/Coalesce), flow (Retry/RetryWithResult). No library code.

## Specs

- S1 (audit finding F5 verbatim 2026-10-23): "slice.md/opt.md/flow.md Which-tables omit the 8 landed exports in their own families (godoc covers them)".
- S2 (launch order verbatim): "Arranca" (parked slice T3; S4 house style: Which rows mirror godoc).
- S3 (gates, standing): prose-only → build gate + review, no RED invented; verifier with review pack (foundry-review, never foundry-craft) before delivery; native review never runs on docs-only; delivery needs new issue + approval, push/PR/merge owner-gated.
- S4 (house style): one row per export, wording mirrors the godoc Which-table; no new prose beyond the rows; second person active present where touched.

## Explicitly NOT in this round

- No new guides, no index or quickstart changes (T1/T2 landed those).
- No rewording of existing rows (mirror only).

## Tasks

- T0 (S1, S2): ODD + mirror + todo + branch `feat/axiom-docs-t3` from main — route: parent inline — commit: pending
- T1 (S2-S4): `slice.md` 1 row (Head/Tail/Initial) — route: parent inline — commit: pending
- T2 (S2-S4): `opt.md` 3 rows (To/Deref/Coalesce) — route: parent inline — commit: pending
- T3 (S2-S4): `flow.md` 2 rows (Retry/RetryWithResult) — route: parent inline — commit: pending
- T4 (S3): verification (read-backs + docs build gate) + memory — route: parent inline — commit: n/a

## Log

- L1 (request verbatim): "Arranca" (parked slice T3).
- L2 (2026-10-23, T0 done): rama `feat/axiom-docs-t3` desde main; ODD + mirror reconciled antes del primer write; nada implementado.
- L3 (2026-10-23, T1-T4 done): 6 filas espejo aplicadas y verificadas por read-back (slice:61, opt:54-56, flow:34-35) + docs suite verde; trabajo inline del parent (unidad mínima, sin triggers de delegación); sin commitear, delivery a orden owner.
