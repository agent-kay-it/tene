## Summary

<!-- 1-3 bullets: what changes in this PR and why -->

## Type

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Docs / README / landing only
- [ ] Refactor / cleanup
- [ ] CI / tooling

## Test plan

<!-- How did you verify this works? Paste commands + output -->

```bash
go test -race ./...
golangci-lint run
# e2e example:
./tene init && ./tene set K v && ./tene list
```

## Breaking changes

<!-- If yes, explain migration path. Otherwise: "None." -->

None.

## Related issue

Closes #

## Checklist

- [ ] Tests pass (`go test -race ./...`)
- [ ] Lint pass (`golangci-lint run`)
- [ ] Docs updated (README / llms.txt / llms-full.txt / cli-reference.md if user-facing)
- [ ] **No secret values pasted anywhere in this PR**
- [ ] No new network calls introduced (tene is local-first)
