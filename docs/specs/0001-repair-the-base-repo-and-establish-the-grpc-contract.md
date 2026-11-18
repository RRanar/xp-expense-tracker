---
id: 0001
title: Repair the base repo and establish the gRPC contract
status: Accepted
date: 2026-11-18
approved_by: owner
approved_date: 2026-11-18
adrs: 0003, 0004, 0005
rfcs: 0001
scope: cmd/, internal/, api/, .github/, go.mod, go.sum, buf.yaml, buf.gen.yaml, Makefile, README.md, .gitignore
---

# Spec 0001: Repair the base repo and establish the gRPC contract

## Goal

Milestone 1 of RFC 0001. Make the repository build and test green on a clean machine, fix the verified defects in the domain and application layers, add CI, and define the v1 expense contract with `buf` linting. The owner gets a trustworthy base for the gRPC and Postgres work in later specs.

## Scope

In:

- Extend the `Expense` entity to the fields the contract needs: `currency` (validated three-letter code, with `Money` = amount in minor units as `int64` plus currency), `merchant`, `expense date` (the day the expense happened, separate from the created time), and an owner `UserID` value object. Rename `description` to `note` everywhere.
- Fix the defects from the audit that live in domain and application code: `CreatedAtFromString` hiding errors, direct `time.Now()` in use cases, a UUID that is never stored.
- Rename directory `expence` to `expense`.
- Inject a `Clock` into use cases.
- Redefine the repository port with `context.Context`, `FindByID`, `List(filter, page)`, `Update`, `Delete`, all scoped to an owner.
- Add `Reconstitute(...)` on the entity for loading from storage (the UUID and created time are preserved).
- Remove the SQLite adapter, the HTTP transport (`internal/infrastructure/http`), `cmd/app` and the `go-sqlite3` dependency (it needs cgo). They were learning scaffolding; the project is now a stand-alone gRPC service. History stays in git. Persistence returns in spec 0002 as a Postgres adapter and transport as a gRPC server. Tests use an in-memory fake of the port.
- **Authorised file operations.** The rename and the deletions above are done with `git mv` and `git rm` (they preserve history and are not covered by the edit hook). Approval of this spec is the owner's authorisation for exactly these operations and no other use of the shell to change code.
- `go mod tidy`; a README with the before/after of the audit and the lesson from the missing `return` after an error in the old HTTP handler.
- CI workflow: build, vet, test with race detector, lint, dependency-rule check, `govulncheck`, secret scanning, `buf lint`, `buf breaking`. Workflow permissions are `contents: read`; third-party actions are pinned to commit SHAs.
- `api/proto/expense/v1/expense.proto`: messages and service definition for expense CRUD, list with page tokens, field-mask update, money type, and `ExportMonth` server stream. Definition only; no server implementation.

Out:

- gRPC server, Postgres adapter, migrations, docker compose (spec 0002).
- Budgets, outbox, Kafka, notifier (later specs).
- Authentication, identity, rate limiting, deployment, Terraform. The owner `UserID` exists in the domain now so the port does not change later; where it comes from is spec work for milestone 4.
- A replacement entry point: `cmd/expense-api` arrives with the gRPC server in spec 0002.

## Acceptance criteria

1. Given a clean clone with the toolchain version in `go.mod`, when `go build ./...` and `go test -race ./...` run, then both succeed, and `CGO_ENABLED=0 go build ./...` passes.
2. Given valid input, when `NewExpense` is called with amount, currency, category, merchant, note, expense date and owner, then the entity holds them; and it returns a domain error for amount not above zero, missing category, a currency that is not a valid three-letter code, or a missing owner.
3. Given an entity, when it is loaded back through `Reconstitute` with the stored id and created time, then every field including the UUID equals the original (tested against the in-memory fake).
4. Given an invalid created-at string, when `CreatedAtFromString` is called, then it returns an error and never substitutes the current time.
5. Given a fake `Clock` returning a fixed instant, when `CreateExpense` runs, then the created time equals that instant; and the linter fails the build if a file under `internal/app` or `internal/application` calls `time.Now`.
6. Given any repository method, when a use case calls it, then it passes a `context.Context`, and the in-memory fake returns the context error for a cancelled context.
7. Given stored expenses of several owners, when `List(filter, page)` runs, then it returns only the caller's expenses; the filter supports category and an inclusive expense-date range; results are ordered by expense date then id; at most the page size (default 50, maximum 200, spec defaults the owner tunes) are returned; and an opaque next-page token is returned when more remain and resumes after the last item.
8. Given an existing expense, when `Update` is applied with a field mask, then only the masked fields change, others keep their values, and invariants are re-checked; `Delete` removes it; both are visible through `FindByID`; an unknown id or an id owned by another user returns the domain not-found error.
9. Given the directory rename, when `go list ./...` runs, then no path contains `expence`, and `git log --follow` on a moved file shows earlier history.
10. Given an import in `domain` of `application` or any adapter package, when the dependency-rule check runs in CI, then it fails.
11. Given `api/proto/expense/v1/expense.proto`, when `buf lint` runs then it passes; and `buf breaking` against the base branch reports no breaking change. If there is no base to compare with yet, the CI file states that reason in a comment and skips only this step.
12. Given a push or pull request, when CI runs, then every gate in the Scope list runs with `contents: read` permission and SHA-pinned actions; a deliberate lint error was shown to fail the run once, recorded in the commit message that fixes it.
13. Given the README, when read, then it contains the before/after of audit findings and states "pet project" plainly, with no production claim.

