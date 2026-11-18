---
id: 0007
title: Transactional outbox moved by Debezium CDC
status: Proposed
date: 2026-11-18
supersedes: none
---

# ADR 0007: Transactional outbox moved by Debezium CDC

## Context

The service must change state and publish an event without a dual write. A failure between the database commit and the broker publish would lose or invent events. The owner wants to demonstrate the outbox pattern with log-based change data capture.

## Decision

We will write an outbox row in the same database transaction as the state change and move it to Kafka with the Debezium PostgreSQL connector using the Outbox Event Router transformation. The service never talks to the broker for domain events. Outbox columns follow the router's conventions (event id, aggregate type, aggregate id, type, payload). The payload format (JSON or Protobuf bytes) and where Debezium runs (a Kafka Connect cluster, Debezium Server, or a managed connector) are settled in the spec after checking the current Debezium and platform documentation.

## Alternatives considered

- Polling relay written in the service: simpler infrastructure and good for learning, but adds code to own and puts polling load on the database; kept as the fallback if Debezium hosting proves too costly.
- Dual write (database then broker): rejected, unsafe.
- Event sourcing as the primary store: a larger design change than the project needs.

## Consequences

Easier: no relay code, low-latency delivery, a faithful ordered change stream. Harder: one more moving part to run and monitor, logical replication prerequisites on the database, duplicates after connector restarts (consumers must be idempotent), schema changes to the outbox table need care.

## Operational impact

Database must allow logical decoding with a dedicated replication role, publication and slot. A stalled connector makes the replication slot retain WAL and can fill storage: alert on slot lag and storage growth, and enable heartbeats. Also alert on outbox backlog age and end-to-end event freshness (`pet-project/standards/alerting`, `pet-project/standards/reliable-messaging`). Runbooks for connector failure and slot cleanup are required. Failure drills: broker down, connector restart, duplicate delivery.
