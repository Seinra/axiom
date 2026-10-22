# Axiom F2 — OrderedMap (+LRU) (declared-future slice F2)

Second declared-future slice from the program (`odd/tasks/axiom-future-slices.md`): insertion-ordered map; LRU as eviction policy on top. Lesson: "maps don't remember order"; caching. stdlib has no ordered map.

## Specs

- S1 (owner orders verbatim 2026-10-22): "Vamos con los slices declarados" + "Y lanza F2 despues" (F2 after F1 per program order; F3-F4 still staged).
- S2 (scope, parent 2026-10-22): `orderedmap` package (full name, no abbreviations), single impl file (Map + LRU share one boundary until it earns a split).
  - `Map[Key comparable, Value any]` + `New() *Map[Key,Value]`; methods `Put(key Key, value Value)`, `Get(key Key) (Value, bool)`, `Delete(key Key)`, `Len() int`, `IsEmpty() bool`, `Keys() []Key` (insertion order), `Values() []Value` (insertion order).
  - Upsert keeps first-insertion position (documented); `Delete` of a missing key is a no-op; re-`Put` after `Delete` takes the back.
  - Zero value ready via lazy init (methods initialize nil backing on first use).
  - `LRU[Key comparable, Value any]` + `NewLRU(capacity int) *LRU[Key,Value]` (`capacity < 1` panics, programmer error); methods `Get(key Key) (Value, bool)` (hit refreshes recency), `Put(key Key, value Value)` (insert/update refreshes; beyond capacity evicts least-recently-used), `Len() int`.
  - No stats surface (Stats-grade observability decided against: singleflight precedent — stats rejected as non-goal, YAGNI; revisit only on explicit owner call).
  - Eviction/recency order pinned by tests (order, never timing); complexity documented in godoc.
  - Do-not-copy documented (headers share backing).
  - No concurrency (F3 owns that); stdlib only.
- S3 (house style, program S5): package doc with Which-table + godoc stdlib-gap justification + compiling Example per type (Map lesson, LRU lesson); full names; OrNull `(V,bool)` shapes; panic policy above; no reflection.
- S4 (gates, standing): `go test -count=1 ./...` green, `go vet` 0, `gofmt` empty, lint 0; bench smoke still green; TDD RED→GREEN recorded.

## Tasks

- T1 (S1, S2, S3): Implement orderedmap package (Map + LRU) with tests + Which-table rows + Examples — route: bounded writer — surfaces: orderedmap/orderedmap.go, orderedmap/orderedmap_test.go, orderedmap/example_test.go — skills: foundry + foundry-testing + cognitive-doc-design — commit: 336f457
- T2 (S4): Independent verifier + close-out — route: parent + gentle-ai-verify; push/PR owner-gated (new issue + approval) — commit: pending

## Log

- L1 (request verbatim): "Y lanza F2 despues"
- L2 (2026-10-22): rama feat/axiom-f2-orderedmap desde main (post-#52); F2 per program order; LRU observability decided: no stats (singleflight precedent); ODD + mirror listos; T1 a writer acotado; delivery needs new issue + approval.
- L3 (2026-10-22, T1 DONE writer mvja81mi-5-9y2i + parent-verified): orderedmap Map + LRU (upsert-keeps-position, lazy zero value, Get-refreshes/Put-evicts, no stats); parent re-run suite 19/19 -count=1 ok, vet 0, gofmt vacío, lint 0, bench smoke PASS; parent read impl file fully; verifier mvjac3pv-6-lcua Go (0 BLOCKER, 1 SUGGESTION isolation-pin + 3 auth-only WARNINGs — all closed by parent: isolation test added + PASS, lint/bench/TDD evidence parent-run); work-unit 336f457 (530+/0, single logical unit over the 400-line heuristic). T2 close + delivery (new issue + approval, native review owner-call).
- L4 (2026-10-22, native review quemado a pedido owner): lineage review-864498bfb3f6a9f2 APPROVED + burned (readability+reliability, 561 líneas, group capture limpio tras un reenvío por token arrastrado); 3 SUGGESTIONs informativas parkeadas sin corrección (R2-001 orderedmap_test.go:73-74, R3-001 orderedmap.go:192-195, R3-002 orderedmap.go:30-39); push/PR owner-gated (nuevo issue + aprobación).
- L5 (2026-10-22, delivery a pedido owner): rama pusheada a origin; issue #53 type:feature creado (sin approved inicial — fail-closed) + status:approved por grant exacto posterior (read-back confirmed); PR #54 type:feature con Closes #53; validate falló solo por approval pendiente, re-disparado vía close/reopen tras el grant → validate+verify pass; merge owner-gated.
