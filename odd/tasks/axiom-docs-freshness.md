# Axiom docs freshness — parked slices from the P1-P5 aftermath audit

Read-only audit found 7 findings (scout + parent spot-check confirmed, no writes). They launch as slices below in rank order unless the owner reorders at launch. Finding 6 is already fixed by the loop edit in this round; finding 7 needs no action.

## Specs

- S1 (audit order verbatim 2026-10-23): "Ok, ahora hace una auditoria nueva sobre docs, a ver si queda algo desactualizado que toque arreglar. Inyecta bien las skills".
- S2 (parked inventory, verified against the tree 2026-10-23 — scout `path:line` evidence + parent spot-check):
  - F1 (rank 1, STALE) — `README.md` count (19 vs 21 lib pkgs: `mathx`, `strings` missing) + Declared-future line listing delivered work (`mathx`?, bridges, strings helpers, retry). First page every reader trusts.
  - F2 (rank 2, STALE) — `docs/guides/flow.md` YAGNI row ("no retry / no backoff") contradicting `flow/retry.go` (Retry/RetryWithResult). Teaches a constraint the tree violates.
  - F3 (rank 3, MISSING) — no `docs/guides/mathx.md`, no `docs/guides/strings.md` (6 exports godoc-only).
  - F4 (rank 4, MISSING) — README guide-index + `docs/quickstart.md` family tables with no `mathx`/`strings` rows (discoverability dead-end).
  - F5 (rank 5, MISSING) — `slice.md`/`opt.md`/`flow.md` Which-tables omit the 8 landed exports in their own families (godoc covers them).
  - F6 (rank 6, STALE, FIXED this round) — `loop-config.yaml` `currentFeature` pointed at the closed guides-missing round; now `odd/tasks/axiom-docs-freshness.md`.
  - F7 (rank 7, NO ACTION) — lint `v2.13.2` mentions live only in ODD Log history; tree pin (`ci.yml` v2.15.0) is current.
- S3 (standing, applies at launch): each slice gets its own ODD (spec + mirror + todo before first write); prose-only → build gate + review, no RED invented; verifier with review pack (foundry-review, never foundry-craft) before delivery; native review never runs on docs-only; delivery needs new issue + approval, push/PR/merge owner-gated.
- S4 (house style, program S5 + trio): guides ship concept → steps → Which-table → runnable Example → next step + CI-executed test (foundry + foundry-docs + cognitive-doc-design trio per REVIEW_PACKS docs units); index rows mirror godoc; second person active present.

## Explicitly NOT in these slices

- `foundations.md` claims nothing false (audit verified) — untouched unless a slice proves otherwise with evidence.
- Rewrites of ratified prose beyond the cited lines (no silent rewording; owner-gated wording stays owner-gated).

## Tasks (all parked — kickoff order, not execution)

- T0 (S1, S2): loop rule v0.3 + this slice doc + STATUS rows — route: parent + loop-engineering (this round) — commit: pending
- T1 (S3, S4, F1+F2): README two cells + `flow.md` YAGNI row — route: per-slice ODD at launch — commit: pending
- T2 (S3, S4, F3+F4): `mathx`/`strings` family guides (house shape + executed tests) + README index + quickstart rows — route: per-slice ODD at launch — commit: pending
- T3 (S3, S4, F5): Which-table rows for the 8 landed exports across `slice.md`/`opt.md`/`flow.md` — route: per-slice ODD at launch — commit: pending

## Log

- L1 (request verbatim): "Ok, ahora hace una auditoria nueva sobre docs, a ver si queda algo desactualizado que toque arreglar. Inyecta bien las skills".
- L2 (order verbatim): "Bueno, asenta en el loop una regla dura de sync con los docs despues de la implementacion de algo, asi no nos vuelve a ocurrir una 3ra vez este problemita... Arma los slices."
- L3 (2026-10-23, parked): all 7 findings ranked with evidence; F6 fixed by the loop edit; F7 no-action; T1-T3 launch one ODD each when the owner picks them, in rank order unless reordered at launch.
