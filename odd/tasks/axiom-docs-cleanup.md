# Axiom docs cleanup — physical paths + parked-vs-done

Single docs-only slice: portable consumer/Refs wording + reconcile freshness parked-vs-done + point loop at this active feature.

## Specs

- S1 (request verbatim 2026-10-23): "Te comento que tenes docs que apuntan a ubicaciones fisicas en mi maquina, como el readme apuntando a la ubicacion fisica del disco D de Catalyst jajaja. Mepa que hay que corregir eso en los docs xD"
- S2 (order verbatim 2026-10-23): "Todo junto en un solo slice, incluso el arreglo de ese despelote de lo parked que en realidad no esta parked porque ya esta hecho"
- S3 (scope verbatim 2026-10-23): "Lee el loop sin olvidarte del review packs y el skill-registry, los artifacts odd y decime en que habíamos quedado."
- S4 (standing, docs-freshness S3): each slice gets its own ODD (spec + mirror + todo before first write); prose-only → build gate + review, no RED invented; verifier with review pack (foundry-review, never foundry-craft) before delivery; native review never runs on docs-only; delivery needs new issue + approval, push/PR/merge owner-gated.
- S5 (standing, house style S4 + trio): trio per REVIEW_PACKS docs units (foundry + foundry-docs + cognitive-doc-design); second person active present; tables stay scannable; doc comments are source, prose links never duplicates.
- S6 (verbatim guard): odd S#/L1 quotes stay verbatim — `odd/tasks/axiom-skeleton-loop.md` S1/S3/L1 keep `D:\Proyectos\Refs` and `D:\Proyectos\AI Tech\Catalyst` untouched; fix applies only to live prose (README, AXIOM_LOOP, REFERENCES, TECH_STACK).
- S7 (tag question verbatim 2026-10-23): "Y fijate que tenemos un tag listo pero no se si hay que hacer algo mas?" — answer only in this slice; no tag/release/commit action without separate explicit owner order.
- S8 (question verbatim 2026-10-23): "Mmm sigue siendo una \"small library\"?"
- S9 (order verbatim 2026-10-23): "Y la comparativa con echo a nivel calidad esta de mas" + "Pone educative library o algo asi, no small porque de small con tanto paquete ya no la describe" — decisions: reword a composable + educative, sacar claim Echo (keep only technical httpx/Echo v5 mentions).

## Tasks

- T0 (S4): track + branch — route: parent inline — commit: pending
- T1 (S1, S5, S6): portable wording in 4 live surfaces — route: parent inline — commit: pending
- T2 (S2, S4): reconcile parked-vs-done (freshness Log close rows + STATUS T1-T3 rows + loop-config pointer to this slice) — route: parent inline — commit: pending
- T3 (S4, S5): verify + review gate + Docs-sync verdict — route: parent inline — commit: 4421cd3
- T4 (S5, S8, S9): reword small->educative/composable + drop Echo quality claim (README tagline, VISION lead, loop objective) — route: parent inline — commit: pending

## Design

- T1 replacements (portable, no local disk):
  - Consumer: `Catalyst (offline-first course app, module catalyst)` + `pilots via scratch replace only — never a Catalyst commit from here`.
  - Refs: `local reference store (uncommitted, not design truth: go/ stdlib usage, echo/ guide+middleware+cookbook, gorm/ CRUD)` + `no local slices/maps/iter/constraints docs — design against stdlib official + pinned external truth`.
  - CATALYST-TREE row: `cloned-repo (module catalyst, go 1.27)` without disk path.
  - ODD-HARNESS/LOOP-ENGINEERING rows: drop `C:\...SKILL.md` absolute, keep `skill name + version + contract`.
- T2 reconciliation (no rewording of ratified prose beyond cited lines):
  - `odd/tasks/axiom-docs-freshness.md` append L4 close-out: T1 (README 21 + flow YAGNI fix, ef090fa / PR #79), T2 (mathx+strings guides + index rows, 93024fd / PR #83), T3 (8 Which-table rows, 6a974c8 / PR #85) — all green + merged; slices done, doc closed.
  - `.loop-modules/STATUS.md` append 3 Log rows (T1/T2/T3 delivery, same SHAs/PRs) — no design prose.
  - `loop-config.yaml` `currentFeature: odd/tasks/axiom-docs-cleanup.md` at T0, verified at close-out.
- T3 gates: `go test ./...` + `go vet ./...` + `gofmt -l .` + `golangci-lint run` (prose-only: build gate + review, no RED per VALIDATION_GATES).

## Log

- L1 (request verbatim): "Hola hermano, inicia sesion en Engram, recupera tu contexto y recorda que tenes acceso a los MCP de Codegraph y zvec-grep (lazy-load codemode) y tenes que hacer uso de ellos durante la sesion. Lee el loop sin olvidarte del review packs y el skill-registry, los artifacts odd y decime en que habíamos quedado. Te comento que tenes docs que apuntan a ubicaciones fisicas en mi maquina, como el readme apuntando a la ubicacion fisica del disco D de Catalyst jajaja. Mepa que hay que corregir eso en los docs xD Y fijate que tenemos un tag listo pero no se si hay que hacer algo mas?"
- L2 (order verbatim): "Todo junto en un solo slice, incluso el arreglo de ese despelote de lo parked que en realidad no esta parked porque ya esta hecho"
- L3 (2026-10-23, T0): codegraph + zvec via codemode done (consumer/replace + physical-path hits); trio loaded (foundry router + foundry-docs + cognitive-doc-design) + foundry-review + work-unit-commits; branch feat/axiom-docs-cleanup from main clean.
- L4 (2026-10-23, T3 verify): `go test ./...` ok 22 pkgs, `go vet ./...` exit 0, `gofmt -l .` empty, `golangci-lint run --timeout=5m ./...` 0 issues; foundry-review gate Go (0 BLOCKER, prose-only, R1 no secrets, R2 portable names, R3 no behavior, R4 no ops); Docs-sync verdict — row1 checked (README 21 == tree 21 lib dirs), row2 N/A (no new package), row3 N/A (no new export), row4 checked (no new no-X/YAGNI lines), row5 checked (Declared-future closed wording untouched+true), row6 checked (currentFeature=odd/tasks/axiom-docs-cleanup.md); tag S7 answer-only, no tag/release action; work-unit committed 4421cd3, branch kept.
- L5 (2026-10-23, T4): user decisions via choice — small: reword a composable, echo: sacar claim; user refines: educative library wording; trio already loaded; implement in README/VISION/loop objective, technical Echo mentions untouched.
