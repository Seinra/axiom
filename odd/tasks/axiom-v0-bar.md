# Axiom v0 maturity bar — define finish line (issue #4)

Decision + authoring track: define what v0 means before tagging anything. No release, no publish, no code freeze yet — borrador para ratificación owner.

## Specs

- S1 (issue #4 verbatim, title `[feat]: Define v0 maturity bar and public-release checklist`): "VISION decides Axiom goes public (MIT/BSD) at maturity, but maturity is unmeasured. Define the bar: API freeze policy, minimum package set, docs/tests/lint gates, versioning scheme, and the release checklist (tag, notes, go proxy availability)."
- S2 (issue #4 Consumer fit verbatim): "Students keep the library after the Catalyst generics DLC; the release is what they `go get`."
- S3 (owner order 2026-10-22 verbatim): "Vamos con #4 entonces"
- S4 (owner question verbatim): "Primero, #4 no es acerca de abrir esta libreria para el uso publico? O de que se trata?"
- S5 (owner question verbatim): "No seria entonces mejor terminar todo lo demas en Axiom ANTES de #4? Porque todavia no tiene consumidores, Catalyst todavia no tiene un DLC de genericos tampoco."
- S6 (owner question verbatim): "Dale... Respecto a API nueva, explicame como seria este tema porque no entiendo... Se necesita una API para hacer uso de una libreria?"
- S7 (acuerdo ratificado 2026-10-22): "#4 borrador ahora (solo docs, sin tag, sin publicar); #36 después; cierre de #4 al final con evidencia y tag solo a orden owner."

## Tasks

- T1 (S1, S4, S5): Explorar base — VISION inv 6+8, STATUS debt ledger, VALIDATION_GATES, SUCCESS_METRICS library bar, ARCHITECTURE coverage/P2 gate, docs/ estado — route: parent inline read-only — commit: n/a
- T2 (S1, S7): Definir bar v0 borrador — freeze policy, paquete mínimo, gates docs/tests/lint, versionado, checklist tag/notes/go proxy, licencia MIT/BSD pendiente — route: parent inline authoring docs-only — commit: this work-unit (docs/v0-bar.md)
- T3 (S7): Verifier independiente + close-out + delivery — route: parent inline (docs-only, Risk: none); push/PR/merge/tag owner-gated — commit: this work-unit

## Log

- L1 (request verbatim): "Vamos con #4 entonces"
- L2 (2026-10-22): contexto previo verbatim — "Primero, #4 no es acerca de abrir esta libreria para el uso publico? O de que se trata?" → aclarado: #4 es definir el finish line, no publicar. "No seria entonces mejor terminar todo lo demas en Axiom ANTES de #4? Porque todavia no tiene consumidores, Catalyst todavia no tiene un DLC de genericos tampoco." → acordado: sin bar el "todo" no existe; #4 borrador barato ahora, #36 después, cierre al final. "Dale... Respecto a API nueva, explicame como seria este tema porque no entiendo... Se necesita una API para hacer uso de una libreria?" → aclarado: API = funciones exportadas (ej. slice.MapSlice); #36 suma Combinations/Permutations; #4 congela la promesa, no el código.
- L3 (2026-10-22): rama feat/axiom-v0-bar creada desde main (incluye fcf8271 local ahead 1, push a orden owner); ODD preparado sin implementación.
- L4 (2026-10-22, T1+T2 DONE + T3 inline Go): base explorada (VISION inv 6+8, STATUS debt, VALIDATION_GATES, SUCCESS_METRICS, ARCHITECTURE P2 gate, README 15 pkgs, docs/ 15 guides + cookbook_test CI, go list 16 pkgs incl. docs, go1.27.0); bar borrador creado en docs/v0-bar.md (freeze aditivo, set mínimo 15, gates test/vet/fmt/lint/doc, semver v0.1.0 propuesto, checklist tag/notes/proxy, licencia y #36-in-v0 abiertas). Evidencia fresca -count=1: suite 16/16 ok, vet 0, gofmt vacío; docs-only (2 md nuevos, 0 go cambiado) → lint previo 0 issues sigue vigente, TDD n/a (prosa, sin rama ni dinero). Review docs inline: 0 BLOCKER/WARNING. Push/PR/tag owner-gated.
