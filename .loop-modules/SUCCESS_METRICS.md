# SUCCESS_METRICS

Definition of Done for Axiom. Measurable, no aspirational checkboxes.

## Skeleton slice (this feature)

- [ ] `go test ./...` green on Go 1.27.0 (paste `command: observed result`).
- [ ] `go vet ./...` clean.
- [ ] `gofmt -l .` empty.
- [ ] `result`, `opt`, `slice` each ship ≥2 exported symbols with godoc + table-driven tests incl. error paths.
- [ ] Zero craft-floor bans (esp. 13–18 naming/`any`, 22 single-export files).
- [ ] README quick path verified by a fresh reader in ≤5 minutes (checklist below).
- [ ] Catalyst pilot note in feature doc Log (scratch replace only, no Catalyst commit).

## Library bar (every future slice)

- [ ] New package cites its REFERENCES ID + stdlib-first justification (what stdlib lacked).
- [ ] Every export has godoc with an example; every rule has a named constraint, never bare `any`.
- [ ] AI-check: a code model can name the right function from the package doc alone (spot-check, recorded).
- [ ] Human-check: naming passes the code-style audit (`audit-code` zero blockers).
