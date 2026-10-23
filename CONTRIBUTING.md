# Contributing to Axiom

External PRs are welcome — and judged by one rule.

## The rule, first

A requested feature or PR is accepted **only if it is genuinely good _and_ it fits the thesis**. Anything else will be declined: please fork it or use another library (`lo` and `mo` are excellent, battle-tested choices for broader coverage).

"Genuinely good" is not about enthusiasm; it is about fit. The thesis lives in three places — read them before writing code:

- [VISION.md](VISION.md) — the 8 invariants (stdlib-first, typed over `any`, grep-able names, green tree, AI-readable, teachable API).
- [.loop-modules/ARCHITECTURE.md](.loop-modules/ARCHITECTURE.md) — frozen decisions, the coverage map, and the rejected list.
- The package godoc — every export documents what the stdlib lacked.

## Checklist for a fitting PR

- [ ] It fills a proven stdlib gap, with the justification written in the godoc.
- [ ] Public API is typed (`any` only at codec edges, wrapped).
- [ ] Names are grep-able: no single-letter vars, no abbreviations, no generic names.
- [ ] One Which-table row per new export, one compiling example per package happy path (every new export ships its own), tests pinning the documented behavior.
- [ ] Pure functions only: inputs never mutated, outputs fresh. No mutable variants, no reflection, no code generation.
- [ ] Discrete scope: one logical unit per PR, no API explosion, no speculative generality.

## What gets declined (with examples)

| Ask | Why it stays out |
| --- | --- |
| Duplicates of stdlib (`sort`/`reverse`/repeat, `sync.OnceValue` memoize, atomic wrappers) | stdlib owns it — never reimplemented. |
| Reflection-first or code-generated APIs | Banned by the thesis; unteachable and unreviewable at this scale. |
| Mutable in-place variants | Callers rely on inputs staying intact; loops and stdlib cover mutation. |
| Domain helpers (dates, URLs, query strings) | Domain of an adapter, not generics lessons. |
| Tuples T2–T9, `Lazy`, gods-style containers | Rejected under the stdlib-gap + teachability gate (see ARCHITECTURE). |
| New packages without a gated need | Package creation needs an explicit owner call — open an issue first. |

A decline cites the row above, links the thesis section, and points at forking. That is the whole appeal process: a fork that proves the thesis wrong with evidence reopens the conversation.

## Process

1. Open an issue first (issue forms choose the template); one logical unit per PR.
2. Reference it with `Closes #N`; keep CI green (`go test`, `go vet`, `gofmt`, lint).
3. Maintainer review decides. The decision is final for this repo; the fork option is always open.
