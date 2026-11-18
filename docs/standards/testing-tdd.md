# Standard: testing and TDD

Platform-agnostic. Tests come from the spec's acceptance criteria and are written before the code.

## The cycle

1. **Red**: write one failing test from an acceptance criterion. Run it and watch it fail for the right reason.
2. **Green**: write the least code that passes.
3. **Refactor**: improve structure with tests green. No behaviour change.
4. Commit in small steps; the first commit of a slice may contain only the failing test. Commit messages reference the spec id.

If the code was written first, the test proves nothing. When code appears without a prior failing test, delete it and start again from the test.

## Test levels

| Level | Scope | Dependencies | Share |
|---|---|---|---|
| Unit | One aggregate, value object or use case | None; pure, in memory | Most |
| Integration | An adapter against the real dependency (database, broker) | Real, ephemeral, started by the test | Some |
| Contract | A boundary: the published API or event schema | Schema and fixtures | Per boundary |
| End to end | A critical journey through the deployed unit | Whole stack, few tests | Few |
| Failure drill | Behaviour under outage, duplicates, restarts | Whole stack | Per risk |

## Principles

- **Test behaviour, not implementation.** A refactor should not break tests.
- **Do not mock what you own.** Use in-memory fakes of your own ports; reserve mocks for true external systems at the edge. Test adapters against the real thing.
- **Deterministic**: injected clock and id generator, seeded randomness, no sleeps (wait on conditions), no shared state, independent test order.
- **One reason to fail** per test; names read as sentences describing the behaviour.
- **Arrange-act-assert**; table-driven cases for variations of one rule.
- **Test the unhappy path**: invalid input, not found, conflict, timeout, dependency down, duplicate delivery, partial failure.
- **Property-based tests** for invariants (for example money arithmetic, ordering, idempotency).
- **Coverage is a signal, not a goal.** Watch it on the domain and use cases; do not chase a number. Mutation testing, where available, says more than line coverage.
- **Flaky tests are bugs.** Quarantine within a day, fix within a week, never ignore.
- **Test data is synthetic** and built with small builders; no production data.

## Failure drills (required for messaging and storage work)

Prove, in tests, that: a message delivered twice has one effect; a consumer restart does not lose or repeat work; a dependency outage does not lose accepted writes; a poison message goes to the dead-letter path and does not block the stream. These drills double as alert tests (`alerting.md`).

## Gates in CI

- All tests pass, run with race/concurrency checks enabled where the platform offers them.
- Static analysis, formatting, dependency-rule check, secret scan, dependency vulnerability scan.
- Contract lint and breaking-change check.
- Test-first evidence: the spec's test plan maps every acceptance criterion to at least one test.

## Review checklist

- [ ] Does each acceptance criterion have a test that failed first?
- [ ] Are time and ids injected?
- [ ] Are adapters tested against real dependencies?
- [ ] Are failure paths and duplicates covered?
- [ ] Are there sleeps or order dependencies?
