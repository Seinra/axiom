# valid — combine small predicates, collect every failure

You validate one candidate against named rules and get every failure back,
not just the first. Define a `Predicate` per condition, wrap each in a
`Rule` with a name and message, collect them with `New`, then call
`Validate` for all failures or `Valid` for the boolean shorthand.

**Version:** `github.com/seinra/axiom`, Go 1.27.
**Prereqs:** `import "github.com/seinra/axiom/valid"`. No dependencies
beyond the standard library.

## Quick path

1. Write one `Predicate[Value]` per condition (`func(Value) bool`).
2. Name each with a `Rule` (`Name` + `Check` + `Message`).
3. Collect with `New`, run `Validate` (all failures) or `Valid` (bool).

## Which function?

The package doc (`go doc ./valid`) is the source; this table mirrors its
"Which function?" table:

| When you want to ... | Reach for ... |
|---|---|
| Express one check as a value | `Predicate` |
| Combine checks | `And` / `Or` / `Not` |
| Name one check with its message | `Rule` (+ `Apply`) |
| Collect rules into a set | `New` |
| Run every rule, collect all failures | `Validator.Validate` |
| Boolean shorthand for one candidate | `Validator.Valid` |

## Example

From `valid/example_test.go` (`ExampleNew` — the exam-score path):

```go fragment: excerpt of ExampleNew — full runnable at valid/example_test.go
passing := valid.Rule[int]{
	Name:    "passing score",
	Check:   func(score int) bool { return score >= 60 },
	Message: "score below passing grade",
}
exam := valid.New(passing)
fmt.Println(exam.Valid(85))
fmt.Println(exam.Valid(42))
fmt.Println(exam.Validate(42))
// Output:
// true
// false
// [axiom valid: rule "passing score" failed: score below passing grade]
```

Run it: `go test ./valid/ -run ExampleNew -v`.

## Contracts and errors

| Contract | What it means for you |
|---|---|
| Empty `And` holds vacuously | `And()` with no legs always reports `true`; `Or()` with no legs always reports `false`. An empty `Validator` (`New()` with no rules) accepts every candidate — `Validate` returns an empty non-nil slice. Build the rule set first; an empty set is "no constraints", never "reject all". |
| Short-circuit, left to right | `And` stops at the first `false`, `Or` at the first `true`. Later legs never run after the outcome is decided — put cheap checks first. |
| `nil` legs never panic | A `nil` leg counts as `false` in `And`/`Or` (skipped in `Or`), and `Not(nil)` always reports `false`. A `Rule` with a `nil` `Check` fails `Apply` with a descriptive error naming the rule. |
| `Validate` collects everything | Unlike fail-fast `if` chains, `Validate` runs every rule and returns all failures, so callers see the full report. `Valid` is only the boolean shorthand over the same run. |
| No domain helpers on purpose | The package ships no email/password/range rules — you write your own. Reach for plain `if`s and errors directly when a single inline check reads clearer than a `Rule`. |

`New` copies the rules slice, so later caller mutations never affect the
`Validator`.

## Next step

When a rule failure must travel as an HTTP 422 envelope, wire the
`Validate` output into `httpx.ValidationWriteError` — see
`docs/guides/httpx.md`. For the absence-to-failure step before validation,
see recipe (a) in `docs/cookbook.md`.
