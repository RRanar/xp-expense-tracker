---
id: {{id}}
title: {{title}}
status: Proposed
date: {{date}}
adrs: none
rfcs: none
scope: TODO-path-prefix/
---

# Spec {{id}}: {{title}}

<!-- A spec is the unit of implementation work. Code may only be written for an Accepted spec whose linked
     ADRs are Accepted (enforced by tools/docgate.py). 'adrs' and 'rfcs' are comma lists of ids.
     'scope' is a comma list of path prefixes this spec may change, for example src/orders/, api/orders/. -->

## Goal

TODO: the outcome in one or two sentences, and who benefits.

## Scope

TODO: what is in, and explicitly what is out.

## Acceptance criteria

<!-- Each criterion is testable and becomes at least one failing test before any code (TDD). -->

1. TODO: given / when / then.

## Domain model

TODO: aggregates, entities, value objects, invariants, domain events, and the ubiquitous-language terms used. See pet-project standard: architecture-ddd.

## Security

TODO: threat model summary, authn/authz rules, input validation, data classification. See pet-project standard: security.

## Performance budget

TODO: latency and throughput targets with percentiles, the workload model, and the benchmark that proves it. See pet-project standard: benchmarking.

## Observability and alerting

### SLIs and SLOs

TODO: the SLIs this work affects and their SLO targets.

### Alerts

| Alert | Condition (symptom) | Severity | Owner | Runbook |
|---|---|---|---|---|
| TODO | TODO | page or ticket | TODO | [runbooks/TODO.md](../runbooks/TODO.md) |

### Logging and tracing

TODO: events logged, correlation id, spans, and which fields are never logged.

<!-- If a change truly has no runtime behaviour, replace the three subsections above with one line:
     Alerting: not applicable — <reason of at least 20 characters> -->

## Test plan

TODO: the test list by level (unit, integration, contract, failure drills) mapped to the acceptance criteria.

## Links

TODO: related ADRs, RFCs, runbooks, benchmark reports.
