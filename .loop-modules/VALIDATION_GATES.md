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
