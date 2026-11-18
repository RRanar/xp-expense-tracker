# Workflow: documents first

This process is enforced where it can be (`tools/docgate.py`) and written down where it cannot.

**Where documents live:** the kit (this repo) holds the process, standards and templates. Every document produced by the process (ADRs, RFCs, specs, runbooks, alert notes, benchmark reports, `docs/work/ACTIVE`) lives in the **target project's** `docs/`, created with `/init-project`. In the sections below, `docs/...` paths mean the target project unless they start with `standards/`, `templates/` or `process/`.

## Which document, when

| Situation | Document |
|---|---|
| A single decision with trade-offs (a database, a protocol, a pattern) | ADR |
| A cross-cutting, risky or hard-to-reverse change, or several options to discuss | RFC, which then yields ADRs and specs |
| A unit of work with testable outcomes | Spec |
| A fix so small it changes no behaviour (typo, comment) | None; docs-only and formatting changes never need a spec |

If you are unsure whether something needs an ADR, it does: decisions that are not written down get re-argued.

## Lifecycle and gates

```
idea -> [RFC] -> ADR(s) -> spec -> review -> APPROVAL -> ACTIVE -> implement (TDD)
     -> alerts + runbooks + benchmarks -> spec Implemented
```

| Gate | Condition to pass | Who |
|---|---|---|
| G0 Ready | Problem is stated; RFC/ADR need is assessed | Author |
| G1 Reviewed | `/review-doc` checklist has no open findings | Author + reviewer skill |
| G2 Approved | Status `Accepted` with `approved_by` and `approved_date` | **The user only** |
| G3 Active | `docs/work/ACTIVE` names the spec; its ADRs/RFCs are Accepted | Gate hook checks |
| G4 Done | Definition of Done below | Author, verified by the user |

## Status values

`Proposed` (draft, under review) -> `Accepted` (approved) -> `Implemented` (specs only, after Done). Also `Rejected`, `Superseded` (name the successor), `Deprecated`. An accepted document is not edited in place for substantive changes: write a new ADR/RFC that supersedes it. Typo fixes are fine.

## Definition of Ready (before a spec can be Accepted)

- Goal and scope are explicit, including what is out of scope.
- Every acceptance criterion is testable.
- The domain model names aggregates, invariants and events in the project's ubiquitous language.
- A security section with a short threat model (see `standards/security.md`).
- A performance budget with percentiles and a named benchmark.
- Observability and alerting: SLIs/SLOs, alerts with severity and owner, a runbook link per alert.
- All decisions it depends on are Accepted ADRs.

## Definition of Done (before a spec becomes Implemented)

1. All acceptance criteria are covered by passing automated tests written first.
2. Lint, static analysis, dependency and secret scans pass in CI.
3. Dependency-rule checks for the architecture pass (`standards/architecture-ddd.md`).
4. The benchmark from the spec ran; the result is recorded in `docs/benchmarks/` and meets the budget, or the miss is documented and accepted by the user.
5. **Alert rules exist as code and each links a runbook; each alert was fired at least once in a drill or test** (`standards/alerting.md`).
6. Logs follow `standards/logging.md`; no secrets or personal data in logs.
7. Runbooks are written for every new alert and reviewed.
8. Docs match the code: ADRs, the spec's Links section, and any diagrams.
9. The user has reviewed the result and set the spec to `Implemented`.

## Change of plan during implementation

If the work needs something the spec does not cover (new path, new dependency, a different approach), **stop and amend the documents first**: edit the spec (and re-approve) or add an ADR. The gate will block out-of-scope edits on purpose.

## Alerting in the workflow (summary)

Alerting is designed at spec time (`/alert-design`), reviewed at G1, implemented and drilled before G4, and revisited after every incident or noisy alert. Details and defaults are in `standards/alerting.md`.
