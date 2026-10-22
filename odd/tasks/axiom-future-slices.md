# Axiom future program — declared futures + thesis-compatible adoptions (no implementation)

Program document: what is declared, what the survey surfaced that fits the thesis, and in which order to attack it. Each slice launches as its own ODD when the owner picks it. Nothing here implements anything.

## Specs

- S1 (owner order verbatim): "Verifica bien que pedo con ellos a ver si les podemos robar algo que si sea compatible con nuestra tesis. Hay que preparar los slices de futuro declarado."
- S2 (owner agreements verbatim): purity invariant ratified ("Me gusta lo de invariante de pureza"); Lazy covered by `sync.OnceValue` and tuples T2-T9 rejected as API explosion are "lo mismo" (agreed split).
- S3 (gate, standing): a candidate enters iff (a) stdlib genuinely lacks it AND (b) it teaches as Catalyst lesson material with a stdlib-first core (godoc + Which-table + compiling Example). Rejections carry the failed gate clause.
- S4 (evidence snapshot 2026-10-22, parent-verified): lo v1.53 (~21k stars) + mo v1.17 (~3.4k) READMEs read live; go-utils README read live (9 pkgs, per-function docs, AI rules, 100% coverage claim); fp-go README read live (fp-ts port, WIP); gods v2 generics noted; Axiom exports verified via `grep ^func` (slice 25, mapx 12 incl. Pick/Omit/Merge/ValueOr, set full algebra, iterx 22 lazy adapters, flow Map/ForEach/Race).
- S5 (owner order verbatim 2026-10-22): "todo lo que robamos se implementa a nuestro estilo" — adopted behavior ships under Axiom house style even where foundry skills already push that way: element-only callbacks (indexed variants explicit, never `(v,i,s)`); Which-table row + godoc + compiling Example per export; purity (frozen decision 8); OrNull `(T,bool)` shapes for absence; panic policy for programmer errors (Chunked/Windows precedent); no reflection; stdlib-gap justification recorded; edge + never-mutate pins in tests. Writer handoffs MUST cite ARCHITECTURE + this S5, not just foundry.

## Adoption list (thesis-compatible only, ranked)

| Rank | Item | Who has it | stdlib gap | DLC lesson | Suggested home |
| --- | --- | --- | --- | --- | --- |
| 1 | Head/Last/Initial/Tail | go-utils, lo | no first/rest/last helpers | "first/rest" FP basics, OrNull-shaped on empty | slice |
| 2 | FindOrElse/FirstOr/LastOr | lo | no OrElse-shaped find | extends the OrNull canon | slice |
| 3 | Option/Result JSON + sql codecs | mo | no Option/Result JSON/sql support | "null vs absent" in JSON/DB | opt, result (codec edge, `any` allowed) |
| 4 | Range/RangeFrom (+Clamp) | lo | no int ranges; no Clamp | "generate, don't allocate" + bounds | slice (or new pkg — owner call) |
| 5 | Every/Some/None | lo, go-utils | no quantifiers | quantifiers over `valid` predicates | slice |
| 6 | FindDuplicates/FindUniques/FindIndexes | lo, go-utils | no | "dedupe", pairs with Distinct | slice |
| 7 | Take/Drop/TakeLast/SkipLast (+While) | lo, go-utils | no eager take/drop | "slicing and bounds", pairs with Chunked/Windows | slice (iterx has lazy Take/Drop/While) |
| 8 | Pluck | lo, go-utils | no field projection | "shape your data" without reflection | slice |
| 9 | Product/Mode | lo | no | "overflow-safe arithmetic", pairs with Average | slice |
| 10 | FromEntries/ToEntries (+ChunkEntries) | lo, go-utils (maputils) | no pairs↔map bridge | "pairs become maps", pairs with Zip | mapx |
| 11 | EnsureUniqueAndAppend | go-utils | no | set-building on slices | slice |
| 12 | CrossJoin | lo | no | pairs, exam use-case cousin of Combinations | slice |
| 13 | Shuffle | lo | thin (stdlib `rand.Shuffle` exists; typed wrapper + lesson) | Fisher-Yates + exam options | slice |
| 14 | IsPrime (+Gcd/Lcm weak) | go-utils mathutils | no | primes sieve with Range+Generate | new `mathx`? (owner package call) |
| 15 | Ptr bridges Deref/Coalesce/To | go-utils ptrutils | no pointer↔value OrNull | "nil-safety at the edges" | opt? or new pkg (owner call) |
| 16 | Retry/RetryWithResult | go-utils excutils | no general retry | resilience, pairs with flow.Race | flow? (owner call; di retries constructors already) |
| 17 | Words/ChunkString/Truncate | lo, go-utils stringutils | no (case converters live in x/text) | "runes vs bytes" | new strings pkg (owner package call) |

