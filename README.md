# Axiom — typed generics for Go 1.27

One dependable, composable, educative generics library: typed helpers for results, options, and slices today; the rest of the ecosystem map declared and built one feature at a time.

## Quick path

1. Point at it (private library — `replace`, not `go get`):
   ```go
   // go.mod of the consumer (e.g. Catalyst)
   require github.com/seinra/axiom v0.0.0
   replace github.com/seinra/axiom => ../Axiom
   ```
2. Use one package:
   ```go
   import (
       "github.com/seinra/axiom/result"
       axiomslice "github.com/seinra/axiom/slice"
   )

   current := result.Ok(21)
   doubled := axiomslice.MapSlice([]int{1, 2, 3}, func(success int) int { return success * 2 })
   _ = current.MapValue(func(success int) int { return success * 2 })
   ```
3. Verify: `go test ./...` — expect `ok` across all packages.

## Which function? (guide index)

Each package doc opens with a "Which function?" table (outcome → function).
Start there, then `go doc` the package:

| If you are holding ... | Start with | Guide |
| --- | --- | --- |
| new to generics syntax (`[T]`, constraints) | the foundations chapter first | [docs/guides/foundations.md](docs/guides/foundations.md) |
| a maybe-present value | `opt` | `go doc ./opt` ([opt/opt.go](opt/opt.go)) |
| a success-or-failure outcome | `result` | `go doc ./result` ([result/result.go](result/result.go)) |
| two typed sides (left / right) | `either` | `go doc ./either` ([either/either.go](either/either.go)) |
| a slice to transform, find, or fold | `slice` | `go doc ./slice` ([slice/slice.go](slice/slice.go)) |
| distinct values to combine or compare | `set` | `go doc ./set` ([set/set.go](set/set.go)) |
| a builtin map to reshape | `mapx` | `go doc ./mapx` ([mapx/mapx.go](mapx/mapx.go)) |
| a lazy `iter.Seq` pipeline | `iterx` | `go doc ./iterx` ([iterx/iterx.go](iterx/iterx.go)) |
| predicates to combine (And / Or / Not) | `valid` | `go doc ./valid` ([valid/valid.go](valid/valid.go)) |
| one call to run once per key | `singleflight` | `go doc ./singleflight` ([singleflight/singleflight.go](singleflight/singleflight.go)) |
| objects to reuse, not rebuild | `pool` | `go doc ./pool` ([pool/pool.go](pool/pool.go)) |
| named singleton services | `di` | `go doc ./di` ([di/di.go](di/di.go)) |
| records addressed by identity | `repo` | `go doc ./repo` ([repo/repo.go](repo/repo.go)) |
| the same over GORM | `repogorm` | `go doc ./repogorm` ([repogorm/repogorm.go](repogorm/repogorm.go)) |
| an Echo JSON envelope | `httpx` | `go doc ./httpx` ([httpx/httpx.go](httpx/httpx.go)) |
| a bounded concurrent map or first-success race | `flow` | `go doc ./flow` ([flow/flow.go](flow/flow.go)) |
| a LIFO work pile (undo, depth-first) | `stack` | `go doc ./stack` ([stack/stack.go](stack/stack.go)) |
| a FIFO line (tickets, breadth-first) | `queue` | `go doc ./queue` ([queue/queue.go](queue/queue.go)) |
| records that must remember insertion order | `orderedmap` | `go doc ./orderedmap` ([orderedmap/orderedmap.go](orderedmap/orderedmap.go)) |
| a background computation to join later | `task` | `go doc ./task` ([task/task.go](task/task.go)) |
| whole numbers to test, divide, or combine | `mathx` | `go doc ./mathx` ([mathx/mathx.go](mathx/mathx.go)) |
| text to split or cut by runes | `strings` | `go doc ./strings` ([strings/strings.go](strings/strings.go)) |

`(T, bool)` is our OrNull across Axiom: the `bool` reports presence — never trust the value without it. Slice helpers (`FindSlice`, `Min`/`Max`, `MinBy`/`MaxBy`, `Average`) and `iterx` callbacks (`Generate`, `FilterMapSeq`, `Unfold`) use this shape.

| If you are holding ... | Reach for |
| --- | --- |
| a loose value that may be missing, your own code | `(T, bool)` |
| an absence to chain or transform | `opt.Option` |
| a failure with a cause | `result.Result` |
| two typed sides | `either.Either` |

## Details

| Topic | Decision |
| --- | --- |
| Module / toolchain | `github.com/seinra/axiom`, `go 1.27` (generic methods on concrete types; no generic interface methods) |
| Ships now (21 pkgs) | `result`, `opt`, `either`, `slice`, `set`, `mapx`, `iterx`, `valid`, `singleflight`, `pool`, `di`, `repo` (+GORM adapter `repogorm`), `httpx` (Echo v5), `flow`, `stack`, `queue`, `orderedmap`, `task`, `mathx`, `strings` — see `.loop-modules/ARCHITECTURE.md` for the map; every package ships godoc + tests green |
| Rules | stdlib-first (`slices`/`maps`/`cmp`/`iter` before hand-rolled); typed API, `any` only at codec edges; grep-able names (no `data`/`item`/`req`/`res`); godoc on every export + one compiling Example per package happy path |
| Loop | `AXIOM_LOOP.md` + `.loop-modules/` + `VISION.md` + `loop-config.yaml`; ODD-tracked in `odd/tasks/` |
| Declared future (owner-gated) | Adoption program closed: P1 Head/Tail/Initial, P2 `mathx`, P3 opt bridges, P4 `flow` retry, P5 `strings` all on main; new packages still need an explicit owner package-creation call before any ODD — see `.loop-modules/ARCHITECTURE.md` |
| Consumer | Catalyst (offline-first course app, module `catalyst`) pilots via scratch `replace` only — never a Catalyst commit from here |

## Checklist

- [ ] `go test ./...` green on Go 1.27
- [ ] `go vet ./...` clean and `gofmt -l .` empty
- [ ] New symbol cites what stdlib lacked + carries godoc with an example
- [ ] No `any` in public signatures except codec edges with typed wrappers

## Why not just use `lo` (or `mo`)?

Use them — they are excellent, battle-tested reference libraries, and Axiom will never try to out-cover them. Axiom is the teachable companion, not a replacement: every package opens with a Which-function table, every export carries godoc plus a compiling example, and every addition documents what the stdlib lacked. It doubles as the Catalyst generics course material, and its v0 API is frozen — the same library students learn on is the one they keep.

## Contributing

External PRs are judged by one rule: genuinely good **and** thesis-fitting, or fork it. See [CONTRIBUTING.md](CONTRIBUTING.md) before opening anything — most well-meaning asks (stdlib duplicates, reflection-first APIs, mutable variants, domain helpers) are declined there with reasons.

## Next step

Read `VISION.md` for the product line, then `.loop-modules/ARCHITECTURE.md` for the coverage map — each future package becomes one ODD feature.
