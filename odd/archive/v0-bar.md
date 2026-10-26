# v0 maturity bar (draft — unreleased, no tag yet)

> [!NOTE]
> Superseded by the v0.2.0 release — kept as history, not a live doc. Do not use it to judge the current tree.

v0 is the first public promise of Axiom: what stays stable, what ships, and how a stranger verifies it before `go get`. This bar decides the finish line for everything built so far; #36 enters v0 only if it meets this bar.

**Version:** Axiom is unreleased — no tag yet; this document tracks the working tree toward `v0.1.0`.
**Prereqs:** Go 1.27 in your `PATH` (`go version` reports `go1.27.x`); clean tree on the release commit.

## Quick path

1. Freeze check: no exported symbol removed, renamed, or re-signed since the bar ratification commit.
2. Package check: all packages in Minimum set present with godoc + Which-table + compiling Example.
3. Run the gates and expect green (commands in Checklist).

## Details

| Topic | Decision |
| --- | --- |
| Status | Ratified 2026-10-22: bar frozen as the v0 gate. #36 Combinations/Permutations included in v0 (first additive slice, meets bar). LICENSE MIT. Tag `v0.1.0`. |
| Freeze policy | Additive-only after ratification: MUST NOT remove, rename, or re-sign any exported symbol. Additions MUST pass the stdlib-gap + teachable gate (ARCHITECTURE P2 gate, Engram `axiom/p2/gate-stdlib-ensenable`): stdlib lacks it AND it teaches as DLC lesson material (godoc + Which-table + Example). Breaking changes need a major bump or an explicit owner exception with a migration note. |
| Minimum package set | 15 library packages: `result`, `opt`, `either`, `slice`, `set`, `mapx`, `iterx`, `valid`, `singleflight`, `pool`, `di`, `repo` (+ adapter `repogorm`), `httpx` (Echo v5), `flow`. Plus narrative docs: `quickstart.md`, 15 family guides, `cookbook.md` with CI-executed recipes (`docs/cookbook_test.go`). #36 `Combinations`/`Permutations` (slice) included in v0 — first addition under the freeze. Declared future stays out of v0: fan-in/out pipelines, Future/Task, OrderedMap/Stack/Queue/LRU, and #36 combinations/permutations until the #36 decision lands it in v0 or v0.x. |
| Docs gate | Every export has godoc with an example; every package doc opens with a Which-table; README index covers all 15 packages plus the canonical `(T, bool)` = OrNull statement; zero stale counts/names; `go doc ./...` renders without warnings. |
| Tests gate | `go test ./...` green on Go 1.27 (all packages `ok`); `go vet ./...` clean; `gofmt -l .` empty. Behavior claims carry `command: observed result` lines; coverage is informational, behavior is the gate. |
| Lint gate | `golangci-lint run --timeout=5m ./...` reports 0 issues (minimal enforcing set: `govet`, `staticcheck`, `errcheck`, `revive:var-naming`, `gocyclo≤10` per TECH_STACK). |
| Versioning | Semver. First public tag proposed `v0.1.0` on module `github.com/seinra/axiom` (`go 1.27`). v0 allows additive growth; breaking exports require v1 or an owner-signed exception. |
| Release checklist | 1) Bar ratified by owner. 2) Tree green on gates above at the release commit. 3) Tag `v0.1.0` pushed. 4) Release notes (freeze scope, package list, gate evidence). 5) Go proxy availability: `GOPROXY=proxy.golang.org go list -m -versions github.com/seinra/axiom` shows the tag; `go get github.com/seinra/axiom@v0.1.0` resolves from a scratch module. 6) Quickstart verified <5 minutes by a fresh reader. |
| License | Decided MIT (owner 2026-10-22, verbatim "Ok, vamos con MIT"). No publish before the LICENSE file + tag land (both owner-gated). |
| Branch protection | Desired (issue #3) but blocked: needs GitHub Pro or a public repo. Until then, PR-gated merges by discipline. |
| Out of scope | Publishing the repo, pushing the tag, opening PRs, Catalyst commits, and #36 implementation. This track only defines the bar. |

## Checklist

- [ ] `go test ./...` reports `ok` for every package on Go 1.27
- [ ] `go vet ./...` exits 0 and `gofmt -l .` prints nothing
- [ ] `golangci-lint run --timeout=5m ./...` reports 0 issues
- [ ] Every package doc has a Which-table and every export has godoc + Example; README index matches the tree
- [ ] `GOPROXY=proxy.golang.org go list -m -versions github.com/seinra/axiom` shows `v0.1.0` after the tag push
- [ ] A fresh reader completes `docs/quickstart.md` in under 5 minutes

## Next step

Owner ratifies this draft (including license pick and the #36 in-v0 vs v0.x call), then T3 runs the gates once against the ratified tree and closes with per-artifact disposition. Tag and publish stay owner-gated.
