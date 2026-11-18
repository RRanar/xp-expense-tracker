---
name: implement
description: Implement an Accepted spec in the active target project, test-first and within its scope, then deliver its alert rules, runbooks and benchmark. Refuses unless the spec and its linked ADRs/RFCs are Accepted. Use only after the user has approved the spec.
argument-hint: <spec id, for example 0003>
---

# implement

Implementation is the last step of a documented decision, not the first. Code is written in the **target project**; Claude must have access to it (started with `claude --add-dir <path>`).

## Preconditions (stop and report if any fails)

1. `python3 -I tools/docgate.py target` names the project, and `status` shows the spec **Accepted** and every ADR/RFC in its `adrs` / `rfcs` **Accepted**.
2. `python3 -I tools/docgate.py check` passes for the spec.
3. The target's `docs/work/ACTIVE` contains this spec id. If not, ask the user to approve setting it. The gate hook blocks code edits otherwise; do not route around it.

If anything is missing, do not write code. Send the user to `/new-spec`, `/new-adr` or `/review-doc`.

## Steps

1. Read the spec, the linked ADRs and the kit standards (architecture-ddd, testing-tdd, security, logging, reliable-messaging as relevant).
2. Plan the work as a short list of slices, each tied to acceptance criteria. Show it to the user before starting.
3. For each slice, TDD:
   1. Write the failing test from an acceptance criterion; run it; confirm it fails for the right reason.
   2. Write the minimum code to pass; run the tests.
   3. Refactor with tests green; keep the dependency rule.
   4. Commit in the target repo (`git -C <path>`): `spec-<id>: <what>`. Only commit when the user has asked for commits in this session; otherwise leave changes staged or uncommitted and say so. Never push.
4. Keep edits inside the spec's `scope`. If more is needed, **stop and amend the spec** (and ask for re-approval) or add an ADR.
5. Add the failure drills the standards require (duplicates, restarts, dependency down) as tests.
6. Alerting (part of the work): create alert rules as code, dashboards, and a runbook per alert in the target's `docs/runbooks/` from the kit template; link them from the spec's alert table. Fire each alert once in a drill or test and record it in the runbook.
7. Benchmark: run the benchmark named in the spec; record the result in the target's `docs/benchmarks/` per the benchmarking standard. Report real numbers, including misses.
8. Run all checks (tests, lint, static analysis, `docgate check`). Update the spec's Links and any diagrams.
9. Report against the Definition of Done in `docs/process/workflow.md`: what passed, what did not, what remains. Ask the user to review. **Only the user sets the spec to `Implemented`.**

## Rules

- No code without a failing test first. If you wrote code first, delete it and start from the test.
- No secrets, no real data. Synthetic data only.
- Never edit code through Bash to dodge the gate hook.
- Do not claim performance, reliability or production readiness beyond what was measured.
