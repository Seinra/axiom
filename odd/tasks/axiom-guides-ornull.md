# Axiom guides + OrNull — documentation track (next session kickoff)

Docs-only track: complete the Swift-style family guides and make OrNull systematic. No API changes. Research (web evidence) runs FIRST next session, then design lock, then docs+examples. Nothing is implemented in this session.

## Specs

- S1: "Ok, ahora hablame de esto:

guías por familia estilo Swift (los godocs por paquete lo cubren a medias) y
 el OrNull sistemático (nuestro (T, bool) es el equivalente, bien).

Que nos falta exactamente para tenerlo completo y como lo podemos completar? Respondeme, nada de planificar implementaciones aun"
- S2: "Anota todo al loop y prepara un ODD, no implementaremos el ODD en esta sesion, en la proxima sesion lo iniciamos pero investigando primero lo que dice el internet asi encontramos las referencias que mejor se adapten a nuestro enfoque."
- S3 (evidence snapshot 2026-10-20, parent-verified via rg): Which-tables in 7/15 package docs (opt, set, result, mapx, slice, either, iterx); missing in valid, singleflight, pool, di, repo, repogorm, httpx, flow; either tabled but unindexed in README; README index covers 6 pkgs + one OrNull paragraph (slice-only); OrNull documented only in slice (FindSlice/Min/Max/MinBy/MaxBy/Average) plus 2 unnamed iterx callback shapes (Generate, FilterMapSeq); compiling examples in 2/15 (httpx, repogorm); README counts stale (14 vs 15 pkgs).
- S4 (completion bar, owner-confirmable): every package doc opens with a Which-table; README index covers all 15 plus a cross-family decision ladder; one canonical OrNull statement with per-site one-line pointers; one compiling Example per happy path; zero stale counts/names; lesson-framing location decided.
- S5 (open scope decision): teachability lesson framing (prerequisites/exercises per VISION DLC) — location undecided: Catalyst DLC material vs Axiom godoc.

## Tasks (next session, in order; T3+ detailed at kickoff)

- T1 (S2): INVESTIGATE — parent-inline web evidence, three lanes: (RQ1) Swift stdlib/guide style references fitting outcome-first tables; (RQ2) OrNull equivalents (Rust Option docs? Kotlin nullable contracts? Swift Optional binding docs? LINQ FirstOrDefault conventions?) and how they document the decision ladder; (RQ3) teachability precedents (stdlib guided tours vs course-material split). Route: read-only research, cited findings with pins. No implementation.
- T2 (S3, S4, S5): design lock — Which-table template, canonical OrNull statement + placement, decision-ladder placement, DLC-scope decision (owner sign-off). Route: parent + owner.
- T3 (S4): implement docs + compiling examples in bounded surface batches. Route: decided at next-session kickoff under the Writer rule. NO native review for docs-only (docs rule); build gate + independent verifier.
- T4: independent verifier (docs lens) + close-out + delivery. Route: parent commits on feature branch; push/PR/merge owner-gated.

## Log

- L1 (request verbatim): "Ok, ahora hablame de esto:

guías por familia estilo Swift (los godocs por paquete lo cubren a medias) y
 el OrNull sistemático (nuestro (T, bool) es el equivalente, bien).

Que nos falta exactamente para tenerlo completo y como lo podemos completar? Respondeme, nada de planificar implementaciones aun"
- L2 (request verbatim): "Anota todo al loop y prepara un ODD, no implementaremos el ODD en esta sesion, en la proxima sesion lo iniciamos pero investigando primero lo que dice el internet asi encontramos las referencias que mejor se adapten a nuestro enfoque."
- L3 (2026-10-20): evidence measured (S3); ODD prepared with research staged; no implementation this session by owner order; next session starts at T1 INVESTIGATE, then owner reference pick.
- L4 (2026-10-20, disposition): candidate af31eba3 (ODD doc + 1 STATUS row, docs-only, sin código) explicitly left unreviewed by owner decision — standing docs-only rule (no native review for prose; build gate + human review apply).
- L5 (2026-10-21, T1 DONE + T2 signed off): repo claims verified via rg (7 Which-tables; README 6-index + either missing + 14-vs-15; OrNull slice-only + 3 unnamed iterx callbacks incl. Unfold; Examples 2/15); web evidence RQ1 Swift guidelines/DocC/Optional, RQ2 Rust Option + Kotlin null-safety + LINQ FirstOrDefault (+ SO ambiguity trap), RQ3 Rust Book/RBE/Rustlings + Kotlin Koans + Swift TSPL + Echo guide/cookbook. Owner lock: tables extended, canon OrNull in README + absence ladder same page, 9 one-line pointers, Axiom standalone teachability (S5 decided → VISION invariant 8). Next: T3 implementation.
- L6 (2026-10-21, T3 DONE + T4 Go): 4 commits docs (README 15-index + canon/ladders, iterx pointers, 8 Which-tables, 13 Examples) + 1 fix (verifier SUGGESTION: either.Either symmetry). Verifier Go, 0 BLOCKER/WARNING. Evidence: suite 15 pkgs ok, vet 0, gofmt clean, lint 0 issues. Push/PR owner-gated.
