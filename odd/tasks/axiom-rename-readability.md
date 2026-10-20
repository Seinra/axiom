# Axiom rename readability-first — design document (next session kickoff)

Owner principles 2026-10-20: (1) human readability first, students, pre-v0 window with zero consumers; (2) convenience outweighs minimalism; (3) readability first again; (4) research dynamic/adaptive bounds + countermeasures; (5) engineer opinion with cross-library panic research; (6) combinatorics parked until the class exists; (7) timing research. No subagents launched yet — all below is parent proposal + staged research, ready to execute next session.

## Specs

- S1: "1) Legibilidad humana primero, necesitamos que los alumnos entienda el codigo que implementan y leen. Se aprovecha ahora que todavia nadie consume Axiom."
- S2: "2) stdlib como preferencia, pero conveniencia pesa mas que minimalismo (ej para ahorrar lineas de codigo, y otros casos de conveniencia), ese nivel de dinamismo tiene que tener esta libreria."
- S3: "3) De nuevo, legibilidad humana primero."
- S4: "4) Investigacion para lograr algo dinamico y adaptable, con comentarios u otras formas de contrarestar/informar el problema/debilidad de lo elegido de acuerdo a la situacion."
- S5: "5) Aca necesitamos opinion de ingeniero con investigacion de como lo resolvieron otras librerias/frameworks."
- S6: "6) Meh, si el problema es que aun no existe clase, se deja parked hasta que exista."
- S7: "7) Invesgicacion."
- S8: "No lances sub-agentes aun, simplemente prepara todo para dejarlo listo y arrancar en la proxima sesion."

## Rename proposal (parent architect call, needs owner sign-off per row)

| Current | Proposed | Why | Price |
|---|---|---|---|
| `either.Map` | `MapRight` | asymmetric partial map is the documented trap; mirrors `MapLeft` | cheap, symmetric |
| `singleflight.Do` | `Call` | `Do` carries zero intent; `flight.Call(key, fn)` reads | cheap |
| `di.Invoke` / `di.MustInvoke` | `Resolve` / `MustResolve` | domain verb; `Invoke` opaque first-read | medium (docs+tests) |
| `either.Bifold` | `Match` | Rust-match intuition for students | cheap |
| `slice.Scan` | `RunningFold` | Kotlin precedent, explicit running semantics | cheap |
| `iterx.TapSeq` | `OnEachSeq` | Kotlin `onEach` precedent; `Tap` opaque | cheap |
| `iterx.DedupSeq` | `DeduplicateSeq` | un-abbreviate (ban-14) | cheap |
| `iterx.Chain` | `ConcatSeq` | explicit concat vocabulary | cheap |
| `slice.Counts` | `Tally` | verb reads; noun read as property | cheap |
| `httpx.JSON` / `httpx.Fail` | `WriteJSON` / `WriteError` | kills `encoding/json` collision; symmetric pair | medium (adapters) |
| `httpx.Bind` | `Decode` | `Decode[Body](ctx)` explicit vs bare verb | medium |
| `httpx.ValidationFail` | `ValidationWriteError` | follow-through consistente de `Fail→WriteError` (confirmado retroactivo L6) | cheap |
| `flow.Each` | `ForEach` | explicit; `Each` empty first-read | cheap |
| `either` params `L/R` | `LeftT/RightT` | readability-first + ban-13; single letters hurt novices/grep — REVERSES #18 `L/R` call, flagged explicitly | cheap, needs explicit sign-off |
| `valid.And/Or/Not` | KEEP + vacuous-truth table | boolean ops fine qualified; churn not worth it | docs only |
| Fallback 5-way stems | KEEP + mapping table | `Else`-suffix encodes eager/lazy — good design once documented | docs only |
| `Filter`/`Flatten` families | KEEP suffixes + tables | suffixes disambiguate; rename churn massive | docs only |
| `result.Must` | KEEP + contrast table | Go `Must*` convention; table Must/MustResolve/Try | docs only |
| `New`×9 | KEEP + constructor guide | Go convention, qualified reads fine | docs only |
| Package names (`di/opt/httpx/…`) | KEEP | ecosystem precedent (`lo/mo/do`); import churn too big | — |