## Domain model

- Aggregate: `Expense` (identity = UUID, owned by one user). Fields: `amount` and `currency` (as `Money`), `category`, `merchant`, `note`, `expense date`, `created at`, `owner`. Invariants: amount above zero, category required, valid currency, owner required, valid dates.
- Value objects: `ID`, `UserID`, `CreatedAt`, `Money` (minor units `int64` plus currency code; arithmetic is checked for overflow when introduced).
- Ports: `Repository` (context-aware, owner-scoped, list with filter and page), `Clock`.
- Domain errors: not found, invalid input; mapped to gRPC codes in a later spec.
- Ubiquitous language: expense, category, merchant, note, expense date, minor units, owner, page token.

## Security

- No secrets, credentials or real data. Synthetic fixtures only.
- Input validation stays in the domain constructor; `CreatedAtFromString` stops hiding bad input.
- Object-level isolation: every repository method takes the owner and a record of another owner is reported as not found, tested with two owners.
- Proto: no field carries secrets; the owner of a write comes from the authenticated context in milestone 4, not from a client-supplied field.
- Data classification: amount, currency, category, merchant, note and dates are financial (synthetic); `UserID` is internal (RFC 0001).
- CI: `govulncheck` for dependencies, secret scanning, read-only workflow permissions, SHA-pinned actions, pinned dependency versions in `go.mod` and `go.sum`.

## Performance budget

Not applicable to runtime in this spec: no service runs. Budget on CI only: the full pipeline completes in a time the owner finds tolerable; record the measured duration in the README once observed, no target invented here. The in-memory `List` is for tests only and has no performance claim.

## Observability and alerting

Alerting: not applicable — this spec changes build, test and contract files only and ships no running service; runtime SLIs, alerts and runbooks start with the service in spec 0002 and are listed in RFC 0001.

## Test plan

| Level | Tests | Criteria |
|---|---|---|
| Unit (domain) | Constructor invariants (table-driven); `Reconstitute` round trip; `CreatedAtFromString` error; `Money` and currency validation, including property-based tests for `Money` invariants | 2, 3, 4 |
| Unit (application) | Create, list, update (field mask), delete with fake repository and fake clock; cancelled context; two-owner isolation; page token resumes correctly | 5, 6, 7, 8 |
| Contract | `buf lint`, `buf breaking` | 11 |
| Build and static | `go build`, `CGO_ENABLED=0 go build`, `go test -race`, `go vet`, lint (including the `time.Now` rule), dependency-rule check, `govulncheck`, secret scan | 1, 5, 10, 12 |
| Repo and docs | `go list` shows no `expence` path; `git log --follow`; README content check | 9, 13 |

Each test is written failing first from its criterion (TDD), then the minimal code, then refactor. Existing passing tests are adapted, not deleted, unless the code they cover is removed (SQLite adapter, HTTP transport); their handler tests are deleted with the handlers.

## Links

- RFC 0001: Expense Ledger gRPC service with transactional outbox on GCP
- ADR 0003 Go, ADR 0004 gRPC with Protobuf, ADR 0005 PostgreSQL
- Gate reminder: ADRs 0003, 0004, 0005 and RFC 0001 are `Proposed`; all must be `Accepted` before `/implement 0001` is allowed. ADRs 0003 to 0005 have not yet been reviewed.
- Open questions, resolved by the owner's approval on 2026-11-18 (accepted as written):
  1. CI provider is GitHub Actions per the source note; confirm, since ADR 0008 covers deployment, not CI.
  2. Linter (`golangci-lint` default configuration), dependency-rule tool and secret scanner are chosen in the first implementation commits; confirm that is acceptable, or name them now.
