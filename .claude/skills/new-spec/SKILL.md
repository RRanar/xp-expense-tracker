---
name: new-spec
description: Write a spec, the unit of implementation work, in the active target project, with testable acceptance criteria, domain model, security, performance budget, alerting and test plan. Status starts Proposed. Use before any code for a feature or change.
argument-hint: <title>
---

# new-spec

No code without an Accepted spec. The spec is also the contract the gate hook enforces: its `adrs`, `rfcs` and `scope` frontmatter decide what may be edited in the target project.

## Steps

1. Resolve the target: `python3 -I tools/docgate.py target`. Read `docs/process/workflow.md` (Definition of Ready), the target's `docs/adr/` and the relevant standards.
2. Scaffold: `python3 -I tools/docgate.py new spec "<title>"` (prints the absolute path in the target).
3. Frontmatter:
   - `adrs`: comma list of ADR ids this work depends on. Each must be Accepted before implementation. If a needed decision has no ADR, create it with `/new-adr` first.
   - `rfcs`: ids, or `none`.
   - `scope`: comma list of path prefixes, relative to the target project root, that the work may change (for example `src/orders/, api/orders/`). Keep it tight; the gate blocks edits outside it.
4. Sections:
   - **Goal**, **Scope** (in and explicitly out).
   - **Acceptance criteria**: numbered, testable, given/when/then; each becomes a failing test first.
   - **Domain model**: aggregates, invariants, events, glossary terms (standard: architecture-ddd).
   - **Security**: threat model summary, authz rules, validation, data classes (standard: security).
   - **Performance budget**: workload model, percentile targets, named benchmark (standard: benchmarking).
   - **Observability and alerting**: run the `/alert-design` method for SLIs/SLOs, the alert table with runbook links, and logging fields.
   - **Test plan**: tests by level mapped to acceptance criteria, plus failure drills for messaging or storage.
   - **Links**: ADRs, RFCs, runbooks, benchmark reports.
5. `python3 -I tools/docgate.py check`. Proposed documents may keep TODOs; list which are open.
6. Leave status `Proposed`; suggest `/review-doc <path>`.

## Rules

- Do not accept the spec, fill approval fields, or edit the target's `docs/work/ACTIVE` yourself. The user approves; then ACTIVE is set (the user may ask you to do it after approval).
- Keep one spec small enough to finish in a few days. Split otherwise.
- No placeholders in an Accepted spec: the gate rejects TODO, TBD and `{{...}}`.
