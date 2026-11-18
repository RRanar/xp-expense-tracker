---
name: new-rfc
description: Draft an RFC in the active target project for a cross-cutting, risky or hard-to-reverse change, with design, security, performance, and mandatory observability and alerting sections. Status starts Proposed. Use before writing ADRs and specs for large changes.
argument-hint: <title>
---

# new-rfc

An RFC explains a significant change and compares options before anything is decided. An accepted RFC produces ADRs (the decisions) and specs (the work). It is written into the **target project's** `docs/rfc/`.

## Use an RFC when

The change spans several components or contexts, is expensive to reverse, changes a contract other parties rely on, or has more than one credible design. Otherwise write an ADR or a spec.

## Steps

1. Resolve the target: `python3 -I tools/docgate.py target`. If none, use `/init-project`.
2. Read `docs/process/workflow.md` and the standards the change touches.
3. Scaffold: `python3 -I tools/docgate.py new rfc "<title>"` (prints the absolute path).
4. Fill the sections:
   - **Summary**, **Motivation** (problem, who is affected, cost of doing nothing).
   - **Detailed design**: enough to implement; name bounded contexts, contracts, data flows; a text diagram.
   - **Security**: STRIDE-style threat summary and mitigations (standard: security).
   - **Performance**: workload model, budget, planned benchmark (standard: benchmarking).
   - **Observability and alerting**: SLIs/SLOs, an alert table with severity, owner and runbook links, logging and tracing. Follow the `/alert-design` method and standard: alerting. If truly no runtime behaviour: `Alerting: not applicable — <reason>`.
   - **Alternatives considered**, **Rollout** (migration, flags, rollback, verification), **Open questions** (each with an owner).
5. List the ADRs the RFC implies and create them with `/new-adr` (Proposed).
6. `python3 -I tools/docgate.py check`, then suggest `/review-doc <path>`.

## Rules

- Status stays `Proposed`. The user alone accepts it.
- No invented numbers: every figure is an assumption or says how it will be measured.
- Alternatives must be real. An RFC with one option is a spec.
