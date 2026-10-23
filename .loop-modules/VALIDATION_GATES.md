# VALIDATION_GATES

Per-layer command tables + evidence rules. Run the narrowest scope first, then the package, then the repo.

## Commands

| Layer | Command | Pins what |
| --- | --- | --- |
| Unit (single package) | `go test ./result/ ./opt/ ./slice/` | behavior of the changed package |
| Repo (all packages) | `go test ./...` | tree green after every iteration |
| Vet | `go vet ./...` | stdlib correctness |
| Lint | `golangci-lint run --timeout=5m ./...` | craft-floor + code-style hygiene |
| Format | `gofmt -l .` (empty = clean) | formatting |
| Complexity (advisory) | `gocyclo -over 10 ./...` | functions >10 complexity |
| Docs | `go doc ./...` renders without warnings | godoc on every export |

## Evidence rules

- Every behavior claim carries ≥1 `command: observed result` line (Result Contract `artifacts`).
- A claim without executed evidence caps at WARNING; passing tests that don't cover the changed behavior don't count.
- `go test` runs once per scope (no re-runs to chase flakes); flakes are quarantined with an owning ticket, never ignored.
- Coverage is informational; the gate is behavior (outputs, errors, observable state), never a percentage.
- Prose/config changes use the build gate + review (no RED invented).

## TDD applicability

- Library logic: YES — RED → GREEN → TRIANGULATE with commands recorded.
- Prose/assets/config: NO — state the one-line why (no branch, no money/security path).

## Standing review orders (owner 2026-10-20)

- Independent verifier (read-only, with the skills fitting the unit) roughly every ~400 authored lines.
- The independent verifier always runs before any native review of the same candidate.
- Native review never runs on docs-only candidates (not designed for prose).
- Commits proceed autonomously; pushes stay owner-gated (permission prompt).
- Findings never overwrite each other (owner 2026-10-20, Engram has no dedup rule): every verifier/review batch is a NEW observation with a source- and commit-scoped topic_key (e.g. axiom/findings/<source>-<short-hash>); cross-reference prior batches, never update-in-place.
- Never wait on a subagent without verifying it is running (owner 2026-10-20): check subagent_list_tasks/subagent_status at the decision boundary before assuming background work is in flight.
- Implementations go to bounded workers (owner 2026-10-20, context hygiene): one package per worker, disjoint Allowed edit surfaces, skills injected per unit; the orchestrator scopes, verifies, commits.

## Docs-sync gate (hard, owner 2026-10-23 — second docs-drift round)

Every implementation work-unit lands with its prose already true, or it does not land. Check each row; record the verdict per row in the ODD Log (one L-row: `checked` or `N/A + reason`).

| # | Row | Check | Evidence |
| --- | --- | --- | --- |
| 1 | Package count | `README` "Ships now (N pkgs)" == dirs with non-test `.go` (test-only dirs excluded) | count command output |
| 2 | New package | README index row + quickstart family row + `docs/guides/<family>.md` (house shape) + executed test | diff stat shows all four, or `N/A + reason` |
| 3 | New export, guided family | Which-table row in godoc AND in the family guide (or drop the guide's mirror claim in the same unit) | grep symbol in both files |
| 4 | No contradiction | grep new symbols across `docs/` + `README.md`: any "no X / out-of-scope / YAGNI" line covering them is updated in the same unit | grep output (hits fixed, or zero hits quoted) |
| 5 | Roadmap lines | `Declared future` / roadmap lines naming delivered scope leave in the same unit; owner-gated wording needs the exact owner wording first (fail-closed: land blocked until granted, like protected labels) | diff or grant quote |
| 6 | Loop pointer | `loop-config.yaml` `currentFeature` names the active feature doc (set at T0, verified at close-out; never points at a closed round when a new one starts) | read-back of the value |

Fail-closed: a code-changing work-unit with zero docs diff is REJECTED at close-out unless every row carries `checked` or `N/A + reason` (e.g., pure internal refactor with no surface change). The verifier scope includes this table's result before delivery. A second consecutive docs-drift round → STOP + owner review of the gate itself.