## Bounds sign-off (owner, R3 — cierra R4-001/R4-002)

| Decision | Sign-off |
|---|---|
| `slice.Chunked/Windows` y `iterx.ChunkedSeq/WindowsSeq` panic en `size <= 0` (paridad stdlib `slices.Chunk`); oversize/empty siguen empty | owner 2026-10-20 L7: "Panic como stdlib, rompo tests" |
| Other type params | KEEP | `T` universal; rest positional-but-harmless; document | docs only |
| bare-`any` publics | KEEP + document | GORM-forced + envelope; revisit at v0 | docs only |
| Seeding verbs | KEEP + mapping table | five verbs, one table | docs only |

## Research staged (evidence gathered, analysis pending next session)

- R-bounds: web_search responseId `mvh4ffk6jftiar` queryIndex 0 (Kotlin eager chunked/windowed vs C++ lazy views + others); fetched `mvh4fmz94ye2mq` urlIndex 2 (KEEP window-sliding). Open: synctest details, Rayon panic docs depth, PLINQ article.
- R-panic: same responseId queryIndex 1 (Rayon catch-panic thread, issues #638/#1270, PLINQ AggregateException SO). Fetched urlIndex 1 (rayon #638). Open: rayon join-panics doc, PLINQ article, Java Streams/LINQ behavior.
- R-timing: same responseId queryIndex 0? No — queryIndex 2 (Go synctest #67434, jMock scheduler, OpenDST, unflake, FakeClock). Fetched urlIndex 0 (synctest issue, 150k chars — get_search_content findText failed, retry with plain offset paging). Open: synctest availability in go1.27, clockwork-style options for plain `go test -race`.
## Research conclusions (parent, 2026-10-20 — web_search mvh4ffk6jftiar + fetches mvh4fmz94ye2mq/mvh4iqcfsaq2bp, SDD agents retired so parent-inline)

### R-bounds: dynamic/adaptive answer
- Kotlin ships BOTH models: eager `Iterable.chunked/windowed` (materialized) + lazy `Sequence` variants, switchable; plus overloads WITH transform (ephemeral window, never materialized). C++ views are lazy non-owning. Our slice(eager)+iterx(lazy) split mirrors this — validated, keep.
- New gap: chunked/windowed WITH transform (compute-without-storing) — allocation saver, teachable, fits S2.
- Package: (a) lazy-allocate ChunkedSeq/WindowsSeq backing on first element (kills one-int OOM); (b) validate positive sizes with sentinel, never silent clamp (Kotlin validates; clamp hides caller bugs); (c) O() cost model in every godoc; (d) page caps are application policy — document guidance + sentinel, never silent clamp.

## Engineer opinion: panic policy (S5 — precedents + judgment)

Precedents: Rayon `join` always runs both closures and propagates the panic to the joiner (`panic_fuse` only stops NEW items — in-flight never preempted); PLINQ/`Parallel` stop issuing new work after first exception and rethrow wrapped in `AggregateException` (in-flight completes); Java ForkJoin may process items past the exception. Consensus: nobody preempts in-flight; everybody stops scheduling new work; everybody propagates, never silent.

Opinion: adopt the consensus in `flow.Map/Race` — recover per runner, latch the FIRST panic payload, cancel siblings (same path as failure today), let in-flight complete, re-panic the original payload in the caller. Matches Rayon + our singleflight/di precedent; keeps no-silent-failure. Do NOT convert panics to errors: signatures stay, programmer bugs stay loud.

## Timing answer (S7)

`testing/synctest` EXISTS in go1.27.0 (`synctest.Test` bubbles, synthetic time — verified via `go doc`). Recommendation: synctest-first for new timing tests (migrate flow/singleflight/di wall-clock suites opportunistically); keep 1-2 real-clock smokes; document margins where wall clock remains; quarantine stays. DST frameworks are heavy artillery — overkill here.

## Tasks (next session, in order)

- R0: owner signs off rename table row by row (esp. `LeftT/RightT` reversal + `Invoke→Resolve` + `WriteJSON/WriteError`).
- R1: finish R-bounds/R-panic/R-timing evidence + engineer opinion (panic policy).
- R2: implement approved renames (bounded workers, tests+docs together).
- R3: implement approved bounds + counters.
- R4: review chain per slice + delivery.

## Log

- L1 (verbatim principles): "1) Legibilidad humana primero, necesitamos que los alumnos entienda el codigo que implementan y leen. Se aprovecha ahora que todavia nadie consume Axiom. 2) stdlib como preferencia, pero conveniencia pesa mas que minimalismo (ej para ahorrar lineas de codigo, y otros casos de conveniencia), ese nivel de dinamismo tiene que tener esta libreria. 3) De nuevo, legibilidad humana primero. 4) Investigacion para lograr algo dinamico y adaptable, con comentarios u otras formas de contrarestar/informar el problema/debilidad de lo elegido de acuerdo a la situacion. 5) Aca necesitamos opinion de ingeniero con investigacion de como lo resolvieron otras librerias/frameworks. 6) Meh, si el problema es que aun no existe clase, se deja parked hasta que exista. 7) Invesgicacion."
- L2 (verbatim): "No lances sub-agentes aun, simplemente prepara todo para dejarlo listo y arrancar en la proxima sesion."
- L3 (2026-10-20, R0 sign-off verbatim): bloque barato APROBADO TODO; bloque medio APROBADO TODO; L/R→LeftT/RightT SI (revierte #18); KEEPs NO congelados — "Reabris el que tengas que reabrir y corregir, aca nada esta congelado hasta que todo este perfecto." R1 conclusiones ya en doc (parent 2026-10-20); opens de profundidad → verifier R4; R2 lanza 3 writers paralelos (either+result / singleflight+di+flow / slice+iterx+httpx+README).
- L4 (2026-10-20, R2 DONE inline): writers rechazados por gateway (exige Spec T<n>, ordenó inline sin reintento); fallback inline con skills leídos (foundry+craft-floor, foundry-testing, cognitive-doc-design); 16 files +340/-332 commit 70955f4 en feat/axiom-rename-readability; evidencia suite 15 pkgs ok, vet 0, gofmt clean, cero refs viejas; slice 672 líneas >400 heuristic — R3 abre slice nuevo.
- L4b (2026-10-20, corrección L4): el slice real es 675 líneas (medición nativa frozen: original_changed_lines 675), no 672.
- L5 (2026-10-20, R4 review quemado): native review-3adf1ae846c2a550 APPROVED + burned (readability, 18 files/675 líneas, 3 advisories informativas R2-001/002/003); verifier Go con 8 SUGGESTIONs aplicadas en 7c4de71 (39+/39-, strings/tests/docs, verificado httpx/either/result/di ok); ValidationFail→ValidationWriteError fue follow-through consistente no listado en R0 — pendiente confirmación explícita del owner; push/PR owner-gated.
- L6 (2026-10-20, owner verbatim): ante "¿Lo confirmo retroactivamente?" eligió "Sí, confirmar fila" — ValidationWriteError queda en tabla R0.
- L7 (2026-10-20, owner verbatim, cierra R4-001/R4-002): ante "¿Qué hago?" eligió "Panic como stdlib, rompo tests" — panic en size<=0 es conducta pedida y firmada (ventana pre-v0), no regresión.
- L8 (2026-10-20, follow-ups verifier R3 aplicados): doc qualifiers (iterx package laziness call-time, Race latched-poison, callTransform panic(nil), Map godoc rejoin) + 3 tests nuevos deterministas (latch-first concurrent con barrier, panic-wins-over-failure, failure-then-panic poison); verificado suite 15 pkgs ok, flow -race PASS (5/5 el gateado), vet 0, gofmt clean, lint 0 issues.
- L8b (2026-10-20, flake real en TestRaceChampionThenPanicPoisons): el drain-on-cancel saltea trabajo no iniciado tras el win, así que el test era racy (0.00s sin panic cuando el sibling arrancaba tarde); fix con gate channel (el champion espera a que el panicker esté in-flight) — determinismo por diseño, sleep solo duración.
