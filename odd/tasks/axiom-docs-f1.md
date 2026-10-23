# Axiom docs T1 — README count + flow YAGNI row (future line waits owner wording)

Fixes freshness findings F1-count and F2. The Declared-future reword is owner-gated text: fail-closed, it waits for the owner's exact wording and is NOT touched here.

## Specs

- S1 (audit findings verbatim 2026-10-23, spot-check confirmed): F1 — "`README.md` count (19 vs 21 lib pkgs: `mathx`, `strings` missing)"; F2 — "`docs/guides/flow.md` YAGNI row ('no retry / no backoff') contradicting `flow/retry.go` (Retry/RetryWithResult)". Evidence: `README.md:68`, `docs/guides/flow.md:64`.
- S2 (launch + scope split verbatim): "Commitea y arranca T1". Fail-closed split: README count + flow YAGNI row ship now; the Declared-future cell (`README.md:71`) stays untouched until the owner supplies the exact replacement text (gate row 5, like protected labels).
- S3 (gates, standing): prose-only → build gate + review, no RED invented; verifier with review pack before delivery; native review never runs on docs-only; delivery needs new issue + approval, push/PR/merge owner-gated.
- S4 (house style): minimal cells, mirror godoc facts (`mathx`/`flow` package docs), no rewording beyond the cited lines.

## Tasks

- T0 (S1, S2): ODD + mirror + branch `feat/axiom-docs-f1` from main — route: parent inline — commit: pending
- T1 (S2, S4): README "Ships now (19 pkgs)" → 21 pkgs + `mathx`, `strings` in the list — route: parent inline — commit: pending
- T2 (S2, S4): `flow.md` YAGNI row drops retry/backoff from out-of-scope — route: parent inline — commit: pending
- T3 (S3): verification (grep read-backs + docs build gate) + memory — route: parent inline — commit: n/a

## Log

- L1 (request verbatim): "Commitea y arranca T1".
- L2 (2026-10-23, T0 done): rama `feat/axiom-docs-f1` desde main; ODD + mirror reconciled antes del primer write; Declared-future cell explicitly OUT (blocked on owner exact wording per gate row 5).
- L3 (2026-10-23, unblocked): owner chose reframing + ordered rewrite with foundry-docs/cognitive-doc-design; cell rewritten (program closed, P1–P5 listed, gate rule kept) and applied; read-back confirmed.
