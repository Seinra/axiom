# pool — reuse transient buffers, never cache

You borrow recyclable values instead of rebuilding them on every call.
The pool hands you a usable buffer — recycled or fresh — and you `Put` it
back when done.

**Version:** `github.com/seinra/axiom`, Go 1.27.
**Prereqs:** `import "github.com/seinra/axiom/pool"`. Keep the `*Pool`
from `New` and share it; never copy a `Pool` value after use.

## Quick path

1. Build with `pool.New(maker)` where `maker` creates a fresh value.
2. `Get` a value, use it, `Put` it back (reset first, e.g. `buffer[:0]`).
3. Assert usability of what `Get` returns, never its identity.

## Which function?

The package doc (`go doc ./pool`) is the source; this table mirrors its
"Which function?" table:

| When you want to ... | Reach for ... |
|---|---|
| Hold recyclables; build with maker | `Pool` / `New` |
| Borrow a value (or zero Item) | `Pool.Get` |
| Return a value for reuse | `Pool.Put` |

## Example

From `pool/example_test.go` (`ExampleNew` — the typed Get/Put path):

```go
buffers := pool.New(func() []byte { return make([]byte, 0, 4) })
buffer := append(buffers.Get(), "ok"...)
fmt.Println(string(buffer))
buffers.Put(buffer[:0])
fmt.Println(len(append(buffers.Get(), 'x')) > 0)
// Output:
// ok
// true
```

Run it: `go test ./pool/ -run ExampleNew -v`.

## Contracts and errors

| Contract | What it means for you |
|---|---|
| Validity, not identity | The runtime may drop any `Put` item at any GC. Never assert `Get` returns the exact item you `Put` — assert it is usable (right type, sufficient capacity) and that the maker ran at least once. |
| Not a cache | Items can vanish between `Put` and `Get`. Anything you must find again belongs in a map or a `repo.Repository`, never in a pool. |
| Zero value is safe | A zero `Pool` with no maker returns the zero `Item` from `Get` instead of panicking. A `nil` maker falls back to a zero-`Item` maker, so `Get` never panics. |
| Reuse must pay for itself | The pool only pays off for transient buffers callers `Put` back. When reuse buys nothing, use a plain value directly. |
| Do not copy after use | The pool embeds a `sync.Pool`; always share the pointer from `New` (or a heap address of a zero `Pool`). |

## Next step

When records need identity and paging instead of throwaway reuse, see
`docs/guides/repo.md`. For collapsing duplicate concurrent work rather
than recycling buffers, see `docs/guides/singleflight.md`.
