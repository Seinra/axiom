# Axiom audit closeouts — implement #22 + #23 + #24

Close the verified-missing items from the 10-issue audit: #22 test-hardening batch (test-only, no behavior change), #23 bounds + error identity (behavior changes with sentinel errors), #24 singleflight close-doc one-liner. #20 already closed; #26/#27/#4 need owner decisions; #13 partial (kept); #6 unmappable; #3 blocked by GitHub plan.

## Specs

- S1: "Implementa ese bloque. Inyecta bien las skills al delegar, arma bien el odd"
- S2 (evidence snapshot 2026-10-21, parent-verified via rg + scout audit): #22 MISSING — di Provide-guard trio + cross-type castCached mismatch; repogorm 5 gaps (Store/Fetch-non-NotFound/Remove-err/List-err DB branches, tied-order paging, concurrent use); nil-func args in either/opt/result/slice/mapx/iterx; set empty-snapshot non-nil + concurrent hammer (Len/Has-on-zero verified); flow nil-transform (workerCount clamp verified); valid nil-legs And/Or + New copy-isolation + failure order; httpx scalar-zero e2e + [nil,nil] guard (nil-entry skips verified). #23 MISSING — repogorm page.Size upper cap; Flatten length-total overflow guard; errors.Is matchability in ValidationFail per-field shape; %w typed panic payloads in Try. #23 DONE — Windows/Chunked cost docs, ChunkedSeq lazy backing, Offset overflow-safe, di sentinels, repogorm sentinel wraps. #24 MISSING — one-line comment naming close(doneCh)-under-guard as deliberate design call (singleflight.go ~line 107).
- S3 (design calls, owner-authorized via S1): page.Size cap enforced with `ErrInvalidPage` sentinel (consistent with existing Size<1 check); Flatten overflow guard panics like size<=0 (pre-v0 signed panic policy, L7 rename doc); ValidationFail keeps HTTP shape, per-field errors stay matchable via wrapped sentinels where the stdlib contract allows; Try carries typed payload with %w.

## Tasks (in order)

- T1 (S2 #22): test-hardening batch, test-only, no behavior change — route: 2 bounded workers, disjoint test surfaces (W-A: di/repogorm/flow/valid/httpx; W-B: either/opt/result/slice/mapx/iterx/set), skills foundry-testing + foundry; parent verifies + commits.
- T2 (S2 #23, S3): bounds + error identity with TDD RED→GREEN per fix — route: bounded worker (surfaces: repogorm, slice/result Flatten, httpx ValidationFail, result Try), skills foundry + foundry-testing; parent verifies + commits.
- T3 (S2 #24): one-line close-doc comment — route: parent inline.
- T4: independent verifier (foundry-review trio) + close-out + delivery — route: parent commits; push/PR owner-gated; merge needs approved labels on #22/#23/#24.

## Log

- L1 (request verbatim): "Implementa ese bloque. Inyecta bien las skills al delegar, arma bien el odd"
- L2 (2026-10-21): audit of 10 open issues completed (#20 closed with evidence; #22/#23/#24 itemized missing; #26/#27/#4 decisions pending; #13 partial kept; #6 unmappable; #3 blocked); ODD prepared; no implementation yet.
