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
| FOUNDRY-TESTING | methodology | `foundry-testing` skill v1.0 | HOW-to-test per language (Go playbook) |
| FOUNDRY-REVIEW | methodology | `foundry-review` skill v1.0 | R1–R4 pre-freeze verdict gate |
| ECO-LO | external-docs | `https://github.com/samber/lo` | collections / functional utilities (target coverage) |
| ECO-MO | external-docs | `https://github.com/samber/mo` | Option / Result / Either / Future / Task (target coverage) |
| ECO-DO | external-docs | `https://github.com/samber/do` | type-safe DI incl. Echo/Gin/Fiber/Chi (target coverage) |
| ECO-SFLIGHT | external-docs | `https://github.com/samber/go-singleflightx` | generic singleflight + batching + sharding (target coverage) |

## Gaps (investigation only)

| ID | Status | Note |
| --- | --- | --- |
| STDLIB-SLICES-MAPS-ITER | missing | no local Refs docs; use official stdlib + GO-REL-1.27 |
| ECO-EXTRA (Set/Pool/Atomic/valid/repo/httpx/pipeline/channels) | missing | declared in ARCHITECTURE as future scope, not yet pinned |
