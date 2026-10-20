# Axiom expand P2-gated — feature document

Stdlib-gap + teachable expansion of done packages plus either/Race, in owner-approved order, autonomous with bounded workers per previous-session rules.

## Specs

- S1: "Vamos en ese orden, en autonomo como la sesion previa, con las mismas reglas."
- S2: "Persistilo ahora, hay que ver si de lo ya hecho toca expandir algo, y de lo que no hicimos toca ver que implementar."
- S3: "Revisa si lo que aprendiste en la sesion previa sobre otras bibliotecas/frameworks de genericos, de otros lenguajes obviamente, no tienen algo sobre lo que expandir de lo hecho, que ahora no se te haya ocurrido"
- S4 (decisión 2026-10-20, gate-stdlib-ensenable): P2 entra si stdlib no lo tiene y es enseñable como material DLC, aunque Catalyst aún no lo use.
- S5 (orden aprobado vía S1): 1) guías + (T,bool), 2) SortBy + Scan, 3) UnwrapOrElse + MapKeys/IsDisjoint, 4) bloque lazy de iterx, 5) either/Race.
- S6 (reglas sesión previa, vigentes): verificador independiente cada ~400 líneas antes que review; review nativo nunca para docs; commitear en autónomo y pushear solo con permiso; findings en observaciones separadas sin pisado; implementaciones a workers con skills inyectadas; verificar subagentes antes de esperarlos.

## Tasks

- T1 (S1, S5.1): guías por familia estilo Swift + declarar (T,bool)=OrNull en docs — route: bounded worker docs-only, no native review — surfaces: README.md, opt/opt.go, result/result.go, slice/slice.go, set/set.go, mapx/mapx.go, iterx/iterx.go — commit: 33f40bb
- T2 (S1, S5.2): slice SortBy convenience + Scan/RunningFold — route: bounded worker — surfaces: slice/* — commit: 725c493
- T3 (S1, S5.3): result UnwrapOrElse (+Flatten) + mapx MapKeys/FilterKeys/Values + set IsDisjoint — route: bounded worker — surfaces: result/*, mapx/*, set/* — commit: db20ebc
- T4 (S1, S5.4): iterx lazy block: Generate/Unfold, FilterMapSeq, ChunkedSeq/WindowsSeq/IntersperseSeq, Chain — route: bounded worker — surfaces: iterx/* — commit: 607acb1
- T5a (S1, S5.5): either package (Either[L,R] + Map/Bifold/ToResult) — route: bounded worker — surfaces: either/* — commit: c1264b3
- T5b (S1, S5.5): flow Race first-success + cancelación — route: bounded worker — surfaces: flow/* — commit: 8d7ee0d
- T6 (S6): verifier independiente + native review non-docs + close-out — route: parent — commit: n/a

## Log

- L1 (request verbatim): "Vamos en ese orden, en autonomo como la sesion previa, con las mismas reglas."
- L2 (2026-10-20): secuencia autónoma — T1 docs primero (fija patrón de guías), luego T2/T3/T4/T5a/T5b en paralelo con superficies disjuntas; verifier antes que review; pushes owner-gated.
- L3 (2026-10-20): T1 DONE verificado por parent (118 inserciones docs-only, build/doc/gofmt OK) — commit 33f40bb, sin review nativo (S6 docs rule).
- L4 (2026-10-20): T2-T5b DONE verificados por parent (spot-checks + fixes: SortBy→SortedStableFunc, IsDisjoint→método, Race→raceRunner por gocyclo; suite 15 pkgs + lint 0 issues) — commits 725c493, db20ebc, 607acb1, c1264b3, 8d7ee0d.
