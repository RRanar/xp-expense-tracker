---
id: 0005
title: PostgreSQL as the primary store
status: Accepted
date: 2026-11-18
approved_by: owner
approved_date: 2026-11-18
supersedes: none
---

# ADR 0005: PostgreSQL as the primary store

## Context

The existing project uses an embedded file database that needs a C toolchain and cannot run as a shared service. The ledger needs transactions, constraints, an outbox table written in the same transaction as state, and a path to change data capture.

## Decision

We will use PostgreSQL as the system of record. Money is stored as integer minor units plus a currency code (`bigint` amount, `char(3)` currency). Schema changes go through versioned migrations run as a separate step, not at application start; the migration tool (goose or golang-migrate) is confirmed in the milestone 2 spec, and migrations follow expand-then-contract so each is reversible. Access uses `pgx` (pure Go) with `sqlc`-generated queries; no ORM that hides transactions. Local development and integration tests run PostgreSQL in Compose and testcontainers with logical decoding enabled, so the outbox flow can be tested locally.

Schema rules: the entity UUID is the primary key. Every owned table has an indexed owner column and every query filters on it. The `outbox` table lives in the same database and is written in the same transaction as the state change. The notifier's `processed_events` table belongs to the notifier's own schema with its own credentials, so the two services share no tables.

Access rules: separate database roles for the API, the notifier, migrations and Debezium's replication role (ADR 0011), each with least privilege. Credentials are in Secret Manager. Connections use TLS, queries are parameterised only, and storage is encrypted at rest by the platform.

## Alternatives considered

- Keep the embedded file database: simplest locally, but single-writer, no shared service, and no logical replication for CDC.
- A document store: flexible schema, weaker relational constraints and a less direct fit for the transactional outbox.
- A managed distributed SQL database: more capability than needed, higher cost and complexity.

## Consequences

Easier: transactions, constraints, mature tooling, logical decoding for the outbox mover. Harder: a managed instance has a continuous cost, and logical replication adds operational duties (replication slots must be monitored so a stalled consumer cannot fill storage).

## Operational impact

Backups and restore tests, connection-pool limits and slow-query visibility are required. Alerts to define in specs: `ReplicationSlotLagHigh` and `DatabaseStorageNearFull` (RFC 0001), and as cause signals connection saturation and backup freshness (`pet-project/standards/alerting`). A restore test runs before the first deployed demo. Hosting option and instance size are decided in the deployment spec; cost to verify before creating resources. ADR 0011 assumes a managed PostgreSQL that supports logical decoding; the deployment spec verifies that against the current provider documentation, and Cloud SQL specifics for encryption, TLS and roles are checked there too.
