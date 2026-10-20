# Axiom skeleton + loop — feature document

Biblioteca de genéricos Go desde 0 (módulo github.com/seinra/axiom, Go 1.27), consumida por Catalyst. Primer slice: solo esqueleto + loop completo, sin implementación total.

## Specs

- S1: "Proyecto nuevo desde 0. En D:\Proyectos\Refs tenes doc de Go actualizados, echo, gin, gorm y varias cositas mas."
- S2: "Este proyecto se va a tratar de crear una Biblioteca de genericos de Go, posiblemente a futuro se podria convertir en framework pero eso es una meta muy a futuro."
- S3: "Este proyecto va a ser consumido tambien por el proyecto Catalyst (D:\Proyectos\AI Tech\Catalyst) que es una app educativa (programacion) con un tutor IA local (offline)."
- S4: "Podes empezar por el git init."
- S5: "La idea es que minimamente esta biblioteca termine cubriendo minimamente todo el contenido de este analisis:" — incluye lo, mo (Option[T] / Result[T] / Either[A,B] / Future[T] / Task[T]), do (do.Provide / do.Invoke con Echo, Gin, Fiber y Chi), go-singleflightx, iter de la stdlib (Go 1.23), constraints, y generic methods de Go 1.27.
- S6: "Pero investiga si faltan mas cosas por añadir a la lista."
- S7: "Arma el loop usando loop-engineering, cognitive-doc-design y foundry, ademas la familia de skills de foundry tiene ciertas reglas que podrias respetar en el tipo de codigo que se hace en la biblioteca de genericos."
- S8: "La idea es apuntar a una biblioteca completa al nivel de echo, pero para genericos, excelente tanto para humanos como para IA, y que sea facil de manejar, con reglas claras, dado que esa es la dificultad de los genericos."
- S9 (decisión 2026-10-20): module path "github.com/seinra/axiom".
- S10 (decisión 2026-10-20): alcance primer slice "Solo esqueleto + loop".
- S11 (decisión 2026-10-20): loop "Loop completo" (AXIOM_LOOP.md + módulos + VISION + config).
- S12: Ecosistema extra a cubrir a futuro (investigado): slices/maps/cmp/iter stdlib, Set[T], OrderedMap/Stack/Queue/LRU genéricos, Pool/Atomic/Singleflight genéricos, validación Validator[T], concurrencia (fan-in/out, pipeline, worker pool, errgroup genérico), canales (Merge/Batch sobre iter.Seq), JSON/DTO (Response[T], Page[T]), Repository[T,ID] + paginación, Echo Binder[T]/Responder[T], testing genérico. No implementar en este slice; solo dejarlo declarado en ARCHITECTURE/REFERENCES.

## Tasks

- T1 (S4, S9): git init + go.mod Go 1.27 + commit inicial — route: inline — commit: b2ff3bb
- T2 (S7, S11): AXIOM_LOOP.md + .loop-modules/* + VISION.md + loop-config.yaml — route: inline — commit: 6385387
- T3 (S1, S2, S3, S5, S12): esqueleto result/opt/slice con godoc + tests verdes + golangci mínimo — route: inline (small path dentro de large) — commit: af7a9cd
- T4 (S7, S8): README con shape cognitive-doc-design + ARCHITECTURE que declara cobertura futura — route: inline — commit: c7e69a6
- T5 (S3): verificación de consumo piloto desde Catalyst (replace directive en scratch, nunca commit en Catalyst) — route: inline — commit: n/a

## Log

- L1 (request original verbatim): "Proyecto nuevo desde 0. En D:\Proyectos\Refs tenes doc de Go actualizados, echo, gin, gorm y varias cositas mas. Este proyecto se va a tratar de crear una Biblioteca de genericos de Go, posiblemente a futuro se podria convertir en framework pero eso es una meta muy a futuro. Este proyecto va a ser consumido tambien por el proyecto Catalyst (D:\Proyectos\AI Tech\Catalyst) que es una app educativa (programacion) con un tutor IA local (offline). Podes empezar por el git init. La idea es que minimamente esta biblioteca termine cubriendo minimamente todo el contenido de este analisis: --- [análisis largo de lo/mo/do/singleflightx/iter/constraints/Go 1.27 generic methods] --- Pero investiga si faltan mas cosas por añadir a la lista. Arma el loop usando loop-engineering, cognitive-doc-design y foundry, ademas la familia de skills de foundry tiene ciertas reglas que podrias respetar en el tipo de codigo que se hace en la biblioteca de genericos. La idea es apuntar a una biblioteca completa al nivel de echo, pero para genericos, excelente tanto para humanos como para IA, y que sea facil de manejar, con reglas claras, dado que esa es la dificultad de los genericos."
- L2 (2026-10-20): decisiones del owner — module github.com/seinra/axiom; alcance Solo esqueleto + loop; loop Loop completo.
- L3 (2026-10-20): scout gentle-ai-explore COMPLETE — Axiom vacío (solo .atl/ + .crystallize/), Catalyst module catalyst go 1.27 sin genéricos propios (solo slices.Sort, maps.Copy, map[string]any en exam.go:400ff), Refs sin docs de slices/maps/iter/constraints (gap a stdlib oficial). Go toolchain real verificado inline: go1.27.0 windows/amd64.
- L4 (2026-10-20): Go 1.27 generic methods confirmados vía fetch (go.dev/blog/generic-methods + go.dev/doc/go1.27): type params en métodos concretos, NO en interfaces; habilita chaining Map/Filter como métodos.
- L5 (2026-10-20): T3 RED→GREEN — `go test ./...` antes de implementar: `no non-test Go files` (FAIL build); después: `ok opt/result/slice`, `go vet ./...` exit 0, `gofmt -l .` vacío. T5 piloto scratch `.tmp/pilot` (replace a ../..): `verdict: approved failure: <nil>`, `failure typed: true`, `hint: no hint cached`, `ready: [0 2 4]`, `grouped: map[even:[2 4] odd:[1 3]]`, exit 0. Catalyst intacto.
- L6 (2026-10-20): DECISION — `.crystallize/keys/` (ed25519.priv) y `.atl/.skill-registry.cache.json` excluidos en `.gitignore`; secretos nunca se commitean (craft-floor ban 2).
