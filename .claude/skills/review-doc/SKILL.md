---
name: review-doc
description: Review an ADR, RFC or spec in a target project against the kit's standards with a checklist, report findings, and move it to Accepted only when the user explicitly approves. Use before any document is relied on for implementation.
argument-hint: <path to ADR, RFC or spec>
---

# review-doc

The review gate (G1) and the approval gate (G2). The reviewer finds problems; **the user decides**.

## Steps

1. Read the document and the kit standards it should follow:
   - Always: `docs/process/workflow.md`.
   - ADR: Context, Decision, Alternatives (at least two real), Consequences, Operational impact.
   - RFC and spec: also security, benchmarking, alerting, architecture-ddd and testing-tdd standards; for messaging, reliable-messaging.
2. Resolve the document's project and run `python3 -I tools/docgate.py check` (use `--project <name>` if the document is not in the active target). Note structural failures.
3. Review against this checklist and report each as pass, finding, or not applicable:
   - **Clarity**: goal and scope unambiguous; terms from the project's glossary.
   - **Testability**: every acceptance criterion testable; test plan maps to each.
   - **Architecture**: dependency rule respected; invariants in aggregates; contracts versioned.
   - **Security**: threat model; authz per object; input validation; secrets; data classification.
   - **Performance**: workload model and percentile budget with a named benchmark; no unmeasured claims.
   - **Observability and alerting**: SLIs/SLOs defined; alerts symptom-based with severity, owner and runbook link; drills planned; logging follows the standard. **An empty or generic alerting section is a blocking finding.**
   - **Consistency**: does not contradict an Accepted ADR in the project; supersession declared; spec `scope` as narrow as possible.
   - **Honesty**: no invented figures, no production claims for a pet project.
4. Report findings ranked: blocking, should-fix, nit. Offer concrete edits; apply the ones the user agrees to, then re-run `check`.
5. Ask the user plainly whether to accept the document. **Only if the user says yes in this conversation**, set `status: Accepted`, `approved_by: <the user's label, or "owner">`, `approved_date: <today>`, and run `check` again. If the user says no or wants changes, leave it `Proposed` (or set `Rejected` if told).

## Rules

- Never accept a document on your own authority, and never because checks pass.
- Do not change substantive content of an already Accepted document; propose a superseding ADR/RFC.
- After accepting a spec, remind the user the next step is setting `docs/work/ACTIVE` in the project and running `/implement <id>`.
