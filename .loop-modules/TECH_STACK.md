# TECH_STACK

Stack, versions, and local docs policy for Axiom. Read before domain work. `modules_read` must include this file when touching toolchain, deps, or docs.

## Toolchain

| Item | Pin | Note |
| --- | --- | --- |
| Go | 1.27.0 (`go 1.27` in `go.mod`) | Verified `go version go1.27.0 windows/amd64` 2026-10-20 |
| Module | `github.com/seinra/axiom` | Owner decision 2026-10-20 |
| golangci-lint | minimal enforcing set (see VALIDATION_GATES) | `govet`, `staticcheck`, `errcheck`, `revive:var-naming`, `gocyclo≤10` |
| Consumer | Catalyst `module catalyst`, `go 1.27` | Offline-first educational app + local AI tutor |

## Language features in use

- Generics since Go 1.18 (functions + types); recursive type params (1.26); **generic methods on concrete types (1.27)** — NO generic interface methods.
- `iter` (`iter.Seq`/`Seq2`, Go 1.23+), `slices`, `maps`, `cmp` from stdlib — always preferred over hand-rolled (foundry Simplicity Ladder rung 3).

## Local docs policy (Refs)

- `D:/Proyectos/Refs` is a Markdown reference store, NOT generic design truth: `go/` (9 files stdlib usage), `echo/` (guide + middleware + cookbook), `gorm/` (8 files CRUD). No local `slices`/`maps`/`iter`/`constraints` docs — design against stdlib official + pinned external truth in REFERENCES.
- Catalyst consumption patterns (from scout): `slices.Sort`, `maps.Copy`, `map[string]any` JSON envelopes (`internal/exam/exam.go:400ff`) — Axiom offers typed alternatives, never `any` in domain (craft-floor ban 17).

## Dependency policy

- Stdlib-first; no new dep for what a few lines do. Every new dependency needs a one-line justification + version pin in REFERENCES.
