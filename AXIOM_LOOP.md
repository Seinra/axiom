# 🔄 Axiom Loop Protocol

One-line identity: Axiom is a Go generics library (module `github.com/seinra/axiom`, Go 1.27); this loop runs its design and implementation; done means tree green after every iteration with one logical feature per iteration.

> **Changelog**: v0.1 (2026-10-20) — initial hub + modules + VISION + config for skeleton slice.

## 🎯 Objective & Success Metrics

- Objective: build a complete, human- and AI-friendly Go generics library at Echo-level quality, with clear rules for the hard parts of generics; consumed by Catalyst (`D:\Proyectos\AI Tech\Catalyst`).
- Definition of Done lives in `.loop-modules/SUCCESS_METRICS.md` — measurable, no aspirational checkboxes.

## 📦 Module Reference

| Module | Purpose |
| --- | --- |
| `.loop-modules/TECH_STACK.md` | Stack, versions, local docs policy |
| `.loop-modules/REFERENCES.md` | Pinned external truth by category |
| `.loop-modules/ARCHITECTURE.md` | Component layout, data flow, frozen decisions |
| `.loop-modules/REVIEW_PACKS.md` | Skills-per-phase pack map |
| `.loop-modules/VALIDATION_GATES.md` | Per-layer command tables + evidence rules |
| `.loop-modules/SUCCESS_METRICS.md` | Definition of Done |
| `.loop-modules/STATUS.md` | Progress log + debt ledger — NEVER in hub |

**Rule**: Always read the relevant module before working in that domain.

## 🔁 Cycle Flow

```
┌──────────┐   ┌────────────┐   ┌───────────┐   ┌──────────┐
│ CONTEXT  │──►│ PLAN (ODD) │──►│ IMPLEMENT │──►│ VALIDATE │
└──────────┘   └────────────┘   └───────────┘   └────┬─────┘
     ▲                                               │
     │            ┌──── Fail (< 3) ──► FIX ◄─────────┤
     └────────────┤                                  │
                  └──── Pass ──► DEPOSIT (memory) ───┘
```

- One logical feature per iteration; the project must be green after each.
- INVESTIGATE phase between CONTEXT and PLAN for research-heavy work (new generic abstraction with stdlib/ecosystem overlap).

## Side Tracks (Referenced, Not Governed)

| Track | Location | Convention | Promotion gate |
| --- | --- | --- | --- |
| Catalyst pilot consumption | scratch `/.tmp/` + Catalyst worktree (never commit in Catalyst) | Single verification note in feature doc Log; owner-gated push/PR | Becomes a phase only when Axiom cuts a versioned release Catalyst pins |

## 🧭 Governance

### Work Routing

| Work Type | Route | Trigger |
| --- | --- | --- |
| Inline Direct | trivial/mechanical | 1–3 files, no unresolved design |
| Delegate scout | exploration | ≥4 files to understand |
| Single writer | implementation | ≥2 non-trivial files |
| Pause + delegate | budget guard | ~20 tool calls / ~5 reads without delegation (self-counted, advisory) |

For ODD work orchestrator triggers win; the numeric table is legacy-only for live content loops without ODD — writers only for a reason (parallel units, context), never file count alone.

### Growth

- Hub over budget (~120 + ~10 per content phase; estimates advisory, wc rules) ⇒ split into modules; hub stays an index.
- Domain knowledge grows ⇒ new domain module + index row, not prose in hub.
- Governance needed elsewhere ⇒ reference this hub — never copy-paste.
- State growth: ODD track adds task rows, never phase files.

### ODD Tracking

Flow: `Authorize → Explore → Resolve uncertainty → Classify (Task Size) → Track → Implement task by task → Close`.
Track: `odd/tasks/<feature>.md` (header + `## Specs` S# quoted verbatim + `## Tasks` with linked S# + `## Log` with L1 verbatim) + Engram mirror (`odd/<feature>/tasks`) + visible `todo` projection rebuilt from the same tasks before the first write.
Work units: ~400 authored changed lines heuristic (advisory only), tests + docs alongside behavior, Conventional Commits, Risk line per change.
Delivery gate: work-unit commit first as review candidate, then `assess` + tiered review; burned approval gates push/PR/merge (owner-commanded HITL), never the commit.
Every task returns a Result Contract: `status`, `executive_summary`, `artifacts`, `next_recommended`, `risks` — formats ODD Tasks/S#, never parallel `phases/NN`; distinct-validator validates before advancing; `artifacts` carries ≥1 `command: observed result` line per behavior claim.
Methodology adapter: the harness methodology is named here (today: ODD) and pinned in REFERENCES as a stable ID; a rename/replacement updates the pin + mapping, never rule prose.

### TDD Scope Matrix

| Component | TDD? | Evidence |
| --- | --- | --- |
| Library logic (`result`, `opt`, `slice`, future `mapx`/`set`/`iterx`/concurrency/repo/httpx/valid) | YES | RED → GREEN → TRIANGULATE; commands recorded |
| Prose/assets/config (`*.md`, `loop-config.yaml`, `.golangci.yaml`) | NO | build gate + review |

Assertion-quality rules: ban tautological tests, ghost loops, implementation-detail coupling.
Matrix is default; per-change ODD applicability wins — no RED → state it; proportionate verification; never invent RED/GREEN.
Review-tier request signals: auth/secrets/guards → risk; public contracts → readability; failure paths/concurrency → resilience.

### Failure Policy

- Fix limit: **≤3 attempts per cycle** (loop-cycle counter, self-counted advisory; Log history rules), then STOP → human-in-the-loop. ODD verifier bound (one correction + bounded re-verify) is verifier-only.
- HITL triggers: structural format changes, security findings, unreproducible divergence, store/policy questions, >3 failures anywhere.

## 🧠 Memory Deposit (end of every cycle)

Typed Engram save per discovery/decision:
`title=[VERIFIED|DECISION|...]`, `type`, `topic_key` namespaced (`axiom/<area>/<slug>`), content with What/Why/Where/Learned. Project facts ≠ user preferences (separate deposits/scopes).

## 📋 Session Protocol

**Start**: read this loop → read the task file (file-first resume, never newest global memory) → check active changes/artifacts → preflight if first ODD-tracked work (feature document + mirror + todo reconciled) → resolve current status → verify prior deposit/mirror exists; if missing-deposit, reconstruct from artifacts.
**During**: confirm the deliverable IS the loop (loops-only guard), follow routing table; distinct-validator validates each phase result before advancing (solo: spot-check, re-run one reported command); each fix retry appends a Log row.
**End**: memory session summary — Goal, Instructions, Discoveries, Accomplished, Next Steps, Relevant Files — plus update STATUS.md if used; paste one externally-measured check (wc / lens / runtime backstop) as evidence at close-out, never asserted. Report `modules_read: [...]`. Mirror failure never blocks: preserve local, mark `mirror pending` in Log, resync later; explicit `session_id` on ambiguous-session writes. Generated artifacts default to English; never translate or paraphrase quoted user strings.

## ⚠️ HITL Triggers

1. Structural change to core formats/schemas
2. Security findings or sandbox escapes
3. Unreproducible state divergence
4. External-policy interpretation (stores, licenses)
5. Any cycle exceeding its fix limit
