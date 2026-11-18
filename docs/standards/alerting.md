# Standard: alerting

Platform-agnostic. Alerting is designed with the feature, not bolted on after launch. This standard defines what a good alert is, how SLOs drive alerts, and how alerting is gated in the workflow.

**Numbers in this document are starting defaults, not rules. The owner tunes them per service and records the choice in the spec or an ADR.**

## Principles

1. **Alert on symptoms, not causes.** Page on what users feel (errors, slowness, wrong or stale data), not on internal causes (CPU, a pod restart). Cause metrics belong on dashboards and in diagnosis.
2. **Every alert is actionable.** If a human cannot do something useful in response, it is not an alert; make it a dashboard panel or a log.
3. **Every alert has an owner, a severity and a runbook.** No runbook, no alert.
4. **Fewer, better alerts.** Each page is paid for in attention. Track and prune noise.
5. **Alerts are code.** Rules live in version control, are reviewed with the change, and are tested.
6. **Absence is a signal.** Alert when expected data stops arriving (dead-man's switch), not only when bad data arrives.

## SLIs and SLOs

Define from the user's point of view, per service or user journey.

| SLI type | Good event | Typical measurement |
|---|---|---|
| Availability | Request served without a server-side error | good requests / valid requests |
| Latency | Request faster than a threshold | share under threshold, per percentile |
| Correctness | Response is right (validated by checks or reconciliation) | good results / checked results |
| Freshness | Data is newer than a threshold | share of data younger than the threshold |
| Async throughput | Work completed within a deadline | share of messages processed within the deadline; consumer lag or oldest-message age |

An SLO is a target for an SLI over a window (for example 99.5% over 30 days). The **error budget** is `1 - SLO`. Pick an SLO you can defend, not the highest one you can imagine: it should be below what the system really achieves and above what makes users unhappy.

## Alert on error-budget burn

A burn rate of 1 consumes the budget exactly over the window. Use multi-window, multi-burn-rate alerts so fast outages page quickly and slow leaks are still caught without flapping. Common starting points for a 30-day window (confirm against current SRE guidance and tune):

| Severity | Burn rate | Long window | Short window (confirm) | Budget consumed if sustained |
|---|---|---|---|---|
| Page | 14.4 | 1 hour | 5 minutes | about 2% |
| Page | 6 | 6 hours | 30 minutes | about 5% |
| Ticket | 3 | 1 day | 2 hours | about 10% |
| Ticket | 1 | 3 days | 6 hours | about 10% |

An alert fires only when both the long and the short window exceed the burn rate; the short window makes it reset quickly after recovery.

For low-traffic services, ratio alerts are noisy: use longer windows, a minimum request count, or synthetic probes.

## Severity

| Severity | Meaning | Response |
|---|---|---|
| Page | Users are affected now, or the budget will be gone soon; human action needed now | Immediate, any hour |
| Ticket | Needs attention, no immediate user impact | Within working days |
| Info | Context only | Dashboard or log, no notification |

For a one-person pet project, "page" can mean a push notification you actually mute at night; the discipline of classification still applies.

## Cause and saturation signals (ticket-level or dashboard)

Use these to diagnose and to warn early, not to page:

- Saturation: CPU, memory, disk, connection pools, queue depth, thread/worker utilisation.
- Dependencies: error rate and latency of each downstream call, circuit breaker state.
- Messaging: consumer lag, oldest unprocessed message age, retry and dead-letter rates (a growing dead-letter queue is a ticket; unbounded lag is a page).
- Data stores: replication lag, slow queries, storage growth, backup freshness.
- Replication/streaming slots and offsets that can silently fill storage if a consumer stops.
- Certificates and credentials approaching expiry.

## Alert rule template (one per alert)

| Field | Content |
|---|---|
| Name | Verb-free, symptom-based, for example `CreateExpenseErrorBudgetBurn` |
| SLI / SLO | Which objective it protects |
| Condition | The query with windows and thresholds |
| Severity | page / ticket |
| Owner | Who is responsible |
| Runbook | Link to `docs/runbooks/<name>.md` |
| Dashboard | Where to look first |
| Test | How it was fired (drill, synthetic failure, replayed data) |

The spec embeds a short table; the full rule is kept as code and named in `docs/alerts/`.

## Runbooks

Each alert links a runbook (template: `docs/templates/runbook.md`): what it means, first checks, diagnosis table, safe mitigations first (roll back, disable a flag, scale), escalation, and a post-incident update step. A runbook is reviewed when written and after every incident in which it was used.

## Testing alerts

- Fire each alert at least once before the feature is Done: induce the failure in a test environment, or replay recorded data against the rule.
- Verify the notification actually arrives at the right destination.
- Verify it resolves.
- Record the drill in the runbook ("last exercised").

## Alert hygiene

- Review alert volume after every release and monthly: for each page, was it actionable, was it real, was the runbook enough?
- Delete or demote alerts that were not actionable.
- Track the share of pages that needed no action; a high share means the alerting is wrong, not that the system is healthy.
- After an incident, ask: did we alert first, or did a user? If a user, add or fix an alert.

## Where alerting sits in the workflow

| Stage | Alerting work |
|---|---|
| Spec / RFC | "Observability and alerting" section: SLIs, SLOs, alert table with severity, owner, runbook link. Run `/alert-design` |
| Review (G1) | Reviewer checks the section against this standard; an empty or generic section blocks acceptance |
| Implement | Alert rules and dashboards are written with the feature, as code |
| Done (G4) | Each alert fired once, runbooks exist, results recorded |
| Operate | Post-incident review updates alerts and runbooks |

The gate (`tools/docgate.py`) refuses to accept a spec or RFC whose alerting section has no SLI/SLO definition or no alert with a runbook link, unless it states `Alerting: not applicable — <reason>`.
