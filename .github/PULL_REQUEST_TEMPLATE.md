# Pull request

## Linked issue (REQUIRED)

Closes #<!--
  Valid keywords: Closes, Fixes, Resolves (case insensitive).
  The linked issue MUST carry the `status:approved` label.
-->

## PR type (check exactly ONE, then add the matching label)

- [ ] Bug fix → `type:bug`
- [ ] New feature → `type:feature`
- [ ] Documentation only → `type:docs`
- [ ] Code refactoring → `type:refactor`
- [ ] Maintenance/tooling → `type:chore`
- [ ] Tests only → `type:test`
- [ ] Breaking change → `type:breaking-change`

## Summary

- <!-- 1-3 bullets on what this PR does and why -->

## Changes

| File | Change |
|------|--------|
| <!-- `path/to/file` --> | <!-- what changed --> |

## Test plan

- [ ] `go test -count=1 ./...` green
- [ ] `go vet ./...` clean
- [ ] `gofmt -l .` empty
- [ ] `golangci-lint run ./...` 0 issues
- [ ] New behavior covered by tests in the same unit (or N/A with reason)

## Contributor checklist

- [ ] Linked an approved issue above
- [ ] Added exactly one `type:*` label
- [ ] Branch is `type/description` (lowercase)
- [ ] Commits follow Conventional Commits
- [ ] No `Co-Authored-By` trailers
- [ ] Docs/godoc updated if behavior changed
