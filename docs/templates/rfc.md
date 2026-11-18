---
id: {{id}}
title: {{title}}
status: Proposed
date: {{date}}
---

# RFC {{id}}: {{title}}

<!-- Use an RFC for cross-cutting, risky or hard-to-reverse changes, or when several options need discussion.
     An accepted RFC produces ADRs (decisions) and specs (work). Status stays Proposed until the user approves. -->

## Summary

TODO: one paragraph a newcomer can read in a minute.

## Motivation

TODO: the problem, who has it, and what happens if we do nothing.

## Detailed design

TODO: the design with enough detail to implement. Include a diagram if it helps. Name the bounded contexts, contracts and data flows touched.

## Security

TODO: threat model summary (STRIDE or similar): assets, trust boundaries, abuse cases, mitigations. See pet-project standard: security.

## Performance

TODO: expected load, latency and throughput budget, and how it will be benchmarked. See pet-project standard: benchmarking.

## Observability and alerting

### SLIs and SLOs

TODO: the user-visible SLIs and their SLO targets (availability, latency, correctness, freshness).

### Alerts

| Alert | Condition (symptom) | Severity | Owner | Runbook |
|---|---|---|---|---|
| TODO | TODO | page or ticket | TODO | [runbooks/TODO.md](../runbooks/TODO.md) |

### Logging and tracing

TODO: the key events logged, the correlation id, and the trace spans. See pet-project standard: logging.

<!-- If a change truly has no runtime behaviour, replace the three subsections above with one line:
     Alerting: not applicable — <reason of at least 20 characters> -->

## Alternatives considered

TODO: the options rejected and why.

## Rollout

TODO: migration, feature flags, rollback plan, and how success is verified.

## Open questions

TODO: unresolved points. An approved RFC must have none left, or each must name an owner and a date.
