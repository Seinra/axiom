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
- L5 (2026-10-20): T6 close — verifier Go (0 BLOCKER/WARNING, 3 SUGGESTIONs parked para docs pass: padding de tablas, params Left/Right, ejemplos compactos). Native review no disponible para el rango T2-T5b (2 START bloqueados candidate-mismatch; facade ofrece solo branch-wide con P0/P1 ya quemados). Evidencia: suite 15 pkgs verde + lint 0 issues + -race en iterx/flow/either. Push a decisión del owner.
- L6 (2026-10-20): review por slices intentado (skill work-unit-commits: 5 commits ya eran work units correctos) — 5 worktrees ../w-exp-* con diffs aislados (196/536/722/422/292); R1 slice START OK sin consent pero capture_group rechazado 3 veces por transporte de bindings (STATUS fresco confirmó sin drift) → STOP por failure policy, worktrees removidos; cierre con verifier+suite+lint.
- L7 (2026-10-20): owner ordenó abandonar lineages activos + probar repo paralelo Axiom-1 — lineage R1 review-266a6029d4739212 abandonado (operator_disposition, committed, sin findings, en cuarentena); Axiom-1 clonado con rama r1-t2-slice (experimento detenido antes del START); worktree w-exp-t2 removido.
- L8 (2026-10-20): R1 slice T2 quemado en Axiom-1 (lineage review-82c353710ed9e967, readability+reliability, APPROVED + authority burned; 3 SUGGESTIONs informativas parkeadas R2-001/002/003). Clave aprendida: capture_group/single fallan si el binding va envuelto en {collectBinding}; el binding crudo del STATUS sí pasa.
- L9 (2026-10-20): R2 extras T3 quemado en Axiom-1 (lineage review-98a70286018a012c, readability+reliability, APPROVED + authority burned; 1 SUGGESTION informativa parkeada R3-flatten-zero-inner-coverage en result.go:136-138). Hallazgo propio: mi transcripción manual mete llave de más antes de "submission" y hashes recortados — verificar hashes completos antes de reenviar.
- L10 (2026-10-20): R3 iterx T4 quemado en Axiom-1 (lineage review-c8924c5717aab643, readability+reliability, APPROVED + authority burned; 1 SUGGESTION informativa parkeada R2-001 en iterx.go:385-386).
- L11 (2026-10-20): R4 either T5a quemado en Axiom-1 (lineage review-612ac359ff6bc350, readability+reliability, APPROVED + authority burned; 4 informativas parkeadas R2-001 WARNING either.go:76-77, R2-002/R2-003 SUGGESTIONs en test, R3-001 WARNING either.go:106-112).
- L12 (2026-10-20): R5 flow Race T5b quemado en Axiom-1 (lineage review-6286a8e56d8db9ea, resilience+readability+reliability, APPROVED + authority burned; 7 informativas parkeadas: R2-001/002/003, R3-cancel-under-lock, R3-doc-race-window WARNING, R4-1/R4-2). Cadena R1-R5 completa y quemada.
- L13 (2026-10-20): delivery — non-goals persistidos, rama pusheada, issues #14 (feature) + #15 (docs polish) + #16 (flatten) + #17 (race) + #18 (either) creados con read-back confirmado; pendiente aprobación de #14 para PR.
