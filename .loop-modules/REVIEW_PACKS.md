# REVIEW_PACKS

Phase → skill routing with load triggers. `loop-engineering` fires only when the deliverable IS the loop, never for domain scouts/writers/verifiers.

## Quick path

1. Match the current phase below.
2. Evaluate its trigger — load the skill only when the trigger holds.
3. Report `packs_loaded`, `code_intel`, `modules_read` on return.

## Core map

| Phase | Skill | Trigger (evaluate each time) | Inputs | Outputs |
| --- | --- | --- | --- | --- |
| CONTEXT | `foundry-explore` | mapping ≥4 files or unknown blast radius | repo tree + question | map with `path:line` evidence |
| INVESTIGATE | `sdd-research`-class external evidence (or coded `fetch_content`) | stdlib/ecosystem overlap or version question | REFERENCES IDs + question | cited findings with pins |
| PLAN (ODD) | `foundry-requirements` | ambiguous scope or missing acceptance criteria | conversation + tree evidence | verifiable requirements (MUST/SHOULD/MAY + Given/When/Then) |
| PLAN (ODD) | `cognitive-doc-design` | any doc a human or reviewer must scan fast | draft doc | outcome-first doc with quick path + checklist |
| IMPLEMENT | `foundry` (exactly one Decision Gate playbook) | designing/implementing/hardening backends | routed playbook + task | verifiable artifact (code/spec/migration/test plan) |
| IMPLEMENT (docs units) | `foundry` + `foundry-docs` + `cognitive-doc-design` (trio) | the unit writes official docs | draft + reference corpus | versioned page (prereqs + runnable example) |
| IMPLEMENT | `loop-engineering` | ONLY when the deliverable IS the loop | loop-template + gap scan | hub/modules/config change + audit |
| VALIDATE | `foundry-testing` | writing/reviewing/adding tests | behavior + boundary | tests + `command: observed result` lines |
| VALIDATE | `foundry-review` | pre-freeze verdict needed | diff + spec | Go/No-Go ledger (BLOCKER/WARNING/SUGGESTION) |
| FIX | bounded writer + `foundry-debug`-class diagnosis | red or finding to fix, ≤3 attempts | failing evidence | fix + re-verify lines |
| DEPOSIT | Engram memory (`mem_save`, `mem_session_summary`) | every cycle end | discoveries/decisions | typed deposits + session summary |

Loop-engineering guard: domain scouts/writers/verifiers NEVER use loop-engineering; it fires only for loop artifacts (`AXIOM_LOOP.md`, `.loop-modules/*`, `loop-config.yaml`, `VISION.md`).

## Five elements (pack-map spec)

| Element | Rule in this loop |
| --- | --- |
| Tiers | core-per-phase: `foundry-explore`, `foundry-requirements`, `cognitive-doc-design`, `foundry`, `foundry-testing`, `foundry-review`, Engram. conditional-trigger: external-evidence (INVESTIGATE only), `loop-engineering` (IMPLEMENT only when deliverable IS loop), bounded writer + `foundry-debug`-class (FIX only), `foundry` + `foundry-docs` + `cognitive-doc-design` trio (IMPLEMENT docs units only). future-roadmap: none — lean loop declares none-mapped with reason when a phase needs no skill. |
| Capability block | compact descriptors (name + one-line trigger + contract pointer) advertised INTO every child envelope — children don't inherit ambient tools. |
| Reporting contract | `packs_loaded`, `code_intel`, `modules_read` lines on return — instruction without reporting evaporates; standing code-intel default (structure-first, honest fallback). |
| Fail-closed | resolution always runs, LOAD stays conditional — absence declares itself (`none-mapped` + reason), never silent. A phase that touched code or docs without loading its mapped packs has an invalid Result Contract. |
| Reference IDs | pack maps cite REFERENCES IDs, never raw versions or URLs: `FOUNDRY`, `FOUNDRY-EXPLORE`, `FOUNDRY-REQUIREMENTS`, `FOUNDRY-DEBUG`, `FOUNDRY-TESTING`, `FOUNDRY-REVIEW`, `FOUNDRY-DOCS`, `COG-DOC`, `LOOP-ENGINEERING`, `ODD-HARNESS`. |

## Checklist

- [ ] Trigger evaluated before loading — no ambient skill inheritance.
- [ ] Exactly one `foundry` Decision Gate playbook per IMPLEMENT task.
- [ ] Docs units load the trio: `foundry` + `foundry-docs` + `cognitive-doc-design`.
- [ ] `loop-engineering` only for loop artifacts — never domain work.
- [ ] Return carries `packs_loaded`, `code_intel`, `modules_read`.

## Next step

Load VALIDATION_GATES for commands and evidence rules before claiming behavior.
