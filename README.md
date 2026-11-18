# xp-expense-tracker

A **pet project**: an expense ledger written in Go, built docs-first (RFC, ADRs and specs under `docs/`) and test-first. It is not a production system and makes no production claims. All data in tests and fixtures is synthetic.

## Status

Spec 0001 (repair the base, establish the contract) is in progress. There is no running service yet: the gRPC server and the Postgres adapter arrive in spec 0002. What exists today:

- `internal/domain`: the `Expense` aggregate, value objects (`Money`, `Currency`, `Date`, `UserID`, `CreatedAt`, `ID`), the `Clock` and `Repository` ports.
- `internal/application/usecase`: create, get, list (filter and page tokens), update (field mask), delete.
- `internal/infrastructure/memory`: an in-memory repository used as a test fake.
- `api/proto/expense/v1/expense.proto`: the v1 contract, definition only.

## Audit: before and after

The first audit of the original code found these, each verified against the repository history:

| Before | After |
|---|---|
| `go build ./...` failed: the SQLite adapter did not compile (`e.CreatedAt.Get` undefined, wrong argument type to `NewExpense`) and needs cgo | SQLite adapter, HTTP transport and `cmd/app` removed (history stays in git); `go build ./...` and `CGO_ENABLED=0 go build ./...` pass |
| `CreatedAtFromString` replaced a bad value with the current time | Returns an error and never substitutes a value |
| Use cases called `time.Now()` directly | A `Clock` is injected; the linter rejects `time.Now` under `internal/application` |
| The entity generated a UUID that no repository stored or reloaded | `Reconstitute` rebuilds an expense from storage keeping its id and created time |
| The SQLite reader ignored a timestamp parse error (`createdTime, _ := ...`) | No adapter ignores parse errors; the in-memory fake rebuilds entities through `Reconstitute`, which re-checks invariants |
| Amount only; no currency, merchant, expense date or owner; no `context.Context` on the repository | `Money` in minor units plus currency, merchant, expense date, owner; every repository method takes a context and is scoped to an owner |
| Directory named `expence` | Renamed to `expense` with `git mv` (history preserved) |

### Lesson: the missing `return`

The old HTTP handler decoded the request body and, on a JSON error, wrote a 400 response but did not `return`:

```go
if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
    http.Error(w, "invalid json", http.StatusBadRequest)
}
output, err := h.useCase.Execute(input)
```

Execution continued with an empty input, ran the use case, and wrote a second response on top of the 400. Writing an error response does not stop a handler. The fix is structural: validation lives in the domain constructor, errors are values returned up the stack, and a transport layer (spec 0002) maps them at the edge in one place, so no handler hand-rolls its own error flow.

## Development

Tooling comes from a Nix devShell defined outside this repository (`xp-expense-tracker` in the owner's `nixos-config`), activated through direnv. Go itself is installed globally. Inside the shell:

| Command | What it runs |
|---|---|
| `make build` | `go build` plus `CGO_ENABLED=0 go build` |
| `make test` | `go test -race ./...`, which includes the dependency-rule check in `internal/archtest` |
| `make lint` | `golangci-lint` (includes the `time.Now` ban in application code) |
| `make vuln` | `govulncheck` |
| `make secrets` | `gitleaks` over the git history |
| `make buf-lint` / `make buf-breaking` | contract lint / breaking-change check against `BASE_BRANCH` |
| `make check` | everything except `buf-breaking` |

CI (`.github/workflows/ci.yml`) runs the same targets with `contents: read` permission and SHA-pinned actions, installing the tools at the same pinned versions. The full pipeline duration has not been observed yet, because CI has not run; it will be recorded here once it has.
