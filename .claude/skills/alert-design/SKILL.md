---
name: alert-design
description: Derive SLIs, SLOs, symptom-based alert rules with severity and owner, runbook stubs and logging fields for a spec or RFC in the active target project, following the kit's alerting standard. Use while writing or reviewing the Observability and alerting section.
argument-hint: <spec or RFC id>
---

# alert-design

Alerting is designed with the feature. This skill turns a spec's behaviour into a reviewable alerting plan. Thresholds are defaults the user tunes; say so.

## Steps

1. Resolve the target (`python3 -I tools/docgate.py target`). Read `docs/standards/alerting.md`, `logging.md`, and (for async work) `reliable-messaging.md`. Read the spec or RFC in the target: goal, acceptance criteria, dependencies and failure modes.
2. **List the user-visible behaviours** the work provides (each user journey or message flow).
3. **Define SLIs** per behaviour from the standard's types: availability, latency, correctness, freshness, async throughput. State the good event and how it is measured.
4. **Propose SLOs** with the reasoning and the window; mark them as defaults to confirm with the user. Compute the error budget.
5. **Design alerts**:
   - Page on error-budget burn with the multi-window defaults from the standard (confirm them).
   - Ticket on early-warning and cause signals (saturation, dependency health, consumer lag, dead-letter growth, backlog age, certificate or credential expiry, stalled replication positions).
   - Add a dead-man's-switch for pipelines whose silence would hide failure.
   - Every alert: name (symptom), condition, severity, owner, dashboard, runbook path, test method.
6. **Write the table** in the spec's "Observability and alerting > Alerts" using runbook links of the form `../runbooks/<name>.md`.
7. **Create runbook stubs** in the target's `docs/runbooks/<name>.md` from the kit's `docs/templates/runbook.md`: what the alert means, first checks, safe mitigations you can justify from the design. Mark parts needing real operational knowledge as open items for the user.
8. **Logging and tracing fields**: events logged, correlation id propagation, spans, what must never be logged.
9. Add the drills that will exercise each alert to the spec's Test plan.
10. Run `python3 -I tools/docgate.py check`.

## Output

Reply with the SLI/SLO table, the alert table, the runbook files created, and the assumptions the user should confirm. If the work has no runtime behaviour, propose `Alerting: not applicable — <reason>` and explain why.

## Rules

- No alert without a runbook and an owner. No alert that a human cannot act on.
- Do not invent traffic or capacity numbers; use stated assumptions and say how they will be measured.
- Do not set the document to Accepted.
