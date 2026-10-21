# httpx — one JSON envelope for Echo handlers

You speak a single JSON envelope: success sets `data`, failure sets
`error`, exactly one per response. Parse with `Decode`, reply with
`WriteJSON` / `WriteError`, and surface rule failures as 422 with
`ValidationWriteError`.

**Version:** `github.com/seinra/axiom`, Go 1.27.
**Prereqs:** `import "github.com/seinra/axiom/httpx"` plus
`github.com/labstack/echo/v5`. Validation itself belongs to `valid` at
the handler edge — this package only renders the outcome.

## Quick path

1. Parse: `payload, err := httpx.Decode[CreateLesson](httpContext)`.
2. Validate with `valid`, then reply: `httpx.WriteJSON(httpContext, http.StatusCreated, lesson)`.
3. On failure: `httpx.WriteError(httpContext, status, err)` — or `httpx.ValidationWriteError(httpContext, failures)` for rule reports.

## Which function?

The package doc (`go doc ./httpx`) is the source; this table mirrors its
"Which function?" table:

| When you want to ... | Reach for ... |
|---|---|
| Parse the request body | `Decode` |
| Reply with a success envelope | `WriteJSON` (`Response`) |
| Reply with a failure envelope | `WriteError` (`Response`) |
| Reply validation failures as 422 | `ValidationWriteError` |
| Shape a paged listing payload | `ListResponse` |

## Example

From `httpx/example_test.go` (`ExampleWriteJSON`):

```go fragment: excerpt of ExampleWriteJSON — full runnable at httpx/example_test.go
engine := echo.New()
httpRequest := httptest.NewRequest(http.MethodGet, "/", nil)
recorder := httptest.NewRecorder()
httpContext := engine.NewContext(httpRequest, recorder)
_ = httpx.WriteJSON(httpContext, http.StatusOK, "ready")
fmt.Println(recorder.Body.String())
// Output: {"data":"ready"}
```

Run it: `go test ./httpx/ -run ExampleWriteJSON -v`.

## Contracts and errors

| Contract | What it means for you |
|---|---|
| Scalar-zero bodies encode keyless | `Data` carries `omitempty`, so a scalar zero body (`0`, `""`, `false`) encodes as `{}` with no `data` key. Read a keyless `{}` success as "zero value", never as failure. Prefer struct, slice, or map bodies: a zero struct still emits `"data":{}`. |
| `Decode` names the target type | Bind failures wrap the Echo error with the `%T` of the target, so the message tells you which payload shape broke. |
| `WriteError` never writes an empty envelope | A `nil` failure becomes a descriptive message instead of an empty `error` key. You choose the status; the envelope carries the text. |
| `ValidationWriteError` is always 422 | It joins the failure messages with `"; "`, skips `nil` entries, and writes the envelope with `http.StatusUnprocessableEntity`. Empty or all-`nil` failures write a descriptive guard message, never an empty envelope. Feed it the `Validate` output from `valid` — see `docs/guides/valid.md`. |
| `ListResponse` shapes paged listings | It carries `Items` with the `repo.Page` cursor and the total count — the same `Page` convention from `docs/guides/repo.md`. |

No middleware, no router helpers, no validation here. For transport
itself, reach for `net/http` or Echo directly.

## Next step

Pair this envelope with `valid.Validate` → `ValidationWriteError` at the
handler edge, and with `repo.Page` + `ListResponse` for listings. For the
absence ladder that feeds validation, see recipe (a) in
`docs/cookbook.md`.
