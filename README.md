# Axiom — typed generics for Go 1.27

One small, dependable generics library at Echo-level quality: typed helpers for results, options, and slices today; the rest of the ecosystem map declared and built one feature at a time.

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
3. Verify: `go test ./...` — expect `ok` for `opt`, `result`, `slice`.

## Details

| Topic | Decision |
| --- | --- |
| Module / toolchain | `github.com/seinra/axiom`, `go 1.27` (generic methods on concrete types; no generic interface methods) |
| Ships now (14 pkgs) | `result`, `opt`, `slice`, `set`, `mapx`, `iterx`, `valid`, `singleflight`, `pool`, `di`, `repo` (+GORM adapter `repogorm`), `httpx` (Echo v5), `flow` — see `.loop-modules/ARCHITECTURE.md` for the map; every package ships godoc + tests green |
| Rules | stdlib-first (`slices`/`maps`/`cmp`/`iter` before hand-rolled); typed API, `any` only at codec edges; grep-able names (no `data`/`item`/`req`/`res`); godoc + example on every export |
| Loop | `AXIOM_LOOP.md` + `.loop-modules/` + `VISION.md` + `loop-config.yaml`; ODD-tracked in `odd/tasks/` |
| Declared future | `mapx`, `set`, `iterx`, concurrency (fan-in/out, pipeline, worker pool), `pool`, `singleflight`, `di`, `repo` (GORM), `httpx` (Echo), `valid` — see `.loop-modules/ARCHITECTURE.md` |
| Consumer | Catalyst (`D:\Proyectos\AI Tech\Catalyst`, offline-first course app) pilots via scratch `replace` only — never a Catalyst commit from here |

## Checklist

- [ ] `go test ./...` green on Go 1.27
- [ ] `go vet ./...` clean and `gofmt -l .` empty
- [ ] New symbol cites what stdlib lacked + carries godoc with an example
- [ ] No `any` in public signatures except codec edges with typed wrappers

## Next step

Read `VISION.md` for the product line, then `.loop-modules/ARCHITECTURE.md` for the coverage map — each future package becomes one ODD feature.
