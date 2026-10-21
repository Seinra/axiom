# REFERENCES

Pinned external truth by category. Packs cite IDs, never raw versions. Prose citing a version without an ID has no standing. `stale`/`missing` rows never support a decision.

## How to use

Every design decision cites a stable ID below. A rename/replacement updates the pin + adapter mapping, never rule prose.

## Pins

| ID | Category | Source | Pin |
| --- | --- | --- | --- |
| GO-TOOLCHAIN | toolchain | local `go version` | `go1.27.0 windows/amd64` (2026-10-20) |
| GO-GENERIC-METHODS | library | `https://go.dev/blog/generic-methods` | generic methods on concrete types only; no generic interface methods (Go 1.27) |
| GO-REL-1.27 | library | `https://go.dev/doc/go1.27` | fetched 2026-10-20; language + toolchain + stdlib changes |
| REFS-GO | external-docs | `D:/Proyectos/Refs/go/` (9 files) | usage docs, not generic design truth |
| REFS-ECHO | external-docs | `D:/Proyectos/Refs/echo/` (guide + middleware + cookbook) | consumption patterns for future `httpx` |
| REFS-GORM | external-docs | `D:/Proyectos/Refs/gorm/` (8 files CRUD) | consumption patterns for future `repo` |
| CATALYST-TREE | cloned-repo | `D:\Proyectos\AI Tech\Catalyst` (`module catalyst`, `go 1.27`) | no own generics; `slices`/`maps` + `map[string]any` envelopes |
| ODD-HARNESS | methodology | el Gentleman orchestrator (this session) | Authorize → Explore → Resolve → Classify → Track → Implement → Close |
| LOOP-ENGINEERING | methodology | `C:\Users\Seinra\.agents\skills\loop-engineering\SKILL.md` v1.9 | loop authoring discipline (this loop complies) |
| COG-DOC | methodology | `cognitive-doc-design` skill v1.0 | lead-with-answer, progressive disclosure, chunking, signposting |
| FOUNDRY | methodology | `foundry` skill v1.0 + `references/craft-floor.md` | absolute bans + quality floor (single policy source) |
| FOUNDRY-EXPLORE | methodology | `foundry-explore` skill v1.0 | read-only mapping, stops before design |
| FOUNDRY-REQUIREMENTS | methodology | `foundry-requirements` skill v1.0 | verifiable requirements, ask before assuming |
| FOUNDRY-DEBUG | methodology | `foundry-debug` skill v1.0 | reproduce → hypothesise → probe → fix → regression-lock |
| FOUNDRY-TESTING | methodology | `foundry-testing` skill v1.0 | HOW-to-test per language (Go playbook) |
| FOUNDRY-REVIEW | methodology | `foundry-review` skill v1.0 | R1–R4 pre-freeze verdict gate |
| FOUNDRY-DOCS | methodology | `foundry-docs` skill v1.0 | official docs layered over cognitive-doc-design (quickstart/guide/cookbook/API ref) |
| WORK-UNIT-COMMITS | methodology | `work-unit-commits` skill v1.0 | one purpose per commit; tests + docs travel with unit |
| BRANCH-PR | methodology | `branch-pr` skill v2.0 | open/prepare PR with approved issue + type label |
| CHAINED-PR | methodology | `chained-pr` skill v1.0 | PRs over 400 lines split into chained review slices |
| AGENT-GIT-COMMIT | methodology | `agent-git-commit` skill v1.0 | safe commits under coding-agent runtime gates |
| RDD-DEFECT | methodology | `rdd-defect-workflow` skill v1.4 | review authority, correction/recovery, delivery gate |
| ISSUE-CREATION | methodology | `issue-creation` skill v1.4 | triage + create from repo evidence with read-back |
| ECO-LO | external-docs | `https://github.com/samber/lo` | collections / functional utilities (target coverage) |
| ECO-MO | external-docs | `https://github.com/samber/mo` | Option / Result / Either / Future / Task (target coverage) |
| ECO-DO | external-docs | `https://github.com/samber/do` | type-safe DI incl. Echo/Gin/Fiber/Chi (target coverage) |
| ECO-SFLIGHT | external-docs | `https://github.com/samber/go-singleflightx` | generic singleflight + batching + sharding (target coverage) |
| ECO-ECHO-V5 | library | `github.com/labstack/echo/v5 v5.4.0` (go.mod direct require) | Echo adapter surface for httpx (Response/Bind/JSON/Fail/ListResponse); justified: Catalyst serves HTTP via Echo, envelope replaces ad-hoc maps |
| ECO-GORM | library | `gorm.io/gorm v1.31.2` + `github.com/glebarez/sqlite v1.11.0` (go.mod direct; pure-Go, no cgo) | GORM adapter surface for repogorm honoring repo.Repository; justified: Catalyst persists via GORM, sqlite driver is test/teaching only |

## Gaps (investigation only)

| ID | Status | Note |
| --- | --- | --- |
| STDLIB-SLICES-MAPS-ITER | missing | no local Refs docs; use official stdlib + GO-REL-1.27 |
| ECO-EXTRA (Set/Pool/Atomic/valid/repo/httpx/pipeline/channels) | missing | declared in ARCHITECTURE as future scope, not yet pinned |
