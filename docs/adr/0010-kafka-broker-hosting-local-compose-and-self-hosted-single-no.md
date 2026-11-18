---
id: 0010
title: Kafka broker hosting: local compose and self-hosted single node
status: Proposed
date: 2026-11-18
supersedes: none
---

# ADR 0010: Kafka broker hosting: local compose and self-hosted single node

## Context

Kafka is the event backbone (ADR 0006). The project needs a broker for local development and for the deployed GCP demo, at a cost a pet project can accept. Provider options and prices change: managed Google Cloud Kafka is native to GCP but billed continuously; Confluent Cloud or Redpanda Cloud serverless tiers may be cheaper; self-hosting is cheapest but is operated by the owner. **Current pricing and free tiers have not been verified for this ADR.**

## Decision

We will use docker compose (Kafka in KRaft mode) for local development and for integration tests with testcontainers, and a single-node self-hosted Kafka (KRaft) on a small GCE VM for the deployed demo, created and destroyed with Terraform. The README states plainly that this is a single node with no replication, suitable for a demo only. The single node is expected to be the cheapest option, which is not yet confirmed. Before the milestone 5 spec is approved, the owner compares the VM against the managed options on current prices and records the date and source in that spec; if a managed option fits the budget and the owner prefers it, this ADR is superseded.

Security and networking: the broker has no public port. The `expense-api`, the `notifier` and Debezium (which shares the VM, ADR 0011) reach it only over a private VPC path; the exact mechanism for Cloud Run is fixed and verified in the milestone 5 spec. The deployed listener requires TLS and SASL (or an equivalent) with per-service credentials in Secret Manager. Local Compose may run without them.

Durability: replication factor 1 on one VM means a lost disk loses unconsumed events, and Debezium does not re-emit events it already read from the WAL. This is accepted for a demo with synthetic data. The recovery path (rebuild the VM, then re-snapshot or replay the outbox) is written in the milestone 3 spec and exercised in the milestone 3b drills.

## Alternatives considered

- Google Cloud Managed Service for Apache Kafka: best fit for a "runs on GCP" claim, but billed continuously; acceptable only for a short demo window.
- Confluent Cloud or Redpanda Cloud serverless/basic: likely cheaper than managed GCP Kafka and less to operate; leaves the GCP network and adds a vendor account. Not chosen as default only because pricing is unverified.
- Run Kafka on GKE Autopilot: more moving parts and cost than a single VM.

## Consequences

Easier: expected low cost, full control, same broker locally and deployed, teaches broker operations. Harder: the owner operates it (disk, upgrades, restarts); no replication factor above 1, so topic durability settings in the demo differ from what real production would use, and this must be documented. The broker and Debezium share one VM, so one failure stops both; the demo accepts this.

## Operational impact

Alerts (ticket-level cause signals unless they threaten the event-freshness SLO in RFC 0001) cover broker down, broker disk usage and consumer lag; drills cover broker down and VM rebuild. The VM has no public Kafka port; clients reach it over a private network. Topic configuration (partitions, replication factor 1, retention, cleanup policy, DLQ topic) is written in the spec. Teardown is part of Terraform. See `reliable-messaging` and `alerting` standards.