## Rejected (with failed gate clause)

- Struct reflection (go-utils structutils): fails purity-adjacent VISION non-goal (no reflection-first APIs).
- Date/URL domain helpers (go-utils dateutils/urlutils): fail gate (b) — domain, not generics-lesson material. (Query-stringify may revisit as httpx follow-up; weak.)
- Either3-5, Pipe1-10, State/Reader/IO/Task (mo, fp-go, go-monad): fail gate (b) — anti-idiomatic Go; course core is stdlib-first.
- Mutable in-place variants (lo/mutable): fail frozen decision 8 (purity).
- sort/reverse/Repeat/Concat/Clone/Compact/IsSorted/Includes/Min/Max-plain/Insert/Delete/Replace/Splice, `sync.OnceValue` memoize, atomic wrappers: fail gate (a) — stdlib-owned, never reimplement. (Axiom Min/Max stay: OrNull `(T,bool)` shape is the differentiator, precedent for thin wrappers.)
- Tuples T2-T9 + Zip2-9/Unzip (lo): fail teachability-per-cost — Go lacks variadic generics; N hand-written types for a weak lesson. `Zip`/`Pair` covers the common case.
- `Lazy` (go-monad): fail gate (a) — covered by `sync.OnceValue`.
- gods containers (OrderedMap/Stack/Queue/LRU): not rejected — already declared future (F1/F2 below).

## Declared-future program (scopes + order)

| Slice | Scope | Lesson | Gate notes |
| --- | --- | --- | --- |
| F1 Stack/Queue | LIFO/FIFO over slices, pure constructors + methods, godoc + Which-table + Example | data structures 101, "structure your work" | smallest; stdlib has `container/list` (untyped) and no generic stack/queue |
| F2 OrderedMap (+LRU) | insertion-ordered map; LRU as eviction policy on top | "maps don't remember order"; caching | stdlib has no ordered map; LRU needs Stats-grade observability decision (see singleflight Stats precedent) |
| F3 fan-in/out pipeline | bounded fan-out + ordered fan-in over `iter.Seq`, errgroup parity with flow | concurrency you can reason about | extends flow; cancellation + backpressure policy required (flow.Map precedent) |
| F4 Future/Task async | deferred async computation with cancel | async done right, last for a reason | weakest stdlib-first story; needs the strongest lesson justification; goes last |

Order: adoption round (table above, slice-first) → F1 → F2 → F3 → F4. One ODD per slice, owner picks each launch. New packages (`mathx`?, strings, ptr) need an explicit owner package-creation call before any ODD.
- L3 (2026-10-22, review disposition wording-fix candidate): sha256:cab5f0bd… (4 palabras, pura prosa) explícitamente sin review nativo por owner — regla docs; archivado sin corrección.

## Log

- L1 (request verbatim): "Verifica bien que pedo con ellos a ver si les podemos robar algo que si sea compatible con nuestra tesis. Hay que preparar los slices de futuro declarado. Me gusta lo de invariante de pureza y coincido con tu juicio. Lo mismo a tu punto 3 de Lazy y tuplas."
- L2 (2026-10-22): go-utils README verified live (9 pkgs, per-function docs, AI rules, `(v,i,s)` callback convention differs from Axiom element-only); corrections applied (Lazy≠tuples split); purity ratified as frozen decision 8; program prepared, no implementation.
