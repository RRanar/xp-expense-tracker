---
id: 0011
title: Debezium hosting: Debezium Server beside the broker
status: Proposed
date: 2026-11-18
supersedes: none
---

# ADR 0011: Debezium hosting: Debezium Server beside the broker

## Context

ADR 0007 moves the outbox with Debezium. Debezium can run as a plugin in a Kafka Connect cluster, as standalone Debezium Server, or as a managed connector if the Kafka provider offers one. Cloud SQL for PostgreSQL supports logical decoding through a database flag and a replication-enabled user; the exact flag and limits must be checked in the current Cloud SQL documentation. **That check has not been done yet.** Also unconfirmed: whether Debezium Server can pass a binary Outbox payload through unchanged, which ADR 0009 depends on (a feature request, DBZ-7512, suggests it was not supported at one point; verify the current release). If it cannot, this ADR is superseded in favour of Kafka Connect.

## Decision

We will run Debezium Server (standalone, no Connect cluster) as a single container, with a Kafka sink, locally in docker compose and in the deployed demo on the same VM as the broker (ADR 0010). It connects to PostgreSQL with a dedicated replication role, a publication and a replication slot using `pgoutput`, with heartbeats enabled. Offsets are stored so a restart resumes from the last position. Where they are stored (a file on the VM disk, a Kafka topic, or a database) is decided in the milestone 3 spec with the trade-off written down. After a VM rebuild the connector may re-snapshot and re-emit events, which is tolerated because consumers are idempotent.

Security: the replication role has only the privileges Debezium needs; its credentials live in Secret Manager; the database connection uses TLS; Debezium authenticates to the broker as in ADR 0010.

Verification gate: this ADR is unproven until two checks pass and are recorded in the specs. (1) A local Compose drill in the milestone 3 spec shows Debezium Server passes a binary Outbox payload through unchanged (the drill of ADR 0009). (2) Before the milestone 5 spec is approved, the current Cloud SQL documentation is checked for the logical-decoding flag, limits and network path. If either fails, this ADR is superseded (Kafka Connect, or the polling relay of ADR 0007).

## Alternatives considered

- Kafka Connect with the Debezium plugin: the more common production shape and richer tooling, but a heavier component to run for one connector.
- Managed CDC connector from a Kafka provider: least to run, least to show, and depends on the provider choice.
- Polling relay in the service: kept as the fallback by ADR 0007 if hosting Debezium proves too costly.

## Consequences

Easier: one small container, simple configuration. Harder: fewer operational features than Connect (for example its management API), and the Cloud SQL logical decoding prerequisites and networking must be confirmed before milestone 5.

## Operational impact

An inactive replication slot makes the database retain WAL and can fill storage: alert on slot lag and storage growth, enable heartbeats, and document slot cleanup in a runbook. Connector restarts can duplicate events, so consumers are idempotent. Schema changes to the outbox table need a documented procedure. Alerts are the ones named in RFC 0001: `ReplicationSlotLagHigh`, `DatabaseStorageNearFull` and `EventFreshnessBreach`, each with a runbook. Failure drill: stop and restart Debezium while the API keeps writing, then verify delivery after recovery.

Teardown order: the replication slot is dropped before Debezium or its host is destroyed, otherwise the database keeps retaining WAL and can fill storage. This rule goes in the teardown runbook and the milestone 5 spec.
