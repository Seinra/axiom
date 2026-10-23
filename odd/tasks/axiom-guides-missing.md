# Axiom missing guides round — foundations (done) + stack/queue/orderedmap/task guides + freshness

Docs-only round: land the ordered capítulo cero and the four family guides missing for F1/F2/F4, plus refresh stale counts. No library code changes.

## Specs

- S1 (owner orders verbatim 2026-10-23): "Hay que aprovechar si todos los docs estan al dia tambien" (freshness audit). Prior: capítulo cero gap + "Ese capítulo hay que escribirlo; es docs, no código." Then: "Si, hacelo" (commit chapter + write the 4 missing guides + README refresh).
- S2 (scope):
  - T0 DONE pre-branch: `docs/guides/foundations.md` (mechanism chapter) + `docs/foundations_test.go` (CI-executed mirror) + quickstart bridge row; parent-written, verifier Go (0 BLOCKERs), 1-word SortBy polish applied + re-green.
  - T1 (this branch): `docs/guides/{stack,queue,orderedmap,task}.md` + CI-executed example tests + quickstart family-table rows for the 4 packages.
  - T2 (freshness): README "Ships now (15 pkgs)" → 19 lib pkgs; "Declared future" line drops delivered items (fan-in/out, Future/Task, OrderedMap/Stack/Queue); loop-config currentFeature if stale.
- S3 (gates, standing): `go test -count=1 ./docs/` green (new tests included), `go vet ./...` 0, `gofmt` empty; prose builds zero packages; verifier Go before delivery; push/PR owner-gated (new issue + approval).
- S4 (house style, program S5 + trio): guide shape concept → steps → Which-table → runnable Example → next step; doc comments are the source, prose links never duplicates; second person active present; every new guide gets a CI-executed test (foundations precedent — stronger than the 15 siblings' scratch-only examples); full package + method names, no abbreviations.

## Tasks

- T0 (S1, S2): Foundations chapter + executed test + bridge row — route: parent + foundry + foundry-docs + cognitive-doc-design + foundry-requirements + foundry-craft; verifier Go — commit: close-out (this branch)
- T1 (S1-S4): 4 family guides + executed tests + quickstart rows — route: parent inline (template + APIs already mapped) — commit: 704159a
- T2 (S2, S3): README + loop-config freshness — route: parent inline — commit: close-out (this branch)
- T3 (S3): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated (new issue + approval) — commit: pending
- T4 (S3): Delivery — route: parent; issue type:docs + status:approved + PR Closes #N type:docs + CI green; merge owner-gated — commit: pending

## Log

- L1 (request verbatim): "Hay que aprovechar si todos los docs estan al dia tambien"
- L2 (audit result, parent-verified 2026-10-23): capítulo cero gap REAL (quickstart mentions `[T]` in passing, 15 guides assume it, no mechanism chapter); docs NOT fully current — guides missing for stack/queue/orderedmap/task; README says 15 pkgs (19 lib pkgs now) + "Declared future" lists delivered items. Verifier pack for T0: review (not craft) per owner correction.
- L3 (2026-10-23): classified large (4 guides + 4 tests + wiring + README, multi-file program); rama feat/axiom-guides-missing desde main (post-#64); ODD + mirror + todo reconciled before first write of T1 (T0 files already in tree, uncommitted).
- L4 (2026-10-23, T0-T3 DONE + verifier Go + close-out): foundations + 4 guides + 2 executed test files + quickstart bridge + 4-row family table + README 19pkgs + Declared-future owner-gated reword + loop-config currentFeature; verifier mvkbo2a3-b-2ort Go all S# (0 BLOCKERs, 1 stacking WARNING closed by single-unit landing, 2 SUGGESTIONs: SortBy-or-keys + dual flow/task link applied inline); parent re-ran go test ./docs/ + vet + gofmt + lint 0. Slice ~700 lines over heuristic = single logical docs unit, no more content commits — delivery strategy applies. T4 delivery.
